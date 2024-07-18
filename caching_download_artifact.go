// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"fmt"
	"io"
)

func (c *cachingDownloader) DownloadArtifact(ctx context.Context, version VersionWithArtifacts, artifactName string) ([]byte, error) {
	if c.config.isDisabled() || c.config.ArtifactCacheTimeout == 0 {
		return c.backingDownloader.DownloadArtifact(ctx, version, artifactName)
	}

	storage := &cacheStorage{
		c.config,
	}

	cachedArtifact, err := c.tryReadArtifactCache(storage, version.ID, artifactName, false)
	if err == nil {
		return cachedArtifact, nil
	}

	artifact, onlineErr := c.backingDownloader.DownloadArtifact(ctx, version, artifactName)
	if onlineErr == nil {
		_ = storage.storeArtifact(version.ID, artifactName, artifact)
		return artifact, nil
	}

	cachedArtifact, err = c.tryReadArtifactCache(storage, version.ID, artifactName, true)
	if err == nil {
		return cachedArtifact, nil
	}
	return nil, onlineErr
}

func (c *cachingDownloader) tryReadArtifactCache(storage *cacheStorage, version Version, artifact string, allowStale bool) ([]byte, error) {
	cacheReader, stale, err := storage.readArtifact(version, artifact)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cacheReader.Close()
	}()
	if stale && !allowStale {
		return nil, fmt.Errorf("resource stale")
	}
	return io.ReadAll(cacheReader)
}
