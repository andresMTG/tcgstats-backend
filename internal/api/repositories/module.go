package repositories

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewUserRepository,
		NewDeckRepository,
		NewTournamentRepository,
		NewOpponentRepository,
		NewMatchRepository,
		NewAnalyticsRepository,
	),
)
