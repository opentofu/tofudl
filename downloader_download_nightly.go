// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (d *downloader) DownloadNightly(ctx context.Context, opts ...DownloadOpt) ([]byte, error) {
	return downloadLatestNightly(ctx, opts, d.config.HTTPClient)
}

func downloadLatestNightly(ctx context.Context, opts []DownloadOpt, httpClient *http.Client) ([]byte, error) {
	downloadOpts := DownloadOptions{}
	for _, opt := range opts {
		if err := opt(&downloadOpts); err != nil {
			return nil, err
		}
	}

	platform, err := downloadOpts.Platform.ResolveAuto()
	if err != nil {
		return nil, err
	}

	architecture, err := downloadOpts.Architecture.ResolveAuto()
	if err != nil {
		return nil, err
	}

	//
	nightlyID := downloadOpts.NightlyID
	if nightlyID == "" {
		metadata, err := fetchLatestNightlyMetadata(ctx, httpClient)
		if err != nil {
			return nil, err
		}
		nightlyID, err = newNightlyID(metadata.Date, metadata.Commit)
		if err != nil {
			return nil, err
		}
	}

	artifactName := fmt.Sprintf("tofu_nightly-%s_%s_%s.tar.gz",
		nightlyID,
		string(platform),
		string(architecture),
	)
	path := fmt.Sprintf("/nightlies/%s/", nightlyID.GetDate())

	// Download the artifact
	artifactURL := fmt.Sprintf("https://nightlies.opentofu.org%s%s", path, artifactName)
	artifact, err := downloadNightlyFile(ctx, httpClient, artifactURL)
	if err != nil {
		return nil, &RequestFailedError{Cause: fmt.Errorf("failed to download artifact %s (%w)", artifactName, err)}
	}

	// Download SHA256SUMS file
	sumsFileName := fmt.Sprintf("tofu_nightly-%s_SHA256SUMS", nightlyID)
	sumsURL := fmt.Sprintf("https://nightlies.opentofu.org%s%s", path, sumsFileName)
	sumsBody, err := downloadNightlyFile(ctx, httpClient, sumsURL)
	if err != nil {
		return nil, &RequestFailedError{Cause: fmt.Errorf("failed to download %s (%w)", sumsFileName, err)}
	}

	// Verify artifact checksum
	if err := verifyArtifactSHAOnly(artifactName, artifact, sumsBody); err != nil {
		return nil, err
	}

	// Extract binary from tar.gz
	binary, err := extractBinaryFromTarGz(artifactName, artifact)
	if err != nil {
		return nil, err
	}

	return binary, nil
}

type nightlyMetadata struct {
	Version   string   `json:"version"`
	Date      string   `json:"date"`
	Commit    string   `json:"commit"`
	Path      string   `json:"path"`
	Artifacts []string `json:"artifacts"`
}

// fetchLatestNightlyMetadata fetches and parses the latest nightly metadata from the https://nightlies.opentofu.org/nightlies/latest.json
func fetchLatestNightlyMetadata(ctx context.Context, httpClient *http.Client) (*nightlyMetadata, error) {
	const metadataURL = "https://nightlies.opentofu.org/nightlies/latest.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct HTTP request (%w)", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed (%w)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body (%w)", err)
	}

	var metadata nightlyMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse nightly metadata (%w)", err)
	}

	return &metadata, nil
}

func downloadNightlyFile(ctx context.Context, httpClient *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct HTTP request (%w)", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed (%w)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body (%w)", err)
	}

	return b, nil
}
