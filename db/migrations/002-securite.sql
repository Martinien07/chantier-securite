-- HSE Security Platform Schema

-- Roles
CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Users
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_active INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);

-- User Roles
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- Sites
CREATE TABLE IF NOT EXISTS sites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    location TEXT,
    description TEXT,
    confidence_config TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Plans (floor plans)
CREATE TABLE IF NOT EXISTS plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id INTEGER,
    level TEXT,
    image_path TEXT,
    scale_factor REAL CHECK (scale_factor > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (site_id) REFERENCES sites(id)
);

-- Cameras
CREATE TABLE IF NOT EXISTS cameras (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plan_id INTEGER,
    name TEXT,
    stream_url TEXT,
    x_plan REAL,
    y_plan REAL,
    orientation REAL,
    fov REAL,
    calibration_matrix TEXT,
    confidence_config TEXT,
    is_webcam INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plan_id) REFERENCES plans(id)
);

-- Camera Calibrations
CREATE TABLE IF NOT EXISTS camera_calibrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    camera_id INTEGER NOT NULL,
    plan_id INTEGER NOT NULL,
    pts_image TEXT NOT NULL,
    pts_plan TEXT NOT NULL,
    homography TEXT NOT NULL,
    reprojection_error REAL,
    is_active INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(camera_id, plan_id),
    FOREIGN KEY (camera_id) REFERENCES cameras(id),
    FOREIGN KEY (plan_id) REFERENCES plans(id)
);

-- Zones
CREATE TABLE IF NOT EXISTS zones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plan_id INTEGER,
    name TEXT,
    type TEXT,
    polygon TEXT NOT NULL,
    risk_level TEXT CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH')),
    is_active INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plan_id) REFERENCES plans(id)
);

-- HSE Rules
CREATE TABLE IF NOT EXISTS hse_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    description TEXT,
    condition_logic TEXT,
    severity INTEGER CHECK (severity BETWEEN 1 AND 5),
    is_active INTEGER DEFAULT 1
);

-- Models (AI/ML)
CREATE TABLE IF NOT EXISTS models (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    type TEXT,
    version TEXT,
    trained_at TIMESTAMP,
    metrics TEXT,
    is_active INTEGER DEFAULT 1
);

-- Detections
CREATE TABLE IF NOT EXISTS detections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    camera_id INTEGER,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    object_class TEXT,
    confidence REAL CHECK (confidence BETWEEN 0 AND 1),
    bbox_x REAL,
    bbox_y REAL,
    bbox_w REAL,
    bbox_h REAL,
    track_id INTEGER,
    FOREIGN KEY (camera_id) REFERENCES cameras(id)
);

-- Activity Windows
CREATE TABLE IF NOT EXISTS activity_windows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    camera_id INTEGER,
    start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP,
    duration INTEGER,
    FOREIGN KEY (camera_id) REFERENCES cameras(id)
);

-- Activities
CREATE TABLE IF NOT EXISTS activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    window_id INTEGER,
    label TEXT,
    confidence REAL,
    model_id INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (window_id) REFERENCES activity_windows(id),
    FOREIGN KEY (model_id) REFERENCES models(id)
);

-- Activity Features
CREATE TABLE IF NOT EXISTS activity_features (
    window_id INTEGER PRIMARY KEY,
    features_json TEXT,
    FOREIGN KEY (window_id) REFERENCES activity_windows(id)
);

-- Risk Events
CREATE TABLE IF NOT EXISTS risk_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    window_id INTEGER,
    zone_id INTEGER,
    rule_id INTEGER,
    risk_score REAL,
    risk_level TEXT,
    explanation TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (window_id) REFERENCES activity_windows(id),
    FOREIGN KEY (zone_id) REFERENCES zones(id),
    FOREIGN KEY (rule_id) REFERENCES hse_rules(id)
);

-- Alerts
CREATE TABLE IF NOT EXISTS alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    risk_event_id INTEGER,
    alert_level TEXT,
    status TEXT DEFAULT 'new',
    sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP,
    FOREIGN KEY (risk_event_id) REFERENCES risk_events(id)
);

-- Human Decisions
CREATE TABLE IF NOT EXISTS human_decisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    alert_id INTEGER,
    user_id INTEGER,
    action TEXT,
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (alert_id) REFERENCES alerts(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Incident Reviews
CREATE TABLE IF NOT EXISTS incident_reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    alert_id INTEGER,
    reviewer_id INTEGER,
    comment TEXT,
    decision_summary TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (alert_id) REFERENCES alerts(id),
    FOREIGN KEY (reviewer_id) REFERENCES users(id)
);

-- Person Zone Events
CREATE TABLE IF NOT EXISTS person_zone_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_track_id INTEGER,
    zone_id INTEGER,
    entry_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    exit_time TIMESTAMP,
    exposure_duration INTEGER,
    FOREIGN KEY (zone_id) REFERENCES zones(id)
);

INSERT OR IGNORE INTO migrations (migration_number, migration_name) VALUES (002, '002-securite');
