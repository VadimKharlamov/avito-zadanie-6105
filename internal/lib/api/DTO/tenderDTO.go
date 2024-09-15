package DTO

import (
	"time"
)

type TenderRequest struct {
	Name            string `json:"name" validate:"required,max=100"`
	Description     string `json:"description,max=500"`
	ServiceType     string `json:"serviceType" validate:"required"`
	OrganizationId  string `json:"organizationId" validate:"required"`
	CreatorUsername string `json:"creatorUsername" validate:"required"`
}

type TenderPatchRequest struct {
	Name            string `json:"name,max=100"`
	Description     string `json:"description,max=500"`
	ServiceType     string `json:"serviceType"`
	OrganizationId  string `json:"organizationId"`
	CreatorUsername string `json:"creatorUsername"`
}

type TenderResponse struct {
	Id          string    `json:"id"`
	Name        string    `json:"name,max=100"`
	Description string    `json:"description,max=500"`
	Status      string    `json:"status"`
	ServiceType string    `json:"serviceType"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}
