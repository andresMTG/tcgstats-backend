package main

import (
	"context"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

func main() {
	startTime := time.Now() 
	ctx := context.Background()
	godotenv.Load()
	app := fx.New(
		module(),
	)

	if err := app.Start(ctx); err != nil {
		panic(fmt.Errorf("unable to start application: %w", err))
	}

	fmt.Printf("\nTime to start: %s!\n\n", time.Since(startTime))

	// Wait for os signal to stop the application.
	sig := <-app.Wait()
	fmt.Printf("Application stopped with exit code: %d\n", sig.ExitCode)

	if err := app.Stop(ctx); err != nil {
		fmt.Println(err)
	}
}
