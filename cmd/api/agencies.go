package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/m5lapp/go-dive-diver-service/internal/data"
	"github.com/m5lapp/go-service-toolkit/persistence/sqldb"
	"github.com/m5lapp/go-service-toolkit/serialisation/jsonz"
	"github.com/m5lapp/go-service-toolkit/validator"
)

func (app *app) createAgencyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CommonName string  `json:"common_name"`
		FullName   string  `json:"full_name"`
		Acronym    *string `json:"acronym"`
		URL        *string `json:"url"`
	}

	err := jsonz.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	agency := &data.Agency{
		CommonName: input.CommonName,
		FullName:   input.FullName,
		Acronym:    input.Acronym,
		URL:        input.URL,
	}

	v := validator.New()

	data.ValidateAgency(v, agency)
	if !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Agencies.Insert(agency)
	if err != nil {
		var errUniqConstraint *sqldb.ErrUniqueConstraintViolation
		switch {
		case errors.As(err, &errUniqConstraint):
			e := make(map[string]string)
			if len(errUniqConstraint.Columns) == 1 {
				e[errUniqConstraint.Columns[0]] = "A record already eists for this value"
			} else {
				cols := strings.Join(errUniqConstraint.Columns, ", ")
				e["form"] = fmt.Sprintf("A record already exists for the values %v", cols)
			}
			app.FailedValidationResponse(w, r, e)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/agencies/%d", agency.ID))

	err = jsonz.WriteJSendSuccess(w, http.StatusCreated, headers, jsonz.Envelope{"agency": agency})
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
