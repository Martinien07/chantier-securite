-- Securite chantier schema

-- Zones du chantier
CREATE TABLE IF NOT EXISTS zones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nom TEXT NOT NULL,
    description TEXT,
    niveau_risque TEXT NOT NULL DEFAULT 'moyen',
    coordonnees TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Cameras de surveillance
CREATE TABLE IF NOT EXISTS cameras (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nom TEXT NOT NULL,
    zone_id INTEGER REFERENCES zones(id),
    emplacement TEXT,
    flux_url TEXT,
    statut TEXT NOT NULL DEFAULT 'active',
    derniere_detection TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Types de risques detectables
CREATE TABLE IF NOT EXISTS types_risque (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    nom TEXT NOT NULL,
    description TEXT,
    severite INTEGER NOT NULL DEFAULT 5,
    couleur TEXT NOT NULL DEFAULT '#ff6b6b',
    icone TEXT
);

-- Alertes detectees
CREATE TABLE IF NOT EXISTS alertes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    camera_id INTEGER REFERENCES cameras(id),
    zone_id INTEGER REFERENCES zones(id),
    type_risque_id INTEGER REFERENCES types_risque(id),
    severite INTEGER NOT NULL,
    description TEXT NOT NULL,
    details_ia TEXT,
    confiance REAL NOT NULL,
    image_url TEXT,
    video_clip_url TEXT,
    statut TEXT NOT NULL DEFAULT 'active',
    acknowledged_at TIMESTAMP,
    acknowledged_by TEXT,
    resolved_at TIMESTAMP,
    resolved_by TEXT,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Detections brutes
CREATE TABLE IF NOT EXISTS detections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    camera_id INTEGER REFERENCES cameras(id),
    type_objet TEXT NOT NULL,
    confiance REAL NOT NULL,
    bbox TEXT,
    attributs TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Statistiques par periode
CREATE TABLE IF NOT EXISTS stats_periode (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date DATE NOT NULL,
    zone_id INTEGER REFERENCES zones(id),
    nb_alertes INTEGER NOT NULL DEFAULT 0,
    nb_detections INTEGER NOT NULL DEFAULT 0,
    nb_personnes_max INTEGER NOT NULL DEFAULT 0,
    taux_conformite_epi REAL,
    incidents_evites INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Migration record
INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (002, '002-securite');
