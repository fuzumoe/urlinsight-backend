package model

import (
	"net/url"
	"time"

	"gorm.io/gorm"
)

const (
	StatusQueued  = "queued"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusError   = "error"
	StatusStopped = "stopped"
)

// URL represents a URL to be analyzed and its processing status.
type URL struct {
	ID              uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint             `gorm:"not null;index" json:"user_id"`
	OriginalURL     string           `gorm:"type:varchar(191);uniqueIndex;not null" json:"original_url"`
	Status          string           `gorm:"type:enum('queued','running','done','error','stopped');default:'queued';not null" json:"status"`
	AnalysisResults []AnalysisResult `gorm:"foreignKey:URLID"`
	Links           []Link           `gorm:"foreignKey:URLID"`
	CreatedAt       time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName returns the name of the table for URL.
func (URL) TableName() string {
	return "urls"
}

// PaginationMetaDTO contains pagination metadata for paginated responses
type PaginationMetaDTO struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// PaginatedResponse is a generic wrapper for any paginated data
type PaginatedResponse[T any] struct {
	Data       []T               `json:"data"`
	Pagination PaginationMetaDTO `json:"pagination"`
}

// URLDTO is the data transfer object for URL.
type URLDTO struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	OriginalURL string    `json:"original_url"`
	Status      string    `json:"status" binding:"omitempty,oneof=queued running done error"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateURLInput defines required fields to create a URL.
type CreateURLInputDTO struct {
	UserID      uint   `json:"user_id" binding:"required"`
	OriginalURL string `json:"original_url" binding:"required,url"`
}
type URLCreateRequestDTO struct {
	OriginalURL string `json:"original_url" binding:"required,url" example:"https://example.com"`
}

type URLResultsDTO struct {
	URL             *URLDTO           `json:"url"`
	AnalysisResults []*AnalysisResult `json:"analysis_results"`
	Links           []*Link           `json:"links"`
}

// AnalysisResult represents a snapshot of the analysis

// ToDTO converts a URL model to a URLDTO.
func (u *URL) ToDTO() *URLDTO {
	return &URLDTO{
		ID:          u.ID,
		UserID:      u.UserID,
		OriginalURL: u.OriginalURL,
		Status:      u.Status,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// FromCreateInput maps CreateURLInput to a URL model.
func URLFromCreateInput(input *CreateURLInputDTO) *URL {
	now := time.Now()
	return &URL{
		UserID:      input.UserID,
		OriginalURL: input.OriginalURL,
		Status:      StatusQueued,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

type UpdateURLInput struct {
	OriginalURL string `json:"original_url" binding:"omitempty,url"`
	Status      string `json:"status"        binding:"omitempty,oneof=queued running done error"`
}

func (u *URL) URL() *url.URL {
	parsed, err := url.Parse(u.OriginalURL)
	if err != nil {
		return nil
	}
	return parsed
}
