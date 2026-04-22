package sync

import (
	"context"
	"encoding/base64"
	"fmt"
	"grpc-apisix-sync/internal/config"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

func (s *Syncer) SyncProto(p config.ProtoDef) error {
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
	return s.Client.Put("/protos/"+p.ID, payload)
}

func (s *Syncer) SyncUpstream(u config.UpstreamDef) error {
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
	return s.Client.Put("/upstreams/"+u.ID, payload)
}

func (s *Syncer) SyncService(svc config.ServiceDef) error {
	payload := map[string]interface{}{
		"upstream_id": svc.Upstream,
	}
	return s.Client.Put("/services/"+svc.ID, payload)
}

func (s *Syncer) SyncRoute(r config.RouteDef) error {
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

	return s.Client.Put("/routes/"+r.ID, payload)
}
