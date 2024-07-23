// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package mockmirror

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

func buildSumsFile(files map[string][]byte) []byte {
	result := ""
	for filename, contents := range files {
		hash := sha256.New()
		hash.Write(contents)
		checksum := hex.EncodeToString(hash.Sum(nil))
		result += checksum + "  " + filename + "\n"
	}
	return []byte(result)
}

func signFile(t *testing.T, contents []byte, key *crypto.Key) []byte {
	msg := crypto.NewPlainMessage(contents)
	keyring, err := crypto.NewKeyRing(key)
	if err != nil {
		t.Fatalf("Failed to construct keyring (%v)", err)
	}
	signature, err := keyring.SignDetached(msg)
	if err != nil {
		t.Fatalf("Failed to sign data (%v)", err)
	}
	return signature.GetBinary()
}
