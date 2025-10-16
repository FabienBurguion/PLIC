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

func WriteHTMLResponseForPasswordReset(w http.ResponseWriter, statusCode int, title, message, password string) error {
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
				background-color: #ffffff;
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
				box-shadow: 0 8px 25px rgba(0,0,0,0.08);
				max-width: 500px;
				text-align: center;
				animation: fadeIn 0.8s ease;
				border-top: 5px solid #FF8C00;
			}
			h1 {
				color: #FF8C00;
				margin-bottom: 20px;
			}
			p {
				font-size: 18px;
				margin-bottom: 25px;
			}
			.password-container {
				display: flex;
				align-items: center;
				justify-content: center;
				gap: 10px;
			}
			.password-box {
				background: #fff7f0;
				padding: 15px 25px;
				border: 2px dashed #FF8C00;
				border-radius: 8px;
				font-size: 22px;
				font-weight: bold;
				color: #FF8C00;
				user-select: all;
				transition: background 0.3s;
			}
			.password-box:hover {
				background: #fff2e0;
			}
			.copy-btn {
				background-color: #FF8C00;
				border: none;
				color: white;
				padding: 12px 18px;
				font-size: 16px;
				font-weight: 600;
				border-radius: 8px;
				cursor: pointer;
				transition: background 0.3s;
			}
			.copy-btn:hover {
				background-color: #e67600;
			}
			.copy-btn:active {
				transform: scale(0.98);
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
			<div class="password-container">
				<div id="password" class="password-box">%s</div>
				<button class="copy-btn" onclick="copyPassword()">Copier</button>
			</div>
			<p class="footer">Conservez ce mot de passe précieusement et changez-le dès votre prochaine connexion.</p>
		</div>

		<script>
			function copyPassword() {
				const pwd = document.getElementById('password').textContent;
				navigator.clipboard.writeText(pwd).then(() => {
					const btn = document.querySelector('.copy-btn');
					btn.textContent = 'Copié !';
					btn.style.backgroundColor = '#28a745';
					setTimeout(() => {
						btn.textContent = 'Copier';
						btn.style.backgroundColor = '#FF8C00';
					}, 2000);
				}).catch(err => {
					alert('Erreur lors de la copie : ' + err);
				});
			}
		</script>
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
