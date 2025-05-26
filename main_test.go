// SPDX-License-Identifier: Apache-2.0
package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("RUN_MAIN_FOR_TEST") == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}
