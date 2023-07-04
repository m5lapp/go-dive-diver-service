package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/m5lapp/go-dive-diver-service/internal/data"
	"github.com/m5lapp/go-service-toolkit/requests"
	"github.com/m5lapp/go-service-toolkit/serialisation/jsonz"
	"github.com/m5lapp/go-service-toolkit/validator"
)

func (app *app) createDiverHandler(w http.ResponseWriter, r *http.Request) {
	var input data.Diver

	err := jsonz.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateDiver(v, &input)
	if !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// du := data.DiverUserResponse{User: data.User{DiverUser: data.DiverUser{Diver: input}}}

	status, body, err := requests.RequestJSend(
		http.MethodGet,
		fmt.Sprintf("%s%s%s", app.cfg.svcUser.Addr, "/v1/user/", input.Email),
		2*time.Second,
		nil,
		// jsonz.Envelope{"user": du},
		struct{}{},
	)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
		return
	}

	// We received an error response, pass it upstream.
	if body.Status == jsonz.JSendStatusError {
		app.ServerErrorResponse(w, r, fmt.Errorf(body.Message))
		return
	}

	// We received a fail response, pass it upstream.
	if body.Status == jsonz.JSendStatusFail {
		switch {
		case status == http.StatusNotFound:
			// We received a 404 error from the user service. So make the error
			// message a bit more contectual and meaningful.
			e := fmt.Sprint(
				"Could not add diver as no active user account could be found for",
				input.Email,
			)
			a := "Check a user account exists, has been activated and is not suspended or deleted"
			data := map[string]string{"error": e, "action": a}
			app.FailResponse(w, r, status, data)
		default:
			app.FailResponse(w, r, status, body.Data)
		}
		return
	}

	err = app.models.Divers.Insert(&input)
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

	// Empty the password so it's not included in the response.
	err = jsonz.WriteJSendSuccess(w, http.StatusAccepted, nil, jsonz.Envelope{"diver": body.Data})
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
