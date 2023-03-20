package models

type Thread struct {
	ID       string
	IsRead   bool
	Messages []Message
}
