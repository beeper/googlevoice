package models

import "time"

type MessageDirection int

const (
	DirectionInbound MessageDirection = iota
	DirectionOutbound
)

type Message struct {
	ID         string
	Timestamp  time.Time
	SenderE164 string
	Direction  MessageDirection
	Status     string
	Body       string
	Thread     *Thread

	// Only for outbound messages
	MessageID string // This is the random value the client used to send the message
}
