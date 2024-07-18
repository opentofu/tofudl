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

type cacheStorage struct {
	config CacheConfig
}

func (c cacheStorage) readAPIFile() (io.ReadCloser, bool, error) {
	apiFile := c.getAPIFileName()
	return c.readCacheFile(apiFile, c.config.APICacheTimeout)
}

func (c cacheStorage) readCacheFile(cacheFile string, timeout time.Duration) (io.ReadCloser, bool, error) {
	stat, err := os.Stat(cacheFile)
	if err != nil {
		return nil, false, err
	}
	stale := false
	if timeout > 0 && stat.ModTime().Add(timeout).Before(time.Now()) {
		stale = true
	}
	fh, err := os.OpenFile(cacheFile, os.O_RDONLY, 0644)
	if err != nil {
		return nil, stale, err
	}
	return fh, stale, nil
}

func (c cacheStorage) getAPIFileName() string {
	apiFile := path.Join(c.config.CacheDirectory, "api.json")
	return apiFile
}

func (c cacheStorage) storeAPIFile(data []byte) error {
	apiFile := c.getAPIFileName()
	return os.WriteFile(apiFile, data, 0644) //nolint:gosec //This is not sensitive
}

func (c cacheStorage) readArtifact(version Version, artifact string) (io.ReadCloser, bool, error) {
	cacheFile := c.getArtifactCacheFileName(c.getArtifactCacheDirectory(version, artifact), artifact)
	return c.readCacheFile(cacheFile, c.config.ArtifactCacheTimeout)
}

func (c cacheStorage) storeArtifact(version Version, artifact string, contents []byte) error {
	cacheDirectory := c.getArtifactCacheDirectory(version, artifact)
	if err := os.MkdirAll(cacheDirectory, 0644); err != nil { //nolint:gosec //This is not sensitive
		return fmt.Errorf("failed to create cache directory %s (%w)", cacheDirectory, err)
	}
	cacheFile := c.getArtifactCacheFileName(cacheDirectory, artifact)
	if err := os.WriteFile(cacheFile, contents, 0644); err != nil {
		return fmt.Errorf("failed to write cache file %s (%w)", cacheFile, err)
	}
	return nil
}

func (c cacheStorage) getArtifactCacheFileName(cacheDirectory string, artifact string) string {
	return path.Join(cacheDirectory, artifact)
}

func (c cacheStorage) getArtifactCacheDirectory(version Version, artifact string) string {
	cacheDirectory := path.Join(c.config.CacheDirectory, "v"+string(version), artifact)
	return cacheDirectory
}
