package server_tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fuzumoe/urlinsight-backend/internal/handler"
	"github.com/fuzumoe/urlinsight-backend/internal/server"
	"github.com/fuzumoe/urlinsight-backend/internal/service"
	"github.com/fuzumoe/urlinsight-backend/tests/utils"
)

func TestRouterIntegration(t *testing.T) {
	// Set up test mode
	gin.SetMode(gin.TestMode)

	// Set up test database
	db := utils.SetupTest(t)

	// Create real services
	healthService := service.NewHealthService(db, "IntegrationTest")

	// Create real handlers
	healthHandler := handler.NewHealthHandler(healthService)

	// Create a new router
	r := gin.New()

	// Register routes with real handlers
	server.RegisterRoutes(
		r,
		"test-secret",
		func(c *gin.Context) { c.Next() },      // Dummy auth middleware for testing
		[]server.RouteRegistrar{healthHandler}, // Use real health handler
		[]server.RouteRegistrar{},              // No protected routes for this test
	)

	// Create a test HTTP server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test the API status endpoint (updated from root)
	t.Run("API Status Endpoint", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/status")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "Hello World!", result["message"])
		assert.Equal(t, "IntegrationTest", result["service"])
		assert.Equal(t, "running", result["status"])
	})

	// Test the health endpoint provided by the health handler
	t.Run("Handler Health Endpoint", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "ok", result["status"])
		assert.Equal(t, "healthy", result["database"])
		assert.Equal(t, "IntegrationTest", result["service"])
		assert.Contains(t, result, "checked")
	})

	// Test non-existent route
	t.Run("Route Not Found", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/non-existent")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
