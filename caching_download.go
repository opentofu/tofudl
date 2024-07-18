// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
)

func (c *cachingDownloader) Download(ctx context.Context, opts ...DownloadOpt) ([]byte, error) {
	return download(ctx, opts, c.ListVersions, c.DownloadVersion)
}
