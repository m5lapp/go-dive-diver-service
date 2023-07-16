package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/m5lapp/go-dive-diver-service/internal/data"
	"github.com/m5lapp/go-service-toolkit/serialisation/jsonz"
	"github.com/m5lapp/go-service-toolkit/validator"
)

func (app *app) createDiverHandler(w http.ResponseWriter, r *http.Request) {
	input := struct {
		data.Diver
		Email string `json:"email"`
	}{}

	err := jsonz.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateDiver(v, &input.Diver)
	validator.ValidateEmail(v, input.Email)
	if !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the User service to see if the given user has a valid account.
	url := fmt.Sprintf("%s%s%s", app.cfg.svcUser.Addr, "/v1/user/email/", input.Email)
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

	// We received a fail response, pass it upstream.
	if res.Status == jsonz.JSendStatusFail {
		switch {
		case httpResp.StatusCode == http.StatusNotFound:
			// We received a 404 error from the user service. So make the error
			// message a bit more contextual and meaningful.
			e := fmt.Sprint(
				"Could not add diver as no active user account could be found for",
				input.Email,
			)
			a := "Check a user account exists, has been activated and is not suspended or deleted"
			data := map[string]string{"error": e, "action": a}
			app.FailResponse(w, r, httpResp.StatusCode, data)
		default:
			app.FailResponse(w, r, httpResp.StatusCode, res.Data)
		}
		return
	}

	// The JSend body contains a success Status, so use a second decoding pass
	// to decode the JSend Data field into targetStruct.
	userResp := &data.UserResponse{User: data.User{}}
	err = jsonz.DecodeJSON(bytes.NewReader(res.Data), userResp, true)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
		return
	}

	input.UserID = userResp.User.UserID
	err = app.models.Divers.Insert(&input.Diver)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a diver with this email address already exists")
			app.FailedValidationResponse(w, r, v.Errors)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}

	app.Logger.Info("New diver successfully registered", "diver", input.Email)

	du := data.DiverUser{
		Email:        input.Email,
		Name:         userResp.User.Name,
		FriendlyName: userResp.User.FriendlyName,
		BirthDate:    userResp.User.BirthDate,
		Gender:       userResp.User.Gender,
		CountryCode:  userResp.User.CountryCode,
		TimeZone:     userResp.User.TimeZone,
		Diver:        input.Diver,
	}
	err = jsonz.WriteJSendSuccess(w, http.StatusAccepted, nil, jsonz.Envelope{"diver": du})
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
