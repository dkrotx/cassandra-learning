package myjournal

import (
	"context"
	"fmt"

	"github.com/dkrotx/cassandra-learning/pkg/entities"
	"github.com/gocql/gocql"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Config struct {
	Hosts       []string `yaml:"hosts"`
	Keyspace    string   `yaml:"keyspace"`
	Consistency string   `yaml:"consistency"`
}

type JournalDB struct {
	session *gocql.Session
	logger  *zap.Logger
}

// Module is the fx-module for the journal database.
var Module = fx.Provide(NewDB)

func NewDB(cfgProvider config.Provider, logger *zap.Logger) (*JournalDB, error) {
	var cfg Config
	if err := cfgProvider.Get("cassandra").Populate(&cfg); err != nil {
		return nil, fmt.Errorf("failed to populate config: %v", err)
	}

	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return &JournalDB{
		session: session,
		logger:  logger,
	}, nil
}

func (db *JournalDB) CreatePost(ctx context.Context, post *entities.Post) error {
	query := `INSERT INTO posts (user, post_id, title, body, created_at, private) VALUES (?, ?, ?, ?, ?, ?)`
	if err := db.session.Query(query,
		post.User,
		post.PostID.String(),
		post.Title,
		post.Body,
		post.CreatedAt,
		post.Private,
	).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to create post: %v", err)
	}
	return nil
}
