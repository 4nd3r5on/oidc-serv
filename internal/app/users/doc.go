// Package users implements user management use cases.
//
// # Use-case structs
//
// Each operation is a small struct with a repository dependency and a logger.
// Constructors (NewCreate, NewGetByID, etc.) handle nil-logger defaulting.
// The structs are intended to be wired once at startup and reused across requests.
//
//   - [Create]              — register a new user (validates input, hashes password)
//   - [GetByID]             — fetch a user by UUID
//   - [GetByUsername]       — fetch a user by username
//   - [Update]              — update username and/or locale
//   - [UpdatePasswordByID]  — change password after verifying the current one
//   - [Delete]              — remove a user by UUID
//   - [Exists]              — check whether a user UUID exists (used by auth cores)
//   - [MatchUserPass]       — verify username + password, returns the user on success
//
// # Self-service
//
// [Me] wires the above into self-service methods (Update, Delete, UpdatePassword)
// that resolve the acting user from context via an [AuthFunc] before delegating
// to the corresponding use-case function.
//
// # API boundary
//
// [User] is the internal entity and includes PasswordHash.
// [GetRes] is the safe, hash-free projection returned to callers outside this package.
package users
