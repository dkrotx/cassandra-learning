CREATE KEYSPACE IF NOT EXISTS myjournal
    WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

CREATE TABLE IF NOT EXISTS myjournal.posts (
    user text,
    post_id uuid,
    title text,
    body text,
    created_at timestamp,
    private boolean,
    PRIMARY KEY ((user), created_at)
);