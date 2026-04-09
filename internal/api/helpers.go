package api

import (
	genapi "github.com/4nd3r5on/oidc-serv/pkg/api"
)

func conflictResponse(err error) *genapi.ResourceConflictError {
	return &genapi.ResourceConflictError{
		Code:  genapi.NewOptResourceConflictErrorCode(genapi.ResourceConflictErrorCodeResourceExists),
		Error: err.Error(),
	}
}
