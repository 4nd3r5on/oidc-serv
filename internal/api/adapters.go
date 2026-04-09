package api

import (
	appusers "github.com/4nd3r5on/oidc-serv/internal/app/users"
	genapi "github.com/4nd3r5on/oidc-serv/pkg/api"
)

func toUser(res *appusers.GetRes) *genapi.User {
	return &genapi.User{
		ID:       res.ID,
		Username: res.Username,
		Locale:   res.Locale,
	}
}

func updateOptsFromReq(req *genapi.UpdateUserRequest) appusers.UpdateOpts {
	return appusers.UpdateOpts{
		Username: optStringToPtr(req.Username),
		Locale:   optStringToPtr(req.Locale),
	}
}

func optStringToPtr(opt genapi.OptString) *string {
	if !opt.IsSet() {
		return nil
	}
	v := opt.Value
	return &v
}
