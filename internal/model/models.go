package models

import (
	"time"

	"github.com/lib/pq"
)

type Post struct {
	PostId      int            `json:"post_id" db:"post_id"`
	Title       string         `json:"title" db:"title"`
	Slug        string         `json:"slug" db:"slug"`
	Content     string         `json:"content" db:"content"`
	Author      string         `json:"author" db:"author"`
	Tags        pq.StringArray `json:"tags" db:"tags"`
	PublishedAt time.Time      `json:"published_at" db:"published_at"`
}
