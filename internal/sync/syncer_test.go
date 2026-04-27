package sync

import (
	"github.com/prasojoam/grpc-apisix-sync/internal/config"
	"testing"
)

func TestQualifyID(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		id       string
		expected string
	}{
		{
			name:     "with prefix",
			prefix:   "user_service",
			id:       "upstream",
			expected: "user_service.upstream",
		},
		{
			name:     "without prefix",
			prefix:   "",
			id:       "upstream",
			expected: "upstream",
		},
		{
			name:     "empty id with prefix",
			prefix:   "user_service",
			id:       "",
			expected: "",
		},
		{
			name:     "empty id without prefix",
			prefix:   "",
			id:       "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Syncer{
				Config: &config.Config{
					IdPrefix: tt.prefix,
				},
			}
			result := s.qualifyID(tt.id)
			if result != tt.expected {
				t.Errorf("qualifyID(%s) = %s; want %s", tt.id, result, tt.expected)
			}
		})
	}
}
