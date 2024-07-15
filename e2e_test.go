// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path"
	"runtime"
	"testing"

	"github.com/opentofu/tofudl"
)

func TestE2E(t *testing.T) {
	dl, err := tofudl.New()
	if err != nil {
		t.Fatal(err)
	}

	binary, err := dl.Download(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	fileName := "tofu"
	if runtime.GOOS == "windows" {
		fileName += ".exe"
	}
	fullPath := path.Join(tmp, fileName)
	if err := os.WriteFile(fullPath, binary, 0755); err != nil { //nolint:gosec //We want the binary to be executable.
		t.Fatal(err)
	}
	stdout := bytes.Buffer{}

	cmd := exec.Command(fullPath, "version")
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	t.Log(stdout.String())
}
