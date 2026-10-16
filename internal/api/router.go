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
		Health     *controllers.Health
		User       *controllers.UserController
		Deck       *controllers.DeckController
		Tournament *controllers.TournamentController
		Opponent   *controllers.OpponentController
		Match      *controllers.MatchController
		Analytics  *controllers.AnalyticsController
	}
)

func NewRouter(
	server *chi.Mux,
	config *Configuration,
	controllers Controllers,
) (*Router, error) {
	corsMiddleware, err := middleware.NewCORSMiddleware(config.CORSAllowedOrigins)
	if err != nil {
		return nil, err
	}
	server.Use(corsMiddleware)

	return &Router{
		server:      server,
		config:      config,
		controllers: controllers,
	}, nil
}

func start(router *Router) {
	router.server.Group(func(r chi.Router) {
		r.Route(router.config.BasePath, func(r chi.Router) {
			r.Get("/health", router.controllers.Health.Health)

			// Rotas Públicas
			r.Post("/users", router.controllers.User.CreateUser)
			r.Post("/users/login", router.controllers.User.Login)
			r.Post("/users/refresh", router.controllers.User.Refresh)
			r.Post("/users/logout", router.controllers.User.Logout)

			// Rotas Protegidas (Exigem JWT)
			r.Group(func(protected chi.Router) {
				protected.Use(middleware.AuthMiddleware(router.config.JWTSecret))

				protected.Get("/users/me", router.controllers.User.GetMe)

				// Decks
				protected.Post("/decks", router.controllers.Deck.CreateDeck)
				protected.Get("/decks", router.controllers.Deck.ListUserDecks)
				protected.Get("/decks/{id}", router.controllers.Deck.GetDeck)

				// Tournaments
				protected.Post("/tournaments", router.controllers.Tournament.CreateTournament)
				protected.Get("/tournaments", router.controllers.Tournament.ListUserTournaments)
				protected.Get("/tournaments/{id}", router.controllers.Tournament.GetTournament)

				// Opponents
				protected.Post("/opponents", router.controllers.Opponent.CreateOpponent)
				protected.Get("/opponents", router.controllers.Opponent.ListUserOpponents)
				protected.Get("/opponents/{id}", router.controllers.Opponent.GetOpponent)

				// Matches
				protected.Post("/matches", router.controllers.Match.CreateMatch)
				protected.Get("/matches/{id}", router.controllers.Match.GetMatch)
				protected.Get("/tournaments/{tournamentId}/matches", router.controllers.Match.ListTournamentMatches)
				protected.Get("/decks/{deckId}/matches", router.controllers.Match.ListDeckMatches)

				// Analytics
				protected.Get("/analytics/decks", router.controllers.Analytics.DeckPerformance)
				protected.Get("/analytics/deck-matchups", router.controllers.Analytics.DeckMatchupPerformance)
				protected.Get("/analytics/opponents", router.controllers.Analytics.OpponentPerformance)
				protected.Get("/analytics/phases", router.controllers.Analytics.PhasePerformance)
				protected.Get("/analytics/tournament-scopes", router.controllers.Analytics.TournamentScopePerformance)
			})
		})
	})
}
