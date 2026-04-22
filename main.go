package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"gopkg.in/yaml.v3"
)

// --- Configuration Models ---
// ... (previous structs)

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
// (rest of the models remain same)

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

// --- Sync Logic ---

type Syncer struct {
	Config Config
	Data   Data
	Client *http.Client
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config file")
	dataPath := flag.String("data", "data.yaml", "Path to data file")
	flag.Parse()

	syncer := &Syncer{
		Client: &http.Client{},
	}

	if err := syncer.LoadFiles(*configPath, *dataPath); err != nil {
		log.Fatalf("Failed to load files: %v", err)
	}

	if err := syncer.Sync(); err != nil {
		log.Fatalf("Sync failed: %v", err)
	}

	fmt.Println("\n🚀 Synchronization complete!")
}

func (s *Syncer) LoadFiles(configPath, dataPath string) error {
	// Load Config
	cfgData, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(cfgData, &s.Config); err != nil {
		return err
	}

	// Load Data
	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(dataBytes, &s.Data); err != nil {
		return err
	}

	return nil
}

func (s *Syncer) Sync() error {
	if s.Config.ResetOnStart {
		fmt.Println("⚠️ reset_on_start is true (Cleanup logic not implemented yet)")
	}

	// 1. Sync Protos
	for _, p := range s.Data.Protos {
		if err := s.syncProto(p); err != nil {
			return err
		}
	}

	// 2. Sync Upstreams
	for _, u := range s.Data.Upstreams {
		if err := s.syncUpstream(u); err != nil {
			return err
		}
	}

	// 3. Sync Services
	for _, svc := range s.Data.Services {
		if err := s.syncService(svc); err != nil {
			return err
		}
	}

	// 4. Sync Routes
	for _, r := range s.Data.Routes {
		if err := s.syncRoute(r); err != nil {
			return err
		}
	}

	return nil
}

func (s *Syncer) syncProto(p ProtoDef) error {
	fmt.Printf("Compiling %s... \n", p.Path)

	// Setup compiler
	importPaths := append([]string{"."}, s.Config.Proto.Includes...)
	// Ensure the directory of the proto file itself is in the import path
	absPath, _ := filepath.Abs(p.Path)
	protoDir := filepath.Dir(absPath)
	importPaths = append(importPaths, protoDir)

	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: importPaths,
		}),
	}

	// Filename relative to import paths
	fileName := filepath.Base(p.Path)

	files, err := compiler.Compile(context.Background(), fileName)
	if err != nil {
		fmt.Println("❌ Compilation Error")
		return fmt.Errorf("failed to compile proto %s: %v", p.Path, err)
	}

	// Bundle into FileDescriptorSet
	fds := &descriptorpb.FileDescriptorSet{}
	for _, f := range files {
		fds.File = append(fds.File, protodesc.ToFileDescriptorProto(f))
	}

	pbBytes, err := proto.Marshal(fds)
	if err != nil {
		return fmt.Errorf("failed to marshal descriptor set: %v", err)
	}

	// Base64 encode for APISIX compatibility
	encoded := base64.StdEncoding.EncodeToString(pbBytes)

	payload := map[string]string{
		"content": encoded,
	}
	return s.put("/protos/"+p.ID, payload)
}

func (s *Syncer) syncUpstream(u UpstreamDef) error {
	nodes := make(map[string]int)
	for _, n := range u.Nodes {
		weight := n.Weight
		if weight == 0 {
			weight = 1
		}
		nodes[fmt.Sprintf("%s:%d", n.Host, n.Port)] = weight
	}

	payload := map[string]interface{}{
		"type":  "roundrobin",
		"nodes": nodes,
	}
	return s.put("/upstreams/"+u.ID, payload)
}

func (s *Syncer) syncService(svc ServiceDef) error {
	payload := map[string]interface{}{
		"upstream_id": svc.Upstream,
	}
	return s.put("/services/"+svc.ID, payload)
}

func (s *Syncer) syncRoute(r RouteDef) error {
	// Apply Defaults
	serviceID := r.Service
	if serviceID == "" {
		serviceID = s.Data.RouteDefaults.Service
	}
	protoID := r.Proto
	if protoID == "" {
		protoID = s.Data.RouteDefaults.Proto
	}

	// Methods
	var methods []string
	if r.Methods != "" {
		methods = strings.Split(r.Methods, ",")
	}

	// gRPC Transcode
	parts := strings.Split(r.Grpc, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid grpc field format for route %s: %s (expected package.Service/Method)", r.ID, r.Grpc)
	}
	serviceName, methodName := parts[0], parts[1]

	payload := map[string]interface{}{
		"uri":        r.URI,
		"service_id": serviceID,
		"plugins": map[string]interface{}{
			"grpc-transcode": map[string]interface{}{
				"proto_id": protoID,
				"service":  serviceName,
				"method":   methodName,
			},
		},
	}
	if len(methods) > 0 {
		payload["methods"] = methods
	}

	return s.put("/routes/"+r.ID, payload)
}

func (s *Syncer) put(path string, body interface{}) error {
	url := fmt.Sprintf("%s/apisix/admin%s", strings.TrimSuffix(s.Config.Apisix.URL, "/"), path)
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("X-API-KEY", s.Config.Apisix.Key)
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("PUT %s... ", url)
	resp, err := s.Client.Do(req)
	if err != nil {
		fmt.Println("❌ Error")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("✅ %d\n", resp.StatusCode)
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("❌ %d: %s\n", resp.StatusCode, string(respBody))
	return fmt.Errorf("request failed with status %d", resp.StatusCode)
}
