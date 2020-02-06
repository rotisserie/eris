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

func Example() error {
	return eris.New("test")
}

func WrapExample() error {
	err := Example()
	if err != nil {
		return eris.Wrap(err, "wrap 1")
	}
	return nil
}

func WrapSecondExample() error {
	err := WrapExample()
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

	err := WrapSecondExample()
	err = eris.Wrap(err, "wrap 3")

	sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})

	sentry.CaptureException(err)
	sentry.Flush(time.Second * 5)
}
