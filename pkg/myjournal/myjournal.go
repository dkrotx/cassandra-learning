package myjournal

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dkrotx/cassandra-learning/pkg/entities"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Config struct {
	Hosts       []string `yaml:"hosts"`
	Keyspace    string   `yaml:"keyspace"`
	LogQueries  bool     `yaml:"log_queries"`
	Consistency string   `yaml:"consistency"`
}

type JournalDB struct {
	session *gocql.Session
	logger  *zap.Logger
}

// Module is the fx-module for the journal database.
var Module = fx.Provide(NewDB)

type verboseQueryObserver struct {
	logger *zap.Logger
}

func (v *verboseQueryObserver) ObserveQuery(ctx context.Context, q gocql.ObservedQuery) {
	vals := make([]string, len(q.Values))
	for i, v := range q.Values {
		switch b := v.(type) {
		case []byte:
			// avoid dumping huge binary; print length
			vals[i] = "<bytes:" + strconv.Itoa(len(b)) + "B>"
		case time.Time:
			vals[i] = b.UTC().Format(time.RFC3339Nano)
		default:
			vals[i] = fmt.Sprintf("%+v", v)
		}
	}

	v.logger.Info("Captured cassandra query",
		zap.String("statement", q.Statement), zap.String("values", strings.Join(vals, ", ")),
	)
}

func NewDB(cfgProvider config.Provider, logger *zap.Logger) (*JournalDB, error) {
	var cfg Config
	if err := cfgProvider.Get("cassandra").Populate(&cfg); err != nil {
		return nil, fmt.Errorf("failed to populate config: %v", err)
	}

	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace

	if cfg.LogQueries {
		cluster.QueryObserver = &verboseQueryObserver{logger: logger}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return &JournalDB{
		session: session,
		logger:  logger,
	}, nil
}

func (db *JournalDB) CreatePost(ctx context.Context, userID entities.UserID, post *entities.PostData) error {
	query := `INSERT INTO posts_by_user (user_id, post_id, post, tags) VALUES (?, ?, {title: ?, body: ?}, ?)`
	if err := db.session.Query(query,
		userID.String(),
		gocql.TimeUUID(),
		post.Title,
		post.Body,
		post.Tags,
	).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to create post: %v", err)
	}
	return nil
}

func (db *JournalDB) ReadPostsByUser(ctx context.Context, userID entities.UserID) ([]*entities.Post, error) {
	query := `SELECT post_id, post.title, post.body, tags FROM posts_by_user WHERE user_id = ? LIMIT ?`
	var posts []*entities.Post

	iter := db.session.Query(query, userID.String(), 3).WithContext(ctx).Iter()

	var postID gocql.UUID
	var tags []string
	var title string
	var body string

	for iter.Scan(&postID, &title, &body, &tags) {
		post := &entities.Post{
			PostID:    uuid.UUID(postID),
			CreatedAt: postID.Time(),
			PostData: entities.PostData{
				Title: title,
				Body:  body,
				Tags:  tags,
			},
		}

		posts = append(posts, post)
	}

	// Check for any errors that occurred during iteration
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to read posts: %v", err)
	}

	return posts, nil
}
