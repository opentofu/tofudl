// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"fmt"
	"os"
	"time"
)

// NewCacheLayer creates a new caching downloader with the corresponding backing downloader.
func NewCacheLayer(config CacheConfig, backingDownloader Downloader) (CachingDownloader, error) {
	if err := os.MkdirAll(config.CacheDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory %s (%w)", config.CacheDirectory, err)
	}

	return &cachingDownloader{
		backingDownloader,
		config,
	}, nil
}

// CachingDownloader is a downloader that caches artifacts. It also supports pre-warming caches by calling the
// PreWarm function. The cache directory can simultaneously also be used as a mirror when served using an HTTP server.
//
// The cache directory layout is as follows:
//
// - api.json
// - v1.2.3/artifact.name
type CachingDownloader interface {
	Downloader

	// PreWarm downloads the last specified number of versions into the cache directory. If versions is negative,
	// all versions are downloaded. Note: the versions include alpha, beta and release candidate versions. Make sure
	// you pre-warm with enough versions for your use case.
	PreWarm(ctx context.Context, versionCount int, progress func(pct int8)) error
}

// CacheConfig is the configuration structure for the caching downloader.
type CacheConfig struct {
	// CacheDirectory is the directory the cache is stored in. This directory must exist or must be creatable.
	CacheDirectory string `json:"cache_directory"`
	// AllowStale enables using stale cached resources if the download fails.
	AllowStale bool `json:"allow_stale"`
	// APICacheTimeout is the time the cached API JSON should be considered valid. A duration of 0 means the API
	// responses should not be cached. A duration of -1 means the API responses should be cached indefinitely.
	APICacheTimeout time.Duration `json:"api_cache_timeout"`
	// ArtifactCacheTimeout is the time the cached artifacts should be considered valid. A duration of 0 means that
	// artifacts should not be cached. A duration of -1 means that artifacts should be cached indefinitely.
	ArtifactCacheTimeout time.Duration `json:"artifact_cache_timeout"`
}

func (c CacheConfig) isDisabled() bool {
	return c.CacheDirectory == ""
}

type cachingDownloader struct {
	backingDownloader Downloader

	config CacheConfig
}

func (c *cachingDownloader) VerifyArtifact(artifactName string, artifactContents []byte, sumsFileContents []byte, signatureFileContent []byte) error {
	return c.backingDownloader.VerifyArtifact(artifactName, artifactContents, sumsFileContents, signatureFileContent)
}

func (c *cachingDownloader) DownloadVersion(ctx context.Context, version VersionWithArtifacts, platform Platform, architecture Architecture) ([]byte, error) {
	return downloadVersion(ctx, version, platform, architecture, c.DownloadArtifact, c.backingDownloader.VerifyArtifact)
}

func (c *cachingDownloader) Download(ctx context.Context, opts ...DownloadOpt) ([]byte, error) {
	return download(ctx, opts, c.ListVersions, c.DownloadVersion)
}

func (c *cachingDownloader) PreWarm(ctx context.Context, versionCount int, progress func(pct int8)) error {
	if c.config.isDisabled() {
		return nil
	}

	versions, err := c.ListVersions(ctx)
	if err != nil {
		return err
	}
	if versionCount > 0 {
		versions = versions[:versionCount]
	}
	totalArtifacts := 0
	for _, version := range versions {
		totalArtifacts += len(version.Files)
	}
	downloadedArtifacts := 0
	for _, version := range versions {
		for _, artifact := range version.Files {
			_, err = c.DownloadArtifact(ctx, version, artifact)
			if err != nil {
				return fmt.Errorf("failed to download artifact %s for version %s (%w)", artifact, version.ID, err)
			}
			downloadedArtifacts += 1
			if progress != nil {
				progress(int8(100 * float64(downloadedArtifacts) / float64(totalArtifacts)))
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
	}
	return nil
}