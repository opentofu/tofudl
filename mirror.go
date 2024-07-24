// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// NewMirror creates a new mirror, optionally acting as a pull-through cache when passing a pullThroughDownloader.
func NewMirror(config MirrorConfig, storage MirrorStorage, pullThroughDownloader Downloader) (Mirror, error) {
	if storage == nil && pullThroughDownloader == nil {
		return nil, fmt.Errorf(
			"no storage and no pull-through downloader passed to NewMirror, cannot create a working mirror",
		)
	}
	return &mirror{
		storage,
		pullThroughDownloader,
		config,
	}, nil
}

// Mirror is a downloader that caches artifacts. It also supports pre-warming caches by calling the
// PreWarm function. You can use this as a handler for an HTTP server in order to act as a mirror to a regular
// Downloader.
type Mirror interface {
	Downloader
	http.Handler

	// PreWarm downloads the last specified number of versions into the storage directory from the pull-through
	// downloader if present. If versions is negative, all versions are downloaded. Note: the versions include alpha,
	// beta and release candidate versions. Make sure you pre-warm with enough versions for your use case.
	//
	// If no pull-through downloader is configured, this function does not do anything.
	PreWarm(ctx context.Context, versionCount int, progress func(pct int8)) error

	// CreateVersion creates a new version in the cache, adding it to the version index. Note that this is not supported
	// when working in pull-through cache mode.
	CreateVersion(ctx context.Context, version Version) error

	// CreateVersionAsset creates a new asset for a version, storing it in the storage and adding it to the version
	// list. Note that this is not supported when working in pull-through cache mode.
	CreateVersionAsset(ctx context.Context, version Version, assetName string, assetData []byte) error
}

// MirrorConfig is the configuration structure for the caching downloader.
type MirrorConfig struct {
	// AllowStale enables using stale cached resources if the download fails.
	AllowStale bool `json:"allow_stale"`
	// APICacheTimeout is the time the cached API JSON should be considered valid. A duration of 0 means the API
	// responses should not be cached. A duration of -1 means the API responses should be cached indefinitely.
	APICacheTimeout time.Duration `json:"api_cache_timeout"`
	// ArtifactCacheTimeout is the time the cached artifacts should be considered valid. A duration of 0 means that
	// artifacts should not be cached. A duration of -1 means that artifacts should be cached indefinitely.
	ArtifactCacheTimeout time.Duration `json:"artifact_cache_timeout"`
}

type mirror struct {
	storage               MirrorStorage
	pullThroughDownloader Downloader
	config                MirrorConfig
}
