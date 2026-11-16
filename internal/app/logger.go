package app

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func newLogger(lc fx.Lifecycle, config Config) (*zap.Logger, error) {
	newFn := zap.NewProduction
	if config.Dev {
		newFn = zap.NewDevelopment
	}

	log, err := newFn()
	if err != nil {
		return nil, err
	}

	lc.Append(fx.StopHook(func(context.Context) {
		// flushing stderr fails on some platforms, which isn't worth reporting
		_ = log.Sync()
	}))
	return log, nil
}
