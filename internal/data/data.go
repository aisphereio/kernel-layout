package data

import (
	"context"

	"github.com/aisphereio/kernel/auditx"
	"github.com/aisphereio/kernel/cachex"
	_ "github.com/aisphereio/kernel/cachex/redis"
	"github.com/aisphereio/kernel/dbx"
	_ "github.com/aisphereio/kernel/dbx/postgres"
	"github.com/aisphereio/kernel/dtmx"
	"github.com/aisphereio/kernel/logx"
	"github.com/aisphereio/kernel/metricsx"
	"github.com/aisphereio/kernel/objectstorex"
	_ "github.com/aisphereio/kernel/objectstorex/minio"

	"github.com/aisphereio/kernel-layout/internal/conf"
)

type ResourceOptions struct {
	Logger  logx.Logger
	Metrics metricsx.Manager
	DTM     dtmx.Manager
}

type Resources struct {
	DB          dbx.DB
	Cache       cachex.Cache
	ObjectStore objectstorex.Client
	Audit       auditx.Recorder
	DTM         dtmx.Manager

	closers []func() error
}

type Data struct {
	Resources *Resources
}

func NewResources(ctx context.Context, cfg conf.Bootstrap, opts ResourceOptions) (*Resources, func(), error) {
	logger := opts.Logger
	if logger == nil {
		logger = logx.DefaultLogger()
	}
	metrics := metricsx.Ensure(opts.Metrics)

	r := &Resources{
		Audit: auditx.NewMemoryStore(),
		DTM:   dtmx.FromContextOr(ctx, opts.DTM),
	}
	if !cfg.Audit.Enabled {
		r.Audit = auditx.Noop()
	}

	if cfg.Data.Database.Enabled {
		dbCfg := cfg.Data.Database.Config
		dbCfg.Logger = logger.Named("data.dbx")
		dbCfg.Metrics = metrics
		dbCfg.MetricsEnabled = dbCfg.MetricsEnabled && cfg.Metrics.Enabled
		db, err := dbx.New(dbCfg)
		if err != nil {
			return nil, nil, err
		}
		r.DB = db
		r.closers = append(r.closers, db.Close)
	}
	if cfg.Data.Cache.Enabled {
		cacheCfg := cfg.Data.Cache.Config
		cacheCfg.Logger = logger.Named("data.cachex")
		cacheCfg.Metrics = metrics
		cacheCfg.MetricsEnabled = cacheCfg.MetricsEnabled && cfg.Metrics.Enabled
		cache, err := cachex.New(cacheCfg)
		if err != nil {
			r.Close()
			return nil, nil, err
		}
		r.Cache = cache
		r.closers = append(r.closers, cache.Close)
	}
	if cfg.Data.ObjectStore.Enabled {
		storeCfg := cfg.Data.ObjectStore.Config
		storeCfg.Logger = logger.Named("data.objectstorex")
		storeCfg.Metrics = metrics
		storeCfg.MetricsEnabled = storeCfg.MetricsEnabled && cfg.Metrics.Enabled
		store, err := objectstorex.New(storeCfg)
		if err != nil {
			r.Close()
			return nil, nil, err
		}
		r.ObjectStore = store
		r.closers = append(r.closers, store.Close)
	}

	return r, func() { _ = r.Close() }, pingEnabled(ctx, r)
}

func NewData(resources *Resources) *Data {
	return &Data{Resources: resources}
}

func pingEnabled(ctx context.Context, r *Resources) error {
	if r.DB != nil {
		if err := r.DB.PingContext(ctx); err != nil {
			return err
		}
	}
	if r.Cache != nil {
		if err := r.Cache.Ping(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resources) Close() error {
	var out error
	for i := len(r.closers) - 1; i >= 0; i-- {
		if err := r.closers[i](); err != nil && out == nil {
			out = err
		}
	}
	return out
}
