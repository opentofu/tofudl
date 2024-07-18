package tofudl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

func (c *cachingDownloader) ListVersions(ctx context.Context, opts ...ListVersionOpt) ([]VersionWithArtifacts, error) {
	if c.config.isDisabled() || c.config.APICacheTimeout == 0 {
		return c.backingDownloader.ListVersions(ctx, opts...)
	}

	storage := &cacheStorage{
		c.config,
	}

	// Fetch non-stale cached version:
	cachedVersions, err := c.tryReadVersionCache(storage, opts, false)
	if err == nil {
		return cachedVersions, nil
	}

	// Fetch online version:
	versions, onlineErr := c.backingDownloader.ListVersions(ctx, opts...)
	if onlineErr == nil {
		marshalledVersions, err := json.Marshal(APIResponse{versions})
		if err == nil {
			_ = storage.storeAPIFile(marshalledVersions)
		}
		return versions, nil
	}

	// Fetch stale cached version:
	cachedVersions, err = c.tryReadVersionCache(storage, opts, true)
	if err == nil {
		return cachedVersions, nil
	}
	return nil, onlineErr
}

func (c *cachingDownloader) tryReadVersionCache(storage *cacheStorage, opts []ListVersionOpt, allowStale bool) ([]VersionWithArtifacts, error) {
	cacheReader, stale, err := storage.readAPIFile()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cacheReader.Close()
	}()
	if stale && !allowStale {
		return nil, fmt.Errorf("resource stale")
	}
	return fetchVersions(opts, func() (io.ReadCloser, error) {
		return cacheReader, nil
	})
}
