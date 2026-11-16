package app

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/sdedovic/wgsltoy-server/internal/controllers"
	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"github.com/sdedovic/wgsltoy-server/internal/persistence"
	"github.com/sdedovic/wgsltoy-server/internal/security"
	"github.com/sdedovic/wgsltoy-server/internal/server"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
		return &fxevent.ZapLogger{Logger: log}
	}),

	fx.Provide(
		NewConfig,
		newLogger,

		fx.Annotate(newTxManager, fx.As(new(persistence.PGXProvider))),
		fx.Annotate(persistence.NewUserRepository, fx.As(new(domain.UserRepository))),
		fx.Annotate(persistence.NewShaderRepository, fx.As(new(domain.ShaderRepository))),
		fx.Annotate(security.NewArgon2Hasher, fx.As(new(domain.PasswordHasher))),
		fx.Annotate(newJWT, fx.As(new(domain.TokenIssuer)), fx.As(new(domain.TokenVerifier))),

		domain.NewUserService,
		domain.NewShaderService,

		controller(controllers.NewHealthController),
		controller(controllers.NewUserController),
		controller(controllers.NewShaderController),
		fx.Annotate(server.NewHandler, fx.ParamTags(``, ``, `group:"controllers"`)),
		newHTTPServer,
	),

	fx.Invoke(func(*http.Server) {}),
)

// Controllers are collected as a group, so server doesn't need to import controllers.
func controller(constructor any) any {
	return fx.Annotate(constructor, fx.As(new(server.Controller)), fx.ResultTags(`group:"controllers"`))
}

func newTxManager(lc fx.Lifecycle, config Config) (*persistence.TxManager, error) {
	txManager, err := persistence.NewTxManager(config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: txManager.Ping,
		OnStop: func(context.Context) error {
			txManager.Close()
			return nil
		},
	})
	return txManager, nil
}

func newJWT(config Config) *security.JWT {
	return security.NewJWT(config.AppSecret)
}

func newHTTPServer(lc fx.Lifecycle, config Config, log *zap.Logger, handler http.Handler) *http.Server {
	srv := server.New(config.Addr, handler)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			listener, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			log.Info("starting server", zap.String("addr", srv.Addr))
			go func() {
				if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
					log.Error("server stopped", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: srv.Shutdown,
	})
	return srv
}
