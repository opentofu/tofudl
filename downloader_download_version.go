// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/opentofu/tofudl/branding"
)

func (d *downloader) DownloadVersion(ctx context.Context, version VersionWithArtifacts, platform Platform, architecture Architecture) ([]byte, error) {
	return downloadVersion(ctx, version, platform, architecture, d.DownloadArtifact, d.VerifyArtifact)
}

func downloadVersion(
	ctx context.Context,
	version VersionWithArtifacts,
	platform Platform,
	architecture Architecture,
	downloadArtifactFunc func(ctx context.Context, version VersionWithArtifacts, artifactName string) ([]byte, error),
	verifyArtifactFunc func(artifactName string, artifactContents []byte, sumsFileContents []byte, signatureFileContent []byte) error,
) ([]byte, error) {
	sumsFileName := branding.ArtifactPrefix + string(version.ID) + "_SHA256SUMS"
	sumsBody, err := downloadArtifactFunc(ctx, version, sumsFileName)
	if err != nil {
		return nil, &RequestFailedError{Cause: fmt.Errorf("failed to download %s (%w)", sumsFileName, err)}
	}

	sumsSigFileName := branding.ArtifactPrefix + string(version.ID) + "_SHA256SUMS.gpgsig"
	sumsSig, err := downloadArtifactFunc(ctx, version, sumsSigFileName)
	if err != nil {
		return nil, &RequestFailedError{Cause: fmt.Errorf("failed to download %s (%w)", sumsSigFileName, err)}
	}

	platform, err = platform.ResolveAuto()
	if err != nil {
		return nil, err
	}
	architecture, err = architecture.ResolveAuto()
	if err != nil {
		return nil, err
	}

	archiveName := branding.ArtifactPrefix + string(version.ID) + "_" + string(platform) + "_" + string(architecture) + ".tar.gz"
	archive, err := downloadArtifactFunc(ctx, version, archiveName)
	if err != nil {
		var noSuchArtifact *NoSuchArtifactError
		if errors.As(err, &noSuchArtifact) {
			return nil, &UnsupportedPlatformOrArchitectureError{
				Platform:     platform,
				Architecture: architecture,
				Version:      version.ID,
			}
		}
	}

	if err := verifyArtifactFunc(archiveName, archive, sumsBody, sumsSig); err != nil {
		return nil, err
	}

	return extractBinaryFromTarGz(archiveName, archive, platform)
}

// extractBinaryFromTarGz extracts the OpenTofu binary from a tar.gz archive
// takes platform as an argument, to determine if we should look for "tofu" or "tofu.exe"
// since it is possible to download for other patforms/archs from a different one
func extractBinaryFromTarGz(archiveName string, archive []byte, platform Platform) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, &ArtifactCorruptedError{
			Artifact: archiveName,
			Cause:    err,
		}
	}
	defer func() {
		_ = gz.Close()
	}()

	binaryName := "tofu"
	if platform == PlatformWindows {
		binaryName = "tofu.exe"
	}
	tarFile := tar.NewReader(gz)
	for {
		current, err := tarFile.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, &ArtifactCorruptedError{
				Artifact: archiveName,
				Cause:    err,
			}
		}

		if current.Name != binaryName || current.Typeflag != tar.TypeReg {
			continue
		}
		buf := &bytes.Buffer{}
		// Protect against a DoS vulnerability by limiting the maximum size of the binary.
		if _, err := io.CopyN(buf, tarFile, branding.MaximumUncompressedFileSize); err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, &ArtifactCorruptedError{
					Artifact: archiveName,
					Cause:    err,
				}
			}
		}
		if buf.Len() == branding.MaximumUncompressedFileSize {
			return nil, &ArtifactCorruptedError{
				Artifact: archiveName,
				Cause:    fmt.Errorf("artifact too large (larger than %d bytes)", branding.MaximumUncompressedFileSize),
			}
		}
		return buf.Bytes(), nil
	}
	return nil, &ArtifactCorruptedError{
		Artifact: archiveName,
		Cause:    fmt.Errorf("file named %s not found", binaryName),
	}
}
