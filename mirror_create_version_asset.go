package tofudl

import (
	"context"
	"encoding/json"
	"fmt"
)

func (m *mirror) CreateVersionAsset(ctx context.Context, version Version, assetName string, assetData []byte) error {
	if m.pullThroughDownloader != nil {
		return fmt.Errorf("cannot use CreateVersionAsset when a pull-through mirror is configured")
	}
	if err := version.Validate(); err != nil {
		return err
	}

	responseData := APIResponse{}

	reader, _, err := m.storage.ReadAPIFile()
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&responseData); err != nil {
		return fmt.Errorf("api.json corrupt in mirror storage (%w)", err)
	}

	foundIndex := -1
	for i, foundVersion := range responseData.Versions {
		if foundVersion.ID == version {
			foundIndex = i
		}
	}
	if foundIndex == -1 {
		return fmt.Errorf("version does not exist: %s", version)
	}

	responseData.Versions[foundIndex].Files = append(responseData.Versions[foundIndex].Files, assetName)

	if err := m.storage.StoreArtifact(version, assetName, assetData); err != nil {
		return fmt.Errorf("cannot store asset %s (%w)", assetName, err)
	}

	marshalled, err := json.Marshal(responseData)
	if err != nil {
		return fmt.Errorf("failed to re-encode api.json (%w)", err)
	}

	if err := m.storage.StoreAPIFile(marshalled); err != nil {
		return fmt.Errorf("failed to store api.json (%w)", err)
	}

	return nil
}
