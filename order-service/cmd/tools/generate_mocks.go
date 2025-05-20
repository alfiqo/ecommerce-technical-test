package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// This tool generates mock implementations for interfaces
func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Define the interfaces to mock
	interfaces := []struct {
		sourceFile string
		destFile   string
		packageName string
	}{
		{
			sourceFile: "./internal/usecase/interface.go",
			destFile:   "./mocks/usecase/inventory_usecase_mock.go",
			packageName: "usecase_mock",
		},
		{
			sourceFile: "./internal/gateway/warehouse/interface.go",
			destFile:   "./mocks/gateway/warehouse/warehouse_gateway_mock.go",
			packageName: "warehouse_mock",
		},
		{
			sourceFile: "./internal/repository/inventory_repository.go",
			destFile:   "./mocks/repository/inventory_repository_mock.go",
			packageName: "repository_mock",
		},
	}

	// Create directories if they don't exist
	dirs := []string{
		"./mocks/usecase",
		"./mocks/repository",
		"./mocks/gateway/warehouse",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.WithError(err).Fatalf("Failed to create directory: %s", dir)
		}
	}

	// Generate mocks
	for _, i := range interfaces {
		// Check if source file exists
		if _, err := os.Stat(i.sourceFile); os.IsNotExist(err) {
			log.Infof("Source file %s does not exist, skipping...", i.sourceFile)
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(i.destFile), 0755); err != nil {
			log.WithError(err).Fatalf("Failed to create directory: %s", filepath.Dir(i.destFile))
		}

		// Build mockgen command
		args := []string{
			"-source", i.sourceFile,
			"-destination", i.destFile,
			"-package", i.packageName,
		}

		// Execute mockgen
		cmd := exec.Command("mockgen", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.WithError(err).Warnf("Failed to generate mock for %s: %s", i.sourceFile, string(out))
			continue
		}

		log.Infof("Generated mock for %s", i.sourceFile)
	}
}