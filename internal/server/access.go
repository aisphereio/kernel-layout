package server

import (
	"context"

	v1 "github.com/aisphereio/kernel-layout/api/todo/v1"
	"github.com/aisphereio/kernel-layout/internal/data"
	"github.com/aisphereio/kernel/accessx"
	"github.com/aisphereio/kernel/middleware"
	mwaccess "github.com/aisphereio/kernel/middleware/access"
	mwauthn "github.com/aisphereio/kernel/middleware/authn"
	"github.com/aisphereio/kernel/middleware/requestinfo"
	"github.com/aisphereio/kernel/requestx"
)

func todoServerMiddlewares(resources *data.Resources) []middleware.Middleware {
	if resources == nil {
		return nil
	}
	return []middleware.Middleware{
		requestinfo.Server(requestinfo.WithResolver(v1.TodoServiceRequestInfoResolver)),
		mwauthn.Server(
			mwauthn.WithAuthenticator(resources.Authn),
			mwauthn.WithAllowAnonymous(true),
		),
		mwaccess.Server(resources.Access, mwaccess.WithResolver(todoAccessResolver)),
	}
}

func todoAccessResolver(ctx context.Context, operation string, req any) (accessx.Check, bool, error) {
	check, ok, err := v1.TodoServiceAccessResolver(ctx, operation, req)
	if err != nil || ok {
		return check, ok, err
	}
	return accessx.Check{}, false, nil
}

var _ requestx.Resolver = v1.TodoServiceRequestInfoResolver
