// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "run", "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1", "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		}

		log.Fatal(err)
	}
}
