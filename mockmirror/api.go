// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package mockmirror

import (
	"encoding/json"
	"testing"

	"github.com/opentofu/tofudl"
)

func buildAPI(t *testing.T, version string, filenames []string) []byte {
	response := tofudl.APIResponse{
		Versions: []tofudl.VersionWithArtifacts{
			{
				ID:    tofudl.Version(version),
				Files: filenames,
			},
		},
	}
	result, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to build API file (%v)", err)
	}
	return result
}
