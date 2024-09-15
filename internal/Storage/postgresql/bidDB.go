package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	bidOpt "zadanie-6105/internal/lib/api/DTO"
)

func (s *Storage) CreateBid(bidData bidOpt.BidRequest) (bidOpt.BidResponse, error) {
	const op = "storage.postgresql.createBid"

	state := `INSERT INTO bid(name, description, tenderId, authorType, authorId) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, description, status, authorType, authorId, version, createdat;`

	res := s.db.QueryRow(context.Background(), state, bidData.Name, bidData.Description, bidData.TenderId,
		bidData.AuthorType, bidData.AuthorId)
	resp := bidOpt.BidResponse{}
	err := res.Scan(&resp.Id, &resp.Name, &resp.Description,
		&resp.Status, &resp.AuthorType, &resp.AuthorId, &resp.Version, &resp.CreatedAt)
	if err != nil {
		return bidOpt.BidResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (s *Storage) GetPersonalBids(limit string, offset string, userId string) ([]bidOpt.BidResponse, error) {
	const op = "storage.postgresql.GetPersonalBids"

	organizationId, err := s.GetUserOrganization(userId)

	var state string
	var res pgx.Rows
	if err != nil {
		state = "SELECT id, name, description, status, authortype, authorid, version, createdat FROM bid WHERE (authorid = $1 AND authortype = 'User') LIMIT $2 OFFSET $3"
		res, err = s.db.Query(context.Background(), state, userId, limit, offset)
	} else {
		state = "SELECT id, name, description, status, authortype, authorid, version, createdat FROM bid WHERE (authorid = $1 AND authortype = 'User') OR (authorid = $2 AND authortype = 'Organization') LIMIT $3 OFFSET $4"
		res, err = s.db.Query(context.Background(), state, userId, organizationId, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	resp := make([]bidOpt.BidResponse, 0)
	for res.Next() {
		bid := bidOpt.BidResponse{}
		err = res.Scan(&bid.Id, &bid.Name, &bid.Description,
			&bid.Status, &bid.AuthorType, &bid.AuthorId, &bid.Version, &bid.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		resp = append(resp, bid)
	}

	return resp, nil
}

func (s *Storage) GetBidList(limit, offset, tenderId string) ([]bidOpt.BidResponse, error) {
	const op = "storage.postgresql.GetBidList"

	state := "SELECT id, name, description, status, authortype, authorid, version, createdat FROM bid WHERE tenderId = $1 AND status = 'Published' LIMIT $2 OFFSET $3"

	res, err := s.db.Query(context.Background(), state, tenderId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	resp := make([]bidOpt.BidResponse, 0)
	for res.Next() {
		bid := bidOpt.BidResponse{}
		err = res.Scan(&bid.Id, &bid.Name, &bid.Description,
			&bid.Status, &bid.AuthorType, &bid.AuthorId, &bid.Version, &bid.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		resp = append(resp, bid)
	}
	return resp, nil
}

func (s *Storage) GetBidStatus(bidId string) (string, error) {
	const op = "storage.postgresql.GetBidStatus"

	state := "SELECT status FROM bid WHERE id::text = $1"

	res := s.db.QueryRow(context.Background(), state, bidId)
	var status string
	err := res.Scan(&status)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return status, nil
}

func (s *Storage) EditBidStatus(bidId, status string) (bidOpt.BidResponse, error) {
	const op = "storage.postgresql.EditBidStatus"

	state := "UPDATE bid SET status = $1 WHERE id::text = $2 RETURNING id, name, description, status, authorType, authorId, version, createdat"

	res := s.db.QueryRow(context.Background(), state, status, bidId)
	resp := bidOpt.BidResponse{}
	err := res.Scan(&resp.Id, &resp.Name, &resp.Description,
		&resp.Status, &resp.AuthorType, &resp.AuthorId, &resp.Version, &resp.CreatedAt)
	if err != nil {
		return bidOpt.BidResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (s *Storage) EditBid(bidData bidOpt.BidPatchRequest, bidId string) (bidOpt.BidResponse, error) {
	const op = "storage.postgresql.EditBid"

	state := "SELECT id, name, description, status, authortype, authorid, version, createdat FROM bid WHERE id::TEXT = $1"

	res := s.db.QueryRow(context.Background(), state, bidId)
	resp := bidOpt.BidResponse{}
	err := res.Scan(&resp.Id, &resp.Name, &resp.Description,
		&resp.Status, &resp.AuthorType, &resp.AuthorId, &resp.Version, &resp.CreatedAt)
	if err != nil {
		return bidOpt.BidResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	state = `INSERT INTO bidrollback(bid_id, name, description, status, authortype, authorid, version) VALUES ($1, $2, $3, $4, $5, $6, $7);`
	res = s.db.QueryRow(context.Background(), state, resp.Id, resp.Name, resp.Description, resp.Status,
		resp.AuthorType, resp.AuthorId, resp.Version)
	err = res.Scan()

	state = `INSERT INTO bid(id, name, description, status, authorType, authorId, version) VALUES ($1, $2, $3, $4, $5, $6, $7, $8;`
	res = s.db.QueryRow(context.Background(), state, resp.Id, resp.Name, resp.Description, resp.Status,
		resp.AuthorType, resp.AuthorId, resp.Version)
	err = res.Scan()

	if bidData.Name != "" {
		resp.Name = bidData.Name
	}
	if bidData.Description != "" {
		resp.Description = bidData.Description
	}
	resp.Version++

	state = "UPDATE bid SET name = $1, description = $2, version = $3 WHERE id::text = $4"

	res = s.db.QueryRow(context.Background(), state, resp.Name, resp.Description, resp.Version, bidId)

	err = res.Scan()

	return resp, nil
}

func (s *Storage) IsHavePerms(userId, bidId string) (bool, error) {
	const op = "storage.postgresql.IsHavePerms"

	state := "SELECT authorId, authorType, tenderid FROM bid WHERE id::text = $1"
	res := s.db.QueryRow(context.Background(), state, bidId)
	var tenderId string
	var authorType string
	var authorId string
	err := res.Scan(&authorId, &authorType, &tenderId)
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

	orgId, err := s.GetTenderOrganizationId(tenderId)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	tenderPerms, err := s.ValidateRules(userId, orgId)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if tenderPerms {
		return true, nil
	}
	return false, nil
}

func (s *Storage) GetTenderId(bidId string) (string, error) {
	const op = "storage.postgresql.GetTenderId"

	state := "SELECT tenderid FROM bid WHERE id::text = $1 AND status = 'Published'"

	res := s.db.QueryRow(context.Background(), state, bidId)
	var tenderId string
	err := res.Scan(&tenderId)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tenderId, nil
}

func (s *Storage) SubmitDecision(userId, bidId, decision, tenderId string) (bidOpt.DecisionResponse, error) {
	const op = "storage.postgresql.SubmitDecision"

	state := `SELECT EXISTS(SELECT 1 FROM decision_list WHERE userid = $1 AND bidid = $2)`
	res := s.db.QueryRow(context.Background(), state, userId, bidId)
	var status bool
	err := res.Scan(&status)
	if err != nil {
		return bidOpt.DecisionResponse{}, fmt.Errorf("%s: %w", op, err)
	}
	if status {
		return bidOpt.DecisionResponse{}, fmt.Errorf("%s: %w", op, "already submitted")
	}

	state = `INSERT INTO decision_list(bidid, decision, userid, tenderid) VALUES ($1, $2, $3, $4) RETURNING id, bidid, decision, userid, tenderid;`

	res = s.db.QueryRow(context.Background(), state, bidId, decision, userId, tenderId)
	response := bidOpt.DecisionResponse{}
	err = res.Scan(&response.Id, &response.BidId, &response.Decision, &response.UserId, &response.TenderId)
	if err != nil {
		return bidOpt.DecisionResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	if response.Decision == "Rejected" {
		state = "UPDATE bid SET bidDecision = 'Rejected' WHERE id::text = $1"
		res = s.db.QueryRow(context.Background(), state, bidId)
		err = res.Scan()
		return response, nil
	}
	state = `SELECT COUNT(*) FROM decision_list WHERE bidid = $1;`

	res = s.db.QueryRow(context.Background(), state, bidId)
	var count int
	err = res.Scan(&count)
	if err != nil {
		return bidOpt.DecisionResponse{}, fmt.Errorf("%s: %w", op, err)
	}
	organizationId, err := s.GetUserOrganization(userId)
	if err != nil {
		return bidOpt.DecisionResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	state = `SELECT COUNT(*) FROM organization_responsible WHERE organization_id = $1;`
	res = s.db.QueryRow(context.Background(), state, organizationId)
	var orgSize int
	err = res.Scan(&orgSize)
	if err != nil {
		return bidOpt.DecisionResponse{}, fmt.Errorf("%s: %w", op, err)
	}
	minimum := min(3, orgSize)

	if count >= minimum {
		state = "UPDATE bid SET bidDecision = 'Approved' WHERE id::text = $1"
		res = s.db.QueryRow(context.Background(), state, bidId)
		err = res.Scan()
		state = "UPDATE tender SET status = 'Closed' WHERE id::text = $1"
		res = s.db.QueryRow(context.Background(), state, tenderId)
		err = res.Scan()
		return response, nil
	}

	return response, nil
}

func (s *Storage) CreateFeedback(userId, bidId, bidFeedback string) (bidOpt.FeedbackResponse, error) {
	const op = "storage.postgresql.CreateFeedback"

	state := `INSERT INTO bidreview(bidid, authorId, bidfeedback) VALUES ($1, $2, $3) RETURNING id, bidfeedback, createdat;`
	res := s.db.QueryRow(context.Background(), state, bidId, userId, bidFeedback)
	response := bidOpt.FeedbackResponse{}
	err := res.Scan(&response.Id, &response.BidFeedback, &response.CreatedAt)
	if err != nil {
		return bidOpt.FeedbackResponse{}, fmt.Errorf("%s: %w", op, err)
	}
	return response, nil
}

func (s *Storage) GetReviews(bids []bidOpt.BidResponse, limit, offset string) ([]bidOpt.FeedbackResponse, error) {
	const op = "storage.postgresql.GetReviews"

	bidId := make([]string, 0)

	for _, bid := range bids {
		bidId = append(bidId, bid.Id)
	}

	args := make([]interface{}, len(bids)+2)
	args[0] = limit
	args[1] = offset
	args[2] = pq.Array(bidId)
	state := "SELECT id, bidfeedback, createdat FROM bidreview WHERE bidid = ANY($3) LIMIT $1 OFFSET $2"

	res, err := s.db.Query(context.Background(), state, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resp := make([]bidOpt.FeedbackResponse, 0)
	for res.Next() {
		bid := bidOpt.FeedbackResponse{}
		err = res.Scan(&bid.Id, &bid.BidFeedback, &bid.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		resp = append(resp, bid)
	}
	return resp, nil
}

func (s *Storage) GetBidRollback(bidId, version string) (bidOpt.BidResponse, error) {
	const op = "storage.postgresql.GetBidRollback"

	state := "SELECT name, description FROM bidrollback WHERE bid_id = $1 AND version::TEXT = $2"

	res := s.db.QueryRow(context.Background(), state, bidId, version)
	resp := bidOpt.BidPatchRequest{}
	err := res.Scan(&resp.Name, &resp.Description)
	if err != nil {
		return bidOpt.BidResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	newBid, err := s.EditBid(resp, bidId)
	if err != nil {
		return bidOpt.BidResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return newBid, nil
}
