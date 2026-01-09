package shorten

import "time"

type ShortLink struct {
	ID        string `json:"short"`
	URL       string `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
	Hits      int64 `json:"hits"`
}
