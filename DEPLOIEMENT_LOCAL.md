# Déploiement Local de SafeSite AI

## Prérequis

- **Go 1.21+** : https://go.dev/dl/
- **Git** : Pour cloner le repository
- **SQLite3** (optionnel) : Pour visualiser la base de données

## Installation

### 1. Cloner le repository

```bash
git clone <url-du-repository> chantier-securite
cd chantier-securite
```

### 2. Compiler l'application

```bash
# Avec Make
make build

# Ou directement avec Go
go build -o safesite ./cmd/srv/
```

### 3. Lancer l'application

```bash
./safesite
```

L'application démarre sur **http://localhost:8000**

## Structure du projet

```
chantier-securite/
├── cmd/srv/           # Point d'entrée (main.go)
├── srv/               # Serveur HTTP
│   ├── server.go      # Configuration routes
│   ├── handlers.go    # Handlers pages
│   ├── handlers_admin.go
│   ├── handlers_api.go
│   ├── templates/     # Templates HTML
│   └── static/        # CSS, JS, images plans
├── db/
│   ├── db.go          # Connexion SQLite + migrations
│   ├── migrations/    # Schéma SQL
│   ├── queries/       # Requêtes SQLC
│   └── dbgen/         # Code Go généré
├── Makefile
├── go.mod
└── srv.service        # Service systemd (production)
```

## Configuration

### Port d'écoute

```bash
./safesite -listen :3000   # Écouter sur le port 3000
```

### Base de données

La base SQLite est créée automatiquement (`db.sqlite3`).
Les migrations s'exécutent au démarrage.

## Développement

### Modifier le schéma SQL

1. Éditer `db/migrations/002-securite.sql`
2. Ajouter des requêtes dans `db/queries/securite.sql`
3. Régénérer le code :

```bash
cd db && go generate
```

### Modifier les templates

Les templates sont dans `srv/templates/`.
Recompiler après modification :

```bash
go build -o safesite ./cmd/srv/ && ./safesite
```

### Hot reload (développement)

Installer [air](https://github.com/cosmtrek/air) :

```bash
go install github.com/cosmtrek/air@latest
air
```

## API REST

### Endpoints principaux

| Méthode | URL | Description |
|---------|-----|-------------|
| GET | `/api/stats` | Statistiques du site |
| GET | `/api/plans` | Liste des plans |
| GET | `/api/plans/{id}/data` | Données plan (caméras, zones) |
| GET | `/api/alerts/active` | Alertes actives |
| PUT | `/api/alerts/{id}` | Modifier une alerte |
| POST | `/api/admin/cameras` | Créer une caméra |
| PUT | `/api/admin/cameras/{id}` | Modifier une caméra |

## Déploiement Production (systemd)

```bash
# Copier le fichier service
sudo cp srv.service /etc/systemd/system/safesite.service

# Activer et démarrer
sudo systemctl daemon-reload
sudo systemctl enable safesite
sudo systemctl start safesite

# Vérifier
systemctl status safesite
```

## Support Webcam

Les webcams locales fonctionnent via l'API `getUserMedia` du navigateur.
Le serveur n'a pas accès aux webcams - tout se passe côté client.

Pour tester : créer une caméra avec `is_webcam = 1` et `stream_url = 'webcam:local'`.

## Dépannage

### Port déjà utilisé

```bash
killall safesite
./safesite -listen :8001
```

### Réinitialiser la base de données

```bash
rm -f db.sqlite3*
./safesite
```

### Logs

```bash
journalctl -u safesite -f
```
