// Package api is the HTTP layer for the OIDC server.
//
// It bridges the ogen-generated handler interface ([genapi.Handler]) with
// application use cases. Each [Handlers] field is a narrow interface
// satisfied by the corresponding use case struct.
//
// [SecurityHandler] sits in front of every request: it runs the auth
// verification pipeline (TMB or session token) and stores the result in
// context via [auth.CtxPutClientData]. The use cases then call
// [auth.CtxGetClientData] through the [Authenticator] when they need the
// caller's identity.
//
// Adapters (adapters.go) handle type conversion between the generated API
// types and the app-layer types. They are intentionally thin — no logic.
package api
