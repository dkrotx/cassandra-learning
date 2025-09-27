package myjournal

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/dkrotx/cassandra-learning/pkg/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/config"
	"go.uber.org/zap"
)

const testConfig = `cassandra:
  hosts: ["localhost:9042"]
  keyspace: "myjournal"`

var testUserUUID = entities.UserID(uuid.MustParse("10000000-1000-f000-f000-000000000000"))

func newTestDB(t *testing.T) *JournalDB {
	configProvider, err := config.NewYAML(config.Source(strings.NewReader(testConfig)))
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	db, err := NewDB(configProvider, logger)
	require.NoError(t, err)
	return db
}

func TestCreatePost(t *testing.T) {
	db := newTestDB(t)
	err := db.CreatePost(context.Background(),
		testUserUUID,
		&entities.PostData{
			Title: "test title",
			Body:  "test body",
			Tags:  []string{"Skanderborg", "school"},
		},
	)

	require.NoError(t, err)
}

func TestReadPostsByUser(t *testing.T) {
	db := newTestDB(t)
	posts, err := db.ReadPostsByUser(context.Background(), testUserUUID)
	require.NoError(t, err)
	fmt.Printf("%v", posts)
}
