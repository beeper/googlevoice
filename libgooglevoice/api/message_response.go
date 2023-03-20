package api

import "time"

type MessageResponse struct {
	ID        string    `json:"threadItemID"`
	Timestamp time.Time `json:"timestampMS"`
}
