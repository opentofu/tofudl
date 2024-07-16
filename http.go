// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (d *downloader) getRequest(ctx context.Context, url string, authorization string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct HTTP request (%w)", err)
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed (%w)", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return resp.Body, nil
}
