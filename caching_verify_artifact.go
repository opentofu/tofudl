// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

func (c *cachingDownloader) VerifyArtifact(artifactName string, artifactContents []byte, sumsFileContents []byte, signatureFileContent []byte) error {
	return c.backingDownloader.VerifyArtifact(artifactName, artifactContents, sumsFileContents, signatureFileContent)
}
