package model

import "time"

type SenderReport struct {
	CompanyName             string `json:"company_name"`
	Email                   string `json:"email"`
	UnsubscribeLink         string `json:"unsubscribe_link"`
	UnsubscribeLinkOneClick bool   `json:"unsubscribe_link_oneclick"`
	UnsubscribeEmail        string `json:"unsubscribe_email"`

	LatestMessage SenderMessage `json:"latest_message"`
	MessageCount  int64         `json:"count"`
}

type SenderMessage struct {
	Date     time.Time `json:"date"`
	Subject  string    `json:"subject"`
	Category string    `json:"category"`
}
