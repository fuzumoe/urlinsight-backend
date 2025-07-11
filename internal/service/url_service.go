package service

import (
	"errors"

	"github.com/fuzumoe/urlinsight-backend/internal/crawler"
	"github.com/fuzumoe/urlinsight-backend/internal/model"
	"github.com/fuzumoe/urlinsight-backend/internal/repository"
)

// URLService defines business operations around URLs.
type URLService interface {
	Create(input *model.CreateURLInput) (uint, error)
	Get(id uint) (*model.URLDTO, error)
	List(userID uint, p repository.Pagination) ([]*model.URLDTO, error)
	Update(id uint, input *model.UpdateURLInput) error
	Delete(id uint) error

	Start(id uint) error
	Stop(id uint) error
	Results(id uint) (*model.URLDTO, error)
}

type urlService struct {
	repo     repository.URLRepository
	crawlers crawler.Pool
}

func (s *urlService) Update(id uint, in *model.UpdateURLInput) error {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if in.OriginalURL != "" {
		u.OriginalURL = in.OriginalURL
	}
	if in.Status != "" {
		switch in.Status {
		case model.StatusQueued, model.StatusRunning,
			model.StatusDone, model.StatusError, model.StatusStopped:
			u.Status = in.Status
		default:
			return errors.New("invalid status value")
		}
	}
	return s.repo.Update(u)
}

// NewURLService constructs a URLService.
func NewURLService(r repository.URLRepository, p crawler.Pool) URLService {
	return &urlService{repo: r, crawlers: p} // ← pass pool
}

// Start: visible to PATCH /urls/:id/start
func (s *urlService) Start(id uint) error {
	if err := s.repo.UpdateStatus(id, model.StatusQueued); err != nil {
		return err
	}
	s.crawlers.Enqueue(id)
	return nil
}

// Stop: flips to "error" status since "stopped" is not in the database schema
func (s *urlService) Stop(id uint) error {
	return s.repo.UpdateStatus(id, model.StatusError)
}

// Results: loads URL with analysis + links eager-loaded
func (s *urlService) Results(id uint) (*model.URLDTO, error) {
	u, err := s.repo.Results(id) // helper added to url_repository.go
	if err != nil {
		return nil, err
	}
	return u.ToDTO(), nil
}

func (s *urlService) Create(input *model.CreateURLInput) (uint, error) {
	u := model.URLFromCreateInput(input)
	if err := s.repo.Create(u); err != nil {
		return 0, err
	}
	return u.ID, nil
}

func (s *urlService) Get(id uint) (*model.URLDTO, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return u.ToDTO(), nil
}

func (s *urlService) List(userID uint, p repository.Pagination) ([]*model.URLDTO, error) {
	urls, err := s.repo.ListByUser(userID, p)
	if err != nil {
		return nil, err
	}
	dtos := make([]*model.URLDTO, len(urls))
	for i, u := range urls {
		dtos[i] = u.ToDTO()
	}
	return dtos, nil
}

func (s *urlService) Delete(id uint) error {
	return s.repo.Delete(id)
}
