package postgresql

import (
	"context"
	"fmt"
)

func (s *Storage) GetUserId(username string) (string, error) {
	const op = "storage.postgresql.getUserId"

	state := `SELECT id::TEXT FROM employee WHERE username = $1`

	res := s.db.QueryRow(context.Background(), state, username)
	var id string
	err := res.Scan(&id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) ValidateRules(userId, organizationId string) (bool, error) {
	const op = "storage.postgresql.validateRules"

	state := `SELECT EXISTS(SELECT 1 FROM organization_responsible WHERE user_id = $1 AND organization_id = $2)`

	res := s.db.QueryRow(context.Background(), state, userId, organizationId)
	var status bool
	err := res.Scan(&status)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return status, nil
}

func (s *Storage) IsExist(authorId, authorType string) (bool, error) {
	const op = "storage.postgresql.isUserExist"

	var state string

	if authorType == "User" {
		state = `SELECT EXISTS(SELECT 1 FROM employee WHERE id = $1)`
	} else {
		state = `SELECT EXISTS(SELECT 1 FROM organization WHERE id = $1)`
	}

	res := s.db.QueryRow(context.Background(), state, authorId)
	var find bool
	err := res.Scan(&find)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return find, nil
}

func (s *Storage) GetUserOrganization(userId string) (string, error) {
	const op = "storage.postgresql.GetUserOrganization"

	state := `SELECT organization_id FROM organization_responsible WHERE user_id = $1`

	res := s.db.QueryRow(context.Background(), state, userId)
	var id string
	err := res.Scan(&id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) IsHaveBidPerms(userId, bidId string) (bool, error) {
	const op = "storage.postgresql.IsHaveBidPerms"

	state := "SELECT authorId, authorType FROM bid WHERE id::text = $1"
	res := s.db.QueryRow(context.Background(), state, bidId)
	var authorType string
	var authorId string
	err := res.Scan(&authorId, &authorType)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if authorType == "User" && authorId == userId {
		return true, nil
	}
	if authorType == "Organization" {
		orgPerms, err := s.ValidateRules(userId, authorId)
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
		if orgPerms {
			return true, nil
		}
	}
	return false, nil
}

func (s *Storage) IsHaveTenderPerms(userId, tenderId string) (bool, error) {
	const op = "storage.postgresql.IsHaveTenderPerms"

	state := "SELECT organizationid FROM tender WHERE id::text = $1"
	res := s.db.QueryRow(context.Background(), state, tenderId)
	var tenderOrgId string
	err := res.Scan(&tenderOrgId)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	userOrgId, err := s.GetUserOrganization(userId)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return userOrgId == tenderOrgId, nil
}
