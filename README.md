# Ichiraku Shop

Boutique en ligne de commande en Go : un serveur HTTP, une interface client et une interface admin, utilisables en local ou en SSH.

## Prérequis

- Docker
- Un client SSH

Go n'a pas besoin d'être installé : tout tourne dans Docker.

## Configuration

```bash
cp .env.example .env
```

Remplir `.env` :

```env
POSTGRES_USER=ichiraku
POSTGRES_PASSWORD=change-moi
POSTGRES_DB=ichiraku
API_URL=http://localhost:8080
ADMIN_EMAIL=admin@test.fr
ADMIN_PASSWORD=change-moi
JWT_SECRET=au-moins-32-caracteres-aleatoires
MAIL_HOST=mailpit
MAIL_PORT=1025
MAIL_FROM=no-reply@ichiraku.shop
```

Générer `JWT_SECRET` : `openssl rand -base64 48`

## Lancement

```bash
docker compose up -d --build
docker compose exec go go build -o bin/ ./cmd/...
docker compose exec go ./bin/server
```

Laisser le serveur tourner. Dans un autre terminal, charger les produits de démo :

```bash
docker compose cp seed.sql db:/seed.sql
docker compose exec db psql -U ichiraku -d ichiraku -f /seed.sql
```

## Interfaces en local

```bash
docker compose exec go ./bin/client
docker compose exec go ./bin/admin
```

## Interfaces en SSH

Démarrer les serveurs SSH, chacun dans un terminal :

```bash
docker compose exec go ./bin/client-ssh
docker compose exec go ./bin/admin-ssh
```

Se connecter :

```bash
ssh -p 23234 localhost   # client
ssh -p 23235 localhost   # admin
```

L'admin se connecte avec `ADMIN_EMAIL` et `ADMIN_PASSWORD`.

## Emails

Les codes de confirmation et de réinitialisation arrivent dans Mailpit : http://localhost:8025

## À savoir

- Mot de passe : 8 à 72 caractères. Le compte doit être confirmé par code avant la première connexion.
- Code : 6 chiffres, valable 15 minutes, 5 essais maximum.
- 5 échecs de connexion bloquent le compte 15 minutes.
- Carte de test : `4242 4242 4242 4242`, expiration `12/30`, CVC `123`.
- TVA 20 %, livraison gratuite.

## Tests avec coverage

```bash
docker compose exec go go test ./internal/domain/... -cover
```

## Remise à zéro

```bash
docker compose down -v
```

Supprime la base et les emails. Relancer ensuite à partir de « Lancement ».