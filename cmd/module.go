package main

import (
	"go.uber.org/fx"
)

const applicationName string = "tcgstats-backend"

func module() fx.Option {
	return fx.Options(
		fx.Supply(applicationName),
		
	)
}