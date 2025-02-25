package httpx

import (
	"PLIC/models"
	"encoding/json"
	"net/http"
)

type httpxError string

const (
	BadRequestError     httpxError = "Bad Request"
	UnauthorizedError   httpxError = "Unauthorized"
	ForbiddenError      httpxError = "Forbidden"
	NotFoundError       httpxError = "Not Found"
	InternalServerError httpxError = "Internal Server Error"
)

func Write(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, statusCode int, message httpxError) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(models.Error{
		Message: string(message),
	})
}
