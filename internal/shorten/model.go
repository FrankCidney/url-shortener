package shorten

import "time"

type ShortLink struct {
	ID        string `json:"short"`
	URL       string `json:"url"`
	Hits      int64 `json:"hits"`
	CreatedAt time.Time `json:"createdAt"`
}