package application

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
	"prolific/config"
	"prolific/debug"
	"prolific/features/common"
	"prolific/features/log"
	"prolific/features/web-hook"
	"time"
)

var (
	elapsed time.Duration
	shutdownTimeout = flag.Duration("shutdown-timeout", 20 * time.Second,
	"shutdown timeout (5s,5m,5h) before connections are cancelled")
)

func init() {
	flag.Parse()
}

type Application struct {
	name	string
	router 	*mux.Router
	server 	*http.Server
}

func NewWithName(name string) *Application {
	module := config.Config.OnModule(name)
	address := fmt.Sprintf("%s:%s",
		module.GetWithDefault("host", "127.0.0.1"),
		module.GetWithDefault("port", "8000"))
	router := mux.NewRouter()
	server := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	debug.Printf("Application named %s initialized.", name)
	return &Application{ name, router, server }
}

func (app *Application) AddRoute(pathPrefix string, route common.IRoute) *Application {
	subRouter := app.router.PathPrefix(pathPrefix).Subrouter()
	debug.Printf("Added: Route of %s", pathPrefix)
	route.Initialise(subRouter)
	return app
}

func (app *Application) ListenAndServe() {
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	start := time.Now()

	go func() {
		debug.Printf("Server listening on %s\n", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			elapsed = time.Since(start)
			debug.Println("Server error, shutdown the server")
			debug.Println("Error: " + err.Error())
			app.shutDown(1)
		}
	}()

	<-stop
	elapsed = time.Since(start)
	debug.Println("Stop command received, gracefully shutdown the server")
	app.shutDown(0)
}

func (app *Application) RegisterRoutes() *Application {
	// List of routes
	app.AddRoute("/log", log.New())
	app.AddRoute("/web-hook", web_hook.New())
	// Not found handler
	app.router.NotFoundHandler = http.HandlerFunc(common.NotFoundHandler)
	return app
}

func (app *Application) shutDown(code int) {
	debug.Printf("Waiting maximum %s for the server to shutdown", (*shutdownTimeout).String())

	ctx, cancel := context.WithTimeout(context.Background(), *shutdownTimeout)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		debug.Println("Server shutdown error.")
		debug.Println(err.Error())
	}

	debug.Println("Server down")
	debug.Printf("Application Up-Time: %s\n", elapsed.String())
	os.Exit(code)
}