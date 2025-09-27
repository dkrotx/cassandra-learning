CREATE KEYSPACE IF NOT EXISTS myjournal
    WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

CREATE TABLE IF NOT EXISTS myjournal.posts_by_user (
    user_id uuid,
    post_id timeuuid,
    title   text,
    body    text,
    tags    set<text>,
    PRIMARY KEY ((user_id), post_id)
) WITH CLUSTERING ORDER BY (post_id DESC);