package domain

type SendgridEvent struct {
	Email       string   `json:"email"`
	Event       string   `json:"event"`
	Timestamp   int64    `json:"timestamp"`
	Category    []string `json:"category"`
	SGEventID   string   `json:"sg_event_id"`
	SGMessageID string   `json:"sg_message_id"`
	IP          string   `json:"ip"`
	UserAgent   string   `json:"useragent"`
	Status      string   `json:"status"`
}
