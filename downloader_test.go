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
	"github.com/opentofu/tofudl/mockmirror"
)

func TestE2E(t *testing.T) {
	mirror := mockmirror.New(t)

	dl, err := tofudl.New(
		tofudl.ConfigGPGKey(mirror.GPGKey()),
		tofudl.ConfigAPIURL(mirror.APIURL()),
		tofudl.ConfigDownloadMirrorURLTemplate(mirror.DownloadMirrorURLTemplate()),
	)
	if err != nil {
		t.Fatal(err)
	}

	binary, err := dl.Download(context.Background())
	if err != nil {
		t.Fatal(err)
	}

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
}
