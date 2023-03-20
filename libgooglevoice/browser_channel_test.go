package libgooglevoice_test

import (
	"github.com/emostar/libgooglevoice/libgooglevoice"
	"testing"
)

func TestBrowserChannel(t *testing.T) {
	eventChannel := make(chan libgooglevoice.BrowserChannelEvent)
	gv := libgooglevoice.NewBrowserChannel(eventChannel, logger.Sugar())
	gv.SetAuth(cookies)

	go gv.StartEventListener()

	var lastEvent libgooglevoice.BrowserChannelEvent

	for {
		select {
		case event := <-eventChannel:
			logger.Sugar().Infof("**** EVENT: %d", event)
			if lastEvent == libgooglevoice.NoopEvent && event == libgooglevoice.NoopEvent {
				logger.Sugar().Infof("2 noops in a row, resetting connection")
				gv.ResetData()
			}
			lastEvent = event
		}
	}
}
