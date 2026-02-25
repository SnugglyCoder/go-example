package main

import (
	"context"
	"fmt"
	gohttp "net/http"
	"os"
	"os/signal"
	"time"

	"github.com/SnugglyCoder/go-example/foo"
	"github.com/SnugglyCoder/go-example/http"
	"github.com/SnugglyCoder/go-example/sqlite"
	"golang.org/x/sync/errgroup"
)

func main() {
	// This sets up a context to catch SIGINT, normally ctrl+c, but also used by other things like Kubernetes to signal the program is going to be forced to shut down soon
	// Doing this will allow for what is called a 'graceful shutdown' which is always nice to try and do
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	// At this point a logger would be setup depending on the exact situation, but for brevity it is excluded.
	// If wanted, can provide this, but for now try to think on what you're seeing in this example and extrapolate the idea into what a logger might look like
	database, err := sqlite.NewClient("connectionString") // Factory pattern to construct an instance of an object based on configurations and attachments passed into a function
	if err != nil {
		fmt.Printf("error connecting to database: %v\n", err)
		return
	}
	fooService := foo.NewService(foo.ServiceAttachments{ // Dependency injection in action. An object that `fooService` depends on is injected into the constructor function
		BarRepository: &database,
	})
	httpServer := http.NewServer(http.ServerAttachments{ // Dependency injection in action. An object that `httpServer` depends on is injected into its constructor function
		FooService: fooService,
	})
	waitGroup, ctx := errgroup.WithContext(ctx) // Wait group, and in this cas an errgroup, is a special construct for waiting for several threads to complete before moving on when the `wait` function is called.
	waitGroup.Go(func() error {
		// logger.PrintInfof("Starting http server!")
		err := httpServer.ListenAndServe()
		if err != gohttp.ErrServerClosed { // RTFM about standard http server to see why this line is the way it is
			// logger.PrintInfof("Error running http server: %v", err)
			return err
		}
		return nil
	})
	waitGroup.Go(func() error {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// logger.PrintInfof("Shutting down http server!")
		err := httpServer.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("error shutting down http server: %v", err)
		}
		// logger.PrintInfof("Shut down http server successfully!")
		return nil
	})
	err = waitGroup.Wait()
	if err != nil {
		// logger.PrintInfof("Encountered an issue: %v", err)
		fmt.Printf("Encountered an issue: %v\n", err)
		return
	}
}
