-- OAuth clients registered with the provider.
-- ClientMeta is a large embedded struct with no benefit to normalising;
-- everything except the sensitive fields is stored as JSONB.
-- secret_enc and registration_token_enc are AES-GCM ciphertext blobs;
-- the plaintext values are never written to this table.
CREATE TABLE clients (
    id                       TEXT    PRIMARY KEY,
    secret_enc               BYTEA,
    registration_token_enc   BYTEA,
    created_at               BIGINT  NOT NULL,
    expires_at               BIGINT  NOT NULL DEFAULT 0,  -- 0 = never
    meta                     JSONB   NOT NULL DEFAULT '{}'
);

-- Durable authorisation grants (consent + refresh token).
-- Sessions and tokens are stored in Redis; only grants live in Postgres
-- because they represent long-lived user consent and must survive restarts.
--
-- expires_at = 0 means the grant is permanent (offline_access).
-- Cleanup queries must guard: WHERE expires_at != 0 AND expires_at < now_epoch.
--
-- refresh_token is the plain opaque string issued by the library.
-- It is high-entropy and stored as-is; add a UNIQUE constraint for O(1) lookup.
--
-- auth_code is kept after exchange for replay-attack revocation
-- (goidc.GrantManager.DeleteByAuthCode). Do not NULL it out post-exchange.
CREATE TABLE grants (
    id                     TEXT    PRIMARY KEY,
    refresh_token          TEXT,
    auth_code              TEXT,
    grant_type             TEXT    NOT NULL,
    subject                TEXT    NOT NULL,
    client_id              TEXT    NOT NULL REFERENCES clients (id) ON DELETE CASCADE,
    scopes                 TEXT,
    nonce                  TEXT,
    jwk_thumbprint         TEXT,
    client_cert_thumbprint TEXT,
    auth_details           JSONB,
    resources              JSONB,
    store                  JSONB,
    created_at             BIGINT  NOT NULL,
    expires_at             BIGINT  NOT NULL DEFAULT 0   -- 0 = never
);

-- Partial indexes keep them small: only rows that actually have the column set.
CREATE UNIQUE INDEX grants_refresh_token_idx ON grants (refresh_token)
    WHERE refresh_token IS NOT NULL;

CREATE INDEX grants_auth_code_idx ON grants (auth_code)
    WHERE auth_code IS NOT NULL;

CREATE INDEX grants_subject_idx   ON grants (subject);
CREATE INDEX grants_client_id_idx ON grants (client_id);
