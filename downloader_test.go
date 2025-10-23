// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl_test

import (
	"context"
	"testing"

	"github.com/opentofu/tofudl"
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

	logTofuVersion(t, binary)
}
