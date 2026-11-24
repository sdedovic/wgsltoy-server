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
	"github.com/sdedovic/wgsltoy-server/internal/views"
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

		newPgClient,
		persistence.NewTxManager,
		persistence.NewUserRepository,
		persistence.NewShaderRepository,
		persistence.NewAuthRepository,
		persistence.NewSessionRepository,
		fx.Annotate(security.NewArgon2Hasher, fx.As(new(domain.PasswordHasher))),
		domain.NewAuthService,
		server.NewAuthMiddleware,

		domain.NewUserService,
		domain.NewShaderService,

		newTemplates,

		controller(controllers.NewHealthController),
		controller(controllers.NewAuthController),
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

func newPgClient(lc fx.Lifecycle, config Config) (persistence.PgClient, error) {
	return persistence.NewPgClient(lc, config.DatabaseURL)
}

func newTemplates(config Config) (*views.Templates, error) {
	return views.NewTemplates(config.TemplateDir)
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
