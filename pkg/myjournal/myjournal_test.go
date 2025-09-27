package myjournal

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/dkrotx/cassandra-learning/pkg/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/config"
	"go.uber.org/zap"
)

const testConfig = `cassandra:
  hosts: ["localhost:9042"]
  keyspace: "myjournal"`

func newTestDB(t *testing.T) *JournalDB {
	configProvider, err := config.NewYAML(config.Source(strings.NewReader(testConfig)))
	require.NoError(t, err)

	db, err := NewDB(configProvider, zap.NewNop())
	require.NoError(t, err)
	return db
}

func TestCreatePost(t *testing.T) {
	db := newTestDB(t)
	err := db.CreatePost(context.Background(), &entities.Post{
		Title:     "test title",
		PostID:    uuid.New(),
		User:      "dkrot",
		Body:      "test body",
		Private:   false,
		CreatedAt: time.Now(),
	})

	require.NoError(t, err)
}
