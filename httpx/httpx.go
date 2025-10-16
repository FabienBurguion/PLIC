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
				--text: #333;
				--muted: #777;
				--card-shadow: 0 8px 25px rgba(0,0,0,0.08);
				--bg-accent: #fff7f0;
				--bg-accent-hover: #fff2e0;
			}
			* { box-sizing: border-box; }
			body {
				font-family: 'Inter', system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif;
				background-color: #ffffff;
				min-height: 100vh;
				margin: 0;
				color: var(--text);
				display: grid;
				place-items: center;
				overflow: hidden;
				position: relative;
			}
			/* --- Fond sport discret (SVG placés en arrière-plan) --- */
			.bg-sports {
				position: absolute;
				inset: 0;
				overflow: hidden;
				pointer-events: none;
				opacity: 0.08; /* très discret */
			}
			.bg-sports svg {
				position: absolute;
				width: 160px;
				height: 160px;
				fill: none;
				stroke: var(--brand);
				stroke-width: 2;
			}
			.bg-sports .ball-1 { top: 5%%; left: 8%%; transform: rotate(-10deg); }
			.bg-sports .ball-2 { top: 20%%; right: -30px; transform: rotate(15deg) scale(1.1); }
			.bg-sports .ball-3 { bottom: 12%%; left: 10%%; transform: rotate(8deg) scale(0.9); }
			.bg-sports .ball-4 { bottom: -20px; right: 12%%; transform: rotate(-6deg) scale(1.2); }

			.app-name {
				font-family: 'Bebas Neue', Arial, sans-serif;
				font-size: 42px;
				letter-spacing: 1px;
				color: var(--brand);
				margin: 0 0 8px 0;
				text-align: center;
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
				padding: 36px 44px;
				border-radius: 16px;
				box-shadow: var(--card-shadow);
				max-width: 560px;
				width: calc(100%% - 32px);
				text-align: center;
				border-top: 5px solid var(--brand);
				position: relative;
				z-index: 1;
				animation: fadeIn 0.8s ease;
			}
			h1 {
				color: var(--brand);
				margin: 0 0 16px 0;
				font-size: 26px;
			}
			p {
				font-size: 18px;
				margin: 0 0 22px 0;
			}
			.password-container {
				display: inline-flex;
				align-items: center;
				justify-content: center;
				gap: 10px;
				margin-bottom: 8px;
			}
			.password-box {
				background: var(--bg-accent);
				padding: 14px 22px;
				border: 2px dashed var(--brand);
				border-radius: 10px;
				font-size: 22px;
				font-weight: 700;
				color: var(--brand);
				user-select: all;
				transition: background 0.3s;
				line-height: 1;
			}
			.password-box:hover { background: var(--bg-accent-hover); }
			.copy-btn {
				background-color: var(--brand);
				border: none;
				color: white;
				padding: 12px 16px;
				font-size: 16px;
				font-weight: 600;
				border-radius: 10px;
				cursor: pointer;
				transition: background 0.3s, transform 0.05s;
				white-space: nowrap;
			}
			.copy-btn:hover { background-color: #e67600; }
			.copy-btn:active { transform: scale(0.98); }
			.copy-hint {
				font-size: 14px;
				color: var(--muted);
				margin-top: 6px;
			}
			.footer {
				margin-top: 24px;
				font-size: 14px;
				color: var(--muted);
			}
			@keyframes fadeIn {
				from { opacity: 0; transform: translateY(14px); }
				to { opacity: 1; transform: translateY(0); }
			}
			/* Mobile tweaks */
			@media (max-width: 420px) {
				.password-box { font-size: 18px; }
				.copy-btn { font-size: 14px; padding: 10px 14px; }
			}
		</style>
	</head>
	<body>
		<!-- Fond “sport” discret -->
		<div class="bg-sports" aria-hidden="true">
			<!-- Basket -->
			<svg class="ball-1" viewBox="0 0 100 100">
				<circle cx="50" cy="50" r="44"></circle>
				<path d="M6,50 H94"></path>
				<path d="M50,6 V94"></path>
				<path d="M16,20 C40,40 60,60 84,80"></path>
				<path d="M84,20 C60,40 40,60 16,80"></path>
			</svg>
			<!-- Foot (simplifié : panneaux) -->
			<svg class="ball-2" viewBox="0 0 100 100">
				<circle cx="50" cy="50" r="44"></circle>
				<polygon points="50,22 62,35 56,52 44,52 38,35" ></polygon>
				<path d="M50,22 L22,40 L30,68 L50,78 L70,68 L78,40 Z"></path>
			</svg>
			<!-- Basket -->
			<svg class="ball-3" viewBox="0 0 100 100">
				<circle cx="50" cy="50" r="44"></circle>
				<path d="M6,50 H94"></path>
				<path d="M50,6 V94"></path>
				<path d="M16,20 C40,40 60,60 84,80"></path>
				<path d="M84,20 C60,40 40,60 16,80"></path>
			</svg>
			<!-- Foot -->
			<svg class="ball-4" viewBox="0 0 100 100">
				<circle cx="50" cy="50" r="44"></circle>
				<polygon points="50,22 62,35 56,52 44,52 38,35" ></polygon>
				<path d="M50,22 L22,40 L30,68 L50,78 L70,68 L78,40 Z"></path>
			</svg>
		</div>

		<div class="container" role="main">
			<div class="app-name">Play The Street</div>
			<div class="app-underline" aria-hidden="true"></div>

			<h1>%s</h1>
			<p>%s</p>

			<div class="password-container">
				<div id="password" class="password-box">%s</div>
				<button class="copy-btn" onclick="copyPassword()">Copier</button>
			</div>
			<div class="copy-hint">Clique pour copier le mot de passe.</div>

			<p class="footer">Conservez ce mot de passe précieusement et changez-le dès votre prochaine connexion.</p>
		</div>

		<script>
			function copyPassword() {
				var el = document.getElementById('password');
				var pwd = el.textContent;
				if (navigator.clipboard && navigator.clipboard.writeText) {
					navigator.clipboard.writeText(pwd).then(function () {
						flashCopied();
					}).catch(function (err) {
						fallbackCopy(pwd);
					});
				} else {
					fallbackCopy(pwd);
				}
			}
			function fallbackCopy(text) {
				var ta = document.createElement('textarea');
				ta.value = text;
				document.body.appendChild(ta);
				ta.select();
				try { document.execCommand('copy'); } catch (e) {}
				document.body.removeChild(ta);
				flashCopied();
			}
			function flashCopied() {
				var btn = document.querySelector('.copy-btn');
				btn.textContent = 'Copié !';
				btn.style.backgroundColor = '#28a745';
				setTimeout(function () {
					btn.textContent = 'Copier';
					btn.style.backgroundColor = getComputedStyle(document.documentElement).getPropertyValue('--brand').trim();
				}, 2000);
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
