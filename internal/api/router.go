package api

import (
	"github.com/andresMTG/tcgstats-backend/internal/api/controllers"
	"github.com/andresMTG/tcgstats-backend/internal/api/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

type (
	Router struct {
		config      *Configuration
		server      *chi.Mux
		controllers Controllers
	}

	Controllers struct {
		fx.In
		Health *controllers.Health
		Users *controllers.Users
	}
)

func NewRouter(
	server *chi.Mux,
	config *Configuration,
	controllers Controllers,
) *Router {
	return &Router{
		server:      server,
		config:      config,
		controllers: controllers,
	}
}

func start(router *Router) {
	router.server.Group(func(r chi.Router) {
		r.Route(router.config.BasePath, func(r chi.Router) {
			r.Get("/health", router.controllers.Health.Health)

			r.Route("/users", func(r chi.Router){
				r.Post("/", router.controllers.Users.Create)
				r.Post("/login", router.controllers.Users.Login)

				r.Group(func(protected chi.Router){
				protected.Use(middleware.AuthMiddleware(router.config.JWTSecret))
				protected.Get("/me", router.controllers.Users.ShowMe)
			})
			})

			
		})
	})
}