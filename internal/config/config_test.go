package config

import (
	"os"
	"testing"
)

func TestLoadFiles_EnvExpansion(t *testing.T) {
	// Create temporary config file
	cfgContent := `
apisix:
  url: "${APISIX_URL}"
  key: "${APISIX_KEY}"
proto:
  includes:
    - /path/to/proto
`
	cfgFile, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfgFile.Name())
	if _, err := cfgFile.WriteString(cfgContent); err != nil {
		t.Fatal(err)
	}
	cfgFile.Close()

	// Create temporary data file
	dataContent := `
upstreams:
  - id: auth-service
    nodes:
      - host: "${AUTH_SERVICE_HOST}"
        port: ${AUTH_SERVICE_PORT}
`
	dataFile, err := os.CreateTemp("", "data.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dataFile.Name())
	if _, err := dataFile.WriteString(dataContent); err != nil {
		t.Fatal(err)
	}
	dataFile.Close()

	// Set environment variables
	os.Setenv("APISIX_URL", "http://localhost:9180")
	os.Setenv("APISIX_KEY", "secret-key")
	os.Setenv("AUTH_SERVICE_HOST", "auth.local")
	os.Setenv("AUTH_SERVICE_PORT", "8080")
	defer func() {
		os.Unsetenv("APISIX_URL")
		os.Unsetenv("APISIX_KEY")
		os.Unsetenv("AUTH_SERVICE_HOST")
		os.Unsetenv("AUTH_SERVICE_PORT")
	}()

	// Load files
	cfg, data, err := LoadFiles(cfgFile.Name(), dataFile.Name())
	if err != nil {
		t.Fatalf("Failed to load files: %v", err)
	}

	// Verify expansion
	if cfg.Apisix.URL != "http://localhost:9180" {
		t.Errorf("Expected APISIX_URL to be http://localhost:9180, got %s", cfg.Apisix.URL)
	}
	if cfg.Apisix.Key != "secret-key" {
		t.Errorf("Expected APISIX_KEY to be secret-key, got %s", cfg.Apisix.Key)
	}
	if data.Upstreams[0].Nodes[0].Host != "auth.local" {
		t.Errorf("Expected host to be auth.local, got %s", data.Upstreams[0].Nodes[0].Host)
	}
	if data.Upstreams[0].Nodes[0].Port != 8080 {
		t.Errorf("Expected port to be 8080, got %d", data.Upstreams[0].Nodes[0].Port)
	}
}

func TestLoadFiles_IdPrefix(t *testing.T) {
	cfgContent := `
apisix:
  url: "http://localhost:9180"
  key: "secret"
id_prefix: "user_service"
`
	cfgFile, _ := os.CreateTemp("", "config.yaml")
	defer os.Remove(cfgFile.Name())
	cfgFile.WriteString(cfgContent)
	cfgFile.Close()

	dataFile, _ := os.CreateTemp("", "data.yaml")
	defer os.Remove(dataFile.Name())
	dataFile.WriteString("upstreams: []")
	dataFile.Close()

	cfg, _, err := LoadFiles(cfgFile.Name(), dataFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if cfg.IdPrefix != "user_service" {
		t.Errorf("Expected id_prefix to be user_service, got %s", cfg.IdPrefix)
	}
}
