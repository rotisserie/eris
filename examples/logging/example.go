package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
)

var (
	// global error values can be useful when wrapping errors or inspecting error types
	ErrInternalServer = eris.New("error internal server")

	// declaring an error with pkg/errors for comparison
	ErrNotFound = errors.New("error not found")
)

type Request struct {
	ID string
}

func (req *Request) Validate() error {
	if req.ID == "" {
		// create a new local error and wrap it with some context
		err := eris.New("error bad request")
		return eris.Wrap(err, "received a request with no ID")
	}
	return nil
}

type Resource struct {
	ID      string
	AbsPath string
}

func GetResource(req Request) (*Resource, error) {
	if req.ID == "res2" {
		return &Resource{
			ID:      req.ID,
			AbsPath: "./some/malformed/absolute/path/data.json", // malformed absolute filepath to simulate a "bug"
		}, nil
	} else if req.ID == "res3" {
		return &Resource{
			ID:      req.ID,
			AbsPath: "/some/correct/path/data.json",
		}, nil
	}

	return nil, errors.Wrapf(ErrNotFound, "failed to get resource '%v'", req.ID)
}

func GetRelPath(base string, path string) (string, error) {
	relPath, err := filepath.Rel(base, path)
	if err != nil {
		// it's generally useful to wrap external errors with a type that you know how to handle
		// first (e.g. ErrInternalServer). this will help if/when you want to do error inspection
		// via eris.Is(err, ErrInternalServer) or eris.Cause(err).
		return "", eris.Wrap(ErrInternalServer, err.Error())
	}
	return relPath, nil
}

type Response struct {
	RelPath string
}

func ProcessResource(req Request) (*Response, error) {
	if err := req.Validate(); err != nil {
		// simply return the error if there's no additional context
		return nil, err
	}

	resource, err := GetResource(req)
	if err != nil {
		return nil, err
	}

	// do some processing on the data
	relPath, err := GetRelPath("/Users/roti/", resource.AbsPath)
	if err != nil {
		// wrap the error if you want to add more context
		return nil, eris.Wrapf(err, "failed to get relative path for resource '%v'", resource.ID)
	}

	return &Response{
		RelPath: relPath,
	}, nil
}

type LogReq struct {
	Method string
	Req    Request
	Res    *Response
	Err    error
}

func LogRequest(logger *logrus.Logger, logReq LogReq) {
	fields := logrus.Fields{
		"method": logReq.Method,
	}
	if logReq.Err != nil {
		// it's generally a good idea to contain error formatting logic inside a utility method like
		// this one to ensure that all errors are logged uniformly. in this case, we're logging with
		// the default format and stack traces enabled.
		fields["error"] = eris.ToJSON(logReq.Err, true)
		logger.WithFields(fields).Error("method completed with error")
	} else {
		fields["response"] = *logReq.Res
		logger.WithFields(fields).Info("method completed successfully")
	}
}

// This example demonstrates how to integrate eris with a JSON logger (e.g. logrus). It's broken
// into several methods to show the formatted output for wrapped errors, and it includes three
// failing cases to demonstrate how your error logs should look in different scenarios.
func main() {
	// setup JSON logger
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// example requests
	reqs := []Request{
		{
			ID: "", // bad request
		},
		{
			ID: "res1", // not found
		},
		{
			ID: "res2", // server error
		},
		{
			ID: "res3", // success
		},
	}

	// process the example requests and log the results
	for _, req := range reqs {
		res, err := ProcessResource(req)
		if req.ID != "res1" {
			// log the eris error
			LogRequest(logger, LogReq{
				Method: "ProcessResource",
				Req:    req,
				Res:    res,
				Err:    err,
			})
		} else {
			// print the pkg/errors error
			fmt.Printf("%+v\n", err)
		}
	}
}
