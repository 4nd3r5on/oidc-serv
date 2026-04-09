// Package auth implements the two-phase authentication pipeline.
//
// # Phases
//
// Phase 1 — extraction ([Verifier]): called at the transport layer (e.g. HTTP
// middleware). It parses and format-validates the raw credential, then stores
// the result as [ClientData] in the request context via [CtxPutClientData].
// It does not touch any backing store.
//
// Phase 2 — resolution ([Core], [Authenticator]): called by application
// handlers when they need a verified user identity. [Authenticator.Auth]
// reads [ClientData] from context, dispatches to the registered [Core] for
// that method, and returns the authenticated user ID.
//
// # Auth methods
//
//   - [MethodNone]    — no credentials; always resolves to unauthenticated.
//
//   - [MethodTMB]     — Trust Me Bro: a UUID presented as "TMB <uuid>".
//     Development/internal only — no secret involved.
//
//   - [MethodSession] — opaque high-entropy session key looked up in the
//     session store.
//
//     // Handler:
//     userID, err := authenticator.Auth(ctx, required)
package auth
