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
	Divers DiverModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Divers: DiverModel{DB: db},
	}
}
