// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl_test

import (
	"context"
	"testing"
	"time"

	"github.com/opentofu/tofudl"
	"github.com/opentofu/tofudl/mockmirror"
)

func TestMirroringE2E(t *testing.T) {
	mirror := mockmirror.New(t)

	dl1, err := tofudl.New(
		tofudl.ConfigGPGKey(mirror.GPGKey()),
		tofudl.ConfigAPIURL(mirror.APIURL()),
		tofudl.ConfigDownloadMirrorURLTemplate(mirror.DownloadMirrorURLTemplate()),
	)
	if err != nil {
		t.Fatal(err)
	}

	cacheDir := t.TempDir()
	storage, err := tofudl.NewFilesystemCachingStorage(cacheDir)
	if err != nil {
		t.Fatal(err)
	}

	cache1, err := tofudl.NewCacheLayer(
		tofudl.CacheConfig{
			AllowStale:           false,
			APICacheTimeout:      time.Minute * 30,
			ArtifactCacheTimeout: time.Minute * 30,
		},
		storage,
		dl1,
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Logf("Pre-warming caches...")
	if err := cache1.PreWarm(ctx, 1, func(pct int8) {
		t.Logf("Pre-warming caches %d%% complete.", pct)
	}); err != nil {
		t.Fatal(err)
	}
	t.Logf("Pre-warming caches complete.")

	// Configure an invalid API and mirror URL and see if the cache works.
	dl2, err := tofudl.New(
		tofudl.ConfigAPIURL("http://127.0.0.1:9999/"),
		tofudl.ConfigDownloadMirrorURLTemplate("http://127.0.0.1:9999/{{ .Version }}/{{ .Artifact }}"),
		tofudl.ConfigGPGKey(mirror.GPGKey()),
	)
	if err != nil {
		t.Fatal(err)
	}

	cache2, err := tofudl.NewCacheLayer(
		tofudl.CacheConfig{
			AllowStale:           false,
			APICacheTimeout:      -1,
			ArtifactCacheTimeout: -1,
		},
		storage,
		dl2,
	)
	if err != nil {
		t.Fatal(err)
	}

	versions, err := cache2.ListVersions(ctx)
	if err != nil {
		t.Fatal(err)
	}

	lastVersion := versions[0]

	binary, err := cache2.DownloadVersion(ctx, lastVersion, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(binary) == 0 {
		t.Fatal("Empty artifact!")
	}
}
