package libgooglevoice_test

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/emostar/libgooglevoice/libgooglevoice"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

var (
	cookies string
	logger  *zap.Logger
)

func init() {
	panic("Set your cookie in init and remove this panic")
	logger, _ = zap.NewDevelopment()
}

func TestClient_GetAccountInfo(t *testing.T) {
	gv := libgooglevoice.NewGoogleVoiceClient(logger.Sugar())
	gv.SetAuth(cookies)

	info, err := gv.GetAccountInfo()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	fmt.Println(info.PrimaryDID)
}

func TestJSON(t *testing.T) {
	src := `[[9,[[[["6",[null,[[[null,"W10=","1678952437065","1678952437245039",["1678952437230373","1678952437220632","1678952437230748","1678952437233544","1678952437245039","1678952437309887","1678952437310084",null,"1678952437310154"],[[[null,"-264184390715715714"]]]]]]]]]]]],[10,[[[["4",[null,[[[null,"W10=","1678952437065","1678952437245039",["1678952437230373","1678952437220632","1678952437230748","1678952437233544","1678952437245039","1678952437309887","1678952437310084",null,"1678952437310200"],[[[null,"-264184390715715714"]]]]]]]]]]]],[11,[[[["7",[null,[[[null,"W10=","1678952437065","1678952437245039",["1678952437230373","1678952437220632","1678952437230748","1678952437233544","1678952437245039","1678952437309887","1678952437310084",null,"1678952437310210"],[[[null,"-264184390715715714"]]]]]]]]]]]],[12,[[[["5",[null,[[[null,"W10=","1678952437065","1678952437245039",["1678952437230373","1678952437220632","1678952437230748","1678952437233544","1678952437245039","1678952437309887","1678952437310084",null,"1678952437310217"],[[[null,"-264184390715715714"]]]]]]]]]]]]]`
	json, err := simplejson.NewJson([]byte(src))
	assert.Nil(t, err)

	n := json.GetIndex(0).GetIndex(110).MustInt()
	assert.Equal(t, 9, n)
}

func TestFetchInbox(t *testing.T) {
	gv := libgooglevoice.NewGoogleVoiceClient(logger.Sugar())
	gv.SetAuth(cookies)

	// Fetch inbox
	threads, err := gv.FetchInbox("", false)
	assert.Nil(t, err)
	assert.NotEmpty(t, threads, "No threads found in inbox")
	assert.NotEmpty(t, threads[0].ID, "No ID found in first thread")
	assert.NotEmpty(t, threads[0].Messages, "No messages found in first thread")
	assert.NotEmpty(t, threads[0].Messages[0].ID, "No ID found in first message")

	logger.Sugar().Infof("Thread count: %d", len(threads))
	logger.Sugar().Infof("First thread: %#v", threads[0])
	logger.Sugar().Infof("First message: %#v", threads[0].Messages[0])
}

func TestFetchThread(t *testing.T) {

}

func TestSendMessage(t *testing.T) {
	gv := libgooglevoice.NewGoogleVoiceClient(logger.Sugar())
	gv.SetAuth(cookies)

	msgResponse, err := gv.SendSMS("t.+17154045634", "Hey there!")
	if err != nil {
		t.Fatal(err.Error())
	}
	logger.Sugar().Infoln("Sent message ID:", msgResponse.ID)
}

func TestNewMessageListener(t *testing.T) {
	gv := libgooglevoice.NewGoogleVoiceClient(logger.Sugar())
	gv.SetAuth(cookies)

	// Start listening for new messages
	_, _ = gv.FetchInbox("", false)
	go gv.StartEventListener()

	// Wait for 1 minute
	time.Sleep(1 * time.Minute)
}
