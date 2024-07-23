// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package mockmirror

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/fs"
	"testing"
	"time"

	"github.com/opentofu/tofudl/branding"
)

func buildTarFile(t *testing.T, binary []byte) []byte {
	buf := &bytes.Buffer{}
	gzipWriter := gzip.NewWriter(buf)

	tarWriter := tar.NewWriter(gzipWriter)

	header, err := tar.FileInfoHeader(&fileInfo{
		name:    branding.PlatformBinaryName,
		size:    int64(len(binary)),
		mode:    0755,
		modTime: time.Now(),
		isDir:   false,
	}, branding.PlatformBinaryName)
	if err != nil {
		t.Fatalf("Failed to construct tar file header (%v)", err)
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write file header to tar file (%v)", err)
	}
	if _, err := tarWriter.Write(binary); err != nil {
		t.Fatalf("Failed to write binary to tar file (%v)", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("Failed to close tar writer (%v)", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer (%v)", err)
	}
	return buf.Bytes()
}

type fileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (f fileInfo) Name() string {
	return f.name
}

func (f fileInfo) Size() int64 {
	return f.size
}

func (f fileInfo) Mode() fs.FileMode {
	return f.mode
}

func (f fileInfo) ModTime() time.Time {
	return f.modTime
}

func (f fileInfo) IsDir() bool {
	return f.isDir
}

func (f fileInfo) Sys() any {
	return nil
}
