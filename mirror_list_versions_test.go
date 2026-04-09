// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// testMirrorStorageStub implements MirrorStorage for API cache tests only.
type testMirrorStorageStub struct {
	apiJSON string
	apiTime time.Time
}

func (s *testMirrorStorageStub) ReadAPIFile() (io.ReadCloser, time.Time, error) {
	return io.NopCloser(strings.NewReader(s.apiJSON)), s.apiTime, nil
}

func (s *testMirrorStorageStub) StoreAPIFile([]byte) error { return nil }

func (s *testMirrorStorageStub) ReadArtifact(Version, string) (io.ReadCloser, time.Time, error) {
	return nil, time.Time{}, &CacheMissError{"", nil}
}

func (s *testMirrorStorageStub) StoreArtifact(Version, string, []byte) error { return nil }

func TestTryReadVersionCacheUsesAPICacheTimeout(t *testing.T) {
	t.Parallel()

	const minimalAPI = `{"versions":[]}`

	t.Run("fresh_by_api_timeout", func(t *testing.T) {
		t.Parallel()
		m := &mirror{
			config: MirrorConfig{
				APICacheTimeout:      time.Hour,
				ArtifactCacheTimeout:   time.Nanosecond, // must not affect API staleness
			},
			storage: &testMirrorStorageStub{
				apiJSON: minimalAPI,
				apiTime: time.Now().Add(-30 * time.Minute),
			},
		}
		_, err := m.tryReadVersionCache(m.storage, nil, false)
		if err != nil {
			t.Fatalf("tryReadVersionCache: %v", err)
		}
	})

	t.Run("stale_by_api_timeout", func(t *testing.T) {
		t.Parallel()
		m := &mirror{
			config: MirrorConfig{
				APICacheTimeout:      time.Hour,
				ArtifactCacheTimeout: -1,
			},
			storage: &testMirrorStorageStub{
				apiJSON: minimalAPI,
				apiTime: time.Now().Add(-2 * time.Hour),
			},
		}
		_, err := m.tryReadVersionCache(m.storage, nil, false)
		var stale *CachedAPIResponseStaleError
		if err == nil {
			t.Fatal("expected CachedAPIResponseStaleError")
		}
		if !errors.As(err, &stale) {
			t.Fatalf("got %T, want *CachedAPIResponseStaleError", err)
		}
	})

	t.Run("indefinite_api_cache_negative_timeout", func(t *testing.T) {
		t.Parallel()
		m := &mirror{
			config: MirrorConfig{
				APICacheTimeout:      -1,
				ArtifactCacheTimeout: time.Nanosecond,
			},
			storage: &testMirrorStorageStub{
				apiJSON: minimalAPI,
				apiTime: time.Now().Add(-365 * 24 * time.Hour),
			},
		}
		_, err := m.tryReadVersionCache(m.storage, nil, false)
		if err != nil {
			t.Fatalf("with APICacheTimeout -1, old file should not be stale: %v", err)
		}
	})

	t.Run("allowStale_skips_api_ttl", func(t *testing.T) {
		t.Parallel()
		m := &mirror{
			config: MirrorConfig{
				APICacheTimeout: time.Second,
			},
			storage: &testMirrorStorageStub{
				apiJSON: minimalAPI,
				apiTime: time.Now().Add(-time.Hour),
			},
		}
		_, err := m.tryReadVersionCache(m.storage, nil, true)
		if err != nil {
			t.Fatalf("allowStale: %v", err)
		}
	})
}
