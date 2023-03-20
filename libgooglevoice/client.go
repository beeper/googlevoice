package libgooglevoice

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/emostar/libgooglevoice/libgooglevoice/api"
	"github.com/emostar/libgooglevoice/libgooglevoice/models"
	"github.com/emostar/libgooglevoice/libgooglevoice/util"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GoogleVoiceClient interfaces with the Google Voice API over HTTPS and works
// similar to how the web client at https://voice.google.com works.
type GoogleVoiceClient struct {
	http *http.Client

	baseURL string
	apiKey  string

	apiSIDHash  string
	cookieParam string

	log *zap.SugaredLogger

	// When a message is received, it is added to this map. Ideally this would
	// be backed up to a database, but for now it's just in memory.
	seenMessages map[string]bool
}

// NewGoogleVoiceClient creates a new GoogleVoiceClient to interact with the API
func NewGoogleVoiceClient(log *zap.SugaredLogger) *GoogleVoiceClient {
	return &GoogleVoiceClient{
		http:         &http.Client{},
		baseURL:      "https://clients6.google.com/voice/v1/voiceclient",
		apiKey:       "AIzaSyDTYc1N4xiODyrQYK0Kl6g_y279LjYkrBg",
		log:          log,
		seenMessages: make(map[string]bool),
	}
}

// SetAuth takes the full list of cookies from the browser and creates the
// necessary auth headers for the API.
func (gv *GoogleVoiceClient) SetAuth(cookieParam string) {
	gv.cookieParam = cookieParam
	gv.apiSIDHash = util.ExtractSID(cookieParam)
}

// IsConnected will return true if Auth has been setup, otherwise false
func (gv *GoogleVoiceClient) IsConnected() bool {
	return gv.apiSIDHash != "" && gv.cookieParam != ""
}

// GetAccountInfo is used to get all the account details of the current user.
func (gv *GoogleVoiceClient) GetAccountInfo() (*api.AccountInfo, error) {
	protobufMessage := "[null,1]"

	u := fmt.Sprintf("%s/account/get?alt=json&key=%s", gv.baseURL, gv.apiKey)
	json, err := gv.doRequest("POST", u, protobufMessage)
	if err != nil {
		return nil, err
	}

	resp := &api.AccountInfo{
		PrimaryDID: json.GetPath("account", "primaryDid").MustString(),
	}

	return resp, nil
}

// SendSMS sends a text message to the given thread. For a single recipient,
// the thread will always be in the format of "t.+1XXXXXXXXXX". For a group,
// there is a different format which is not yet supported. If you have an
// existing thread for a group message, you can send a message to that thread
// and it will work properly.
func (gv *GoogleVoiceClient) SendSMS(threadID, msg string) (*api.MessageResponse, error) {
	protobufMessage := fmt.Sprintf(
		"[null,null,null,null,\"%s\",\"%s\",[],null,[%d]]",
		msg, threadID, rand.Int63(),
	)
	gv.log.Debugf("protobufMessage: %s", protobufMessage)

	u := fmt.Sprintf("%s/api2thread/sendsms?alt=json&key=%s", gv.baseURL, gv.apiKey)
	json, err := gv.doRequest("POST", u, protobufMessage)
	if err != nil {
		return nil, err
	}

	timestampStr, err := json.Get("timestampMs").String()
	if err != nil {
		return nil, err
	}
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, err
	}

	resp := &api.MessageResponse{
		ID:        json.Get("threadItemId").MustString(),
		Timestamp: time.UnixMilli(timestamp),
	}
	return resp, nil
}

// FetchInbox fetches the inbox for the current user. The fetchPage parameter is
// for pagination, which is not implemented yet. The alertNew parameter is set
// to false on the first call to cache the messages. Subsequent calls should
// be true, so new messages can be detected.
//
// In the future we should support inbox type, thread count, and max recent
// parameters, that get set in the protobufMessage.
func (gv *GoogleVoiceClient) FetchInbox(fetchPage string, alertNew bool) ([]models.Thread, error) {
	// Format:
	// Inbox type (2 = SMS, 3 = CallHistory)
	// Thread count
	// Max recent messages, can be all in one thread or spread out
	protobufMessage := fmt.Sprintf(
		"[2,20,30,\"%s\",null,[null,true,true]]",
		fetchPage,
	)
	gv.log.Debugf("outbound protobufMessage: %s", protobufMessage)

	u := fmt.Sprintf("%s/api2thread/list?alt=json&key=%s", gv.baseURL, gv.apiKey)
	json, err := gv.doRequest("POST", u, protobufMessage)
	if err != nil {
		return nil, err
	}

	jsonThreads := json.Get("thread").MustArray()
	threads := make([]models.Thread, 0, len(jsonThreads))

	for idx := range jsonThreads {
		threadObj := json.Get("thread").GetIndex(idx).MustMap()
		jsonItems := json.Get("thread").GetIndex(idx).Get("item").MustArray()

		thread := models.Thread{
			ID:       threadObj["id"].(string),
			IsRead:   threadObj["read"].(bool),
			Messages: make([]models.Message, 0, len(jsonItems)),
		}

		for itemIndex := range jsonItems {
			itemObj := json.Get("thread").GetIndex(idx).Get("item").GetIndex(itemIndex).MustMap()
			message := models.Message{
				ID:         itemObj["id"].(string),
				Timestamp:  util.StringMilliTimestampToTime(itemObj["startTime"].(string)),
				SenderE164: itemObj["did"].(string),
				Status:     itemObj["status"].(string),
				Body:       itemObj["messageText"].(string),
				Thread:     &thread,
			}
			if itemObj["type"].(string) == "smsOut" {
				message.Direction = models.DirectionOutbound
			} else {
				message.Direction = models.DirectionInbound
			}
			if itemObj["messageId"] != nil {
				message.MessageID = itemObj["messageId"].(string)
			}

			// Check if we should alert that we have a new inbound message
			if alertNew {
				if _, ok := gv.seenMessages[message.ID]; !ok {
					gv.log.Infof("new message: %s %#v", message.ID, message)
					gv.seenMessages[message.ID] = true
				}
			}

			// Save that we have seen this message
			gv.seenMessages[message.ID] = true

			thread.Messages = append(thread.Messages, message)
		}

		threads = append(threads, thread)
	}

	// TODO - handle pagination
	nextPageToken := json.Get("paginationToken").MustString()
	_ = nextPageToken

	return threads, nil
}

// StartEventListener is meant to be run in a goroutine. It will listen for
// events from the browser channel. Currently, it only prints out any new
// incoming messages. In the future we should support pushing new inbound
// messages over a channel, so the caller can be alerted in real-time of new
// messages.
func (gv *GoogleVoiceClient) StartEventListener() {
	eventChannel := make(chan BrowserChannelEvent)
	bc := NewBrowserChannel(eventChannel, gv.log)
	bc.SetAuth(gv.cookieParam)

	go bc.StartEventListener()

	var lastEvent BrowserChannelEvent

	for {
		select {
		case event := <-eventChannel:
			gv.log.Infof("**** EVENT: %d", event)
			if lastEvent == NoopEvent && event == NoopEvent {
				gv.log.Infof("2 noops in a row, resetting connection")
				bc.ResetData()
			}
			lastEvent = event

			if event == RefreshInboxEvent {
				_, _ = gv.FetchInbox("", true)
			}
		}
	}
}

// buildRequest builds a new HTTP request, adding the required headers.
func (gv *GoogleVoiceClient) buildRequest(method, rawURL string, body io.Reader, contentType string) (*http.Request, error) {
	if gv.apiSIDHash == "" {
		return nil, errors.New("missing SID hash")
	}

	if gv.cookieParam == "" {
		return nil, errors.New("missing auth cookie")
	}

	var bodyw io.Reader
	if contentType != "" {
		bodyw = body
	}
	req, err := http.NewRequest(method, rawURL, bodyw)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// If we want to change the User-Agent and then encode it into
	// X-ClientDetails, as well as update the User-Agent header

	if parsedURL.Host == "clients6.google.com" {
		req.Header.Add("X-ClientDetails", "appVersion=5.0%20(X11%3B%20Ubuntu)&platform=Linux%20x86_64&userAgent=Mozilla%2F5.0%20(X11%3B%20Ubuntu%3B%20Linux%20x86_64%3B%20rv%3A109.0)%20Gecko%2F20100101%20Firefox%2F109.0")
		req.Header.Add("X-Requested-With", "XMLHttpRequest")
		req.Header.Add("X-JavaScript-User-Agent", "google-api-javascript-client/1.1.0")
		req.Header.Add("X-Client-Version", "512793257")
		req.Header.Add("X-Origin", "https://voice.google.com")
		req.Header.Add("X-Referer", "https://voice.google.com")
		req.Header.Add("X-Goog-Encode-Response-If-Executable", "base64")
		req.Header.Add("Origin", "https://clients6.google.com")
		req.Header.Add("Referer", "https://clients6.google.com/static/proxy.html?usegapi=1")
		req.Header.Add("Sec-Fetch-Site", "same-origin")
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:68.0) Gecko/20100101 Firefox/68.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Authorization", gv.apiSIDHash)
	req.Header.Add("X-Goog-AuthUser", "0")
	req.Header.Add("DNT", "1")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-GPC", "1")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("TE", "trailers")
	req.Header.Add("Cookie", gv.cookieParam)
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	return req, nil
}

// doRequest performs an HTTP request, and returns the response as a JSON
func (gv *GoogleVoiceClient) doRequest(method, rawURL, body string) (*simplejson.Json, error) {
	req, err := gv.buildRequest(method, rawURL, bytes.NewBufferString(body), "application/json+protobuf")
	if err != nil {
		return nil, err
	}

	resp, err := gv.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	resp.Body = http.MaxBytesReader(nil, resp.Body, 1<<20) // 1MB max
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	json, err := simplejson.NewJson(respBody)
	if err != nil {
		return nil, err
	}

	if apiError, ok := json.CheckGet("error"); ok {
		return nil, errors.New(apiError.Get("message").MustString())
	}

	return json, nil
}
