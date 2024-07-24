// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/opentofu/tofudl/branding"
)

// ReleaseBuilder is a tool to build a release and add it to a mirror. Note that this does not (yet) produce a full
// release suitable for end users as this does not support signing with cosign and does not produce other artifacts.
type ReleaseBuilder interface {
	// PackageBinary creates a .tar.gz file for the specific platform and architecture based on the binary contents.
	// You may pass extra files to package, such as LICENSE, etc. in extraFiles.
	PackageBinary(platform Platform, architecture Architecture, contents []byte, extraFiles map[string][]byte) error

	// AddArtifact adds an artifact to the release, adds it to the checksum file and signs the checksum file.
	AddArtifact(artifactName string, data []byte) error

	// Build builds the release and adds it to the specified mirror. Note that the ReleaseBuilder should not be
	// reused after calling Build.
	Build(ctx context.Context, version Version, mirror Mirror) error
}

// NewReleaseBuilder creates a new ReleaseBuilder with the given gpgKey to sign the release.
func NewReleaseBuilder(gpgKey *crypto.Key) (ReleaseBuilder, error) {
	return &releaseBuilder{
		gpgKey:    gpgKey,
		binaries:  nil,
		artifacts: map[string][]byte{},
	}, nil
}

type releaseBinary struct {
	platform     Platform
	architecture Architecture
	contents     []byte
	extraFiles   map[string][]byte
}

type releaseBuilder struct {
	gpgKey    *crypto.Key
	binaries  []releaseBinary
	artifacts map[string][]byte
}

func (r *releaseBuilder) PackageBinary(platform Platform, architecture Architecture, contents []byte, extraFiles map[string][]byte) error {
	var err error
	platform, err = platform.ResolveAuto()
	if err != nil {
		return err
	}
	architecture, err = architecture.ResolveAuto()
	if err != nil {
		return err
	}
	r.binaries = append(r.binaries, releaseBinary{
		platform:     platform,
		architecture: architecture,
		contents:     contents,
		extraFiles:   extraFiles,
	})
	return nil
}

func (r *releaseBuilder) AddArtifact(artifactName string, data []byte) error {
	r.artifacts[artifactName] = data
	return nil
}

func (r *releaseBuilder) Build(ctx context.Context, version Version, mirror Mirror) error {
	if err := version.Validate(); err != nil {
		return err
	}
	for _, binary := range r.binaries {
		tarFile, err := buildTarFile(binary.contents, binary.extraFiles)
		if err != nil {
			return fmt.Errorf("failed to build archive for %s / %s (%w)", binary.platform, binary.architecture, err)
		}
		if err := r.AddArtifact(branding.ArtifactPrefix+string(version)+"_"+string(binary.platform)+"_"+string(binary.architecture)+".tar.gz", tarFile); err != nil {
			return fmt.Errorf("failed to add tar file as artifact (%w)", err)
		}
	}

	sums := r.buildSumsFile()
	sumsSig, err := r.signFile(sums)
	if err != nil {
		return fmt.Errorf("cannot sign checksum file (%w)", err)
	}
	if err := r.AddArtifact(branding.ArtifactPrefix+string(version)+"_SHA256SUMS", sums); err != nil {
		return fmt.Errorf("failed to add checksum file (%w)", err)
	}
	if err := r.AddArtifact(branding.ArtifactPrefix+string(version)+"_SHA256SUMS.gpgsig", sumsSig); err != nil {
		return fmt.Errorf("failed to add checksum signature file (%w)", err)
	}

	if err := mirror.CreateVersion(ctx, version); err != nil {
		return fmt.Errorf("failed to create version on mirror (%w)", err)
	}
	for artifactName, artifact := range r.artifacts {
		if err := mirror.CreateVersionAsset(ctx, version, artifactName, artifact); err != nil {
			return fmt.Errorf("cannot create version asset %s in mirror (%w)", artifactName, err)
		}
	}
	return nil
}

func (r *releaseBuilder) buildSumsFile() []byte {
	result := ""
	for filename, contents := range r.artifacts {
		hash := sha256.New()
		hash.Write(contents)
		checksum := hex.EncodeToString(hash.Sum(nil))
		result += checksum + "  " + filename + "\n"
	}
	return []byte(result)
}

func (r *releaseBuilder) signFile(contents []byte) ([]byte, error) {
	msg := crypto.NewPlainMessage(contents)
	keyring, err := crypto.NewKeyRing(r.gpgKey)
	if err != nil {
		return nil, fmt.Errorf("failed to construct keyring (%w)", err)
	}
	signature, err := keyring.SignDetached(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data (%w)", err)
	}
	return signature.GetBinary(), nil
}

func buildTarFile(binary []byte, extraFiles map[string][]byte) ([]byte, error) {
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
		return nil, fmt.Errorf("failed to construct tar file header (%w)", err)
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("failed to write file header to tar file (%w)", err)
	}
	if _, err := tarWriter.Write(binary); err != nil {
		return nil, fmt.Errorf("failed to write binary to tar file (%w)", err)
	}

	for file, contents := range extraFiles {
		header, err = tar.FileInfoHeader(
			&fileInfo{
				name:    file,
				size:    int64(len(contents)),
				mode:    0644,
				modTime: time.Now(),
				isDir:   false,
			}, file)
		if err != nil {
			return nil, fmt.Errorf("failed to construct tar file header (%w)", err)
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("failed to write file header to tar file (%w)", err)
		}
		if _, err := tarWriter.Write(binary); err != nil {
			return nil, fmt.Errorf("failed to write binary to tar file (%w)", err)
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer (%w)", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer (%w)", err)
	}
	return buf.Bytes(), nil
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
