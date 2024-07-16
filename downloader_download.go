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
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/opentofu/tofudl/branding"
)

// DownloadOptions describes the settings for downloading. They default to the current architecture and platform.
type DownloadOptions struct {
	Platform         Platform
	Architecture     Architecture
	Version          Version
	MinimumStability *Stability
}

// DownloadOpt is a function that modifies the download options.
type DownloadOpt func(spec *DownloadOptions) error

// DownloadOptPlatform specifies the platform to download for. This defaults to the current platform.
func DownloadOptPlatform(platform Platform) DownloadOpt {
	return func(spec *DownloadOptions) error {
		if err := platform.Validate(); err != nil {
			return err
		}
		spec.Platform = platform
		return nil
	}
}

// DownloadOptArchitecture specifies the architecture to download for. This defaults to the current architecture.
func DownloadOptArchitecture(architecture Architecture) DownloadOpt {
	return func(spec *DownloadOptions) error {
		if err := architecture.Validate(); err != nil {
			return err
		}
		spec.Architecture = architecture
		return nil
	}
}

// DownloadOptVersion specifies the version to download. Defaults to the latest version with the specified minimum
// stability.
func DownloadOptVersion(version Version) DownloadOpt {
	return func(spec *DownloadOptions) error {
		if err := version.Validate(); err != nil {
			return err
		}
		if spec.MinimumStability != nil {
			return &InvalidOptionsError{
				fmt.Errorf("the stability and version constraints for download are mutually exclusive"),
			}
		}
		spec.Version = version
		return nil
	}
}

// DownloadOptMinimumStability specifies the minimum stability of the version to download. This is mutually exclusive
// with setting the Version.
func DownloadOptMinimumStability(stability Stability) DownloadOpt {
	return func(spec *DownloadOptions) error {
		if err := stability.Validate(); err != nil {
			return err
		}
		if spec.Version != "" {
			return &InvalidOptionsError{
				fmt.Errorf("the stability and version constraints for download are mutually exclusive"),
			}
		}
		spec.MinimumStability = &stability
		return nil
	}
}

func (d *downloader) Download(ctx context.Context, opts ...DownloadOpt) ([]byte, error) {
	downloadOpts := DownloadOptions{}
	for _, opt := range opts {
		if err := opt(&downloadOpts); err != nil {
			return nil, err
		}
	}
	var listOptions []ListVersionOpt
	if downloadOpts.MinimumStability != nil {
		listOptions = append(listOptions, ListVersionOptMinimumStability(*downloadOpts.MinimumStability))
	}
	listResult, err := d.ListVersions(ctx, listOptions...)
	if err != nil {
		return nil, err
	}
	if len(listResult) == 0 {
		return nil, &RequestFailedError{
			Cause: fmt.Errorf("the API request returned no versions"),
		}
	}

	var foundVer *VersionWithArtifacts
	if downloadOpts.Version != "" {
		for _, ver := range listResult {
			if ver.ID == downloadOpts.Version {
				ver := ver
				foundVer = &ver
				break
			}
		}
		if foundVer == nil {
			return nil, &NoSuchVersionError{downloadOpts.Version}
		}
	} else {
		foundVer = &listResult[0]
	}
	return d.DownloadVersion(ctx, *foundVer, downloadOpts.Platform, downloadOpts.Architecture)
}

func (d *downloader) DownloadVersion(ctx context.Context, version VersionWithArtifacts, platform Platform, architecture Architecture) ([]byte, error) {
	sumsFileName := "tofu_" + string(version.ID) + "_SHA256SUMS"
	sumsBody, err := d.downloadArtifact(ctx, version, sumsFileName)
	if err != nil {
		return nil, &RequestFailedError{Cause: fmt.Errorf("failed to download %s (%w)", sumsFileName, err)}
	}

	sumsSigFileName := "tofu_" + string(version.ID) + "_SHA256SUMS.gpgsig"
	sumsSig, err := d.downloadArtifact(ctx, version, sumsSigFileName)
	if err != nil {
		return nil, &RequestFailedError{Cause: fmt.Errorf("failed to download %s (%w)", sumsSigFileName, err)}
	}

	if err := d.keyRing.VerifyDetached(
		crypto.NewPlainMessage(sumsBody),
		crypto.NewPGPSignature(sumsSig),
		crypto.GetUnixTime(),
	); err != nil {
		return nil, &SignatureError{
			"Signature verification failed",
			err,
		}
	}

	platform, err = platform.ResolveAuto()
	if err != nil {
		return nil, err
	}
	architecture, err = architecture.ResolveAuto()
	if err != nil {
		return nil, err
	}

	archiveName := "tofu_" + string(version.ID) + "_" + string(platform) + "_" + string(architecture) + ".tar.gz"
	archive, err := d.downloadArtifact(ctx, version, archiveName)
	if err != nil {
		var noSuchArtifact *NoSuchArtifactError
		if errors.As(err, noSuchArtifact) {
			return nil, &UnsupportedPlatformOrArchitectureError{
				Platform:     platform,
				Architecture: architecture,
				Version:      version.ID,
			}
		}
	}
	hash := sha256.New()
	hash.Write(archive)
	sum := hex.EncodeToString(hash.Sum(nil))

	found := false
	for _, line := range strings.Split(string(sumsBody), "\n") {
		if strings.HasSuffix(strings.TrimSpace(line), "  "+archiveName) {
			parts := strings.Split(strings.TrimSpace(line), " ")
			expectedSum := parts[0]
			found = true
			if expectedSum != sum {
				return nil, &ArtifactCorruptedError{
					archiveName,
					fmt.Errorf(
						"invalid checksum, expected %s found %s",
						expectedSum,
						sum,
					),
				}
			}
		}
	}
	if !found {
		return nil, &SignatureError{
			Message: fmt.Sprintf(
				"No checksum found for artifact %s",
				archiveName,
			),
			Cause: nil,
		}
	}

	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, &ArtifactCorruptedError{
			Artifact: archiveName,
			Cause:    nil,
		}
	}
	defer func() {
		_ = gz.Close()
	}()

	binaryName := branding.BinaryName
	if platform == PlatformWindows {
		binaryName += ".exe"
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
