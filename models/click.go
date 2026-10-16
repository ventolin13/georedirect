package models

import "time"

type ClientIP struct {
	IP string
}

type Click struct {
	ClickID     string    `json:"click_id"`
	ClientIP    string    `json:"client_ip"`
	UserAgent   string    `json:"user_agent"`
	Fingerprint string    `json:"fingerprint"`
	Path        string    `json:"path"`
	Query       string    `json:"query"`
	Blocked     bool      `json:"blocked"`
	At          time.Time `json:"at"`
}
