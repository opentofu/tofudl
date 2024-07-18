// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
)

func (c *cachingDownloader) DownloadVersion(ctx context.Context, version VersionWithArtifacts, platform Platform, architecture Architecture) ([]byte, error) {
	return downloadVersion(ctx, version, platform, architecture, c.DownloadArtifact, c.backingDownloader.VerifyArtifact)
}
