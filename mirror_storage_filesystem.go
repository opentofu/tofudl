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

// NewFilesystemStorage returns a mirror storage that relies on files and modification timestamps. This storage
// can also be used to mirror the artifacts for air-gapped usage. The filesystem layout is as follows:
//
// - api.json
// - v1.2.3/artifact.name
//
// Note: when used as a pull-through cache, the underlying filesystem must support modification timestamps or the
// cache timeout must be set to -1 to prevent the mirror from re-fetching every time.
func NewFilesystemStorage(directory string) (MirrorStorage, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory %s (%w)", directory, err)
	}

	return &filesystemStorage{
		directory,
	}, nil
}

type filesystemStorage struct {
	directory string
}

func (c filesystemStorage) ReadAPIFile() (io.ReadCloser, time.Time, error) {
	apiFile := c.getAPIFileName()
	return c.readCacheFile(apiFile)
}

func (c filesystemStorage) readCacheFile(cacheFile string) (io.ReadCloser, time.Time, error) {
	if c.directory == "" {
		return nil, time.Time{}, &CacheMissError{cacheFile, nil}
	}

	stat, err := os.Stat(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, time.Time{}, &CacheMissError{cacheFile, err}
		}
		return nil, time.Time{}, err
	}
	fh, err := os.OpenFile(cacheFile, os.O_RDONLY, 0644)
	if err != nil {
		return nil, stat.ModTime(), &CacheMissError{cacheFile, err}
	}
	return fh, stat.ModTime(), nil
}

func (c filesystemStorage) getAPIFileName() string {
	apiFile := path.Join(c.directory, "api.json")
	return apiFile
}

func (c filesystemStorage) StoreAPIFile(data []byte) error {
	apiFile := c.getAPIFileName()
	return os.WriteFile(apiFile, data, 0644) //nolint:gosec //This is not sensitive
}

func (c filesystemStorage) ReadArtifact(version Version, artifact string) (io.ReadCloser, time.Time, error) {
	cacheFile := c.getArtifactCacheFileName(c.getArtifactCacheDirectory(version), artifact)
	return c.readCacheFile(cacheFile)
}

func (c filesystemStorage) StoreArtifact(version Version, artifact string, contents []byte) error {
	cacheDirectory := c.getArtifactCacheDirectory(version)
	if err := os.MkdirAll(cacheDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory %s (%w)", cacheDirectory, err)
	}
	cacheFile := c.getArtifactCacheFileName(cacheDirectory, artifact)
	if err := os.WriteFile(cacheFile, contents, 0644); err != nil { //nolint:gosec // This is not sensitive
		return fmt.Errorf("failed to write cache file %s (%w)", cacheFile, err)
	}
	return nil
}

func (c filesystemStorage) getArtifactCacheFileName(cacheDirectory string, artifact string) string {
	return path.Join(cacheDirectory, artifact)
}

func (c filesystemStorage) getArtifactCacheDirectory(version Version) string {
	cacheDirectory := path.Join(c.directory, "v"+string(version))
	return cacheDirectory
}
