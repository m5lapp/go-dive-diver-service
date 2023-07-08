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

	userResponse := &data.UserResponse{User: data.User{}}
	httpStatus, res, err := requests.RequestJSend(
		http.MethodGet,
		fmt.Sprintf("%s%s%s", app.cfg.svcUser.Addr, "/v1/user/", input.Email),
		2*time.Second,
		nil,
		userResponse,
	)
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
		case httpStatus == http.StatusNotFound:
			// We received a 404 error from the user service. So make the error
			// message a bit more contextual and meaningful.
			e := fmt.Sprint(
				"Could not add diver as no active user account could be found for",
				input.Email,
			)
			a := "Check a user account exists, has been activated and is not suspended or deleted"
			data := map[string]string{"error": e, "action": a}
			app.FailResponse(w, r, httpStatus, data)
		default:
			app.FailResponse(w, r, httpStatus, res)
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

	du := data.DiverUser{
		Diver: input,
		User:  userResponse.User,
	}

	err = jsonz.WriteJSendSuccess(w, http.StatusAccepted, nil, jsonz.Envelope{"diver": du})
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
