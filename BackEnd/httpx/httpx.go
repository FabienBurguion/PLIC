package httpx

import (
	"PLIC/models"
	"encoding/json"
	"fmt"
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

func WriteHTMLResponse(w http.ResponseWriter, statusCode int, title string, message string) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="fr">
	<head>
		<meta charset="UTF-8">
		<title>%s</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				background-color: #f4f4f4;
				display: flex;
				align-items: center;
				justify-content: center;
				height: 100vh;
				margin: 0;
			}
			.container {
				background: white;
				padding: 40px;
				border-radius: 10px;
				box-shadow: 0 0 15px rgba(0,0,0,0.1);
				max-width: 500px;
				text-align: center;
			}
			h1 {
				color: #007BFF;
			}
			p {
				font-size: 18px;
				color: #333;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>%s</h1>
			<p>%s</p>
		</div>
	</body>
	</html>	
	`, title, title, message)

	_, err := w.Write([]byte(html))
	return err
}

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
