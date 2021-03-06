package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gitlab.com/spacewalker/geotracker/internal/app/history/core/domain"
	"gitlab.com/spacewalker/geotracker/internal/app/history/core/port"
	"gitlab.com/spacewalker/geotracker/internal/pkg/errpack"
	"gitlab.com/spacewalker/geotracker/internal/pkg/geo"
)

const (
	// RecordsTable contains name of the records database table.
	RecordsTable = "records"

	constraintRecordsALongitudeValid = "records_a_longitude_valid"
	constraintRecordsALatitudeValid  = "records_a_latitude_valid"
	constraintRecordsBLongitudeValid = "records_b_longitude_valid"
	constraintRecordsBLatitudeValid  = "records_b_latitude_valid"
)

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository returns pointer to new PostgresRepository instance.
func NewPostgresRepository(db *sql.DB) port.HistoryRepository {
	return &postgresRepository{db: db}
}

var addRecordQuery = fmt.Sprintf(
	`
INSERT INTO %s
(user_id, a, b, timestamp)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, a, b, timestamp
`,
	RecordsTable,
)

// AddRecord adds a history record into records table.
//
// It returns added record and any error encountered.
//
// `ErrInvalidArgument` is returned in case any of provided geo points contains
// invalid latitude or longitude.
//
// `ErrInternalError` is returned in case of any other error.
func (r postgresRepository) AddRecord(ctx context.Context, req port.HistoryRepositoryAddRecordRequest) (domain.Record, error) {
	var record domain.Record
	var a, b geo.PostgresPoint

	if err := r.db.QueryRowContext(ctx, addRecordQuery, req.UserID, geo.PostgresPoint(req.A), geo.PostgresPoint(req.B), req.Timestamp).Scan(
		&record.ID,
		&record.UserID,
		&a,
		&b,
		&record.Timestamp,
	); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Constraint {
			case constraintRecordsALongitudeValid:
				fallthrough
			case constraintRecordsALatitudeValid:
				fallthrough
			case constraintRecordsBLongitudeValid:
				fallthrough
			case constraintRecordsBLatitudeValid:
				return domain.Record{}, fmt.Errorf("%w", errpack.ErrInvalidArgument)
			}
		}
		return domain.Record{}, fmt.Errorf("%w: %v", errpack.ErrInternalError, err)
	}

	record.A = geo.Point(a)
	record.B = geo.Point(b)

	return record, nil
}

var getDistanceQuery = fmt.Sprintf(
	`
SELECT coalesce(SUM(a <@> b), 0.00) * 1609.344
FROM %s
WHERE user_id = $1 AND timestamp >= $2 AND timestamp <= $3
`,
	RecordsTable,
)

// GetDistance returns distance a user with the provided ID passed in
// a provided period of time.
//
// It returns distance and any error occurred.
//
// If there is no user with provided ID, 0 is returned as distance.
//
// `ErrInternalError` is returned in case of any error.
func (r postgresRepository) GetDistance(ctx context.Context, req port.HistoryRepositoryGetDistanceRequest) (float64, error) {
	var distance float64
	if err := r.db.QueryRowContext(ctx, getDistanceQuery, req.UserID, req.From, req.To).Scan(&distance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("%w: %v", errpack.ErrInternalError, err)
	}

	return distance, nil
}
