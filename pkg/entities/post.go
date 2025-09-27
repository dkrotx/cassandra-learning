package entities

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	Title     string
	PostID    uuid.UUID
	User      string
	Body      string
	Private   bool
	CreatedAt time.Time
}
