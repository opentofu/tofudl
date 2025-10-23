// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl_test

import (
	"context"
	"testing"

	"github.com/opentofu/tofudl"
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
