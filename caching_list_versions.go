// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

func (c *cachingDownloader) ListVersions(ctx context.Context, opts ...ListVersionOpt) ([]VersionWithArtifacts, error) {
	if c.storage == nil || c.config.APICacheTimeout == 0 {
		return c.backingDownloader.ListVersions(ctx, opts...)
	}

	// Fetch non-stale cached version:
	cachedVersions, err := c.tryReadVersionCache(c.storage, opts, false)
	if err == nil {
		return cachedVersions, nil
	}

	// Fetch online version:
	versions, onlineErr := c.backingDownloader.ListVersions(ctx, opts...)
	if onlineErr == nil {
		marshalledVersions, err := json.Marshal(APIResponse{versions})
		if err == nil {
			_ = c.storage.StoreAPIFile(marshalledVersions)
		}
		return versions, nil
	}

	// Fetch stale cached version:
	cachedVersions, err = c.tryReadVersionCache(c.storage, opts, true)
	if err == nil {
		return cachedVersions, nil
	}
	return nil, onlineErr
}

func (c *cachingDownloader) tryReadVersionCache(storage CachingStorage, opts []ListVersionOpt, allowStale bool) ([]VersionWithArtifacts, error) {
	cacheReader, storeTime, err := storage.ReadAPIFile()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cacheReader.Close()
	}()
	if !allowStale && c.config.ArtifactCacheTimeout > 0 && storeTime.Add(c.config.ArtifactCacheTimeout).Before(time.Now()) {
		return nil, &CachedAPIResponseStaleError{}
	}
	return fetchVersions(opts, func() (io.ReadCloser, error) {
		return cacheReader, nil
	})
}
