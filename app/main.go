package main

import (
	"fmt"
	"github.com/theshamuel/gemini-proxy/app/config"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
	"github.com/theshamuel/gemini-proxy/app/cmd"
)

var version = "unknown"

// Opts structure represent options to start application
type Opts struct {
	ServerCmd cmd.ServerCmd `command:"server"`
	Config    struct {
		Enabled  bool   `long:"enabled" env:"ENABLED" description:"enable getting parameters from config. In that case all parameters will be read only form config"`
		FileName string `long:"file-name" env:"FILE_NAME" default:"gemini-proxy.yml" description:"config file name"`
	} `group:"config" namespace:"config" env-namespace:"CONFIG"`
}

var opts Opts

func main() {
	parseFlags()

}

func parseFlags() {
	p := flags.NewParser(&opts, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		c := command.(cmd.ServerCommand)

		var cnf *config.Config
		if opts.Config.Enabled {
			cnf = &config.Config{
				FileName: opts.Config.FileName,
			}
		}

		if opts.Config.Enabled {
			var err error
			var co *config.CommonOpts

			if co, err = cnf.GetCommon(); err != nil {
				panic(fmt.Errorf("[ERROR] can not read config file, %w", err))
			}
			opts.ServerCmd.GeminiAPIKey = co.GeminiAPIKey
			opts.ServerCmd.DelayRequests = co.DelayRequests
			opts.ServerCmd.TLS.Enabled = co.TLS.Enabled
			opts.ServerCmd.TLS.CertPath = co.TLS.CertPath
			opts.ServerCmd.TLS.PrivateKeyPath = co.TLS.PrivateKeyPath
			opts.ServerCmd.Debug = co.Debug
			opts.ServerCmd.Version = version
		}

		setupLogLevel(opts.ServerCmd.Debug)

		err := c.Execute(args)
		if err != nil {
			log.Printf("[ERROR] failed with %+v", err)
		}
		return err
	}
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func setupLogLevel(debug bool) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stdout,
	}
	log.SetFlags(log.Ldate | log.Ltime)

	if debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		filter.MinLevel = "DEBUG"
	}
	log.SetOutput(filter)
}

func getStackTrace() string {
	maxSize := 7 * 1024 * 1024
	stacktrace := make([]byte, maxSize)
	length := runtime.Stack(stacktrace, true)
	if length > maxSize {
		length = maxSize
	}
	return string(stacktrace[:length])
}

func init() {
	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			log.Printf("[INFO] Singal QUITE is cought , stacktrace [\n%s", getStackTrace())
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}
