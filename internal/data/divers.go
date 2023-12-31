package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/m5lapp/go-service-toolkit/serialisation/jsonz"
	"github.com/m5lapp/go-service-toolkit/validator"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// Diver represents a human diver who is a user of the go-dive system. It embeds
// a standard User struct and adds some additional fields.
type Diver struct {
	UserID               string          `json:"user_id"`
	Version              int             `json:"-"`
	DivingSince          *jsonz.DateOnly `json:"diving_since"`
	DiveNumberOffset     int             `json:"dive_number_offset"`
	DefaultDivingCountry *string         `json:"default_diving_country"`
	DefaultDivingTZ      *string         `json:"default_diving_timezone"`
}

// DiverUser contains the base User fields from the User sevice, combined with
// the Diver-specific fields for passing to clients.
type DiverUser struct {
	Email        string          `json:"email"`
	Name         string          `json:"name"`
	FriendlyName *string         `json:"friendly_name,omitempty"`
	BirthDate    *jsonz.DateOnly `json:"birth_date,omitempty"`
	Gender       *string         `json:"gender,omitempty"`
	CountryCode  *string         `json:"country_code,omitempty"`
	TimeZone     *string         `json:"time_zone,omitempty"`
	Diver
}

type DiverModel struct {
	DB *sql.DB
}

// ValidateDiver validates a Diver struct and stores any errors in the provided
// validator.Validator struct. Only the additional fields will be validated. The
// base User fields can be validated by the go-user-service service.
func ValidateDiver(v *validator.Validator, diver *Diver) {
	if diver.DivingSince != nil {
		inPast := diver.DivingSince.Before(time.Now())
		v.Check(inPast, "diving_since", "Must not be in the future")
		// if diver.BirthDate != nil {
		// 	tooYoung := diver.DivingSince.After(diver.BirthDate.AddDate(8, 0, 0))
		// 	v.Check(tooYoung, "diving_since", "Must not be before your eight birthday")
		// }
	}

	v.Check(diver.DiveNumberOffset >= 0, "dive_number_offset", "Must not be negative")
	v.Check(diver.DiveNumberOffset <= 32767, "dive_number_offset", "Must be less than 32768")

	if diver.DefaultDivingCountry != nil {
		// TODO: Ensure the country code is a valid option.
		v.Check(len(*diver.DefaultDivingCountry) == 2, "default_diving_country", "Must be exactly two bytes long")
	}

	if diver.DefaultDivingTZ != nil {
		_, err := time.LoadLocation(*diver.DefaultDivingTZ)
		v.Check(err == nil, "default_diving_timezone", "Must be a valid time zone name")
	}
}

// Insert adds the given Diver into the database. If the email address (case
// insensitive) already exists in the database, then an ErrDuplicateEmail
// response will be returned.
func (m DiverModel) Insert(diver *Diver) error {
	// The INSERT query returns the automatically generated values so that they
	// can be added to the User struct.
	query := `
		insert into divers (
			user_id, diving_since, dive_number_offset, default_diving_country,
			default_diving_timezone
		)
		values ($1, $2, $3, $4, $5)
	 returning version
	`

	args := []any{
		diver.UserID,
		diver.DivingSince,
		diver.DiveNumberOffset,
		diver.DefaultDivingCountry,
		diver.DefaultDivingTZ,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, query, args...)
	err := row.Scan(&diver.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "divers_pkey"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// GetByID queries the database for a diver record with the given User ID.
// If no matching record exists, ErrRecordNotFound is returned.
func (m DiverModel) GetByID(id string) (*Diver, error) {
	query := `
		select
		    user_id, version, diving_since, dive_number_offset,
			default_diving_country, default_diving_timezone
		  from divers
		 where user_id = $1
	`

	var diver Diver

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&diver.UserID,
		&diver.Version,
		&diver.DivingSince,
		&diver.DiveNumberOffset,
		&diver.DefaultDivingCountry,
		&diver.DefaultDivingTZ,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &diver, nil
}
