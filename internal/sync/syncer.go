package sync

import (
	"fmt"
	"github.com/prasojoam/grpc-apisix-sync/internal/apisix"
	"github.com/prasojoam/grpc-apisix-sync/internal/config"
)

type Syncer struct {
	Config *config.Config
	Data   *config.Data
	Client *apisix.Client
}

func NewSyncer(cfg *config.Config, data *config.Data, client *apisix.Client) *Syncer {
	return &Syncer{
		Config: cfg,
		Data:   data,
		Client: client,
	}
}

func (s *Syncer) qualifyID(id string) string {
	if s.Config.IdPrefix == "" || id == "" {
		return id
	}
	return s.Config.IdPrefix + "." + id
}

func (s *Syncer) Sync() error {
	if s.Config.ResetOnStart {
		if err := s.Cleanup(); err != nil {
			return fmt.Errorf("cleanup failed: %v", err)
		}
	}

	// 1. Sync Protos
	for _, p := range s.Data.Protos {
		if err := s.SyncProto(p); err != nil {
			return err
		}
	}

	// 2. Sync Upstreams
	for _, u := range s.Data.Upstreams {
		if err := s.SyncUpstream(u); err != nil {
			return err
		}
	}

	// 3. Sync Services
	for _, svc := range s.Data.Services {
		if err := s.SyncService(svc); err != nil {
			return err
		}
	}

	// 4. Sync Routes
	for _, r := range s.Data.Routes {
		if err := s.SyncRoute(r); err != nil {
			return err
		}
	}

	return nil
}

func (s *Syncer) Cleanup() error {
	fmt.Println("🧹 Starting cleanup...")

	// Order matters to avoid dependency issues
	types := []string{"routes", "services", "upstreams", "protos"}
	prefix := ""
	if s.Config.IdPrefix != "" {
		prefix = s.Config.IdPrefix + "."
	}

	for _, t := range types {
		if err := s.Client.DeleteAll(t, prefix); err != nil {
			return err
		}
	}
	fmt.Println("✨ Cleanup finished.")
	return nil
}
