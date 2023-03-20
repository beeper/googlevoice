package main

import (
	"maunium.net/go/mautrix/bridge/commands"
	"strings"
	"time"
)

type WrappedCommandEvent struct {
	*commands.Event
	Bridge *GVBridge
	User   *User
	Portal *Portal
}

func (br *GVBridge) RegisterCommands() {
	proc := br.CommandProcessor.(*commands.Processor)
	proc.AddHandlers(
		cmdLogin,
		cmdLatestMessage,
		cmdSendMessage,
	)
}

func wrapCommand(handler func(*WrappedCommandEvent)) func(*commands.Event) {
	return func(ce *commands.Event) {
		user := ce.User.(*User)
		var portal *Portal
		if ce.Portal != nil {
			portal = ce.Portal.(*Portal)
		}
		br := ce.Bridge.Child.(*GVBridge)
		handler(&WrappedCommandEvent{ce, br, user, portal})
	}
}

var cmdLogin = &commands.FullHandler{
	Func: wrapCommand(fnLogin),
	Name: "login",
	Help: commands.HelpMeta{
		Section:     commands.HelpSectionAuth,
		Description: "Link the bridge to your Google Voice account.",
	},
}

func fnLogin(ce *WrappedCommandEvent) {
	if len(ce.Args) == 0 {
		ce.Reply("Please provide your Google cookies as an argument")
		return
	}

	cookie := strings.Join(ce.Args, " ")
	ce.User.GVoice.SetAuth(cookie)
	info, err := ce.User.GVoice.GetAccountInfo()
	if err != nil {
		ce.Reply("Failed to log in to Google Voice: %v", err)
		return
	}
	ce.Reply("Successfully logged in to Google Voice as %s", info.PrimaryDID)
}

var cmdLatestMessage = &commands.FullHandler{
	Func: wrapCommand(fnLatestMessage),
	Name: "latest-message",
	Help: commands.HelpMeta{
		Section:     commands.HelpSectionGeneral,
		Description: "Get the latest message from your Google Voice account.",
	},
}

func fnLatestMessage(ce *WrappedCommandEvent) {
	if !ce.User.GVoice.IsConnected() {
		ce.Reply("Please login first")
		return
	}

	threads, err := ce.User.GVoice.FetchInbox("", false)
	if err != nil {
		ce.Reply("Failed to fetch inbox: %v", err)
		return
	}

	for _, thread := range threads {
		for _, message := range thread.Messages {
			if len(message.Body) > 0 {
				ce.Reply("[%s] %s: %s",
					message.Timestamp.Format(time.Kitchen), message.SenderE164,
					message.Body,
				)
				return
			}
		}
	}

	ce.Reply("No messages found")
}

var cmdSendMessage = &commands.FullHandler{
	Func: wrapCommand(fnSendMessage),
	Name: "send-message",
	Help: commands.HelpMeta{
		Section:     commands.HelpSectionGeneral,
		Description: "Send a message over Google Voice.",
	},
}

func fnSendMessage(ce *WrappedCommandEvent) {
	if !ce.User.GVoice.IsConnected() {
		ce.Reply("Please login first")
		return
	}

	if len(ce.Args) < 2 {
		ce.Reply("Please provide a phone number in E164 format and a message as arguments")
		return
	}

	threadID := "t." + ce.Args[0]
	resp, err := ce.User.GVoice.SendSMS(threadID, strings.Join(ce.Args[1:], " "))
	if err != nil {
		ce.Reply("Failed to send message: %v", err)
		return
	}
	ce.Reply("Sent message with ID %s", resp.ID)
}
