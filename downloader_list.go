// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
)

// ListVersionsOptions are the options for listing versions.
type ListVersionsOptions struct {
	Stability *Stability
}

// ListVersionOpt is an option for the ListVersions call.
type ListVersionOpt func(options *ListVersionsOptions) error

// ListVersionOptMinimumStability sets the minimum stability for listing versions.
func ListVersionOptMinimumStability(stability Stability) ListVersionOpt {
	return func(options *ListVersionsOptions) error {
		options.Stability = &stability
		return nil
	}
}

type listVersionsResponse struct {
	Versions []VersionWithArtifacts `json:"versions"`
}

func (d *downloader) ListVersions(ctx context.Context, opts ...ListVersionOpt) ([]VersionWithArtifacts, error) {
	options := ListVersionsOptions{}
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, &InvalidOptionsError{err}
		}
	}
	body, err := d.getRequest(ctx, d.config.APIURL, d.config.APIURLAuthorization)
	if err != nil {
		return nil, &RequestFailedError{
			err,
		}
	}
	defer func() {
		_ = body.Close()
	}()

	responseData := listVersionsResponse{}
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&responseData); err != nil {
		return nil, &RequestFailedError{
			fmt.Errorf("failed to decode JSON response from API endpoint (%w)", err),
		}
	}

	var versions []VersionWithArtifacts
	for _, version := range responseData.Versions {
		if options.Stability == nil || options.Stability.Matches(version.ID) {
			versions = append(versions, version)
		}
	}

	slices.SortStableFunc(versions, func(a, b VersionWithArtifacts) int {
		return -1 * a.ID.Compare(b.ID)
	})

	return versions, nil
}
