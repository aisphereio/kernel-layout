package server

import (
	"context"

	"github.com/aisphereio/kernel-layout/internal/conf"
	"github.com/aisphereio/kernel-layout/internal/data"
	"github.com/aisphereio/kernel/middleware"
	"github.com/aisphereio/kernel/securityx"
	"github.com/aisphereio/kernel/serverx"
)

func todoServerMiddlewares(resources *data.Resources, cfg conf.SecurityConfig) []middleware.Middleware {
	if resources == nil {
		return nil
	}
	securityRuntime := mustSecurityRuntime(cfg)
	providers := TodoCatalog().RuntimeProviders(serverx.RuntimeProviders{
		Security:    securityRuntime,
		AccessGuard: &resources.Access,
	})
	return serverx.ServerMiddlewareFromProviders(context.Background(), providers)
}

func mustSecurityRuntime(cfg conf.SecurityConfig) *securityx.Runtime {
	runtime, err := securityx.NewRuntime(context.Background(), securityx.Config{
		Authn: securityx.AuthnBoundaryConfig{
			Enabled:        cfg.Authn.Enabled,
			Mode:           cfg.Authn.Mode,
			Provider:       cfg.Authn.Provider,
			OIDC:           cfg.Authn.OIDC,
			InternalCall:   cfg.InternalCall,
			CacheTTL:       cfg.Authn.CacheTTL,
			AllowAnonymous: true,
		},
		InternalCall: cfg.InternalCall,
		Access:       cfg.Access,
	}, nil)
	if err != nil {
		panic(err)
	}
	return runtime
}
