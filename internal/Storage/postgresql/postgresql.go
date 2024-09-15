package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *pgx.Conn
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgresql.New"
	db, err := pgx.Connect(context.Background(), storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s - %s", op, err)
	}

	state := `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`

	_, err = db.Exec(context.Background(), state)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	state =
		`
	CREATE TABLE IF NOT EXISTS tender (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(100) NOT NULL,
		description TEXT,
		serviceType TEXT,
		status tender_status DEFAULT 'Created',
		organizationId UUID NOT NULL,
		creatorUsername VARCHAR(50) NOT NULL,
	    version INTEGER NOT NULL DEFAULT 1,         
		createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS bid (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(100) NOT NULL,
		description TEXT,
		status tender_status DEFAULT 'Created',
		tenderId UUID NOT NULL,
		authorType author_status NOT NULL,
		authorId UUID NOT NULL,
		bidDecision bid_decision DEFAULT 'Active',
	    version INTEGER NOT NULL DEFAULT 1,         
		createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS decision_list (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		bidId UUID REFERENCES bid(id) ON DELETE CASCADE,
		decision decision_variant,
		userId UUID REFERENCES employee(id) ON DELETE CASCADE,
		tenderId UUID REFERENCES tender(ID) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS bidreview (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		bidId UUID NOT NULL,
		authorId UUID NOT NULL,
		bidFeedback TEXT,
		createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tenderRollback (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tender_id UUID NOT NULL,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		serviceType TEXT,
		status tender_status NOT NULL,
	    version INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS bidRollback (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		bid_id UUID NOT NULL,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		status tender_status NOT NULL,
		authorType author_status NOT NULL,
		authorId UUID NOT NULL,
	    version INTEGER NOT NULL         
	);
`

	_, err = db.Exec(context.Background(), state)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}
