package app

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestHandleFatalAndExit(t *testing.T) {
	log.SetFlags(0)

	testCases := []struct {
		name string // For t.Run clarity
		err  error
	}{
		{
			"ErrConfigInvalid",
			fmt.Errorf("%w: specific validation detail", ErrConfigInvalid),
		},
		{
			"ErrFileNotFound",
			ErrFileNotFound,
		},
		{
			"ErrUnknown",
			errors.New("some other specific error"),
		},
		{
			"Nil Error",
			nil,
		},
	}

	for _, tc := range testCases {
		currentErr := tc.err

		t.Run(tc.name, func(t *testing.T) {
			var expectedLog string
			switch {
			case errors.Is(currentErr, ErrConfigInvalid):
				expectedLog = fmt.Sprintf("%s: %s", ErrConfigInvalid.Error(), currentErr.Error())
			case errors.Is(currentErr, ErrFileNotFound):
				expectedLog = fmt.Sprintf("%s: %s", ErrFileNotFound.Error(), currentErr.Error())
			default:
				errStr := "<nil>"
				if currentErr != nil {
					errStr = currentErr.Error()
				}
				expectedLog = fmt.Sprintf("%s: %s", ErrUnknown.Error(), errStr)
			}

			if os.Getenv("BE_EXITER") == "1" {
				HandleFatalAndExit(currentErr)
				return
			}

			cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$")
			cmd.Env = append(os.Environ(), "BE_EXITER=1")

			var stderrBuf bytes.Buffer
			cmd.Stderr = &stderrBuf

			err := cmd.Run()

			e, ok := err.(*exec.ExitError)
			if !ok || e.Success() {
				t.Fatalf("Expected process to exit with non-zero status, but got err: %v", err)
			}

			output := strings.TrimSpace(stderrBuf.String())
			if output != expectedLog {
				t.Errorf("Log output mismatch:\nExpected: %q\nActual:   %q", expectedLog, output)
			}
		})
	}
}
