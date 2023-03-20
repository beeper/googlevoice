package libgooglevoice

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
)

type BrowserChannelEvent int

const (
	RefreshInboxEvent BrowserChannelEvent = iota
	PingEvent
	NoopEvent
)

type BrowserChannel struct {
	*GoogleVoiceClient

	gsessionID  string
	sid         string
	lastArrayID int
	requestID   int
	tryCount    int

	eventChannel chan BrowserChannelEvent
	log          *zap.SugaredLogger
}

func NewBrowserChannel(eventChannel chan BrowserChannelEvent, log *zap.SugaredLogger) *BrowserChannel {
	return &BrowserChannel{
		GoogleVoiceClient: &GoogleVoiceClient{
			http:    &http.Client{},
			baseURL: "https://signaler-pa.clients6.google.com/punctual",
			apiKey:  "AIzaSyDTYc1N4xiODyrQYK0Kl6g_y279LjYkrBg",
		},
		eventChannel: eventChannel,
		log:          log,
	}
}

func (bc *BrowserChannel) generateRandomZX() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	r := make([]rune, 11)
	for i := range r {
		r[i] = letters[rand.Intn(len(letters))]
	}
	return string(r)
}

func (bc *BrowserChannel) ResetData() {
	bc.gsessionID = ""
	bc.sid = ""
	bc.lastArrayID = 0
	bc.requestID = 0
	bc.tryCount = 0

	// TODO Force streaming connection to restart
}

func (bc *BrowserChannel) StartEventListener() {
	var err error

	bc.ResetData()

	// Run a loop that connects to the browser channel and listens for events
	for {
		// We can clear the gsessionID and/or sid to force a refresh when we
		// need to reconnect.

		if bc.gsessionID == "" {
			bc.gsessionID, err = bc.chooseServer()
			if err != nil {
				bc.log.Errorw("error choosing server", "error", err)
				return
			}
		}

		if bc.sid == "" {
			bc.sid, err = bc.getSID()
			if err != nil {
				bc.log.Errorw("error getting SID", "error", err)
				return
			}
		}

		// The browser channel is now set up and ready to start the main listening
		// request

		rawURL := fmt.Sprintf(
			"%s/multi-watch/channel?VER=8&gsessionid=%s&key=%s&RID=rpc&SID=%s&CI=0&AID=%d&TYPE=xmlhttp&zx=%s&t=%d",
			bc.baseURL, bc.gsessionID, bc.apiKey, bc.sid, bc.lastArrayID, bc.generateRandomZX(), bc.tryCount,
		)
		_, err := bc.doRawRequest("GET", rawURL, "", "", true)
		if err != nil {
			bc.log.Errorw("error in browser channel request", "error", err)
			return
		}
	}
}

func (bc *BrowserChannel) chooseServer() (string, error) {
	rawURL := fmt.Sprintf("%s/v1/chooseServer?key=%s",
		bc.baseURL, bc.apiKey,
	)
	protobufMessage := `[[null,null,null,[7,5],null,[null,[null,true],[[["1"]]]]]]`

	resp, err := bc.doRawRequest("POST", rawURL, protobufMessage, "application/json+protobuf", false)
	if err != nil {
		return "", err
	}

	sessionRegex := regexp.MustCompile(`^\["(.*)",`)
	sessionMatch := sessionRegex.FindStringSubmatch(resp)
	if len(sessionMatch) != 2 {
		return "", errors.New(fmt.Sprintf("failed to parse session ID: %s", resp))
	}

	return sessionMatch[1], nil
}

func (bc *BrowserChannel) getSID() (string, error) {
	// Start at a new random request ID
	bc.requestID = rand.Intn(30000)

	rawURL := fmt.Sprintf(
		"%s/multi-watch/channel?VER=8&gsessionid=%s&key=%s&RID=%d&CVER=22&zx=%s&t=1",
		bc.baseURL, bc.gsessionID, bc.apiKey, bc.requestID, bc.generateRandomZX(),
	)
	postBody := `count=7&ofs=0&req0___data__=[[["1",[null,null,null,[7,5],null,[null,[null,true],[[["2"]]]],null,null,1],null,3]]]&req1___data__=[[["2",[null,null,null,[7,5],null,[null,[null,true],[[["3"]]]],null,null,1],null,3]]]&req2___data__=[[["3",[null,null,null,[7,5],null,[null,[null,true],[[["3"]]]],null,null,1],null,3]]]&req3___data__=[[["4",[null,null,null,[7,5],null,[null,[null,true],[[["1"]]]],null,null,1],null,3]]]&req4___data__=[[["5",[null,null,null,[7,5],null,[null,[null,true],[[["1"]]]],null,null,1],null,3]]]&req5___data__=[[["6",[null,null,null,[7,5],null,[null,[null,true],[[["1"]]]],null,null,1],null,3]]]&req6___data__=[[["7",[null,null,null,[7,5],null,[null,[null,true],[[["1"]]]],null,null,1],null,3]]]`
	resp, err := bc.doRawRequest("POST", rawURL, postBody,
		"application/x-www-form-urlencoded", false,
	)
	if err != nil {
		return "", err
	}

	sidRegex := regexp.MustCompile(`\[\[\d,\["[A-Za-z0-9].*","(.*)",".*"`)
	sidMatch := sidRegex.FindStringSubmatch(resp)
	if len(sidMatch) != 2 {
		return "", errors.New(fmt.Sprintf("failed to parse SID: %s", resp))
	}

	return sidMatch[1], nil
}

func (bc *BrowserChannel) addHeaders(req *http.Request, contentType string) {
	req.Header.Add("Referer", "https://voice.google.com/")
	req.Header.Add("Sec-Fetch-Site", "same-site")
	req.Header.Add("Origin", "https://voice.google.com")
	req.Header.Add("accept-encoding", "gzip, deflate, br")

	if contentType == "application/x-www-form-urlencoded" {
		req.Header.Add("X-WebChannel-Content-Type", "application/json+protobuf")
	}
}

func (bc *BrowserChannel) doRawRequest(method, rawURL, body, contentType string, streaming bool) (string, error) {
	req, err := bc.buildRequest(method, rawURL, bytes.NewBufferString(body), contentType)
	if err != nil {
		return "", err
	}

	// Add the custom headers for the browser channel
	bc.addHeaders(req, contentType)

	resp, err := bc.http.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	lineLengthRegex := regexp.MustCompile(`(?m)([0-9]+)\n`)

	var r string
	for {
		line := make([]byte, 4096)
		n, err := resp.Body.Read(line)
		bc.log.Debugf("n: %d, line: %s", n, line[:n])

		thisLine := string(line[:n])
		r += thisLine

		if streaming {
			lineLengthMatches := lineLengthRegex.FindAllStringSubmatch(thisLine, -1)

			if n > 0 && len(lineLengthMatches) == 0 {
				// There is no data, we need to reset the entire connection
				bc.log.Infoln("No data, resetting connection")
				bc.ResetData()
				return "", nil
			}

			var firstJson *simplejson.Json
			for _, lineLengthMatch := range lineLengthMatches {
				if len(lineLengthMatch) == 2 {
					skip := len(lineLengthMatch[1])
					lineLength, err := strconv.Atoi(lineLengthMatch[1])
					if err != nil {
						bc.log.Warnf("Error parsing line length: %s (%s)",
							lineLengthMatch[1], err,
						)
						continue
					}

					// Ensure that we have the entire line to parse
					if n < skip+lineLength+1 {
						// TODO We should save this line and append to it from
						// the next read to ensure we don't lose data
						bc.log.Warnf("Line length mismatch, skipping: %d < %d",
							n, skip+lineLength+1,
						)
						continue
					}

					rawData := line[skip : skip+lineLength+1]
					json, err := simplejson.NewJson(rawData)
					if err != nil {
						bc.log.Warnf("Error parsing JSON: %s (%s)",
							string(rawData), err,
						)
						continue
					}

					// Save for processing after the loops
					if firstJson == nil {
						firstJson = json.GetIndex(0)
					}

					for idx := range json.MustArray() {
						item := json.GetIndex(idx)
						arrayID := item.GetIndex(0).MustInt()
						if arrayID > 0 {
							bc.lastArrayID = arrayID + 1
						}
					}

				}
			}

			// Only send one event, as the loops seem to contain
			// multiple events
			if firstJson != nil {
				bc.eventChannel <- getEventType(firstJson)
				firstJson = nil
			}

			if err != nil {
				// Return no error so we can reconnect
				bc.log.Infof("Reconnect with lastArrayID: %d",
					bc.lastArrayID,
				)
				return "", nil
			}
		}

		if err == io.EOF {
			return r, nil
		} else if err != nil {
			return r, err
		}
	}
}

func getEventType(json *simplejson.Json) BrowserChannelEvent {
	array := json.GetIndex(1).MustArray()
	if len(array) == 1 {
		if val, ok := array[0].(string); ok && val == "noop" {
			return NoopEvent
		}
	}

	innerArray := json.GetIndex(1).GetIndex(0).GetIndex(0)
	innerArrayLength := len(innerArray.MustArray())
	if innerArrayLength > 1 {
		return PingEvent
	} else {
		// The initial response array has an empty inner array, so filter those
		// out as noops. It may be setting the initial state, but we don't rely
		// on state.
		deepArrayLen := len(innerArray.GetIndex(0).GetIndex(1).MustArray())
		if deepArrayLen == 1 {
			return NoopEvent
		}

		return RefreshInboxEvent
	}
}
