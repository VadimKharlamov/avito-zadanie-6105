package DTO

import (
	"time"
)

type BidRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"required,max=500"`
	TenderId    string `json:"tenderId" validate:"required"`
	AuthorType  string `json:"authorType" validate:"required"`
	AuthorId    string `json:"authorId" validate:"required"`
}

type BidPatchRequest struct {
	Name        string `json:"name,max=100"`
	Description string `json:"description,max=500"`
}

type BidResponse struct {
	Id          string    `json:"id"`
	Name        string    `json:"name,max=100"`
	Description string    `json:"description,max=500"`
	Status      string    `json:"status"`
	AuthorType  string    `json:"authorType"`
	AuthorId    string    `json:"authorId"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}
