// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"text/template"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// Downloader describes the functions the downloader provides.
type Downloader interface {
	// ListVersions lists all versions matching the filter options in descending order.
	ListVersions(ctx context.Context, opts ...ListVersionOpt) ([]VersionWithArtifacts, error)

	// DownloadArtifact downloads an artifact for a version.
	DownloadArtifact(ctx context.Context, version VersionWithArtifacts, artifactName string) ([]byte, error)

	// VerifyArtifact verifies a named artifact against a checksum file with SHA256 hashes and the checksum file against a GPG signature file.
	VerifyArtifact(artifactName string, artifactContents []byte, sumsFileContents []byte, signatureFileContent []byte) error

	// DownloadVersion downloads the OpenTofu binary from a specific artifact obtained from ListVersions.
	DownloadVersion(ctx context.Context, version VersionWithArtifacts, platform Platform, architecture Architecture) ([]byte, error)

	// Download downloads the OpenTofu binary and provides it as a byte slice.
	Download(ctx context.Context, opts ...DownloadOpt) ([]byte, error)
}

func New(opts ...ConfigOpt) (Downloader, error) {
	cfg := Config{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	cfg.ApplyDefaults()

	tpl := template.New("url")
	tpl, err := tpl.Parse(cfg.DownloadMirrorURLTemplate)
	if err != nil {
		return nil, &InvalidConfigurationError{
			Message: "Cannot parse download mirror URL template",
			Cause:   err,
		}
	}

	key, err := crypto.NewKeyFromArmored(cfg.GPGKey)
	if err != nil {
		return nil, &InvalidConfigurationError{
			Message: "Failed to decode GPG key",
			Cause:   err,
		}
	}
	if !key.CanVerify() {
		return nil, &InvalidConfigurationError{Message: "The provided key cannot be used for verification."}
	}

	keyRing, err := crypto.NewKeyRing(key)
	if err != nil {
		return nil, &InvalidConfigurationError{Message: "Cannot create keyring", Cause: err}
	}

	return &downloader{
		cfg,
		tpl,
		keyRing,
	}, nil
}

type downloader struct {
	config                    Config
	downloadMirrorURLTemplate *template.Template
	keyRing                   *crypto.KeyRing
}
