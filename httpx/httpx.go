package httpx

import (
	"PLIC/models"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	BadRequestError     = "Bad Request"
	UnauthorizedError   = "Unauthorized"
	InternalServerError = "Internal Server Error"
)

func WriteHTMLResponse(w http.ResponseWriter, statusCode int, title, message, password string) error {
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
				font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
				background: linear-gradient(135deg, #007BFF 0%%, #00BFFF 100%%);
				display: flex;
				align-items: center;
				justify-content: center;
				height: 100vh;
				margin: 0;
				color: #333;
			}
			.container {
				background: #fff;
				padding: 40px 50px;
				border-radius: 16px;
				box-shadow: 0 10px 30px rgba(0,0,0,0.15);
				max-width: 500px;
				text-align: center;
				animation: fadeIn 0.8s ease;
			}
			h1 {
				color: #007BFF;
				margin-bottom: 20px;
			}
			p {
				font-size: 18px;
				margin-bottom: 25px;
			}
			.password-box {
				display: inline-block;
				background: #f8f9fa;
				padding: 15px 25px;
				border: 2px dashed #007BFF;
				border-radius: 8px;
				font-size: 20px;
				font-weight: bold;
				color: #007BFF;
				user-select: all;
				transition: background 0.3s;
			}
			.password-box:hover {
				background: #e9f2ff;
			}
			.footer {
				margin-top: 30px;
				font-size: 14px;
				color: #777;
			}
			@keyframes fadeIn {
				from { opacity: 0; transform: translateY(20px); }
				to { opacity: 1; transform: translateY(0); }
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>%s</h1>
			<p>%s</p>
			<div class="password-box">%s</div>
			<p class="footer">Conservez ce mot de passe précieusement et changez-le dès votre prochaine connexion.</p>
		</div>
	</body>
	</html>
	`, title, title, message, password)

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
