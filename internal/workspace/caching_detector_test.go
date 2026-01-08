package workspace

import (
	"context"
	"testing"
	"time"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
)

func TestNewCachingDetector(t *testing.T) {
	fs := core.NewMockFileSystem()
	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	t.Run("default TTL", func(t *testing.T) {
		cd := NewCachingDetector(detector, 0)
		if cd.ttl != DefaultCacheTTL {
			t.Errorf("expected default TTL %v, got %v", DefaultCacheTTL, cd.ttl)
		}
	})

	t.Run("custom TTL", func(t *testing.T) {
		customTTL := 10 * time.Second
		cd := NewCachingDetector(detector, customTTL)
		if cd.ttl != customTTL {
			t.Errorf("expected custom TTL %v, got %v", customTTL, cd.ttl)
		}
	})
}

func TestCachingDetector_DetectContext(t *testing.T) {
	ctx := context.Background()

	t.Run("caches results", func(t *testing.T) {
		fs := core.NewMockFileSystem()
		_ = fs.WriteFile(ctx, "/project/.version", []byte("1.0.0\n"), 0644)

		cfg := &config.Config{}
		detector := NewDetector(fs, cfg)
		cd := NewCachingDetector(detector, 1*time.Minute)

		// First call - should detect
		result1, err := cd.DetectContext(ctx, "/project")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result1.Mode != SingleModule {
			t.Errorf("expected SingleModule, got %v", result1.Mode)
		}

		// Verify cache is populated
		info := cd.GetCacheInfo()
		if !info.HasCache {
			t.Error("expected cache to be populated")
		}

		// Second call - should return cached result
		result2, err := cd.DetectContext(ctx, "/project")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should be the same pointer (cached)
		if result1 != result2 {
			t.Error("expected cached result to be returned")
		}
	})

	t.Run("invalidates on different root", func(t *testing.T) {
		fs := core.NewMockFileSystem()
		_ = fs.WriteFile(ctx, "/project1/.version", []byte("1.0.0\n"), 0644)
		_ = fs.WriteFile(ctx, "/project2/.version", []byte("2.0.0\n"), 0644)

		cfg := &config.Config{}
		detector := NewDetector(fs, cfg)
		cd := NewCachingDetector(detector, 1*time.Minute)

		// First call for project1
		result1, err := cd.DetectContext(ctx, "/project1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Second call for project2 - should not use cache
		result2, err := cd.DetectContext(ctx, "/project2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result1.Path == result2.Path {
			t.Error("expected different paths for different roots")
		}

		// Cache should now be for project2
		info := cd.GetCacheInfo()
		if info.Root != "/project2" {
			t.Errorf("expected cache root /project2, got %s", info.Root)
		}
	})

	t.Run("expires after TTL", func(t *testing.T) {
		fs := core.NewMockFileSystem()
		_ = fs.WriteFile(ctx, "/project/.version", []byte("1.0.0\n"), 0644)

		cfg := &config.Config{}
		detector := NewDetector(fs, cfg)
		cd := NewCachingDetector(detector, 10*time.Millisecond)

		// First call
		_, err := cd.DetectContext(ctx, "/project")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Wait for TTL to expire
		time.Sleep(20 * time.Millisecond)

		// Cache should be expired
		info := cd.GetCacheInfo()
		if info.ExpiresIn != 0 {
			t.Errorf("expected cache to be expired, expiresIn: %v", info.ExpiresIn)
		}
	})
}

func TestCachingDetector_InvalidateCache(t *testing.T) {
	ctx := context.Background()
	fs := core.NewMockFileSystem()
	_ = fs.WriteFile(ctx, "/project/.version", []byte("1.0.0\n"), 0644)

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)
	cd := NewCachingDetector(detector, 1*time.Minute)

	// Populate cache
	_, err := cd.DetectContext(ctx, "/project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify cache exists
	info := cd.GetCacheInfo()
	if !info.HasCache {
		t.Error("expected cache to be populated")
	}

	// Invalidate
	cd.InvalidateCache()

	// Verify cache is cleared
	info = cd.GetCacheInfo()
	if info.HasCache {
		t.Error("expected cache to be cleared")
	}
}

func TestCachingDetector_DiscoverModules(t *testing.T) {
	ctx := context.Background()
	fs := core.NewMockFileSystem()
	_ = fs.MkdirAll(ctx, "/project/moduleA", 0755)
	_ = fs.MkdirAll(ctx, "/project/moduleB", 0755)
	_ = fs.WriteFile(ctx, "/project/moduleA/.version", []byte("1.0.0\n"), 0644)
	_ = fs.WriteFile(ctx, "/project/moduleB/.version", []byte("2.0.0\n"), 0644)

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)
	cd := NewCachingDetector(detector, 1*time.Minute)

	// DiscoverModules should not use cache
	modules, err := cd.DiscoverModules(ctx, "/project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(modules))
	}

	// Cache should not be populated by DiscoverModules
	info := cd.GetCacheInfo()
	if info.HasCache {
		t.Error("DiscoverModules should not populate cache")
	}
}

func TestCachingDetector_Concurrent(t *testing.T) {
	ctx := context.Background()
	fs := core.NewMockFileSystem()
	_ = fs.WriteFile(ctx, "/project/.version", []byte("1.0.0\n"), 0644)

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)
	cd := NewCachingDetector(detector, 1*time.Minute)

	// Run multiple goroutines concurrently
	done := make(chan bool, 10)
	for range 10 {
		go func() {
			_, err := cd.DetectContext(ctx, "/project")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	// Should have valid cache
	info := cd.GetCacheInfo()
	if !info.HasCache {
		t.Error("expected cache to be populated after concurrent access")
	}
}

func TestCacheInfo(t *testing.T) {
	ctx := context.Background()
	fs := core.NewMockFileSystem()
	_ = fs.WriteFile(ctx, "/project/.version", []byte("1.0.0\n"), 0644)

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)
	cd := NewCachingDetector(detector, 1*time.Second)

	// Before any detection
	info := cd.GetCacheInfo()
	if info.HasCache {
		t.Error("expected no cache initially")
	}

	// After detection
	_, _ = cd.DetectContext(ctx, "/project")
	info = cd.GetCacheInfo()

	if !info.HasCache {
		t.Error("expected cache after detection")
	}
	if info.Root != "/project" {
		t.Errorf("expected root /project, got %s", info.Root)
	}
	if info.Age < 0 || info.Age > 100*time.Millisecond {
		t.Errorf("unexpected age: %v", info.Age)
	}
	if info.ExpiresIn <= 0 || info.ExpiresIn > 1*time.Second {
		t.Errorf("unexpected expiresIn: %v", info.ExpiresIn)
	}
}
