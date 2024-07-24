// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"io"
	"time"
)

// MirrorStorage is responsible for handling the low-level storage of caches.
type MirrorStorage interface {
	// ReadAPIFile reads the API file cache and returns a reader for it. It also returns the time when the cached
	// response was written. It will return a CacheMissError if the API response is not cached.
	ReadAPIFile() (io.ReadCloser, time.Time, error)
	// StoreAPIFile stores the API file in the cache.
	StoreAPIFile(apiFile []byte) error

	// ReadArtifact reads a binary artifact from the cache for a specific version and returns a reader to it.
	// It also returns the time the artifact was stored as the second parameter. It will return a CacheMissError if
	// there is no such artifact in the cache.
	ReadArtifact(version Version, artifactName string) (io.ReadCloser, time.Time, error)
	// StoreArtifact stores a binary artifact in the cache for a specific version.
	StoreArtifact(version Version, artifactName string, contents []byte) error
}
