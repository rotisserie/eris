package main

import (
	"encoding/json"
	"fmt"

	pkgerrors "github.com/pkg/errors"
	"github.com/rotisserie/eris"
)

var (
	errExternal = pkgerrors.New("external error")
)

// example method that returns an external error (e.g. pkg/errors).
func getResource(id string) error {
	return pkgerrors.Wrap(errExternal, "resource not found")
}

// example method that wraps an external error using eris.
func readResource(id string) error {
	err := getResource(id)
	if err != nil {
		return eris.Wrapf(err, "failed to get resource '%v'", id)
	}
	return nil
}

func processResource(id string) error {
	err := readResource(id)
	if err != nil {
		return eris.Wrapf(err, "failed to process resource '%v'", id)
	}
	return nil
}

// This example demonstrates how error wrapping works with external error handling libraries
// (e.g. pkg/errors). When an external error is wrapped, eris attempts to unwrap it and returns
// a new error containing the external error chain, the new context, and an eris stack trace.
func main() {
	err := processResource("res1")
	fmt.Printf("%+v\n", err)
	jsonErr := eris.ToJSON(err, true)
	jsonStr, _ := json.Marshal(jsonErr)
	fmt.Printf("%v\n", string(jsonStr))
}
