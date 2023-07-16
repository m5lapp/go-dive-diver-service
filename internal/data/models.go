package data

import (
	"database/sql"
	"errors"
)

var (
	ErrEditConflict   = errors.New("edit conflict")
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Buddies BuddyModel
	Divers  DiverModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Buddies: BuddyModel{DB: db},
		Divers:  DiverModel{DB: db},
	}
}
