package myjournal

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/config"
	"go.uber.org/zap"
)

type Config struct {
	Hosts      []string `yaml:"hosts"`
	Keyspace   string   `yaml:"keyspace"`
	LogQueries bool     `yaml:"log_queries"`
}

type JournalDB struct {
	session *gocql.Session
	logger  *zap.Logger
}

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
