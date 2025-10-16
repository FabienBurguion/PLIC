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
				--brand: #FF8C00;   /* Orange Play The Street */
				--gold:  #EFBF04;   /* Or un peu plus chaud que #FFD700 */
				--text:  #333;
				--muted: #777;
				--card-shadow: 0 10px 30px rgba(0,0,0,0.10);
				--bg-accent: #fff7f0;
				--bg-accent-hover: #fff2e0;
			}
			* { box-sizing: border-box; }
			body {
				font-family: 'Inter', system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif;
				background: #fff;
				min-height: 100vh;
				margin: 0;
				color: var(--text);
				display: grid;
				place-items: center;
				overflow: hidden;
				position: relative;
			}

			/* === FOND SPORT plus présent (ballons + couronnes “propres”) === */
			.bg-sports {
				position: absolute;
				inset: 0;
				pointer-events: none;
				/* un peu plus voyant qu'avant */
				opacity: 0.16;
			}
			.bg-sports svg {
				position: absolute;
				width: 180px;
				height: 180px;
				fill: none;
			}

			/* Variations positions / tailles */
			.ball-1   { top: 5%%;   left: 6%%;   transform: rotate(-8deg)  scale(1.0);  }
			.ball-2   { top: 22%%;  right: 8%%;  transform: rotate(12deg)  scale(1.15); }
			.ball-3   { bottom: 8%%; left: 12%%; transform: rotate(7deg)   scale(0.95); }
			.crown-1  { top: 14%%;  left: 32%%;  transform: rotate(-6deg)  scale(1.1);  }
			.crown-2  { bottom: 10%%; right: 14%%;transform: rotate(10deg)  scale(1.0);  }
			.crown-3  { bottom: 28%%; left: 4%%;  transform: rotate(-12deg) scale(0.9);  }

			/* Titre appli */
			.app-name {
				font-family: 'Bebas Neue', Arial, sans-serif;
				font-size: 46px;
				letter-spacing: 1px;
				color: var(--brand);
				margin: 0 0 8px 0;
				text-align: center;
			}
			.app-underline {
				width: 172px;
				height: 6px;
				background: var(--brand);
				margin: 0 auto 24px auto;
				border-radius: 999px;
				opacity: 0.32;
			}

			.container {
				background: #fff;
				padding: 40px 50px;
				border-radius: 18px;
				box-shadow: var(--card-shadow);
				max-width: 600px;
				width: calc(100%% - 32px);
				text-align: center;
				border-top: 6px solid var(--brand);
				position: relative;
				z-index: 1;
				animation: fadeIn 0.7s ease;
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
				color: #fff;
				padding: 12px 16px;
				font-size: 16px;
				font-weight: 600;
				border-radius: 10px;
				cursor: pointer;
				transition: background 0.25s, transform 0.05s;
				white-space: nowrap;
			}
			.copy-btn:hover  { background-color: #e67600; }
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
				to   { opacity: 1; transform: translateY(0); }
			}

			/* Mobile */
			@media (max-width: 420px) {
				.password-box { font-size: 18px; }
				.copy-btn     { font-size: 14px; padding: 10px 14px; }
			}
		</style>
	</head>
	<body>
		<!-- === FOND SVG: Ballons + Couronnes stylées (inline, pas d’assets externes) === -->
		<svg class="bg-sports" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
			<!-- Défs de dégradés pour un trait plus chic -->
			<defs>
				<linearGradient id="gradBrand" x1="0" y1="0" x2="1" y2="1">
					<stop offset="0%%"  stop-color="#FF8C00" stop-opacity="0.9"/>
					<stop offset="100%%" stop-color="#FF8C00" stop-opacity="0.7"/>
				</linearGradient>
				<linearGradient id="gradGold" x1="0" y1="0" x2="1" y2="1">
					<stop offset="0%%"  stop-color="#F6D04D" stop-opacity="0.95"/>
					<stop offset="100%%" stop-color="#D9A400" stop-opacity="0.8"/>
				</linearGradient>

				<!-- Symbole ballon de basket propre -->
				<symbol id="basketball" viewBox="0 0 100 100">
					<circle cx="50" cy="50" r="44" stroke="url(#gradBrand)" stroke-width="2" fill="none"/>
					<path d="M6,50 H94 M50,6 V94" stroke="url(#gradBrand)" stroke-width="2" fill="none"/>
					<path d="M16,20 C40,40 60,60 84,80" stroke="url(#gradBrand)" stroke-width="2" fill="none"/>
					<path d="M84,20 C60,40 40,60 16,80" stroke="url(#gradBrand)" stroke-width="2" fill="none"/>
				</symbol>

				<!-- Symbole couronne stylée (arches + gemmes) -->
				<symbol id="crown" viewBox="0 0 120 100">
					<!-- base -->
					<path d="M10 72 H110 Q112 74 110 78 H10 Q8 74 10 72 Z" stroke="url(#gradGold)" stroke-width="2" fill="none"/>
					<!-- arches principales -->
					<path d="M10 72 L28 32 Q40 18 52 32 L60 46 L68 32 Q80 18 92 32 L110 72"
						  stroke="url(#gradGold)" stroke-width="2.4" fill="none" stroke-linejoin="round"/>
					<!-- pointes -->
					<circle cx="28" cy="32" r="3" fill="#FFD54A" stroke="url(#gradGold)" stroke-width="1"/>
					<circle cx="68" cy="32" r="3" fill="#FFD54A" stroke="url(#gradGold)" stroke-width="1"/>
					<circle cx="92" cy="32" r="3" fill="#FFD54A" stroke="url(#gradGold)" stroke-width="1"/>
					<!-- gemmes -->
					<rect x="54" y="52" width="12" height="10" rx="2" fill="#FFE37A" stroke="url(#gradGold)" stroke-width="1"/>
					<rect x="22" y="56" width="10" height="8"  rx="2" fill="#FFE37A" stroke="url(#gradGold)" stroke-width="1"/>
					<rect x="88" y="56" width="10" height="8"  rx="2" fill="#FFE37A" stroke="url(#gradGold)" stroke-width="1"/>
				</symbol>
			</defs>

			<!-- Instances positionnées (on joue sur transform/opacity pour du relief) -->
			<use href="#basketball" x="-6"  y="0"   width="36%%" height="36%%" style="transform:rotate(-8deg); transform-origin: center; opacity:0.9;"></use>
			<use href="#basketball" x="68%%" y="16%%" width="34%%" height="34%%" style="transform:rotate(12deg);  transform-origin: center; opacity:0.85;"></use>
			<use href="#basketball" x="10%%" y="62%%" width="30%%" height="30%%" style="transform:rotate(7deg);   transform-origin: center; opacity:0.9;"></use>

			<use href="#crown" x="28%%" y="10%%" width="28%%" height="28%%" style="transform:rotate(-6deg);  transform-origin: center; opacity:0.9;"></use>
			<use href="#crown" x="8%%"  y="48%%" width="24%%" height="24%%" style="transform:rotate(-12deg); transform-origin: center; opacity:0.9;"></use>
			<use href="#crown" x="68%%" y="58%%" width="26%%" height="26%%" style="transform:rotate(10deg);  transform-origin: center; opacity:0.9;"></use>
		</svg>

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
					navigator.clipboard.writeText(pwd).then(flashCopied).catch(function(){ fallbackCopy(pwd); });
				} else { fallbackCopy(pwd); }
			}
			function fallbackCopy(text) {
				var ta = document.createElement('textarea');
				ta.value = text; document.body.appendChild(ta); ta.select();
				try { document.execCommand('copy'); } catch(e) {}
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
				}, 1800);
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
