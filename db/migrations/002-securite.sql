-- HSE Security Platform Schema

-- Roles
CREATE TABLE IF NOT EXISTS roles (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Users
CREATE TABLE IF NOT EXISTS users (
    id            INT PRIMARY KEY AUTO_INCREMENT,
    username      VARCHAR(100) NOT NULL UNIQUE,
    email         VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_active     TINYINT DEFAULT 1,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login    TIMESTAMP NULL
);

-- User Roles
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- Sites
CREATE TABLE IF NOT EXISTS sites (
    id                INT PRIMARY KEY AUTO_INCREMENT,
    name              VARCHAR(100) NOT NULL,
    location          VARCHAR(150),
    description       TEXT,
    confidence_config TEXT,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Plans (floor plans)
CREATE TABLE IF NOT EXISTS plans (
    id           INT PRIMARY KEY AUTO_INCREMENT,
    site_id      INT,
    level        VARCHAR(50),
    image_path   TEXT,
    scale_factor FLOAT CHECK (scale_factor > 0),
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (site_id) REFERENCES sites(id)
);

-- Cameras
CREATE TABLE IF NOT EXISTS cameras (
    id                 INT PRIMARY KEY AUTO_INCREMENT,
    plan_id            INT,
    name               VARCHAR(100),
    stream_url         TEXT,
    x_plan             FLOAT,
    y_plan             FLOAT,
    orientation        FLOAT,
    fov                FLOAT,
    calibration_matrix TEXT,
    confidence_config  TEXT,
    is_webcam          TINYINT DEFAULT 0,
    created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plan_id) REFERENCES plans(id)
);

-- Camera Calibrations
CREATE TABLE IF NOT EXISTS camera_calibrations (
    id                INT PRIMARY KEY AUTO_INCREMENT,
    camera_id         INT NOT NULL,
    plan_id           INT NOT NULL,
    pts_image         TEXT NOT NULL,
    pts_plan          TEXT NOT NULL,
    homography        TEXT NOT NULL,
    reprojection_error FLOAT,
    is_active         TINYINT DEFAULT 1,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(camera_id, plan_id),
    FOREIGN KEY (camera_id) REFERENCES cameras(id),
    FOREIGN KEY (plan_id) REFERENCES plans(id)
);

-- Zones
CREATE TABLE IF NOT EXISTS zones (
    id         INT PRIMARY KEY AUTO_INCREMENT,
    plan_id    INT,
    name       VARCHAR(100),
    type       VARCHAR(50),
    polygon    TEXT NOT NULL,
    risk_level VARCHAR(10) CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH')),
    is_active  TINYINT DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plan_id) REFERENCES plans(id)
);

-- HSE Rules
CREATE TABLE IF NOT EXISTS hse_rules (
    id              INT PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(100),
    description     TEXT,
    condition_logic TEXT,
    severity        INT CHECK (severity BETWEEN 1 AND 5),
    is_active       TINYINT DEFAULT 1
);

-- Models (AI/ML)
CREATE TABLE IF NOT EXISTS models (
    id         INT PRIMARY KEY AUTO_INCREMENT,
    name       VARCHAR(100),
    type       VARCHAR(50),
    version    VARCHAR(20),
    trained_at TIMESTAMP NULL,
    metrics    TEXT,
    is_active  TINYINT DEFAULT 1
);

-- Detections
CREATE TABLE IF NOT EXISTS detections (
    id           INT PRIMARY KEY AUTO_INCREMENT,
    camera_id    INT,
    timestamp    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    object_class VARCHAR(50),
    confidence   FLOAT CHECK (confidence BETWEEN 0 AND 1),
    bbox_x       FLOAT,
    bbox_y       FLOAT,
    bbox_w       FLOAT,
    bbox_h       FLOAT,
    track_id     INT,
    FOREIGN KEY (camera_id) REFERENCES cameras(id)
);

-- Activity Windows
CREATE TABLE IF NOT EXISTS activity_windows (
    id         INT PRIMARY KEY AUTO_INCREMENT,
    camera_id  INT,
    start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_time   TIMESTAMP NULL,
    duration   INT,
    FOREIGN KEY (camera_id) REFERENCES cameras(id)
);

-- Activities
CREATE TABLE IF NOT EXISTS activities (
    id         INT PRIMARY KEY AUTO_INCREMENT,
    window_id  INT,
    label      VARCHAR(50),
    confidence FLOAT,
    model_id   INT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (window_id) REFERENCES activity_windows(id),
    FOREIGN KEY (model_id) REFERENCES models(id)
);

-- Activity Features
CREATE TABLE IF NOT EXISTS activity_features (
    window_id     INT PRIMARY KEY,
    features_json TEXT,
    FOREIGN KEY (window_id) REFERENCES activity_windows(id)
);

-- Risk Events
CREATE TABLE IF NOT EXISTS risk_events (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    window_id   INT,
    zone_id     INT,
    rule_id     INT,
    risk_score  FLOAT,
    risk_level  VARCHAR(10),
    explanation TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (window_id) REFERENCES activity_windows(id),
    FOREIGN KEY (zone_id) REFERENCES zones(id),
    FOREIGN KEY (rule_id) REFERENCES hse_rules(id)
);

-- Alerts
CREATE TABLE IF NOT EXISTS alerts (
    id              INT PRIMARY KEY AUTO_INCREMENT,
    risk_event_id   INT,
    alert_level     VARCHAR(20),
    status          VARCHAR(20) DEFAULT 'new',
    sent_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP NULL,
    camera_id       INT,
    FOREIGN KEY (risk_event_id) REFERENCES risk_events(id),
    FOREIGN KEY (camera_id) REFERENCES cameras(id)
);

-- Human Decisions
CREATE TABLE IF NOT EXISTS human_decisions (
    id         INT PRIMARY KEY AUTO_INCREMENT,
    alert_id   INT,
    user_id    INT,
    action     VARCHAR(100),
    comment    TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (alert_id) REFERENCES alerts(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Incident Reviews
CREATE TABLE IF NOT EXISTS incident_reviews (
    id               INT PRIMARY KEY AUTO_INCREMENT,
    alert_id         INT,
    reviewer_id      INT,
    comment          TEXT,
    decision_summary TEXT,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (alert_id) REFERENCES alerts(id),
    FOREIGN KEY (reviewer_id) REFERENCES users(id)
);

-- Person Zone Events
CREATE TABLE IF NOT EXISTS person_zone_events (
    id                INT PRIMARY KEY AUTO_INCREMENT,
    person_track_id   INT,
    zone_id           INT,
    entry_time        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    exit_time         TIMESTAMP NULL,
    exposure_duration INT,
    FOREIGN KEY (zone_id) REFERENCES zones(id)
);

INSERT IGNORE INTO migrations (migration_number, migration_name) VALUES (002, '002-securite');
