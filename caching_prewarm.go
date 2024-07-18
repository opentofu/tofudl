// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"fmt"
)

func (c *cachingDownloader) PreWarm(ctx context.Context, versionCount int, progress func(pct int8)) error {
	if c.config.isDisabled() {
		return nil
	}

	versions, err := c.ListVersions(ctx)
	if err != nil {
		return err
	}
	if versionCount > 0 {
		versions = versions[:versionCount]
	}
	totalArtifacts := 0
	for _, version := range versions {
		totalArtifacts += len(version.Files)
	}
	downloadedArtifacts := 0
	for _, version := range versions {
		for _, artifact := range version.Files {
			_, err = c.DownloadArtifact(ctx, version, artifact)
			if err != nil {
				return fmt.Errorf("failed to download artifact %s for version %s (%w)", artifact, version.ID, err)
			}
			downloadedArtifacts += 1
			if progress != nil {
				progress(int8(100 * float64(downloadedArtifacts) / float64(totalArtifacts)))
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
	}
	return nil
}
