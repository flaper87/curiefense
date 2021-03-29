package main

import (
	"context"
	"curielog/pkg"
)

func main() {
	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			pkg.LoadConfig,
			pkg.NewLogSender,
			newServer,
		),
		fx.Invoke(grpcInit),
	)
	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
}
