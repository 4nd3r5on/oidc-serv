// Package session contains use cases for managing user auth sessions.
//
// A session is an opaque random token tied to a user ID with an expiry time.
// It is stored externally (Redis) and looked up on each authenticated request.
//
// Use cases:
//   - [IssueSession]: verifies credentials via [UserVerifyFunc], generates a
//     session token, and persists it via [Storer].
//   - [Verify]: retrieves a session by token and checks that it has not expired.
//   - [Delete]: removes a session (logout).
//
// Dependencies are injected as interfaces or function types so the package
// has no knowledge of the storage backend.
package session
