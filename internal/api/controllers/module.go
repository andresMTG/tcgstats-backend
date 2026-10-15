package controllers

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewCookieSettings,
		NewUserController,
		NewDeckController,
		NewTournamentController,
		NewOpponentController,
		NewMatchController,
		NewAnalyticsController,
	),
)
