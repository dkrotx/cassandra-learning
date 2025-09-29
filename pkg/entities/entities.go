package entities

import (
	"time"

	"github.com/google/uuid"
)

type UserID = uuid.UUID
type PostID = uuid.UUID

type PostData struct {
	Title string
	Body  string
	Tags  []string
}

type Post struct {
	PostData
	PostID    uuid.UUID
	CreatedAt time.Time
}
