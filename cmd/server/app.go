package main

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	appauth "github.com/4nd3r5on/oidc-serv/internal/app/auth"
	appclients "github.com/4nd3r5on/oidc-serv/internal/app/clients"
	appsession "github.com/4nd3r5on/oidc-serv/internal/app/session"
	appusers "github.com/4nd3r5on/oidc-serv/internal/app/users"
)

// App is the wired set of use cases for this binary.
// It is assembled once in [initApp] and its fields are distributed to the
// API layer and OIDC provider. It is intentionally cmd-local — the app
// packages expose individual use cases; App is just how main groups them.
type App struct {
	// clients
	CreateClient *appclients.Create
	GetClient    *appclients.GetByID
	DeleteClient *appclients.Delete

	// users
	CreateUser        *appusers.Create
	GetUser           *appusers.GetByID
	GetUserByUsername *appusers.GetByUsername
	UpdateUser        *appusers.Update
	UpdatePassword    *appusers.UpdatePasswordByID
	DeleteUser        *appusers.Delete
	MatchUserPass     *appusers.MatchUserPass
	Me                *appusers.Me

	// sessions
	IssueSession  *appsession.IssueSession
	DeleteSession *appsession.Delete

	// auth
	TMBVerifier     *appauth.TMBVerifier
	SessionVerifier *appauth.VerifySession
	Authenticator   *appauth.Authenticator
}

func initApp(repos *Repos, logger *slog.Logger) *App {
	// clients
	createClient := appclients.NewCreate(repos.Clients, logger)
	getClient := appclients.NewGetByID(repos.Clients, logger)
	deleteClient := appclients.NewDelete(repos.Clients, logger)

	// users
	createUser := appusers.NewCreate(repos.Users, logger)
	getUser := appusers.NewGetByID(repos.Users, logger)
	getUserByUsername := appusers.NewGetByUsername(repos.Users, logger)
	updateUser := appusers.NewUpdate(repos.Users, logger)
	updatePassword := appusers.NewUpdatePasswordByID(repos.Users, logger)
	deleteUser := appusers.NewDelete(repos.Users, logger)
	matchUserPass := appusers.NewMatchUserPass(repos.Users, logger)

	// sessions
	verifySession := appsession.NewVerify(repos.UserSessions, logger)
	issueSession := appsession.NewIssueSession(
		func(ctx context.Context, username, password string) (uuid.UUID, error) {
			user, err := matchUserPass.MatchUserPass(ctx, username, password)
			if err != nil {
				return uuid.Nil, err
			}
			return user.ID, nil
		},
		repos.UserSessions,
		0,
		logger,
	)
	deleteSession := appsession.NewDelete(repos.UserSessions, logger)

	// auth
	userExists := appusers.NewExists(repos.Users, logger)
	tmbVerifier := appauth.NewTMBVerifier(logger)
	sessionVerifier := appauth.NewVerifySession(logger)
	authenticator := appauth.NewAuthenticator(logger, map[appauth.Method]appauth.Core{
		appauth.MethodTMB:     &appauth.TMBCore{Users: userExists},
		appauth.MethodSession: &appauth.SessionCore{Verifier: verifySession},
	})

	// self-service
	me := &appusers.Me{
		Auth:               authenticator.Auth,
		GetCore:            getUser.Get,
		UpdateCore:         updateUser.Update,
		DeleteCore:         deleteUser.Delete,
		UpdatePasswordCore: updatePassword.UpdatePasswordByID,
	}

	return &App{
		CreateClient: createClient,
		GetClient:    getClient,
		DeleteClient: deleteClient,

		CreateUser:        createUser,
		GetUser:           getUser,
		GetUserByUsername: getUserByUsername,
		UpdateUser:        updateUser,
		UpdatePassword:    updatePassword,
		DeleteUser:        deleteUser,
		MatchUserPass:     matchUserPass,
		Me:                me,

		IssueSession:  issueSession,
		DeleteSession: deleteSession,

		TMBVerifier:     tmbVerifier,
		SessionVerifier: sessionVerifier,
		Authenticator:   authenticator,
	}
}
