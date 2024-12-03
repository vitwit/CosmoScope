package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var GlobalConfig Config

func Load() Config {
	file, err := os.ReadFile("configs/config.json")
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}

	if err := json.Unmarshal(file, &GlobalConfig); err != nil {
		panic(fmt.Sprintf("Error parsing config file: %v", err))
	}

	return GlobalConfig
}

func LoadIBCAssets(filepath string) (map[string]*IBCAsset, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading IBC assets file: %v", err)
	}

	var assets []IBCAsset
	if err := json.Unmarshal(file, &assets); err != nil {
		return nil, fmt.Errorf("error parsing IBC assets file: %v", err)
	}

	assetMap := make(map[string]*IBCAsset)
	for _, asset := range assets {
		if asset.Type == "ibc" {
			assetCopy := asset
			assetMap[asset.Denom] = &assetCopy
			if _, exists := assetMap[asset.Symbol]; !exists {
				assetMap[asset.Symbol] = &assetCopy
			}
		}
	}

	return assetMap, nil
}
