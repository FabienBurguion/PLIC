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
		<link href="https://fonts.googleapis.com/css2?family=Bebas+Neue&family=Inter:wght@400;600&display=swap" rel="stylesheet">
		<style>
			:root {
				--brand: #FF8C00;
				--gold: #FFD700;
				--text: #333;
				--muted: #777;
				--card-shadow: 0 8px 25px rgba(0,0,0,0.08);
				--bg-accent: #fff7f0;
				--bg-accent-hover: #fff2e0;
			}
			body {
				font-family: 'Inter', sans-serif;
				background-color: #fff;
				min-height: 100vh;
				margin: 0;
				display: grid;
				place-items: center;
				color: var(--text);
				overflow: hidden;
				position: relative;
			}

			/* --- Fond basket + couronnes --- */
			.bg-sports {
				position: absolute;
				inset: 0;
				overflow: hidden;
				pointer-events: none;
				opacity: 0.08;
			}
			.bg-sports svg {
				position: absolute;
				width: 160px;
				height: 160px;
				fill: none;
				stroke-width: 2;
			}
			.ball-1 { top: 5%%; left: 6%%; stroke: var(--brand); transform: rotate(-10deg); }
			.ball-2 { bottom: 12%%; right: 10%%; stroke: var(--brand); transform: rotate(15deg) scale(1.1); }
			.crown-1 { top: 15%%; right: 5%%; stroke: var(--gold); transform: rotate(-8deg) scale(1.1); }
			.crown-2 { bottom: 5%%; left: 15%%; stroke: var(--gold); transform: rotate(12deg); }

			.app-name {
				font-family: 'Bebas Neue', sans-serif;
				font-size: 44px;
				color: var(--brand);
				margin: 0 0 8px;
				text-align: center;
				letter-spacing: 1px;
			}
			.app-underline {
				width: 160px;
				height: 6px;
				background: var(--brand);
				margin: 0 auto 22px auto;
				border-radius: 999px;
				opacity: 0.3;
			}
			.container {
				background: #fff;
				padding: 40px 50px;
				border-radius: 16px;
				box-shadow: var(--card-shadow);
				max-width: 560px;
				width: calc(100%% - 32px);
				text-align: center;
				border-top: 5px solid var(--brand);
				z-index: 1;
				animation: fadeIn 0.8s ease;
			}
			h1 {
				color: var(--brand);
				margin-bottom: 18px;
				font-size: 26px;
			}
			p {
				font-size: 18px;
				margin-bottom: 25px;
			}
			.password-container {
				display: inline-flex;
				align-items: center;
				justify-content: center;
				gap: 10px;
				margin-bottom: 10px;
			}
			.password-box {
				background: var(--bg-accent);
				padding: 15px 25px;
				border: 2px dashed var(--brand);
				border-radius: 8px;
				font-size: 22px;
				font-weight: bold;
				color: var(--brand);
				user-select: all;
				transition: background 0.3s;
			}
			.password-box:hover {
				background: var(--bg-accent-hover);
			}
			.copy-btn {
				background-color: var(--brand);
				border: none;
				color: white;
				padding: 12px 18px;
				font-size: 16px;
				font-weight: 600;
				border-radius: 8px;
				cursor: pointer;
				transition: background 0.3s, transform 0.05s;
			}
			.copy-btn:hover { background-color: #e67600; }
			.copy-btn:active { transform: scale(0.98); }
			.copy-hint {
				font-size: 14px;
				color: var(--muted);
				margin-top: 4px;
			}
			.footer {
				margin-top: 28px;
				font-size: 14px;
				color: var(--muted);
			}
			@keyframes fadeIn {
				from { opacity: 0; transform: translateY(14px); }
				to { opacity: 1; transform: translateY(0); }
			}
		</style>
	</head>
	<body>
		<div class="bg-sports" aria-hidden="true">
			<!-- 🏀 Ballon de basket -->
			<svg class="ball-1" viewBox="0 0 100 100">
				<circle cx="50" cy="50" r="44"></circle>
				<path d="M6,50 H94 M50,6 V94 M16,20 C40,40 60,60 84,80 M84,20 C60,40 40,60 16,80"></path>
			</svg>
			<!-- 👑 Couronne -->
			<svg class="crown-1" viewBox="0 0 100 100">
				<path d="M10 70 L25 30 L50 60 L75 30 L90 70 Z"></path>
				<path d="M10 70 H90 V80 H10 Z"></path>
			</svg>
			<!-- 👑 Couronne -->
			<svg class="crown-2" viewBox="0 0 100 100">
				<path d="M10 70 L25 30 L50 60 L75 30 L90 70 Z"></path>
				<path d="M10 70 H90 V80 H10 Z"></path>
			</svg>
		</div>

		<div class="container">
			<div class="app-name">Play The Street</div>
			<div class="app-underline"></div>

			<h1>%s</h1>
			<p>%s</p>

			<div class="password-container">
				<div id="password" class="password-box">%s</div>
				<button class="copy-btn" onclick="copyPassword()">Copier</button>
			</div>
			<div class="copy-hint">Clique pour copier le mot de passe</div>

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
