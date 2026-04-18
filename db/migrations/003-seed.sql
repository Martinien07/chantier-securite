-- Seed data

-- Roles
INSERT IGNORE INTO roles (id, name, description) VALUES
    (1, 'admin', 'Administration système'),
    (2, 'hse', 'Responsable sécurité'),
    (3, 'supervisor', 'Superviseur chantier'),
    (4, 'viewer', 'Consultation seule'),
    (5, 'auditor', 'Audit sécurité');

-- Users
INSERT IGNORE INTO users (id, username, email, password_hash, is_active) VALUES
    (1, 'admin', 'admin@safesite.com', '$2a$10$dummy', 1),
    (2, 'hse_manager', 'hse@safesite.com', '$2a$10$dummy', 1),
    (3, 'supervisor1', 'sup@safesite.com', '$2a$10$dummy', 1);

INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (1, 1), (2, 2), (3, 3);

-- Sites
INSERT IGNORE INTO sites (id, name, location, description) VALUES
    (1, 'Tour Horizon', 'Paris La Défense', 'Construction immeuble R+25'),
    (2, 'Centre Commercial', 'Lyon Part-Dieu', 'Rénovation centre commercial'),
    (3, 'Pont Urbain', 'Bordeaux', 'Construction pont piéton');

-- Plans
INSERT IGNORE INTO plans (id, site_id, level, image_path, scale_factor) VALUES
    (1, 1, 'RDC', '/static/plans/horizon_rdc.svg', 0.05),
    (2, 1, 'R+1', '/static/plans/horizon_r1.svg', 0.05),
    (3, 1, 'R+2', '/static/plans/horizon_r2.svg', 0.05),
    (4, 2, 'RDC', '/static/plans/centre_rdc.svg', 0.06);

-- Cameras (sans is_webcam - colonne absente dans cette base)
INSERT IGNORE INTO cameras (id, plan_id, name, stream_url, x_plan, y_plan, orientation, fov) VALUES
    (1, 1, 'CAM-A1', 'rtsp://192.168.1.10/stream1', 150, 200, 45, 90),
    (2, 1, 'CAM-A2', 'rtsp://192.168.1.11/stream1', 400, 150, 180, 120),
    (3, 1, 'CAM-B1', 'rtsp://192.168.1.12/stream1', 600, 300, 270, 90),
    (4, 2, 'CAM-C1', 'rtsp://192.168.1.13/stream1', 200, 250, 0, 100),
    (5, 1, 'Webcam Test', 'webcam://local', 300, 350, 90, 60);

-- Zones
INSERT IGNORE INTO zones (id, plan_id, name, type, polygon, risk_level, is_active) VALUES
    (1, 1, 'Zone Gros Oeuvre', 'construction', '[[100,100],[400,100],[400,300],[100,300]]', 'HIGH', 1),
    (2, 1, 'Zone Échafaudage', 'height', '[[450,50],[650,50],[650,250],[450,250]]', 'HIGH', 1),
    (3, 1, 'Zone Stockage', 'storage', '[[50,350],[250,350],[250,500],[50,500]]', 'LOW', 1),
    (4, 1, 'Circulation Engins', 'traffic', '[[300,320],[700,320],[700,380],[300,380]]', 'MEDIUM', 1),
    (5, 1, 'Base Vie', 'safe', '[[720,400],[900,400],[900,550],[720,550]]', 'LOW', 1);

-- HSE Rules
INSERT IGNORE INTO hse_rules (id, name, description, condition_logic, severity, is_active) VALUES
    (1, 'Absence de casque', 'Personne détectée sans casque de sécurité', 'no_hardhat AND person', 4, 1),
    (2, 'Absence de gilet', 'Personne sans gilet haute visibilité', 'no_vest AND person', 3, 1),
    (3, 'Travail en hauteur sans harnais', 'Risque de chute de hauteur', 'height_zone AND no_harness', 5, 1),
    (4, 'Proximité engin', 'Personne trop proche d''un engin en mouvement', 'machine_distance < 3', 5, 1),
    (5, 'Zone interdite', 'Intrusion dans zone dangereuse', 'forbidden_zone AND person', 4, 1);

-- Models
INSERT IGNORE INTO models (id, name, type, version, metrics, is_active) VALUES
    (1, 'YOLOv8-PPE', 'detection', 'v8n', '{"mAP": 0.89, "precision": 0.91}', 1),
    (2, 'YOLOv8-Person', 'detection', 'v8m', '{"mAP": 0.94, "precision": 0.96}', 1),
    (3, 'Activity-Classifier', 'classification', 'v1', '{"accuracy": 0.87}', 1);

-- Sample Risk Events and Alerts
INSERT IGNORE INTO risk_events (id, zone_id, rule_id, risk_score, risk_level, explanation, created_at) VALUES
    (1, 1, 1, 0.85, 'HIGH', 'Personne détectée sans casque dans zone gros oeuvre', DATE_SUB(NOW(), INTERVAL 10 MINUTE)),
    (2, 2, 3, 0.95, 'HIGH', 'Travailleur en hauteur sans équipement de protection', DATE_SUB(NOW(), INTERVAL 5 MINUTE)),
    (3, 4, 4, 0.78, 'HIGH', 'Proximité dangereuse entre piéton et chariot élévateur', DATE_SUB(NOW(), INTERVAL 2 MINUTE));

INSERT IGNORE INTO alerts (id, risk_event_id, alert_level, status, sent_at) VALUES
    (1, 1, 'HIGH', 'new', DATE_SUB(NOW(), INTERVAL 10 MINUTE)),
    (2, 2, 'HIGH', 'acknowledged', DATE_SUB(NOW(), INTERVAL 5 MINUTE)),
    (3, 3, 'HIGH', 'new', DATE_SUB(NOW(), INTERVAL 2 MINUTE));

INSERT IGNORE INTO migrations (migration_number, migration_name) VALUES (003, '003-seed');
