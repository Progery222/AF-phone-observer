package domain

import "time"

type CacheClearResult struct {
	Serial        string    `json:"serial"`
	Cleared       bool      `json:"cleared"`
	ScreenCleared bool      `json:"screen_cleared"`
	UICleared     bool      `json:"ui_cleared"`
	ClearedAt     time.Time `json:"cleared_at"`
}
