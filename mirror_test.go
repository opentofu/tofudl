// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/opentofu/tofudl"
	"github.com/opentofu/tofudl/branding"
	"github.com/opentofu/tofudl/internal/helloworld"
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
	storage, err := tofudl.NewFilesystemStorage(cacheDir)
	if err != nil {
		t.Fatal(err)
	}

	cache1, err := tofudl.NewMirror(
		tofudl.MirrorConfig{
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

	cache2, err := tofudl.NewMirror(
		tofudl.MirrorConfig{
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

func TestMirrorStandalone(t *testing.T) {
	binaryContents := helloworld.Build(t)

	ctx := context.Background()
	key, err := crypto.GenerateKey(branding.ProductName+" Test", "noreply@example.org", "rsa", 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubKey, err := key.GetArmoredPublicKey()
	if err != nil {
		t.Fatal(err)
	}
	builder, err := tofudl.NewReleaseBuilder(key)
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.PackageBinary(tofudl.PlatformAuto, tofudl.ArchitectureAuto, binaryContents, nil); err != nil {
		t.Fatalf("failed to package binary (%v)", err)
	}

	mirrorStorage, err := tofudl.NewFilesystemStorage(t.TempDir())
	if err != nil {
		t.Fatalf("failed to set up TofuDL mirror")
	}
	downloader, err := tofudl.NewMirror(
		tofudl.MirrorConfig{
			GPGKey: pubKey,
		},
		mirrorStorage,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.Build(ctx, "1.9.0", downloader); err != nil {
		t.Fatal(err)
	}
	_, err = downloader.Download(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure all file handles are closed.
	if runtime.GOOS == "windows" {
		runtime.GC()
	}
}
