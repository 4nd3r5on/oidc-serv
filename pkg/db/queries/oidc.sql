-- name: UpsertClient :exec
INSERT INTO clients (id, secret_enc, registration_token_enc, created_at, expires_at, meta)
VALUES (@id, @secret_enc, @registration_token_enc, @created_at, @expires_at, @meta)
ON CONFLICT (id) DO UPDATE SET
    secret_enc             = EXCLUDED.secret_enc,
    registration_token_enc = EXCLUDED.registration_token_enc,
    created_at             = EXCLUDED.created_at,
    expires_at             = EXCLUDED.expires_at,
    meta                   = EXCLUDED.meta;

-- name: GetClient :one
SELECT id, secret_enc, registration_token_enc, created_at, expires_at, meta
FROM clients
WHERE id = @id;

-- name: DeleteClient :exec
DELETE FROM clients WHERE id = @id;

-- name: UpsertGrant :exec
INSERT INTO grants (
    id, refresh_token, auth_code, grant_type, subject, client_id,
    scopes, nonce, jwk_thumbprint, client_cert_thumbprint,
    auth_details, resources, store, created_at, expires_at
) VALUES (
    @id, @refresh_token, @auth_code, @grant_type, @subject, @client_id,
    @scopes, @nonce, @jwk_thumbprint, @client_cert_thumbprint,
    @auth_details, @resources, @store, @created_at, @expires_at
)
ON CONFLICT (id) DO UPDATE SET
    refresh_token          = EXCLUDED.refresh_token,
    auth_code              = EXCLUDED.auth_code,
    grant_type             = EXCLUDED.grant_type,
    subject                = EXCLUDED.subject,
    client_id              = EXCLUDED.client_id,
    scopes                 = EXCLUDED.scopes,
    nonce                  = EXCLUDED.nonce,
    jwk_thumbprint         = EXCLUDED.jwk_thumbprint,
    client_cert_thumbprint = EXCLUDED.client_cert_thumbprint,
    auth_details           = EXCLUDED.auth_details,
    resources              = EXCLUDED.resources,
    store                  = EXCLUDED.store,
    created_at             = EXCLUDED.created_at,
    expires_at             = EXCLUDED.expires_at;

-- name: GetGrant :one
SELECT id, refresh_token, auth_code, grant_type, subject, client_id,
       scopes, nonce, jwk_thumbprint, client_cert_thumbprint,
       auth_details, resources, store, created_at, expires_at
FROM grants
WHERE id = @id;

-- name: GetGrantByRefreshToken :one
SELECT id, refresh_token, auth_code, grant_type, subject, client_id,
       scopes, nonce, jwk_thumbprint, client_cert_thumbprint,
       auth_details, resources, store, created_at, expires_at
FROM grants
WHERE refresh_token = @refresh_token;

-- name: DeleteGrant :exec
DELETE FROM grants WHERE id = @id;

-- name: DeleteGrantByAuthCode :exec
DELETE FROM grants WHERE auth_code = @auth_code;
