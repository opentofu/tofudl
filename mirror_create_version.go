// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

func (m *mirror) CreateVersion(_ context.Context, version Version) error {
	if m.pullThroughDownloader != nil {
		return fmt.Errorf("cannot use CreateVersionAsset when a pull-through mirror is configured")
	}
	if err := version.Validate(); err != nil {
		return err
	}

	responseData := APIResponse{}

	reader, _, err := m.storage.ReadAPIFile()
	if err != nil {
		var notFound *CacheMissError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("cannot read api.json from mirror storage (%w)", err)
		}
	} else {
		decoder := json.NewDecoder(reader)
		if err := decoder.Decode(&responseData); err != nil {
			return fmt.Errorf("api.json corrupt in mirror storage (%w)", err)
		}
	}

	for _, foundVersion := range responseData.Versions {
		if foundVersion.ID == version {
			return fmt.Errorf("version %s already exists", version)
		}
	}
	responseData.Versions = append([]VersionWithArtifacts{
		{
			ID:    version,
			Files: []string{},
		},
	}, responseData.Versions...)

	marshalled, err := json.Marshal(responseData)
	if err != nil {
		return fmt.Errorf("failed to re-encode api.json (%w)", err)
	}

	if err := m.storage.StoreAPIFile(marshalled); err != nil {
		return fmt.Errorf("failed to store api.json (%w)", err)
	}

	return nil
}
