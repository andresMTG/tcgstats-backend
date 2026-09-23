package main


import (
	"context"
	"fmt"

	"go.uber.org/fx"
)

func main() {
ctx := context.Background()
	app := fx.New(
		module(),
	)

	if err := app.Start(ctx); err != nil {
		panic(fmt.Errorf("unable to start application: %w", err))
	}

	// Wait for os signal to stop the application.
	sig := <-app.Wait()
	fmt.Printf("Application stopped with exit code: %d\n", sig.ExitCode)

	if err := app.Stop(ctx); err != nil {
		fmt.Println(err)
	}
}
