package pkg

import (
	"fmt"
	clogger "interviews/pkg/logger"
	"net/http"
)

type CustomErrors struct {
	helper Helper
}

func (e *CustomErrors) logError(r *http.Request, err error) {
	clogger.ErrorCtx(err, clogger.Ctx{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

type envelope map[string]any

func (e *CustomErrors) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {

	env := envelope{"error": message}

	err := e.helper.WriteJSON(w, status, env, nil)
	if err != nil {
		e.logError(r, err)
		w.WriteHeader(500)
	}
}

func (e *CustomErrors) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	e.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (e *CustomErrors) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	e.errorResponse(w, r, http.StatusNotFound, message)
}

func (e *CustomErrors) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	e.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (e *CustomErrors) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (e *CustomErrors) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	e.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (e *CustomErrors) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	e.errorResponse(w, r, http.StatusConflict, message)
}

func (e *CustomErrors) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	e.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (e *CustomErrors) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	e.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (e *CustomErrors) InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	e.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (e *CustomErrors) AuthenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	e.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (e *CustomErrors) InactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	e.errorResponse(w, r, http.StatusForbidden, message)
}

func (e *CustomErrors) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	e.errorResponse(w, r, http.StatusForbidden, message)
}
