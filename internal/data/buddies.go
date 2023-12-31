package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/m5lapp/go-service-toolkit/validator"
)

// Buddy represents a diver's buddy.
type Buddy struct {
	ID           int64     `json:"-"`
	Version      int       `json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
	UserID       string    `json:"user_id"`
	BuddyUserID  *string   `json:"-"`
	Name         string    `json:"name"`
	Email        *string   `json:"email"`
	PhoneNumber  *string   `json:"phone_number"`
	Organisation *string   `json:"organisation"`
	OrgMemberID  *string   `json:"org_member_id"`
	Notes        *string   `json:"notes"`
}

type BuddyModel struct {
	DB *sql.DB
}

// ValidateBuddy validates a Buddy struct and stores any errors in the provided
// validator.Validator struct.
func ValidateBuddy(v *validator.Validator, buddy *Buddy) {
	v.Check(buddy.Name != "", "name", "Must be provided")
	v.Check(len(buddy.Name) <= 500, "name", "Must not be more than 500 bytes long")

	if buddy.Email != nil {
		validator.ValidateEmail(v, *buddy.Email)
	}

	if buddy.PhoneNumber != nil {
		validator.ValidateStrLenByte(v, *buddy.PhoneNumber, "phone_number", 7, 24)
	}

	if buddy.Organisation != nil {
		// TODO: Ensure the organisation is a valid option.
		validator.ValidateStrLenRune(v, *buddy.Organisation, "organisation", 2, 64)
	}

	if buddy.OrgMemberID != nil {
		validator.ValidateStrLenRune(v, *buddy.OrgMemberID, "org_member_id", 2, 32)
		v.Check(buddy.Organisation != nil, "org_member_id",
			"Cannot be supplied without an organisation")
	}

	if buddy.Notes != nil {
		validator.ValidateStrLenRune(v, *buddy.Notes, "notes", 0, 65535)
	}
}

// Insert adds the given Buddy into the database. If the email address (case
// insensitive) already exists in the database, then an ErrDuplicateEmail
// response will be returned.
func (m BuddyModel) Insert(buddy *Buddy) error {
	// The INSERT query returns the automatically generated values so that they
	// can be added to the User struct.
	query := `
		insert into buddies (
			user_id, name, email, phone_number, buddy_user_id, organisation,
			org_member_id, notes
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	 returning id, version, created_at, updated_at
	`

	args := []any{
		buddy.UserID,
		buddy.Name,
		buddy.Email,
		buddy.PhoneNumber,
		buddy.BuddyUserID,
		buddy.Organisation,
		buddy.OrgMemberID,
		buddy.Notes,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, query, args...)
	err := row.Scan(&buddy.ID, &buddy.Version, &buddy.CreatedAt, &buddy.UpdatedAt)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "buddies_user_id_buddy_user_id_key"`,
			err.Error() == `pq: duplicate key value violates unique constraint "buddies_user_id_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// GetAllForDiver queries the database for all the buddies of the Diver with the
// given UserID.
func (m BuddyModel) GetAllForDiver(userID string) ([]*Buddy, error) {
	query := `
		select
		    id, version, created_at, updated_at, user_id, buddy_user_id, name,
			email, phone_number, organisation, org_member_id, notes
		  from buddies
		 where user_id = $1
	  order by name desc
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buddies := []*Buddy{}
	for rows.Next() {
		var buddy Buddy

		err := rows.Scan(
			&buddy.ID,
			&buddy.Version,
			&buddy.CreatedAt,
			&buddy.UpdatedAt,
			&buddy.UserID,
			&buddy.BuddyUserID,
			&buddy.Name,
			&buddy.Email,
			&buddy.PhoneNumber,
			&buddy.Organisation,
			&buddy.OrgMemberID,
			&buddy.Notes,
		)
		if err != nil {
			return nil, err
		}

		buddies = append(buddies, &buddy)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return buddies, nil
}
