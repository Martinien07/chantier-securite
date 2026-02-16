-- SafeSite AI - MySQL Schema
-- Compatible avec MySQL 8.0+

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

-- --------------------------------------------------------
-- Roles
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `roles` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL UNIQUE,
    `description` TEXT,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Users
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(50) NOT NULL UNIQUE,
    `email` VARCHAR(100) NOT NULL UNIQUE,
    `password_hash` TEXT NOT NULL,
    `is_active` TINYINT(1) DEFAULT 1,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `last_login` TIMESTAMP NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- User Roles
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `user_roles` (
    `user_id` BIGINT UNSIGNED NOT NULL,
    `role_id` BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`user_id`, `role_id`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Sites (Chantiers)
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `sites` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL,
    `location` VARCHAR(150),
    `description` TEXT,
    `confidence_config` JSON,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Plans (Niveaux)
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `plans` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `site_id` BIGINT UNSIGNED,
    `level` VARCHAR(50),
    `image_path` TEXT,
    `scale_factor` FLOAT DEFAULT 0.05,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`site_id`) REFERENCES `sites`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Cameras
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `cameras` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `plan_id` BIGINT UNSIGNED,
    `name` VARCHAR(100),
    `stream_url` TEXT,
    `x_plan` FLOAT,
    `y_plan` FLOAT,
    `orientation` FLOAT DEFAULT 0,
    `fov` FLOAT DEFAULT 90,
    `calibration_matrix` JSON,
    `confidence_config` JSON,
    `is_webcam` TINYINT(1) DEFAULT 0,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`plan_id`) REFERENCES `plans`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Camera Calibrations
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `camera_calibrations` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `camera_id` BIGINT UNSIGNED NOT NULL,
    `plan_id` BIGINT UNSIGNED NOT NULL,
    `pts_image` JSON NOT NULL,
    `pts_plan` JSON NOT NULL,
    `homography` JSON NOT NULL,
    `reprojection_error` FLOAT,
    `is_active` TINYINT(1) DEFAULT 1,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `camera_plan` (`camera_id`, `plan_id`),
    FOREIGN KEY (`camera_id`) REFERENCES `cameras`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`plan_id`) REFERENCES `plans`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Zones
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `zones` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `plan_id` BIGINT UNSIGNED,
    `name` VARCHAR(100),
    `type` VARCHAR(50),
    `polygon` JSON NOT NULL,
    `risk_level` ENUM('LOW', 'MEDIUM', 'HIGH') DEFAULT 'MEDIUM',
    `is_active` TINYINT(1) DEFAULT 1,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`plan_id`) REFERENCES `plans`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- HSE Rules
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `hse_rules` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(100),
    `description` TEXT,
    `condition_logic` TEXT,
    `severity` TINYINT CHECK (`severity` BETWEEN 1 AND 5),
    `is_active` TINYINT(1) DEFAULT 1,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Models (IA)
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `models` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(100),
    `type` VARCHAR(50),
    `version` VARCHAR(20),
    `trained_at` TIMESTAMP,
    `metrics` JSON,
    `is_active` TINYINT(1) DEFAULT 1,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Detections
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `detections` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `camera_id` BIGINT UNSIGNED,
    `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `object_class` VARCHAR(50),
    `confidence` FLOAT CHECK (`confidence` BETWEEN 0 AND 1),
    `bbox_x` FLOAT,
    `bbox_y` FLOAT,
    `bbox_w` FLOAT,
    `bbox_h` FLOAT,
    `track_id` INT,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`camera_id`) REFERENCES `cameras`(`id`) ON DELETE SET NULL,
    INDEX `idx_camera_timestamp` (`camera_id`, `timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Activity Windows
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `activity_windows` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `camera_id` BIGINT UNSIGNED,
    `start_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `end_time` TIMESTAMP NULL,
    `duration` INT,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`camera_id`) REFERENCES `cameras`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Activities
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `activities` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `window_id` BIGINT UNSIGNED,
    `label` VARCHAR(50),
    `confidence` FLOAT,
    `model_id` BIGINT UNSIGNED,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`window_id`) REFERENCES `activity_windows`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`model_id`) REFERENCES `models`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Activity Features
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `activity_features` (
    `window_id` BIGINT UNSIGNED NOT NULL,
    `features_json` JSON,
    PRIMARY KEY (`window_id`),
    FOREIGN KEY (`window_id`) REFERENCES `activity_windows`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Risk Events
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `risk_events` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `window_id` BIGINT UNSIGNED,
    `zone_id` BIGINT UNSIGNED,
    `rule_id` BIGINT UNSIGNED,
    `risk_score` FLOAT,
    `risk_level` ENUM('LOW', 'MEDIUM', 'HIGH'),
    `explanation` TEXT,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`window_id`) REFERENCES `activity_windows`(`id`) ON DELETE SET NULL,
    FOREIGN KEY (`zone_id`) REFERENCES `zones`(`id`) ON DELETE SET NULL,
    FOREIGN KEY (`rule_id`) REFERENCES `hse_rules`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Alerts
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `alerts` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `risk_event_id` BIGINT UNSIGNED,
    `camera_id` BIGINT UNSIGNED,
    `alert_level` ENUM('LOW', 'MEDIUM', 'HIGH'),
    `status` ENUM('new', 'acknowledged', 'closed') DEFAULT 'new',
    `sent_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `acknowledged_at` TIMESTAMP NULL,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`risk_event_id`) REFERENCES `risk_events`(`id`) ON DELETE SET NULL,
    FOREIGN KEY (`camera_id`) REFERENCES `cameras`(`id`) ON DELETE SET NULL,
    INDEX `idx_status_level` (`status`, `alert_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Human Decisions
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `human_decisions` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `alert_id` BIGINT UNSIGNED,
    `user_id` BIGINT UNSIGNED,
    `action` VARCHAR(100),
    `comment` TEXT,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`alert_id`) REFERENCES `alerts`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Incident Reviews
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `incident_reviews` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `alert_id` BIGINT UNSIGNED,
    `reviewer_id` BIGINT UNSIGNED,
    `comment` TEXT,
    `decision_summary` TEXT,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`alert_id`) REFERENCES `alerts`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`reviewer_id`) REFERENCES `users`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- Person Zone Events
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `person_zone_events` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `person_track_id` INT,
    `zone_id` BIGINT UNSIGNED,
    `entry_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `exit_time` TIMESTAMP NULL,
    `exposure_duration` INT,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`zone_id`) REFERENCES `zones`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
