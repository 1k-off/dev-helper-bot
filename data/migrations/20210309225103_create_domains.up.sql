CREATE TABLE domains (
    fqdn text PRIMARY KEY,
    ip TEXT NOT NULL,
    user_id TEXT NOT NULL UNIQUE,
    user_name TEXT NOT NULL,
    created_at TEXT NOT NULL,
    delete_at TEXT NOT NULL,
    basic_auth bool DEFAULT true
);