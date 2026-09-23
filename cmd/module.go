package main

import (
	"github.com/andresMTG/tcgstats-backend/internal/api"
	"go.uber.org/fx"
)

const applicationName string = "tcgstats-backend"

func module() fx.Option {
	return fx.Options(
		fx.Supply(applicationName),
		
		api.Module(),
	)
}