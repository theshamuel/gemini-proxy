package cmd

import (
	"context"
	"github.com/theshamuel/gemini-proxy/app/config"
	"github.com/theshamuel/gemini-proxy/app/rest/api"
	"github.com/theshamuel/gemini-proxy/app/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ServerCommand interface {
	Execute(args []string) error
}

// ServerCmd represent arguments that can be used to start server (application)
type ServerCmd struct {
	config.CommonOpts
	Port    int `long:"port" env:"SERVER_PORT" default:"9443" description:"application port"`
	Version string
}

type application struct {
	ServerCmd
	rest       *api.Rest
	terminated chan struct{}
}

// Execute is the entry point for server command
func (sc ServerCmd) Execute(_ []string) error {
	log.Printf("[INFO] start app server")
	log.Printf("[INFO] server args:\n"+
		"                     port: %d;\n"+
		"                     TLS enabled: %t;\n"+
		"                         Cert path: %s;\n"+
		"                         Pribate key path: %s;\n",
		sc.Port, sc.TLS.Enabled, sc.TLS.CertPath, sc.TLS.PrivateKeyPath)

	log.Printf("[DEBUG] common options: %+v", sc.CommonOpts)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] Get interrupt signal")
		cancel()
	}()
	app := sc.bootstrapApp()
	log.Printf("[INFO] Starting Gemini Proxy:[version: %s] ...\n", sc.Version)
	if err := app.run(ctx); err != nil {
		log.Printf("[ERROR] Server terminated with error %v", err)
		return err
	}
	log.Printf("[INFO] Server terminated")
	return nil
}

func (app *application) run(ctx context.Context) error {

	go func() {
		<-ctx.Done()
		app.rest.Shutdown()
		log.Print("[INFO] shutdown is completed")
	}()

	app.rest.Run(app.Port)
	close(app.terminated)
	return nil
}

func (sc ServerCmd) bootstrapApp() *application {
	rest := &api.Rest{
		Version:       sc.Version,
		DelayRequests: sc.DelayRequests,
		Service: &service.GeminiProxy{
			OriginURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=",
			Client: http.Client{
				Timeout: 20 * time.Second,
			},
			APIKey: sc.GeminiAPIKey,
		},
		TLSEnabled:     sc.TLS.Enabled,
		CertPath:       sc.TLS.CertPath,
		PrivateKeyPath: sc.TLS.PrivateKeyPath,
	}

	return &application{
		ServerCmd:  sc,
		rest:       rest,
		terminated: make(chan struct{}),
	}
}

// Wait for application completion (termination)
func (app *application) Wait() {
	<-app.terminated
}
