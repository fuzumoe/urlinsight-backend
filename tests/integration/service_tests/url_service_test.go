package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fuzumoe/urlinsight-backend/internal/analyzer"
	"github.com/fuzumoe/urlinsight-backend/internal/crawler"
	"github.com/fuzumoe/urlinsight-backend/internal/model"
	"github.com/fuzumoe/urlinsight-backend/internal/repository"
	"github.com/fuzumoe/urlinsight-backend/internal/service"
	"github.com/fuzumoe/urlinsight-backend/tests/utils"
)

func TestURLService_Integration(t *testing.T) {
	// Setup test database.
	db := utils.SetupTest(t)
	defer utils.CleanTestData(t)

	// Create repositories.
	userRepo := repository.NewUserRepo(db)
	urlRepo := repository.NewURLRepo(db)

	// Create a mock analyzer for testing
	htmlAnalyzer := analyzer.NewHTMLAnalyzer()

	// Create a crawler pool with 1 worker for testing
	crawlerPool := crawler.New(urlRepo, htmlAnalyzer, 1, 5)
	go crawlerPool.Start()
	defer crawlerPool.Shutdown()

	// Create URLService with the crawler pool
	urlService := service.NewURLService(urlRepo, crawlerPool)

	// Create a test user.
	testUser := &model.User{
		Username:  "testuser", // using the field from the user model.
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := userRepo.Create(testUser)
	require.NoError(t, err, "Should create test user without error.")
	require.NotZero(t, testUser.ID, "User ID should be set after creation.")

	t.Run("Create and Get", func(t *testing.T) {
		// Create a URL through URLService.
		createInput := &model.CreateURLInput{
			UserID:      testUser.ID,
			OriginalURL: "https://example.com",
		}
		createdID, err := urlService.Create(createInput)
		require.NoError(t, err, "Should create URL without error.")
		require.NotZero(t, createdID, "Created URL ID should be set.")

		// Get the URL back.
		urlDTO, err := urlService.Get(createdID)
		require.NoError(t, err, "Should get URL without error.")
		assert.Equal(t, "https://example.com", urlDTO.OriginalURL, "OriginalURL should match the input.")
	})

	t.Run("List", func(t *testing.T) {
		// Create several URLs for the test user.
		urlsToCreate := []string{
			"https://example.com/1",
			"https://example.com/2",
			"https://example.com/3",
		}
		for _, orig := range urlsToCreate {
			input := &model.CreateURLInput{
				UserID:      testUser.ID,
				OriginalURL: orig,
			}
			_, err := urlService.Create(input)
			require.NoError(t, err, "Should create URL without error.")
		}

		// List URLs with pagination.
		pagination := repository.Pagination{
			Page:     1,
			PageSize: 10,
		}
		urlList, err := urlService.List(testUser.ID, pagination)
		require.NoError(t, err, "Should list URLs without error.")
		assert.GreaterOrEqual(t, len(urlList), 3, "Should return at least 3 URLs.")
	})

	t.Run("Update", func(t *testing.T) {
		// Create a URL to update.
		createInput := &model.CreateURLInput{
			UserID:      testUser.ID,
			OriginalURL: "https://example.com/old",
		}
		createdID, err := urlService.Create(createInput)
		require.NoError(t, err, "Should create URL without error.")

		// Update the URL.
		updateInput := &model.UpdateURLInput{
			OriginalURL: "https://example.com/new",
			Status:      model.StatusRunning, // allowed status value.
		}
		err = urlService.Update(createdID, updateInput)
		require.NoError(t, err, "Should update URL without error.")

		// Verify the update.
		updatedDTO, err := urlService.Get(createdID)
		require.NoError(t, err, "Should get URL without error.")
		assert.Equal(t, "https://example.com/new", updatedDTO.OriginalURL, "OriginalURL should be updated.")
		assert.Equal(t, model.StatusRunning, updatedDTO.Status, "Status should be updated to running.")
	})

	t.Run("Delete", func(t *testing.T) {
		// Create a URL to delete.
		createInput := &model.CreateURLInput{
			UserID:      testUser.ID,
			OriginalURL: "https://example.com/delete",
		}
		createdID, err := urlService.Create(createInput)
		require.NoError(t, err, "Should create URL without error.")
		require.NotZero(t, createdID, "Created URL ID should be set.")

		// Delete the URL.
		err = urlService.Delete(createdID)
		require.NoError(t, err, "Should delete URL without error.")

		// Attempt to get the deleted URL.
		_, err = urlService.Get(createdID)
		assert.Error(t, err, "Getting a deleted URL should return an error.")
	})

	t.Run("Start", func(t *testing.T) {
		// Create a URL to start crawling
		createInput := &model.CreateURLInput{
			UserID:      testUser.ID,
			OriginalURL: "https://example.com/start",
		}
		createdID, err := urlService.Create(createInput)
		require.NoError(t, err, "Should create URL without error.")

		// Start crawling the URL
		err = urlService.Start(createdID)
		require.NoError(t, err, "Should start crawling without error.")

		// Verify the URL status is updated to queued
		urlDTO, err := urlService.Get(createdID)
		require.NoError(t, err, "Should get URL without error.")
		assert.Equal(t, model.StatusQueued, urlDTO.Status, "Status should be queued after starting.")
	})

	t.Run("Stop", func(t *testing.T) {
		// Create a URL to stop crawling
		createInput := &model.CreateURLInput{
			UserID:      testUser.ID,
			OriginalURL: "https://example.com/stop",
		}
		createdID, err := urlService.Create(createInput)
		require.NoError(t, err, "Should create URL without error.")

		// First set status to running
		updateInput := &model.UpdateURLInput{
			Status: model.StatusRunning,
		}
		err = urlService.Update(createdID, updateInput)
		require.NoError(t, err, "Should update URL status to running without error.")

		// Now stop crawling the URL
		err = urlService.Stop(createdID)
		require.NoError(t, err, "Should stop crawling without error.")

		// Verify the URL status is updated to error (not stopped)
		urlDTO, err := urlService.Get(createdID)
		require.NoError(t, err, "Should get URL without error.")

		// Use 'error' status instead of 'stopped' since that's what's allowed in the DB
		assert.Equal(t, model.StatusError, urlDTO.Status,
			"Status should be 'error' after stopping (since 'stopped' is not allowed in the DB).")
	})

	t.Run("ErrorCases", func(t *testing.T) {
		// Create a URL first.
		createInput := &model.CreateURLInput{
			UserID:      testUser.ID,
			OriginalURL: "https://example.com/error",
		}
		createdID, err := urlService.Create(createInput)
		require.NoError(t, err, "Should create URL without error.")

		// Try to update with an invalid status value.
		updateInput := &model.UpdateURLInput{
			OriginalURL: "https://example.com/error-updated",
			Status:      "invalid_status",
		}
		err = urlService.Update(createdID, updateInput)
		assert.Error(t, err, "Updating with an invalid status should return an error.")

		// Try to start a URL that doesn't exist
		err = urlService.Start(9999)
		assert.Error(t, err, "Starting a non-existent URL should return an error.")

		// Try to stop a URL that doesn't exist
		err = urlService.Stop(9999)
		assert.Error(t, err, "Stopping a non-existent URL should return an error.")
	})
}
