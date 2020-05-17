CREATE TABLE IF NOT EXISTS health_authorities (
    authority_id    BYTEA NOT NULL,
    name            TEXT NOT NULL,
    api_key         BYTEA NOT NULL,
    UNIQUE(api_key),
    PRIMARY KEY(authority_id, api_key)
);

CREATE TABLE IF NOT EXISTS authorization_keys (
    authorization_key BYTEA PRIMARY KEY,
    api_key           BYTEA REFERENCES health_authorities(api_key),
    key_type          TEXT,
    permitted_start   TIMESTAMP NOT NULL,
    permitted_end     TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS reported_keys (
    TEK                BYTEA NOT NULL PRIMARY KEY,
    ENIN               TIMESTAMP NOT NULL,
    authorization_key  BYTEA NOT NULL REFERENCES authorization_keys(authorization_key),
    uploaded_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX enin_idx ON reported_keys(ENIN);
CREATE INDEX hak_idx ON reported_keys(authorization_key);

INSERT INTO health_authorities(authority_id, name, api_key) VALUES (
    decode('da250d7fbffca634bf9b38e9430508bb', 'hex'),
    'Fake Health Authority #1',
    decode('c3b9b61b687b895aff09eb072fb07d33', 'hex')
);
