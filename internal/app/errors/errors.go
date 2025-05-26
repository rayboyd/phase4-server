// SPDX-License-Identifier: Apache-2.0
package errors

import (
	"fmt"
	"os"
)

func HandleFatalAndExit(err error) {
	if err == nil {
		return
	}

	if fatal, ok := err.(*FatalError); ok {
		fmt.Fprintf(os.Stderr, "%s\n", fatal.Error())
	} else {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
	}

	os.Exit(1)
}

func (f *FatalError) Error() string {
	if f.Err != nil {
		return fmt.Sprintf("FATAL: %s: %v", f.Message, f.Err)
	}
	return fmt.Sprintf("FATAL: %s", f.Message)
}

func (c *CommandCompleted) Error() string {
	return c.Message
}
