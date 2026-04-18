Système Intégré de Surveillance HSE - Chantier Sécurité

Ce projet est une plateforme full-stack combinant une Intelligence Artificielle de pointe et une Interface Web temps réel pour la gestion de la sécurité sur les chantiers. Il permet de détecter les risques (EPI, zones interdites) et de diffuser les flux vidéos via un tableau de bord interactif.

# 🌐 Interface de Gestion HSE - Chantier Sécurité

Serveur de supervision développé en **Go (Golang)** permettant de visualiser les alertes, de gérer le parc de caméras et de visionner les flux en direct.

## 📂 Organisation du projet

### 🚀 Serveur Core (`cmd/srv/`)
- **`main.go`** : Point d'entrée du serveur. Initialise les routes et la connexion à la base de données.
- **`templates/`** : Pages HTML dynamiques (Tableau de bord, configuration des zones, historique).
- **`static/`** : Fichiers JavaScript et CSS. Contient également les plans de chantier uploadés par l'utilisateur.

### 🗄️ Base de Données (`db/`)
- **`migrations/`** : Scripts de création des tables pour MariaDB.
- **`queries/`** : Requêtes SQL pour récupérer les statistiques de sécurité et les dernières alertes.

### 🎥 Streaming MJPEG (`srv/`)
Le serveur implémente un endpoint de streaming qui surveille les fichiers générés par l'IA. 
- Il lit `data/frames/cam_{id}.jpg` en boucle et l'envoie au format `multipart/x-mixed-replace` pour un affichage fluide en direct.

## 🛠️ Installation
1. S'assurer que Go est installé (v1.20+).
2. Installer les modules : `go mod tidy`.
3. Lancer le serveur : `go run cmd/srv/main.go`.

## ⚙️ Configuration SQL
Créez l'utilisateur avec les droits nécessaires pour permettre au serveur d'écrire les logs d'alertes :
```sql
CREATE USER 'admin'@'localhost' IDENTIFIED BY 'admin';
GRANT ALL PRIVILEGES ON *.* TO 'admin'@'localhost' WITH GRANT OPTION;