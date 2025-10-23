// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/opentofu/tofudl"
	"github.com/opentofu/tofudl/branding"
)

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

	// Verify the binary is executable by running tofu version
	tmp := t.TempDir()
	fileName := branding.PlatformBinaryName
	fullPath := path.Join(tmp, fileName)
	if err := os.WriteFile(fullPath, binary, 0755); err != nil { //nolint:gosec //We want the binary to be executable for the test purposes
		t.Fatal(err)
	}
	stdout := bytes.Buffer{}

	cmd := exec.Command(fullPath, "version")
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run tofu version: %v\nOutput: %s", err, stdout.String())
	}

	t.Logf("Nightly build version: %s", stdout.String())
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
}
