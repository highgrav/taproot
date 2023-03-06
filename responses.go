package taproot

import (
	"fmt"
	"highgrav/taproot/v1/validation"
	"net/http"
)

func (srv *AppServer) ErrorResponse(w http.ResponseWriter, r *http.Request, status int, msg any) {
	env := DataEnvelope{
		"ok":    false,
		"error": msg,
	}
	err := srv.WriteJSON(w, true, status, env, nil)
	if err != nil {
		// srv.Log(r, logging.LevelError, err)
		w.WriteHeader(500)
	}
}

func (srv *AppServer) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the required resource could not be found"
	srv.ErrorResponse(w, r, http.StatusNotFound, message)
}

func (srv *AppServer) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported on this resource", r.Method)
	srv.ErrorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (srv *AppServer) ValidationErrorResponse(w http.ResponseWriter, r *http.Request, v *validation.Validator) {
	srv.ErrorResponse(w, r, http.StatusUnprocessableEntity, v.Errors)
}

func (srv *AppServer) EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	srv.ErrorResponse(w, r, http.StatusConflict, "unable to update the entity due to a merge conflict, please try again")
}

func (srv *AppServer) ServerErrorResponse(w http.ResponseWriter, r *http.Request) {
	srv.ErrorResponse(w, r, http.StatusInternalServerError, "a AppServer error has occurred")
}

func (srv *AppServer) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	srv.ErrorResponse(w, r, http.StatusTooManyRequests, "too many requests, please slow your roll")
}

func (srv *AppServer) InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	srv.ErrorResponse(w, r, http.StatusUnauthorized, "missing or invalid authentication token")
}

func (srv *AppServer) ForbiddenResponse(w http.ResponseWriter, r *http.Request) {
	srv.ErrorResponse(w, r, http.StatusForbidden, "you are not authorized to access this resource")
}

func (srv *AppServer) UserRequiresActivationResponse(w http.ResponseWriter, r *http.Request) {
	srv.ErrorResponse(w, r, http.StatusForbidden, "you must activate your account to access this resource")
}

func (srv *AppServer) ReauthenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	srv.ErrorResponse(w, r, http.StatusProxyAuthRequired, "you must reauthenticate to access this resource")
}
