CREATE ROLE IF NOT EXISTS oreo_user WITH PASSWORD = 'oreo_password' AND LOGIN = true AND SUPERUSER = true;
CREATE KEYSPACE IF NOT EXISTS oreo WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE oreo;

CREATE TABLE IF NOT EXISTS items (
    key text,
    value text,
    group_key_list text,
    txn_state int,
    t_valid bigint,
    t_lease timestamp,
    prev text,
    linked_len int,
    is_deleted boolean,
    version text,
    PRIMARY KEY (key)
) WITH gc_grace_seconds = 172800;

CREATE TABLE IF NOT EXISTS kv(
    key text,
    value text,
    PRIMARY KEY ( key )
);


