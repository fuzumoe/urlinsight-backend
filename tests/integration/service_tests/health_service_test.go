package service_test

import (
	"testing"
	"time"

	"github.com/fuzumoe/urlinsight-backend/internal/service"
	"github.com/fuzumoe/urlinsight-backend/tests/utils"
)

func TestHealthServiceIntegration(t *testing.T) {
	// Setup test database.
	db := utils.SetupTest(t)
	defer utils.CleanTestData(t)

	t.Run("LiveHealthy", func(t *testing.T) {
		hs := service.NewHealthService(db, "LiveHealthTest")
		status := hs.Check()

		t.Logf("Health status: %+v", status)
		if status.Database != "healthy" {
			t.Errorf("expected database 'healthy', got %s", status.Database)
		}
		if !status.Healthy {
			t.Errorf("expected Healthy to be true")
		}
		// Ensure the checked timestamp is recent.
		if time.Since(status.Checked) > 5*time.Second {
			t.Errorf("unexpected Checked timestamp: %v", status.Checked)
		}
	})

	t.Run("NilDB", func(t *testing.T) {
		hs := service.NewHealthService(nil, "LiveNilTest")
		status := hs.Check()

		t.Logf("Health status (nil db): %+v", status)
		if status.Database != "disconnected" {
			t.Errorf("expected database 'disconnected', got %s", status.Database)
		}
		if status.Healthy {
			t.Errorf("expected Healthy to be false")
		}
	})
}
