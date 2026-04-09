// Package redis contains Redis-backed repository implementations.
//
// Repositories:
//   - [UserSessionRepo]: user auth sessions. Key scheme: user_session:{token}.
//     Implements [session.Storer], [session.Getter], [session.Deleter].
//   - [SessionRepo], [GrantRepoCached], [TokenRepo], [LogoutSessionRepo]:
//     OIDC provider storage used by go-oidc.
//
// All repositories accept a *redis.Client and are safe for concurrent use.
package redis
