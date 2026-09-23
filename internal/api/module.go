package api

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/controllers"
	"github.com/andresMTG/tcgstats-backend/pkg/env"



	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

type Configuration struct {
	ApplicationName string `env:"APP_NAME"         envDefault:"tcgstats-backend"`
	Port            int    `env:"SERVER_PORT"      envDefault:"8080"`
	BasePath        string `env:"SERVER_BASE_PATH" envDefault:"/api/"`
	PprofEnabled    bool   `env:"PPROF_ENABLED"    envDefault:"false"`
}

func Module() fx.Option {
	return fx.Module("api", fx.Options(
		fx.Provide(
			// Config
			env.LoadEnvConfiguration[Configuration],
			chi.NewRouter,
			NewRouter,

			// Controllers
			controllers.NewHealth,

			// Services

		),
		fx.Invoke(
			start,
			registerHooks,
		),
	))
}

func registerHooks(
	lc fx.Lifecycle,
	shutdown fx.Shutdowner,
	config *Configuration,
	router *Router,
) {
	server := &http.Server{
		Addr:    config.getAddress(),
		Handler: router.server,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				err := server.ListenAndServe()
				if err != nil && err != http.ErrServerClosed {
					log.Println(context.TODO(), "server stopped", "error", err)
					_ = shutdown.Shutdown()
				}
			}()

			for {
				log.Println(context.TODO(), "waiting for server to start", "address", config.getAddress())
				_, err := net.Dial("tcp", config.getAddress())
				if err == nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}

// getAddress returns the server address based on the configured port.
func (c *Configuration) getAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}
