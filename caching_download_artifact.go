// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"io"
	"time"
)

func (c *cachingDownloader) DownloadArtifact(ctx context.Context, version VersionWithArtifacts, artifactName string) ([]byte, error) {
	if c.storage == nil || c.config.ArtifactCacheTimeout == 0 {
		return c.backingDownloader.DownloadArtifact(ctx, version, artifactName)
	}

	cachedArtifact, err := c.tryReadArtifactCache(c.storage, version.ID, artifactName, false)
	if err == nil {
		return cachedArtifact, nil
	}

	artifact, onlineErr := c.backingDownloader.DownloadArtifact(ctx, version, artifactName)
	if onlineErr == nil {
		_ = c.storage.StoreArtifact(version.ID, artifactName, artifact)
		return artifact, nil
	}

	cachedArtifact, err = c.tryReadArtifactCache(c.storage, version.ID, artifactName, true)
	if err == nil {
		return cachedArtifact, nil
	}
	return nil, onlineErr
}

func (c *cachingDownloader) tryReadArtifactCache(storage CachingStorage, version Version, artifact string, allowStale bool) ([]byte, error) {
	cacheReader, storeTime, err := storage.ReadArtifact(version, artifact)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cacheReader.Close()
	}()
	if !allowStale && c.config.ArtifactCacheTimeout > 0 && storeTime.Add(c.config.ArtifactCacheTimeout).Before(time.Now()) {
		return nil, &CachedArtifactStaleError{Version: version, Artifact: artifact}
	}
	return io.ReadAll(cacheReader)
}
