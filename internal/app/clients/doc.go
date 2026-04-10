// Package clients implements OIDC client management use cases.
//
// # Use-case structs
//
// Each operation is a small struct with a repository dependency and a logger.
// Constructors (NewCreate, NewGetByID, NewDelete) handle nil-logger defaulting.
// The structs are intended to be wired once at startup and reused across requests.
//
//   - [Create]   — register a new OIDC client (generates secret if absent, checks for conflicts)
//   - [GetByID]  — fetch a client by ID
//   - [Delete]   — remove a client by ID
//
// # API boundary
//
// [Client] is the safe, secret-free projection returned to callers outside this package.
// [CreateRes] carries the plaintext secret and is returned by [Create] only once —
// the secret cannot be recovered after that call.
package clients
