// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package helloworld

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"runtime"
	"testing"
)

// Build creates a hello-world binary for the current platform you can use to test.
func Build(t *testing.T) []byte {
	fakeName := "fake"
	if runtime.GOOS == "windows" {
		fakeName += ".exe"
	}

	dir := path.Join(os.TempDir(), fakeName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	if err := os.WriteFile(path.Join(dir, "go.mod"), []byte(gomod), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(dir, "main.go"), []byte(code), 0600); err != nil {
		t.Fatal()
	}

	cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", fakeName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() != 0 {
			t.Fatalf("Build failed (exit code %d)", exitErr.ExitCode())
		} else {
			t.Fatalf("Build failed (%v)", err)
		}
	}

	contents, err := os.ReadFile(path.Join(dir, fakeName))
	if err != nil {
		t.Fatalf("Failed to read compiled fake (%v)", err)
	}
	return contents
}

var code = `package main

func main() {
	print("Hello world!")
}
`

var gomod = `module fake

go 1.21`
