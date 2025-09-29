CREATE KEYSPACE IF NOT EXISTS myjournal
    WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

CREATE TYPE IF NOT EXISTS myjournal.post (
    title   text,
    body    text
);

CREATE TABLE IF NOT EXISTS myjournal.users (
    name text,
    email text,
    user_id uuid,
    data map<text, text>,
    PRIMARY KEY ((name))
);

CREATE TABLE IF NOT EXISTS myjournal.posts_by_user (
    user_id uuid,
    post_id timeuuid,
    post    frozen<post>,
    tags    set<text>,
    PRIMARY KEY ((user_id), post_id)
) WITH CLUSTERING ORDER BY (post_id DESC);

-- reverse lookup table to be able to locate user_id by post_id
CREATE TABLE IF NOT EXISTS myjournal.user_by_postid (
    post_id    timeuuid,
    user_id    uuid,
    PRIMARY KEY ((post_id))
);