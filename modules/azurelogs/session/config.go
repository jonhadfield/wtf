package session

import (
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// open yaml files in the config directory
const (
	TestDataFile    = "testdata.yaml"
	AzureConfigFile = "azure.yaml"
)

type QueryFile struct {
	Title          string   `yaml:"title"`
	Style          string   `yaml:"style"`
	SubscriptionID string   `yaml:"azure_subscription_id"`
	WorkspaceID    string   `yaml:"azure_workspace_id"`
	Columns        []string `yaml:"columns"`
	Query          string   `yaml:"query"`
}

func readQueryFile(queryPath string) error {
	file, err := os.OpenFile(queryPath, os.O_RDONLY, 0o600)
	if err != nil {
		return err
	}

	if file.Name() != AzureConfigFile && file.Name() != TestDataFile && file.Name()[len(file.Name())-5:] == ".yaml" {
		var configFile QueryFile
		configFile, err = readConfigFile(queryPath)
		if err != nil {
			return err
		}

		S.QueryFile = configFile
	}

	return nil
}

// readConfigFile reads a single config file and returns a QueryFile struct
func readConfigFile(filePath string) (QueryFile, error) {
	var configFile QueryFile
	data, err := os.ReadFile(filePath)
	if err != nil {
		return configFile, fmt.Errorf("failed to read query config file %s: %w", filePath, err)
	}

	err = yaml.Unmarshal(data, &configFile)
	if err != nil {
		return configFile, fmt.Errorf("failed to parse YAML in config file %s: %w", filePath, err)
	}

	return configFile, nil
}
