package changeloggenerator

import (
	"testing"
)

func TestResolveRemote_FromConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "gitlab",
		Host:     "gitlab.com",
		Owner:    "mygroup",
		Repo:     "myproject",
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remote, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if remote.Provider != "gitlab" {
		t.Errorf("Provider = %q, want 'gitlab'", remote.Provider)
	}
	if remote.Host != "gitlab.com" {
		t.Errorf("Host = %q, want 'gitlab.com'", remote.Host)
	}
	if remote.Owner != "mygroup" {
		t.Errorf("Owner = %q, want 'mygroup'", remote.Owner)
	}
}

func TestResolveRemote_FillDefaults(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Owner:    "owner",
		Repo:     "repo",
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remote, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Host should be filled from provider
	if remote.Host != "github.com" {
		t.Errorf("Host = %q, want 'github.com'", remote.Host)
	}
}

func TestResolveRemote_AutoDetect(t *testing.T) {
	// Save and restore original function
	originalFn := GetRemoteInfoFn
	defer func() { GetRemoteInfoFn = originalFn }()

	// Mock GetRemoteInfoFn
	GetRemoteInfoFn = func() (*RemoteInfo, error) {
		return &RemoteInfo{
			Provider: "github",
			Host:     "github.com",
			Owner:    "autodetected",
			Repo:     "repo",
		}, nil
	}

	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		AutoDetect: true,
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remote, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if remote.Owner != "autodetected" {
		t.Errorf("Owner = %q, want 'autodetected'", remote.Owner)
	}
}

func TestResolveRemote_NoConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = nil
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = g.resolveRemote()
	if err == nil {
		t.Error("expected error when repository config is nil")
	}
}

func TestResolveRemote_Cached(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First call
	remote1, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call should return cached
	remote2, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if remote1 != remote2 {
		t.Error("expected cached remote to be returned")
	}
}

func TestResolveRemote_FillProviderFromHost(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Host:  "github.com",
		Owner: "owner",
		Repo:  "repo",
		// Provider not set - should be filled from host
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remote, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Provider should be filled from host
	if remote.Provider != "github" {
		t.Errorf("Provider = %q, want 'github'", remote.Provider)
	}
}
