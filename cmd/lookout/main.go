package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/ipfs/go-log/v2"
	"github.com/ipni/lookout"
	"github.com/ipni/lookout/check"
	"github.com/ipni/lookout/sample"
)

var logger = log.Logger("ipni/lookout/cmd")

func main() {

	checkInterval := flag.Duration("checkInterval", 10*time.Minute, "The interval at which checks are run.")
	logLevel := flag.String("logLevel", "info", "The logging level. Only applied if GOLOG_LOG_LEVEL environment variable is unset.")
	flag.Parse()

	if _, set := os.LookupEnv("GOLOG_LOG_LEVEL"); !set {
		_ = log.SetLogLevel("*", *logLevel)
	}

	checkerWithDhtCascade, err := check.NewIpniNonStreamingChecker(
		check.WithName("cid_contact_with_cascade"),
		check.WithIpniEndpoint("https://cid.contact"),
		check.WithCheckTimeout(30*time.Second),
		check.WithIpfsDhtCascade(true),
	)
	if err != nil {
		logger.Fatalw("Failed to instantiate checker", "err", err)
	}
	checkerWithoutDhtCascade, err := check.NewIpniNonStreamingChecker(
		check.WithName("cid_contact"),
		check.WithIpniEndpoint("https://cid.contact"),
		check.WithCheckTimeout(30*time.Second),
		check.WithIpfsDhtCascade(false),
	)
	if err != nil {
		logger.Fatalw("Failed to instantiate checker", "err", err)
	}
	l, err := lookout.New(
		lookout.WithCheckers(checkerWithDhtCascade, checkerWithoutDhtCascade),
		lookout.WithSamplers(&sample.SaturnTopCidsSampler{}),
		lookout.WithCheckInterval(*checkInterval),
	)
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
