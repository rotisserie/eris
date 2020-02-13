package main

import (
	"flag"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rotisserie/eris"
)

var dsn string

func init() {
	flag.StringVar(&dsn, "dsn", "", "Sentry DSN for logging stack traces")
}

func example() error {
	return eris.New("test")
}

func wrapExample() error {
	err := example()
	if err != nil {
		return eris.Wrap(err, "wrap 1")
	}
	return nil
}

func wrapSecondExample() error {
	err := wrapExample()
	if err != nil {
		return eris.Wrap(err, "wrap 2")
	}
	return nil
}

func main() {
	flag.Parse()
	if dsn == "" {
		log.Fatal("Sentry DSN is a required flag, please pass it with '-dsn'")
	}

	err := wrapSecondExample()
	err = eris.Wrap(err, "wrap 3")

	initErr := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
	if initErr != nil {
		log.Fatalf("failed to initialize Sentry: %v", initErr)
	}

	sentry.CaptureException(err)
	sentry.Flush(time.Second * 5)
}
