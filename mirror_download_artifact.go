// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"io"
	"time"
)

func (m *mirror) DownloadArtifact(ctx context.Context, version VersionWithArtifacts, artifactName string) ([]byte, error) {
	if m.pullThroughDownloader == nil {
		return m.tryReadArtifactCache(m.storage, version.ID, artifactName, true)
	}

	if m.storage == nil || m.config.ArtifactCacheTimeout == 0 {
		return m.pullThroughDownloader.DownloadArtifact(ctx, version, artifactName)
	}

	cachedArtifact, err := m.tryReadArtifactCache(m.storage, version.ID, artifactName, false)
	if err == nil {
		return cachedArtifact, nil
	}

	artifact, onlineErr := m.pullThroughDownloader.DownloadArtifact(ctx, version, artifactName)
	if onlineErr == nil {
		_ = m.storage.StoreArtifact(version.ID, artifactName, artifact)
		return artifact, nil
	}

	cachedArtifact, err = m.tryReadArtifactCache(m.storage, version.ID, artifactName, true)
	if err == nil {
		return cachedArtifact, nil
	}
	return nil, onlineErr
}

func (m *mirror) tryReadArtifactCache(storage MirrorStorage, version Version, artifact string, allowStale bool) ([]byte, error) {
	cacheReader, storeTime, err := storage.ReadArtifact(version, artifact)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cacheReader.Close()
	}()
	if !allowStale && m.config.ArtifactCacheTimeout > 0 && storeTime.Add(m.config.ArtifactCacheTimeout).Before(time.Now()) {
		return nil, &CachedArtifactStaleError{Version: version, Artifact: artifact}
	}
	return io.ReadAll(cacheReader)
}
