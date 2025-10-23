// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"fmt"
)

func (m *mirror) DownloadNightly(_ context.Context, opts ...DownloadOpt) ([]byte, error) {
	return nil, fmt.Errorf("downloading nightly builds is not supported through mirrors")
}
