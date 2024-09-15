package postgresql

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	tenderOpt "zadanie-6105/internal/lib/api/DTO"
)

func (s *Storage) CreateTender(tenderData tenderOpt.TenderRequest) (tenderOpt.TenderResponse, error) {
	const op = "storage.postgresql.createTender"

	state := `INSERT INTO tender(name, description, servicetype, organizationid, creatorusername) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, description, servicetype, status, version, createdat;`

	res := s.db.QueryRow(context.Background(), state, tenderData.Name, tenderData.Description, tenderData.ServiceType, tenderData.OrganizationId, tenderData.CreatorUsername)
	resp := tenderOpt.TenderResponse{}
	err := res.Scan(&resp.Id, &resp.Name, &resp.Description, &resp.ServiceType,
		&resp.Status, &resp.Version, &resp.CreatedAt)
	if err != nil {
		return tenderOpt.TenderResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (s *Storage) GetAllTenders(limit string, offset string, filters []string) ([]tenderOpt.TenderResponse, error) {
	const op = "storage.postgresql.GetAllTenders"

	args := make([]interface{}, len(filters)+2)
	args[0] = limit
	args[1] = offset
	var state string
	if len(filters) > 0 {
		args[2] = pq.Array(filters)
		state = "SELECT id, name, description, status, servicetype, createdat, version FROM tender WHERE servicetype = ANY($3) AND status = 'Published' LIMIT $1 OFFSET $2"
	} else {
		state = "SELECT id, name, description, status, servicetype, createdat, version FROM tender WHERE status = 'Published' LIMIT $1 OFFSET $2"
	}

	res, err := s.db.Query(context.Background(), state, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	resp := make([]tenderOpt.TenderResponse, 0)
	for res.Next() {
		tender := tenderOpt.TenderResponse{}
		err = res.Scan(&tender.Id, &tender.Name, &tender.Description,
			&tender.Status, &tender.ServiceType, &tender.CreatedAt, &tender.Version)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		resp = append(resp, tender)
	}
	return resp, nil
}

func (s *Storage) GetPersonalTenders(limit string, offset string, username string) ([]tenderOpt.TenderResponse, error) {
	const op = "storage.postgresql.GetPersonalTenders"

	state := "SELECT id, name, description, status, servicetype, createdat, version FROM tender WHERE creatorusername = $3 LIMIT $1 OFFSET $2"

	res, err := s.db.Query(context.Background(), state, limit, offset, username)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	resp := make([]tenderOpt.TenderResponse, 0)
	for res.Next() {
		tender := tenderOpt.TenderResponse{}
		err = res.Scan(&tender.Id, &tender.Name, &tender.Description,
			&tender.Status, &tender.ServiceType, &tender.CreatedAt, &tender.Version)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		resp = append(resp, tender)
	}
	return resp, nil
}

func (s *Storage) GetTenderStatus(tenderId string) (string, error) {
	const op = "storage.postgresql.GetTenderStatus"

	state := "SELECT status FROM tender WHERE id::text = $1"

	res := s.db.QueryRow(context.Background(), state, tenderId)
	var status string
	err := res.Scan(&status)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return status, nil
}

func (s *Storage) EditTenderStatus(tenderId, status, username string) (tenderOpt.TenderResponse, error) {
	const op = "storage.postgresql.EditTenderStatus"

	state := "UPDATE tender SET status = $1 WHERE id::text = $2 AND creatorusername = $3 RETURNING id, name, description, status, servicetype, createdat, version"

	res := s.db.QueryRow(context.Background(), state, status, tenderId, username)
	resp := tenderOpt.TenderResponse{}
	err := res.Scan(&resp.Id, &resp.Name, &resp.Description,
		&resp.Status, &resp.ServiceType, &resp.CreatedAt, &resp.Version)
	if err != nil {
		return tenderOpt.TenderResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (s *Storage) EditTender(tenderData tenderOpt.TenderPatchRequest, tenderId, username string) (tenderOpt.TenderResponse, error) {
	const op = "storage.postgresql.EditTender"

	state := "SELECT id, name, description, status, servicetype, createdat, version FROM tender WHERE creatorusername = $2 AND id::TEXT = $1"

	res := s.db.QueryRow(context.Background(), state, tenderId, username)
	resp := tenderOpt.TenderResponse{}
	err := res.Scan(&resp.Id, &resp.Name, &resp.Description,
		&resp.Status, &resp.ServiceType, &resp.CreatedAt, &resp.Version)
	if err != nil {
		return tenderOpt.TenderResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	state = `INSERT INTO tenderrollback(tender_id, name, description, status, servicetype, version) VALUES ($1, $2, $3, $4, $5, $6);`
	res = s.db.QueryRow(context.Background(), state, resp.Id, resp.Name, resp.Description, resp.Status,
		resp.ServiceType, resp.Version)
	err = res.Scan()

	if tenderData.Name != "" {
		resp.Name = tenderData.Name
	}
	if tenderData.Description != "" {
		resp.Description = tenderData.Description
	}
	if tenderData.ServiceType != "" {
		resp.ServiceType = tenderData.ServiceType
	}
	resp.Version++

	state = "UPDATE tender SET name = $1, description = $2,servicetype = $3, version = $4 WHERE id::text = $5 AND creatorusername = $6"

	res = s.db.QueryRow(context.Background(), state, resp.Name, resp.Description, resp.ServiceType, resp.Version, tenderId, username)

	err = res.Scan()

	return resp, nil
}

func (s *Storage) GetTenderOrganizationId(tenderId string) (string, error) {
	const op = "storage.postgresql.GetTenderOrganizationId"

	state := "SELECT organizationid FROM tender WHERE id::text = $1"

	res := s.db.QueryRow(context.Background(), state, tenderId)
	var id string
	err := res.Scan(&id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) IsTenderPublic(tenderId string) (bool, error) {
	const op = "storage.postgresql.isTenderPublic"

	state := "SELECT status FROM tender WHERE id::text = $1"

	res := s.db.QueryRow(context.Background(), state, tenderId)
	var status string
	err := res.Scan(&status)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if status == "Published" {
		return true, nil
	}
	return false, nil
}

func (s *Storage) GetTenderRollback(tenderId, version, username string) (tenderOpt.TenderResponse, error) {
	const op = "storage.postgresql.GetTenderRollback"

	state := "SELECT name, description, servicetype FROM tenderrollback WHERE tender_id = $1 AND version::TEXT = $2"

	res := s.db.QueryRow(context.Background(), state, tenderId, version)
	resp := tenderOpt.TenderPatchRequest{}
	err := res.Scan(&resp.Name, &resp.Description, &resp.ServiceType)
	if err != nil {
		return tenderOpt.TenderResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	newTender, err := s.EditTender(resp, tenderId, username)
	if err != nil {
		return tenderOpt.TenderResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return newTender, nil
}
