// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"time"
)

// NewCacheLayer creates a new caching downloader with the corresponding backing downloader.
func NewCacheLayer(config CacheConfig, cachingStorage CachingStorage, backingDownloader Downloader) (CachingDownloader, error) {
	return &cachingDownloader{
		cachingStorage,
		backingDownloader,
		config,
	}, nil
}

// CachingDownloader is a downloader that caches artifacts. It also supports pre-warming caches by calling the
// PreWarm function. The cache directory can simultaneously also be used as a mirror when served using an HTTP server.
type CachingDownloader interface {
	Downloader

	// PreWarm downloads the last specified number of versions into the cache directory. If versions is negative,
	// all versions are downloaded. Note: the versions include alpha, beta and release candidate versions. Make sure
	// you pre-warm with enough versions for your use case.
	PreWarm(ctx context.Context, versionCount int, progress func(pct int8)) error
}

// CacheConfig is the configuration structure for the caching downloader.
type CacheConfig struct {
	// AllowStale enables using stale cached resources if the download fails.
	AllowStale bool `json:"allow_stale"`
	// APICacheTimeout is the time the cached API JSON should be considered valid. A duration of 0 means the API
	// responses should not be cached. A duration of -1 means the API responses should be cached indefinitely.
	APICacheTimeout time.Duration `json:"api_cache_timeout"`
	// ArtifactCacheTimeout is the time the cached artifacts should be considered valid. A duration of 0 means that
	// artifacts should not be cached. A duration of -1 means that artifacts should be cached indefinitely.
	ArtifactCacheTimeout time.Duration `json:"artifact_cache_timeout"`
}

type cachingDownloader struct {
	storage           CachingStorage
	backingDownloader Downloader
	config            CacheConfig
}
