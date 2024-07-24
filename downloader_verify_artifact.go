// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

func (d *downloader) VerifyArtifact(artifactName string, artifactContents []byte, sumsFileContents []byte, signatureFileContent []byte) error {
	return verifyArtifact(d.keyRing, artifactName, artifactContents, sumsFileContents, signatureFileContent)
}

func verifyArtifact(keyRing *crypto.KeyRing, artifactName string, artifactContents []byte, sumsFileContents []byte, signatureFileContent []byte) error {
	if err := keyRing.VerifyDetached(
		crypto.NewPlainMessage(sumsFileContents),
		crypto.NewPGPSignature(signatureFileContent),
		crypto.GetUnixTime(),
	); err != nil {
		return &SignatureError{
			"Signature verification failed",
			err,
		}
	}

	hash := sha256.New()
	hash.Write(artifactContents)
	sum := hex.EncodeToString(hash.Sum(nil))

	found := false
	for _, line := range strings.Split(string(sumsFileContents), "\n") {
		if strings.HasSuffix(strings.TrimSpace(line), "  "+artifactName) {
			parts := strings.Split(strings.TrimSpace(line), " ")
			expectedSum := parts[0]
			found = true
			if expectedSum != sum {
				return &ArtifactCorruptedError{
					artifactName,
					fmt.Errorf(
						"invalid checksum, expected %s found %s",
						expectedSum,
						sum,
					),
				}
			}
		}
	}
	if !found {
		return &SignatureError{
			Message: fmt.Sprintf(
				"No checksum found for artifact %s",
				artifactName,
			),
			Cause: nil,
		}
	}
	return nil
}
