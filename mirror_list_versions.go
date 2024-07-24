// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

func (m *mirror) ListVersions(ctx context.Context, opts ...ListVersionOpt) ([]VersionWithArtifacts, error) {
	if m.pullThroughDownloader == nil {
		return m.tryReadVersionCache(m.storage, opts, true)
	}
	if m.storage == nil || m.config.APICacheTimeout == 0 {
		return m.pullThroughDownloader.ListVersions(ctx, opts...)
	}

	// Fetch non-stale cached version:
	cachedVersions, err := m.tryReadVersionCache(m.storage, opts, false)
	if err == nil {
		return cachedVersions, nil
	}

	// Fetch online version:
	versions, onlineErr := m.pullThroughDownloader.ListVersions(ctx, opts...)
	if onlineErr == nil {
		marshalledVersions, err := json.Marshal(APIResponse{versions})
		if err == nil {
			_ = m.storage.StoreAPIFile(marshalledVersions)
		}
		return versions, nil
	}

	// Fetch stale cached version:
	cachedVersions, err = m.tryReadVersionCache(m.storage, opts, true)
	if err == nil {
		return cachedVersions, nil
	}
	return nil, onlineErr
}

func (m *mirror) tryReadVersionCache(storage MirrorStorage, opts []ListVersionOpt, allowStale bool) ([]VersionWithArtifacts, error) {
	cacheReader, storeTime, err := storage.ReadAPIFile()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cacheReader.Close()
	}()
	if !allowStale && m.config.ArtifactCacheTimeout > 0 && storeTime.Add(m.config.ArtifactCacheTimeout).Before(time.Now()) {
		return nil, &CachedAPIResponseStaleError{}
	}
	return fetchVersions(opts, func() (io.ReadCloser, error) {
		return cacheReader, nil
	})
}
