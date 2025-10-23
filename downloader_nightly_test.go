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
	"github.com/opentofu/tofudl/branding"
)

// Given a tofu binary, calls "version" subcommand and logs the output
// This also calls runtime.GC() to cleanup any file handle still held after the test is done (THANKS WINDOWS!)
func logTofuVersion(t *testing.T, binary []byte) {
	t.Helper()
	tmp := t.TempDir()
	fileName := branding.PlatformBinaryName
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
	runtime.GC()
}

// Default options, no options passed in
func TestNightlyDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping nightly download test in short mode")
	}

	dl, err := tofudl.New()
	if err != nil {
		t.Fatal(err)
	}

	binary, err := dl.DownloadNightly(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(binary) == 0 {
		t.Fatal("Downloaded binary is empty")
	}

	logTofuVersion(t, binary)
}

// TestNightlyDownloadWithOptions tests downloading with specific platform/architecture
// We are not testing specific build ID, since those are cleaned up ofter
func TestNightlyDownloadWithHostOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping nightly download test in short mode")
	}

	dl, err := tofudl.New()
	if err != nil {
		t.Fatal(err)
	}

	// Test downloading for linux/amd64 specifically
	binary, err := dl.DownloadNightly(
		context.Background(),
		tofudl.DownloadOptPlatform(tofudl.PlatformLinux),
		tofudl.DownloadOptArchitecture(tofudl.ArchitectureAMD64),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(binary) == 0 {
		t.Fatal("Downloaded binary is empty")
	}

	logTofuVersion(t, binary)
}
