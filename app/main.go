package main

import (
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
	"github.com/theshamuel/gemini-proxy/app/cmd"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var version = "unknown"

// Opts structure represent options to start application
type Opts struct {
	ServerCmd cmd.ServerCommand `command:"server"`
	Debug     bool              `long:"debug" env:"DEBUG" description:"debug mode"`
}

func main() {
	log.Printf("[INFO] Starting gemini-proxy version: %s\n", version)
	var opts Opts
	p := flags.NewParser(&opts, flags.Default)
	setupLogLevel(opts.Debug)
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		log.Printf("[ERROR] during parsing flags: %+v", err)
		os.Exit(3)
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
