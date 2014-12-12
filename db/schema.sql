BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS node (
	id INTEGER PRIMARY KEY NOT NULL,
    parent_id INTEGER NOT NULL DEFAULT 0,
    domain_id smallint DEFAULT 0,
    title character varying(150) DEFAULT '',
    vote int DEFAULT 0,
    tripcode character varying(10) DEFAULT '',
    body text,
    rendered text,
    level smallint DEFAULT 0,
    status smallint DEFAULT 1,
    created timestamp,
    updated timestamp
);

CREATE INDEX IF NOT EXISTS parent_id_ndx ON node(parent_id);
CREATE INDEX IF NOT EXISTS domain_id_ndx ON node(domain_id);
CREATE INDEX IF NOT EXISTS vote_ndx ON node(vote DESC);
CREATE INDEX IF NOT EXISTS status_ndx ON node(status DESC);
CREATE INDEX IF NOT EXISTS created_ndx ON node(created DESC);
CREATE INDEX IF NOT EXISTS updated_ndx ON node(updated DESC);

COMMIT TRANSACTION;