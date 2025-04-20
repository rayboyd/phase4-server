package app

import (
	"errors"
	"log"
	"os"
)

var (
	ErrConfigInvalid = errors.New("config invalid")
	ErrFileNotFound  = errors.New("file not found")
	ErrUnknown       = errors.New("unknown error")
)

func HandleFatalAndExit(err error) {
	log.SetOutput(os.Stderr)

	switch {
	case errors.Is(err, ErrConfigInvalid):
		log.Printf("%v: %v\n", ErrConfigInvalid, err)
	case errors.Is(err, ErrFileNotFound):
		log.Printf("%v: %v\n", ErrFileNotFound, err)
	default:
		log.Printf("%v: %v\n", ErrUnknown, err)
	}

	os.Exit(1)
}
