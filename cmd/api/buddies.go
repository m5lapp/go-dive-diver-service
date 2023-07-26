package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/m5lapp/go-dive-diver-service/internal/data"
	"github.com/m5lapp/go-service-toolkit/serialisation/jsonz"
	"github.com/m5lapp/go-service-toolkit/validator"
)

func (app *app) createBuddyHandler(w http.ResponseWriter, r *http.Request) {
	input := &data.Buddy{}

	err := jsonz.ReadJSON(w, r, input)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateBuddy(v, input)
	if !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// If the buddy's email has been provided, see if it resolves to an active
	// user account.
	if input.Email != nil {
		// Call the User service to see if the given user has a valid account.
		url := fmt.Sprintf("%s%s%s", app.cfg.svcUser.Addr, "/v1/user/email/", *input.Email)
		httpResp, res, err := jsonz.RequestJSend(http.MethodGet, url, 2*time.Second, nil)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		// We received an error response, pass it upstream.
		if res.Status == jsonz.JSendStatusError {
			app.ServerErrorResponse(w, r, fmt.Errorf(res.Message))
			return
		}

		if res.Status == jsonz.JSendStatusSuccess {
			// The JSend body contains a success Status, so use a second
			// decoding pass to decode the JSend Data field into a user struct.
			userResp := &data.UserResponse{User: data.User{}}
			err = jsonz.DecodeJSON(bytes.NewReader(res.Data), userResp, true)
			if err != nil {
				app.ServerErrorResponse(w, r, err)
				return
			}

			// Attempt to get the diver record from the database if one exists.
			d, err := app.models.Divers.GetByID(userResp.User.UserID)
			if err != nil && !errors.Is(err, data.ErrRecordNotFound) {
				app.ServerErrorResponse(w, r, err)
				return
			}

			// If a record was found, assign their user ID to the BuddyUserID
			// field. Also use the name provided on their account to avoid any
			// confusion.
			if !errors.Is(err, data.ErrRecordNotFound) {
				// Check that the account belonging to the provided email is not
				// that of the requesting user.
				if input.UserID == d.UserID {
					e := map[string]string{"email": "Must not be for your own account"}
					app.FailedValidationResponse(w, r, e)
					return
				}

				input.BuddyUserID = &d.UserID
				input.Name = userResp.User.Name
			}
		}

		if res.Status == jsonz.JSendStatusFail {
			// We received a fail response, pass it upstream. If the response
			// was a 404 then don't do anything as that just means the buddy is
			// not a registered user.
			switch {
			case httpResp.StatusCode != http.StatusNotFound:
				app.FailResponse(w, r, httpResp.StatusCode, res.Data)
				return
			}
		}
	}

	err = app.models.Buddies.Insert(input)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "A buddy with this email address already exists")
			app.FailedValidationResponse(w, r, v.Errors)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}

	app.Logger.Info("New buddy successfully added", "user", input.UserID,
		"buddy", input.Name, "account_linked", input.BuddyUserID != nil)

	b := jsonz.Envelope{"buddy": input}
	err = jsonz.WriteJSendSuccess(w, http.StatusCreated, nil, b)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

func (app *app) listBuddiesHandler(w http.ResponseWriter, r *http.Request) {
	var userID string

	params := httprouter.ParamsFromContext(r.Context())
	userID = params.ByName("id")
	v := validator.New()

	v.Check(validator.Matches(userID, validator.BetterGUIDRX),
		"user-id", "must be a valid BetterGUID")

	if !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// TODO: Currently returns a 200 and an empty result set if the diverID does
	// not exist. Might want to check the diverID first.
	buddies, err := app.models.Buddies.GetAllForDiver(userID)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
		return
	}

	envelope := jsonz.Envelope{"buddies": buddies}
	err = jsonz.WriteJSON(w, http.StatusOK, nil, envelope)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
