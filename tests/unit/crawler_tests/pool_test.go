package crawler_test

import (
	"context"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fuzumoe/urlinsight-backend/internal/crawler"
	"github.com/fuzumoe/urlinsight-backend/internal/model"
	"github.com/fuzumoe/urlinsight-backend/internal/repository"
)

// mockPRepo implements repository.URLRepository for testing.
type mockPRepo struct {
	mu                sync.Mutex
	statusUpdates     map[uint][]string
	findByIDCalls     []uint
	saveResultsCalled bool
}

func newMockPRepo() *mockPRepo {
	return &mockPRepo{
		statusUpdates: make(map[uint][]string),
	}
}

func (r *mockPRepo) UpdateStatus(id uint, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statusUpdates[id] = append(r.statusUpdates[id], status)
	return nil
}

func (r *mockPRepo) FindByID(id uint) (*model.URL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findByIDCalls = append(r.findByIDCalls, id)
	return &model.URL{
		OriginalURL: "http://example.com",
	}, nil
}

func (r *mockPRepo) SaveResults(id uint, res *model.AnalysisResult, links []model.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saveResultsCalled = true
	return nil
}

// Implement Create with the correct signature.
func (r *mockPRepo) Create(u *model.URL) error {
	return nil
}

// Delete method.
func (r *mockPRepo) Delete(id uint) error {
	return nil
}

// ListByUser with the correct signature.
func (r *mockPRepo) ListByUser(userID uint, p repository.Pagination) ([]model.URL, error) {
	return []model.URL{}, nil
}

// Results returns a dummy URL.
func (r *mockPRepo) Results(id uint) (*model.URL, error) {
	return &model.URL{
		OriginalURL: "http://example.com",
	}, nil
}

// Update method.
func (r *mockPRepo) Update(u *model.URL) error {
	return nil
}

// mockPAnalyzer implements analyzer.Analyzer for testing.
type mockPAnalyzer struct{}

func (a *mockPAnalyzer) Analyze(ctx context.Context, u *url.URL) (*model.AnalysisResult, []model.Link, error) {

	result := &model.AnalysisResult{
		HTMLVersion: "HTML 5",
		Title:       "Test Page",
	}
	links := []model.Link{
		{Href: "http://example.com/page1", StatusCode: 200},
		{Href: "http://example.com/page2", StatusCode: 404},
	}
	return result, links, nil
}

func TestPool_ProcessTasks(t *testing.T) {
	// Create a pool with the mock repository and analyzer.
	mockPRepo := newMockPRepo()
	mockAnal := &mockPAnalyzer{}

	// Create a pool with 2 workers and a buffer size of 10.
	pool := crawler.New(mockPRepo, mockAnal, 2, 10)

	t.Run("Start and Enqueue Tasks", func(t *testing.T) {
		pool.Start()

		// Enqueue several tasks.
		taskIDs := []uint{1, 2, 3}
		for _, id := range taskIDs {
			pool.Enqueue(id)
		}

		// Allow some time for tasks to be picked up.
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Shutdown Pool", func(t *testing.T) {
		// Shutdown the pool.
		pool.Shutdown()
	})

	// Verify that tasks were processed.
	t.Run("Verify Task Processing", func(t *testing.T) {
		// Check that for each task enqueued, UpdateStatus was called.
		mockPRepo.mu.Lock()
		defer mockPRepo.mu.Unlock()
		for _, id := range []uint{1, 2, 3} {
			statuses, ok := mockPRepo.statusUpdates[id]
			require.True(t, ok, "Expected task id %d to have status updates", id)
			// Assuming worker sets at least two statuses: one for "running" and one for "done".
			require.GreaterOrEqual(t, len(statuses), 2, "Expected at least two status updates for task id %d", id)
		}

		// Check that FindByID was called for each task.
		for _, id := range []uint{1, 2, 3} {
			assert.Contains(t, mockPRepo.findByIDCalls, id, "Expected FindByID to be called for task id %d", id)
		}

		// Verify SaveResults was called at least once.
		assert.True(t, mockPRepo.saveResultsCalled, "Expected SaveResults to be called")
	})
}
