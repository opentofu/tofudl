// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"fmt"
)

// DownloadOptions describes the settings for downloading. They default to the current architecture and platform.
type DownloadOptions struct {
	Platform         Platform
	Architecture     Architecture
	Version          Version
	NightlyID        NightlyID
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

// DownloadOptNighlyBuildID specified id of the nightly build in format "${build_date}-${commit_hash}" (20251006-f839281c15)
// If the id isn't in the correct regex format we return error. This option is specifically for nightly download and does not interfere with other version options.
func DownloadOptNightlyBuildID(id string) DownloadOpt {
	return func(spec *DownloadOptions) error {
		nighlyID := NightlyID(id)
		if err := nighlyID.Validate(); err != nil {
			return err
		}
		spec.NightlyID = nighlyID
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
	return download(ctx, opts, d.ListVersions, d.DownloadVersion)
}

func download(ctx context.Context, opts []DownloadOpt, listVersionsFunc func(ctx context.Context, opts ...ListVersionOpt) ([]VersionWithArtifacts, error), downloadVersionFunc func(ctx context.Context, version VersionWithArtifacts, platform Platform, architecture Architecture) ([]byte, error)) ([]byte, error) {
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
	listResult, err := listVersionsFunc(ctx, listOptions...)
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
	return downloadVersionFunc(ctx, *foundVer, downloadOpts.Platform, downloadOpts.Architecture)
}
