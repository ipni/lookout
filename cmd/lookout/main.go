package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/ipfs/go-log/v2"
	"github.com/ipni/lookout"
	"github.com/ipni/lookout/cmd/lookout/internal"
)

var logger = log.Logger("ipni/lookout/cmd")

func main() {

	config := flag.String("config", "config.yaml", "The path to lookout YAML config file.")
	logLevel := flag.String("logLevel", "info", "The logging level. Only applied if GOLOG_LOG_LEVEL environment variable is unset.")
	flag.Parse()

	if _, set := os.LookupEnv("GOLOG_LOG_LEVEL"); !set {
		_ = log.SetLogLevel("*", *logLevel)
	}

	cfg, err := internal.NewConfig(*config)
	if err != nil {
		logger.Fatalw("Failed to load config from path", "path", *config, "err", err)
	}
	opts, err := cfg.ToOptions()
	if err != nil {
		logger.Fatalw("Failed to generate options from config", "path", *config, "err", err)
	}

	l, err := lookout.New(opts...)
	if err != nil {
		logger.Fatalw("Failed to instantiate lookout", "err", err)
	}
	ctx := context.Background()
	if err := l.Start(ctx); err != nil {
		logger.Fatalw("Failed to start lookout", "err", err)
	}
	sch := make(chan os.Signal, 1)
	signal.Notify(sch, os.Interrupt)

	<-sch
	logger.Info("Terminating...")
	if err := l.Shutdown(ctx); err != nil {
		logger.Warnw("Failure occurred while shutting down server.", "err", err)
	} else {
		logger.Info("Shut down server successfully.")
	}
}
