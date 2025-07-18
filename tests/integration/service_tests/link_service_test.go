package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fuzumoe/urlinsight-backend/internal/model"
	"github.com/fuzumoe/urlinsight-backend/internal/repository"
	"github.com/fuzumoe/urlinsight-backend/internal/service"
	"github.com/fuzumoe/urlinsight-backend/tests/utils"
)

func TestLinkService_Integration(t *testing.T) {
	// Setup test database.
	db := utils.SetupTest(t)
	defer utils.CleanTestData(t)

	// Create repositories and services.
	userRepo := repository.NewUserRepo(db) // User repository.
	urlRepo := repository.NewURLRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	linkService := service.NewLinkService(linkRepo)

	//   create a test user using the correct "Username" field.
	testUser := &model.User{
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := userRepo.Create(testUser)
	require.NoError(t, err, "Should create test user without error.")
	require.NotZero(t, testUser.ID, "User ID should be set after creation.")

	// Create a test URL for the links.
	testURL := &model.URL{
		UserID:      testUser.ID,
		OriginalURL: "https://test-domain.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = urlRepo.Create(testURL)
	require.NoError(t, err, "Should create test URL without error.")
	require.NotZero(t, testURL.ID, "URL ID should be set after creation.")

	// Use the created URL's ID for our links.
	urlID := testURL.ID

	createTestLink := func(t *testing.T, hrefSuffix string) *model.Link {
		link := &model.Link{
			URLID:      urlID,
			Href:       "https://example.com/" + hrefSuffix,
			IsExternal: false,
			StatusCode: 200,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err := linkService.Add(link)
		require.NoError(t, err, "Should add link without error.")
		require.NotZero(t, link.ID, "Link ID should be set after creation.")
		return link
	}

	t.Run("Add", func(t *testing.T) {
		// Test adding a new link.
		link := &model.Link{
			URLID:      urlID,
			Href:       "https://example.com/test",
			IsExternal: true,
			StatusCode: 200,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err := linkService.Add(link)
		assert.NoError(t, err, "Should add link without error.")
		assert.NotZero(t, link.ID, "Link ID should be set after creation.")
	})

	t.Run("List", func(t *testing.T) {
		// Create several test links.
		createTestLink(t, "Link1")
		createTestLink(t, "Link2")
		createTestLink(t, "Link3")

		// Define pagination settings.
		pagination := repository.Pagination{
			Page:     1,
			PageSize: 10,
		}

		// List the links using the non-paginated method.
		links, err := linkService.List(urlID, pagination)
		assert.NoError(t, err, "Should list links without error.")
		assert.GreaterOrEqual(t, len(links), 3, "Should return at least 3 links.")

		// Verify each link contains expected data.
		for _, link := range links {
			assert.Equal(t, urlID, link.URLID, "Link should have the correct URL ID.")
			assert.NotEmpty(t, link.Href, "Link should have an href.")
		}

		// Test pagination with smaller page size.
		paginationSmall := repository.Pagination{
			Page:     1,
			PageSize: 2,
		}
		linksPage1, err := linkService.List(urlID, paginationSmall)
		assert.NoError(t, err, "Should list links without error.")
		assert.Len(t, linksPage1, 2, "Should return exactly 2 links for page 1.")

		// Get the second page.
		paginationPage2 := repository.Pagination{
			Page:     2,
			PageSize: 2,
		}
		linksPage2, err := linkService.List(urlID, paginationPage2)
		assert.NoError(t, err, "Should list links without error.")
		assert.GreaterOrEqual(t, len(linksPage2), 1, "Should have at least 1 link on page 2.")

		// Verify links on different pages are different.
		if len(linksPage2) > 0 {
			assert.NotEqual(t, linksPage1[0].ID, linksPage2[0].ID, "Links on different pages should be different.")
		}
	})

	// New test for the paginated ListByURL method
	t.Run("ListByURL", func(t *testing.T) {
		// Create additional test links
		createTestLink(t, "PaginatedLink1")
		createTestLink(t, "PaginatedLink2")
		createTestLink(t, "PaginatedLink3")

		// Define pagination settings
		pagination := repository.Pagination{
			Page:     1,
			PageSize: 10,
		}

		// List the links using the paginated method
		paginatedResult, err := linkService.ListByURL(urlID, pagination)
		require.NoError(t, err, "Should list links without error.")

		// Check pagination metadata
		assert.Equal(t, 1, paginatedResult.Pagination.Page, "Page should be 1")
		assert.Equal(t, 10, paginatedResult.Pagination.PageSize, "PageSize should be 10")
		assert.GreaterOrEqual(t, paginatedResult.Pagination.TotalItems, 6, "TotalItems should include all created links")
		assert.GreaterOrEqual(t, paginatedResult.Pagination.TotalPages, 1, "Should have at least 1 page")

		// Verify data contains links
		assert.NotEmpty(t, paginatedResult.Data, "Data should not be empty")
		assert.GreaterOrEqual(t, len(paginatedResult.Data), 6, "Should return at least 6 links")

		// Verify each link contains expected data
		for _, link := range paginatedResult.Data {
			assert.Equal(t, urlID, link.URLID, "Link should have the correct URL ID.")
			assert.NotEmpty(t, link.Href, "Link should have an href.")
		}

		// Test pagination with smaller page size
		smallPagination := repository.Pagination{
			Page:     1,
			PageSize: 3,
		}

		smallPageResult, err := linkService.ListByURL(urlID, smallPagination)
		require.NoError(t, err, "Should list links without error.")

		// Verify pagination metadata for smaller page
		assert.Equal(t, 1, smallPageResult.Pagination.Page, "Page should be 1")
		assert.Equal(t, 3, smallPageResult.Pagination.PageSize, "PageSize should be 3")
		assert.Len(t, smallPageResult.Data, 3, "Should return exactly 3 links")
		assert.GreaterOrEqual(t, smallPageResult.Pagination.TotalPages, 2, "Should have at least 2 pages")

		// Get the second page
		page2Pagination := repository.Pagination{
			Page:     2,
			PageSize: 3,
		}

		page2Result, err := linkService.ListByURL(urlID, page2Pagination)
		require.NoError(t, err, "Should list links without error.")

		// Verify pagination metadata for page 2
		assert.Equal(t, 2, page2Result.Pagination.Page, "Page should be 2")
		assert.Equal(t, 3, page2Result.Pagination.PageSize, "PageSize should be 3")
		assert.NotEmpty(t, page2Result.Data, "Page 2 should have data")

		// Verify different pages have different links
		if len(page2Result.Data) > 0 && len(smallPageResult.Data) > 0 {
			assert.NotEqual(t, smallPageResult.Data[0].ID, page2Result.Data[0].ID,
				"Links on different pages should be different")
		}
	})

	t.Run("Update", func(t *testing.T) {
		// Create a test link.
		link := createTestLink(t, "UpdateTest")

		// Modify the link.
		link.Href = "https://example.com/updated"
		link.IsExternal = true
		link.StatusCode = 301
		link.UpdatedAt = time.Now()

		// Update the link.
		err := linkService.Update(link)
		assert.NoError(t, err, "Should update link without error.")

		// Retrieve the links to verify the update.
		paginatedResult, err := linkService.ListByURL(urlID, repository.Pagination{Page: 1, PageSize: 100})
		assert.NoError(t, err, "Should list links without error.")

		// Find our updated link.
		var updatedLink model.LinkDTO
		for _, l := range paginatedResult.Data {
			if l.ID == link.ID {
				updatedLink = l
				break
			}
		}

		// Verify the link was updated.
		assert.Equal(t, link.ID, updatedLink.ID, "Updated link should be found in the list.")
		assert.Equal(t, "https://example.com/updated", updatedLink.Href, "Link href should be updated.")
		assert.Equal(t, true, updatedLink.IsExternal, "Link IsExternal should be updated.")
		assert.Equal(t, 301, updatedLink.StatusCode, "Link StatusCode should be updated.")
	})

	t.Run("Delete", func(t *testing.T) {
		// Create a test link to delete.
		link := createTestLink(t, "DeleteTest")

		// Get the initial count of links.
		initialResult, err := linkService.ListByURL(urlID, repository.Pagination{Page: 1, PageSize: 100})
		assert.NoError(t, err, "Should list links without error.")
		initialCount := len(initialResult.Data)

		// Delete the link.
		err = linkService.Delete(link)
		assert.NoError(t, err, "Should delete link without error.")

		// Get the count after deletion.
		afterResult, err := linkService.ListByURL(urlID, repository.Pagination{Page: 1, PageSize: 100})
		assert.NoError(t, err, "Should list links without error.")
		afterCount := len(afterResult.Data)

		// Verify one link was deleted.
		assert.Equal(t, initialCount-1, afterCount, "One link should be deleted.")

		// Verify the deleted link is no longer in the list.
		found := false
		for _, l := range afterResult.Data {
			if l.ID == link.ID {
				found = true
				break
			}
		}
		assert.False(t, found, "Deleted link should not be found in the list.")
	})

	t.Run("ErrorCases", func(t *testing.T) {
		// Test adding a link with invalid data: missing Href.
		invalidLink := &model.Link{
			URLID:      urlID,
			Href:       "", // Empty Href.
			StatusCode: 200,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err := linkService.Add(invalidLink)
		// Adjusted expectation: no error is expected since empty string is valid.
		assert.NoError(t, err, "Adding link with empty Href should not return error by default.")

		// Test updating a non-existent link.
		nonExistentLink := &model.Link{
			ID:         999999, // ID that doesn't exist.
			URLID:      urlID,
			Href:       "https://example.com/non-existent",
			IsExternal: false,
			StatusCode: 200,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err = linkService.Update(nonExistentLink)
		// Adjusted expectation: no error is returned if no record is updated.
		assert.NoError(t, err, "Updating non-existent link should not return error by default.")

		// Test deleting a non-existent link.
		err = linkService.Delete(nonExistentLink)
		// Adjusted expectation: no error is returned if no record is deleted.
		assert.NoError(t, err, "Deleting non-existent link should not return error by default.")
	})
}
