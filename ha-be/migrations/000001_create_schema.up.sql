CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(64) PRIMARY KEY,
    email VARCHAR(256) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    confirm_token VARCHAR(256),
    confirm_token_expires_at TIMESTAMP,
    reset_token VARCHAR(256),
    reset_token_expires_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rooms (
    id VARCHAR(64) PRIMARY KEY,
    slug VARCHAR(128) NOT NULL UNIQUE,
    name VARCHAR(256) NOT NULL,
    owner_id VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS room_actions (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(128) NOT NULL,
    payload BYTEA,
    created BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_room_actions_slug ON room_actions (slug);

CREATE TABLE IF NOT EXISTS plugins (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    visible BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS custom_cards (
    id VARCHAR(64) PRIMARY KEY,
    plugin_id VARCHAR(64) NOT NULL,
    name VARCHAR(256) NOT NULL,
    data JSONB
);

CREATE TABLE IF NOT EXISTS lfg_posts (
    id VARCHAR(64) PRIMARY KEY,
    plugin_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    text TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(64) PRIMARY KEY,
    message TEXT NOT NULL,
    level VARCHAR(32) NOT NULL
);
