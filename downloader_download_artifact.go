// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
)

var artifactRe = regexp.MustCompile(`^[a-zA-Z0-9._\-]+$`)

func (d *downloader) DownloadArtifact(ctx context.Context, version VersionWithArtifacts, artifactName string) ([]byte, error) {
	found := false
	for _, file := range version.Files {
		if file == artifactName {
			found = true
		}
	}
	if !found {
		return nil, &NoSuchArtifactError{
			artifactName,
		}
	}

	if !artifactRe.MatchString(artifactName) {
		return nil, &InvalidOptionsError{
			Cause: fmt.Errorf("invalid artifact name: " + artifactName),
		}
	}

	wr := &bytes.Buffer{}
	if err := d.downloadMirrorURLTemplate.Execute(wr, &MirrorURLTemplateParameters{
		Version:  version.ID,
		Artifact: artifactName,
	}); err != nil {
		return nil, &InvalidConfigurationError{
			Message: "Failed to construct mirror URL",
			Cause:   err,
		}
	}

	reader, err := d.getRequest(ctx, wr.String(), d.config.DownloadMirrorAuthorization)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, &RequestFailedError{Cause: err}
	}
	return b, nil
}
