# Migration vers MySQL

## Comparaison SQLite vs MySQL

| Aspect | SQLite | MySQL |
|--------|--------|-------|
| **Installation** | Aucune (fichier) | Serveur requis |
| **Performance** | Bonne pour < 100k lignes | Excellente |
| **Concurrence** | Limitée | Excellente |
| **Backup** | Copier le fichier | mysqldump |
| **Hébergement** | Local uniquement | Local ou distant |

## Prérequis

- MySQL 8.0+ ou MariaDB 10.5+
- Go 1.21+

## Étapes de migration

### 1. Installer MySQL

```bash
# Ubuntu/Debian
sudo apt install mysql-server

# macOS
brew install mysql

# Windows
# Télécharger depuis https://dev.mysql.com/downloads/
```

### 2. Créer la base de données

```bash
mysql -u root -p
```

```sql
CREATE DATABASE safesite CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'safesite'@'localhost' IDENTIFIED BY 'votre_mot_de_passe';
GRANT ALL PRIVILEGES ON safesite.* TO 'safesite'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 3. Exécuter le schéma

```bash
mysql -u safesite -p safesite < db/mysql/schema.sql
```

### 4. Générer le code Go pour MySQL

```bash
cd db/mysql
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate
```

### 5. Modifier le code Go

Remplacer dans `cmd/srv/main.go` :

```go
// Avant (SQLite)
import "srv.exe.dev/db"

database, err := db.Open("db.sqlite3")

// Après (MySQL)
import "github.com/go-sql-driver/mysql"

dsn := "safesite:votre_mot_de_passe@tcp(127.0.0.1:3306)/safesite?parseTime=true"
database, err := sql.Open("mysql", dsn)
```

### 6. Installer le driver MySQL

```bash
go get github.com/go-sql-driver/mysql
```

### 7. Variables d'environnement (recommandé)

```bash
export DB_HOST=127.0.0.1
export DB_PORT=3306
export DB_USER=safesite
export DB_PASSWORD=votre_mot_de_passe
export DB_NAME=safesite
```

## Migration des données

Pour migrer les données de SQLite vers MySQL :

```bash
# Exporter de SQLite
sqlite3 db.sqlite3 ".mode insert" ".output data.sql" "SELECT * FROM sites;"
sqlite3 db.sqlite3 ".mode insert" ".output -" "SELECT * FROM plans;" >> data.sql
# etc.

# Importer dans MySQL
mysql -u safesite -p safesite < data.sql
```

Ou utilisez un outil comme [pgloader](https://pgloader.io/) qui supporte SQLite -> MySQL.

## Différences de syntaxe SQL

| SQLite | MySQL |
|--------|-------|
| `AUTOINCREMENT` | `AUTO_INCREMENT` |
| `INTEGER` | `BIGINT UNSIGNED` |
| `TEXT` | `TEXT` ou `VARCHAR(n)` |
| `CURRENT_TIMESTAMP` | `CURRENT_TIMESTAMP` |
| `json_valid()` | Native JSON support |

## Configuration SQLC

Le fichier `db/mysql/sqlc.yaml` est configuré pour MySQL :

```yaml
version: "2"
sql:
  - engine: "mysql"
    queries: "queries.sql"
    schema: "schema.sql"
    gen:
      go:
        package: "dbmysql"
        out: "dbmysql/"
```

## Déploiement avec Docker

```yaml
# docker-compose.yml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpass
      MYSQL_DATABASE: safesite
      MYSQL_USER: safesite
      MYSQL_PASSWORD: safesitepass
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./db/mysql/schema.sql:/docker-entrypoint-initdb.d/schema.sql

  app:
    build: .
    environment:
      DB_HOST: mysql
      DB_PORT: 3306
      DB_USER: safesite
      DB_PASSWORD: safesitepass
      DB_NAME: safesite
    ports:
      - "8000:8000"
    depends_on:
      - mysql

volumes:
  mysql_data:
```

## Performance

Pour optimiser MySQL :

```sql
-- Index recommandés (déjà dans schema.sql)
CREATE INDEX idx_camera_timestamp ON detections(camera_id, timestamp);
CREATE INDEX idx_status_level ON alerts(status, alert_level);
```

## Backup

```bash
# Backup
mysqldump -u safesite -p safesite > backup.sql

# Restore
mysql -u safesite -p safesite < backup.sql
```
