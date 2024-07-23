// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

// NewFilesystemCachingStorage returns a cache storage that relies on files and modification timestamps. This storage
// can also be used to mirror the artifacts for airgapped usage. The filesystem layout is as follows:
//
// - api.json
// - v1.2.3/artifact.name
//
// Note: the filesystem must support modification timestamps for this storage to work. Alterantively, you can set
// the cache timeout to infinite (-1), which will always fetch the stored files.
func NewFilesystemCachingStorage(directory string) (CachingStorage, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory %s (%w)", directory, err)
	}

	return &cacheStorage{
		directory,
	}, nil
}

type cacheStorage struct {
	directory string
}

func (c cacheStorage) ReadAPIFile() (io.ReadCloser, time.Time, error) {
	apiFile := c.getAPIFileName()
	return c.readCacheFile(apiFile)
}

func (c cacheStorage) readCacheFile(cacheFile string) (io.ReadCloser, time.Time, error) {
	if c.directory == "" {
		return nil, time.Time{}, &CacheMissError{cacheFile, nil}
	}

	stat, err := os.Stat(cacheFile)
	if err != nil {
		return nil, time.Time{}, err
	}
	fh, err := os.OpenFile(cacheFile, os.O_RDONLY, 0644)
	if err != nil {
		return nil, stat.ModTime(), &CacheMissError{cacheFile, err}
	}
	return fh, stat.ModTime(), nil
}

func (c cacheStorage) getAPIFileName() string {
	apiFile := path.Join(c.directory, "api.json")
	return apiFile
}

func (c cacheStorage) StoreAPIFile(data []byte) error {
	apiFile := c.getAPIFileName()
	return os.WriteFile(apiFile, data, 0644) //nolint:gosec //This is not sensitive
}

func (c cacheStorage) ReadArtifact(version Version, artifact string) (io.ReadCloser, time.Time, error) {
	cacheFile := c.getArtifactCacheFileName(c.getArtifactCacheDirectory(version, artifact), artifact)
	return c.readCacheFile(cacheFile)
}

func (c cacheStorage) StoreArtifact(version Version, artifact string, contents []byte) error {
	cacheDirectory := c.getArtifactCacheDirectory(version, artifact)
	if err := os.MkdirAll(cacheDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory %s (%w)", cacheDirectory, err)
	}
	cacheFile := c.getArtifactCacheFileName(cacheDirectory, artifact)
	if err := os.WriteFile(cacheFile, contents, 0644); err != nil { //nolint:gosec // This is not sensitive
		return fmt.Errorf("failed to write cache file %s (%w)", cacheFile, err)
	}
	return nil
}

func (c cacheStorage) getArtifactCacheFileName(cacheDirectory string, artifact string) string {
	return path.Join(cacheDirectory, artifact)
}

func (c cacheStorage) getArtifactCacheDirectory(version Version, artifact string) string {
	cacheDirectory := path.Join(c.directory, "v"+string(version), artifact)
	return cacheDirectory
}
