package httpx

import (
	"PLIC/models"
	"encoding/json"
	"net/http"
)

const (
	BadRequestError       = "Bad Request"
	UnauthorizedError     = "Unauthorized"
	ForbiddenError        = "Forbidden"
	NotFoundError         = "Not Found"
	InternalServerError   = "Internal Server Error"
	MethodNotAllowedError = "Method Not Allowed"
)

func Write(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, statusCode int, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(models.Error{
		Message: message,
	})
}
