package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/m5lapp/go-service-toolkit/persistence/sqldb"
	"github.com/m5lapp/go-service-toolkit/validator"
)

// Agency represents a diving certification agency.
type Agency struct {
	ID         int64   `json:"id"`
	CommonName string  `json:"common_name"`
	FullName   string  `json:"full_name"`
	Acronym    *string `json:"acronym"`
	URL        *string `json:"url"`
}

type AgencyModel struct {
	DB *sql.DB
}

// ValidateAgency validates an Agency struct and stores any errors in the
// provided validator.Validator struct.
func ValidateAgency(v *validator.Validator, agency *Agency) {
	v.Check(agency.CommonName != "", "common_name", "Must be provided")
	validator.ValidateStrLenRune(v, agency.CommonName, "common_name", 2, 256)

	v.Check(agency.FullName != "", "full_name", "Must be provided")
	validator.ValidateStrLenRune(v, agency.FullName, "full_name", 2, 256)

	if agency.Acronym != nil {
		validator.ValidateStrLenRune(v, *agency.Acronym, "acronym", 2, 12)
	}

	if agency.URL != nil {
		validator.ValidateURLHTTP(v, *agency.URL, "url")
	}
}

// Insert adds the given dive certification Agency into the database. If the email address (case
// insensitive) already exists in the database, then an ErrDuplicateEmail
// response will be returned.
func (m AgencyModel) Insert(agency *Agency) error {
	// The INSERT query returns the automatically generated values so that they
	// can be added to the User struct.
	query := `
		insert into agencies (
			common_name, full_name, acronym, url
		)
		values ($1, $2, $3, $4)
	 returning id
	`

	args := []any{
		agency.CommonName,
		agency.FullName,
		agency.Acronym,
		agency.URL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, query, args...)
	err := row.Scan(&agency.ID)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "agencies_common_name_key"`:
			return sqldb.NewUniqueConstraintErr("agencies", "common_name")
		case err.Error() == `pq: duplicate key value violates unique constraint "agencies_full_name_key"`:
			return sqldb.NewUniqueConstraintErr("agencies", "full_name")
		default:
			return err
		}
	}

	return nil
}

// GetAll queries the database for all the dive certification agencies.
func (m AgencyModel) GetAll() ([]*Agency, error) {
	query := `
		select
		    id, common_name, full_name, acronym, url
		  from agencies
	  order by common_name desc
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agencies := []*Agency{}
	for rows.Next() {
		var agency Agency

		err := rows.Scan(
			&agency.ID,
			&agency.CommonName,
			&agency.FullName,
			&agency.Acronym,
			&agency.URL,
		)
		if err != nil {
			return nil, err
		}

		agencies = append(agencies, &agency)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return agencies, nil
}
