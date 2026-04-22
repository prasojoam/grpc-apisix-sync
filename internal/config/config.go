package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// --- Configuration Models ---

type ApisixConfig struct {
	URL string `yaml:"url"`
	Key string `yaml:"key"`
}

type ProtoConfig struct {
	Includes []string `yaml:"includes"`
}

type Config struct {
	Apisix       ApisixConfig `yaml:"apisix"`
	Proto        ProtoConfig  `yaml:"proto"`
	ResetOnStart bool         `yaml:"reset_on_start"`
}

// --- Data Models ---

type ProtoDef struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

type UpstreamDef struct {
	ID    string `yaml:"id"`
	Nodes []Node `yaml:"nodes"`
}

type Node struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Weight int    `yaml:"weight"`
}

type ServiceDef struct {
	ID       string `yaml:"id"`
	Upstream string `yaml:"upstream"`
}

type RouteDefaults struct {
	Service string `yaml:"service"`
	Proto   string `yaml:"proto"`
}

type RouteDef struct {
	ID      string `yaml:"id"`
	URI     string `yaml:"uri"`
	Service string `yaml:"service"`
	Proto   string `yaml:"proto"`
	Methods string `yaml:"methods"`
	Grpc    string `yaml:"grpc"`
}

type Data struct {
	Protos        []ProtoDef    `yaml:"protos"`
	Upstreams     []UpstreamDef `yaml:"upstreams"`
	Services      []ServiceDef  `yaml:"services"`
	RouteDefaults RouteDefaults `yaml:"route_defaults"`
	Routes        []RouteDef    `yaml:"routes"`
}

func LoadFiles(configPath, dataPath string) (*Config, *Data, error) {
	cfg := &Config{}
	data := &Data{}

	// Load Config
	cfgData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, err
	}
	if err := yaml.Unmarshal(cfgData, cfg); err != nil {
		return nil, nil, err
	}

	// Load Data
	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, nil, err
	}
	if err := yaml.Unmarshal(dataBytes, data); err != nil {
		return nil, nil, err
	}

	return cfg, data, nil
}
