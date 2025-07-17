# PlayTheStreet - Backend (Go)

Ce dépôt contient le backend de l'application mobile **PlayTheStreet**, développé en Go.  
Il expose des API HTTP via AWS Lambda, avec une prise en charge de Swagger pour la documentation, des tests automatisés, et un environnement de développement Dockerisé.

## Pré-requis

- [Go 1.24](https://golang.org/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
- [GolangCI-Lint](https://golangci-lint.run/)
- [Swag (Swagger generator)](https://github.com/swaggo/swag)


## Commandes utiles (Makefile)

| Commande                | Description                                                      |
|-------------------------|------------------------------------------------------------------|
| `make build`            | Compile le backend dans le dossier `bin/`                        |
| `make docker-up`        | Démarre le service Docker utilisé pour les tests                 |
| `make docker-down`      | Arrête et supprime les conteneurs Docker de tests                |
| `make test`             | Lance les tests Go avec couverture                               |
| `make generate-swagger` | Génère la documentation Swagger à partir des commentaires Go     |
| `make lint`             | Analyse statique du code avec golangci-lint                      |
| `make zip-windows`      | Crée un zip `function.zip` pour déploiement AWS Lambda (Windows) |
| `make deploy-windows`   | Déploie le zip vers AWS Lambda (fonction `backend-go-lambda`)    |
| `make clean-windows`    | Nettoie le dossier `bin/` et les artefacts de build Windows      |

## Déploiement

Pour déployer la dernière version du backend sur AWS Lambda (sous Windows) :

```sh
make deploy-windows
```
## Documentation API

La documentation Swagger est générée avec la commande :

```sh
make generate-swagger
```

## Configuration (.env)

Le backend utilise un fichier `.env` à la racine pour stocker les variables sensibles et de configuration.

### Exemple de fichier `.env`

```env
# Connexion à la base de données (PostgreSQL, MySQL, etc.)
DATABASE_URL=postgres://user:password@localhost:5432/playthestreet?sslmode=disable

# Clé secrète pour la génération et la vérification des JWT
JWT_SECRET=changeme_supersecretjwtkey

# Configuration SMTP (envoi d'e-mails)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your_smtp_username
SMTP_PASSWORD=your_smtp_password
SMTP_FROM=no-reply@playthestreet.com
```