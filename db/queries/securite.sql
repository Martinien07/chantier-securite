-- Stats
-- name: CountActiveAlerts :one
SELECT COUNT(*) FROM alerts WHERE status = 'new';

-- name: CountActiveAlertsBySite :one
SELECT COUNT(*) FROM alerts a
JOIN risk_events re ON a.risk_event_id = re.id
JOIN zones z ON re.zone_id = z.id
JOIN plans p ON z.plan_id = p.id
WHERE a.status = 'new' AND p.site_id = ?;

-- name: CountAlertsToday :one
SELECT COUNT(*) FROM alerts WHERE date(sent_at) = date('now');

-- name: CountActiveCameras :one
SELECT COUNT(*) FROM cameras;

-- name: CountCamerasBySite :one
SELECT COUNT(*) FROM cameras c
JOIN plans p ON c.plan_id = p.id
WHERE p.site_id = ?;

-- name: CountHighRiskZones :one
SELECT COUNT(*) FROM zones WHERE risk_level = 'HIGH' AND is_active = 1;

-- name: CountHighRiskZonesBySite :one
SELECT COUNT(*) FROM zones z
JOIN plans p ON z.plan_id = p.id
WHERE z.risk_level = 'HIGH' AND z.is_active = 1 AND p.site_id = ?;

-- name: CountDetectionsToday :one
SELECT COUNT(*) FROM detections WHERE date(timestamp) = date('now');

-- name: CountDetectionsTodayBySite :one
SELECT COUNT(*) FROM detections d
JOIN cameras c ON d.camera_id = c.id
JOIN plans p ON c.plan_id = p.id
WHERE date(d.timestamp) = date('now') AND p.site_id = ?;

-- Sites
-- name: ListSites :many
SELECT * FROM sites ORDER BY name;

-- name: GetSite :one
SELECT * FROM sites WHERE id = ?;

-- name: SearchSites :many
SELECT * FROM sites WHERE name LIKE ? OR location LIKE ? ORDER BY name LIMIT 20;

-- name: CreateSite :one
INSERT INTO sites (name, location, description) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateSite :exec
UPDATE sites SET name = ?, location = ?, description = ? WHERE id = ?;

-- name: DeleteSite :exec
DELETE FROM sites WHERE id = ?;

-- Plans
-- name: ListPlans :many
SELECT p.*, s.name as site_name FROM plans p
JOIN sites s ON p.site_id = s.id
ORDER BY s.name, p.level;

-- name: ListPlansBySite :many
SELECT * FROM plans WHERE site_id = ? ORDER BY level;

-- name: GetPlan :one
SELECT * FROM plans WHERE id = ?;

-- name: CreatePlan :one
INSERT INTO plans (site_id, level, image_path, scale_factor) VALUES (?, ?, ?, ?) RETURNING *;

-- name: UpdatePlan :exec
UPDATE plans SET level = ?, image_path = ?, scale_factor = ? WHERE id = ?;

-- name: DeletePlan :exec
DELETE FROM plans WHERE id = ?;

-- Cameras
-- name: ListCameras :many
SELECT c.*, p.level as plan_level, s.name as site_name FROM cameras c
LEFT JOIN plans p ON c.plan_id = p.id
LEFT JOIN sites s ON p.site_id = s.id
ORDER BY s.name, c.name;

-- name: ListCamerasBySite :many
SELECT c.*, p.level as plan_level FROM cameras c
JOIN plans p ON c.plan_id = p.id
WHERE p.site_id = ?
ORDER BY c.name;

-- name: ListCamerasByPlan :many
SELECT * FROM cameras WHERE plan_id = ?;

-- name: GetCamera :one
SELECT * FROM cameras WHERE id = ?;

-- name: CreateCamera :one
INSERT INTO cameras (plan_id, name, stream_url, x_plan, y_plan, orientation, fov, is_webcam)
VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateCamera :exec
UPDATE cameras SET name = ?, stream_url = ?, x_plan = ?, y_plan = ?, orientation = ?, fov = ?, is_webcam = ? WHERE id = ?;

-- name: DeleteCamera :exec
DELETE FROM cameras WHERE id = ?;

-- Zones
-- name: ListZones :many
SELECT z.*, p.level as plan_level, s.name as site_name FROM zones z
LEFT JOIN plans p ON z.plan_id = p.id
LEFT JOIN sites s ON p.site_id = s.id
ORDER BY s.name, z.name;

-- name: ListZonesBySite :many
SELECT z.*, p.level as plan_level FROM zones z
JOIN plans p ON z.plan_id = p.id
WHERE p.site_id = ?
ORDER BY z.name;

-- name: ListZonesByPlan :many
SELECT * FROM zones WHERE plan_id = ?;

-- name: GetZone :one
SELECT * FROM zones WHERE id = ?;

-- name: CreateZone :one
INSERT INTO zones (plan_id, name, type, polygon, risk_level, is_active)
VALUES (?, ?, ?, ?, ?, 1) RETURNING *;

-- name: UpdateZone :exec
UPDATE zones SET name = ?, type = ?, polygon = ?, risk_level = ?, is_active = ? WHERE id = ?;

-- name: DeleteZone :exec
DELETE FROM zones WHERE id = ?;

-- Alerts
-- name: ListAlerts :many
SELECT a.*, re.explanation, z.name as zone_name FROM alerts a
LEFT JOIN risk_events re ON a.risk_event_id = re.id
LEFT JOIN zones z ON re.zone_id = z.id
ORDER BY a.sent_at DESC LIMIT ?;

-- name: ListAlertsBySite :many
SELECT a.id, a.risk_event_id, a.alert_level, a.status, a.sent_at, a.acknowledged_at, a.camera_id,
       re.explanation, z.name as zone_name, 
       c.name as camera_name, c.x_plan as camera_x, c.y_plan as camera_y, 
       c.orientation as camera_orientation, c.fov as camera_fov, c.plan_id as camera_plan_id,
       c.is_webcam as camera_is_webcam
FROM alerts a
LEFT JOIN risk_events re ON a.risk_event_id = re.id
LEFT JOIN zones z ON re.zone_id = z.id
LEFT JOIN cameras c ON a.camera_id = c.id
LEFT JOIN plans p ON c.plan_id = p.id
WHERE p.site_id = ?
ORDER BY a.sent_at DESC LIMIT 50;

-- name: ListActiveAlerts :many
SELECT a.*, re.explanation, z.name as zone_name FROM alerts a
LEFT JOIN risk_events re ON a.risk_event_id = re.id
LEFT JOIN zones z ON re.zone_id = z.id
WHERE a.status = 'new'
ORDER BY a.sent_at DESC;

-- name: AcknowledgeAlert :exec
UPDATE alerts SET status = 'acknowledged', acknowledged_at = ? WHERE id = ?;

-- name: CloseAlert :exec
UPDATE alerts SET status = 'closed' WHERE id = ?;

-- Detections
-- name: GetRecentDetections :many
SELECT * FROM detections ORDER BY timestamp DESC LIMIT ?;

-- name: GetRecentDetectionsBySite :many
SELECT d.* FROM detections d
JOIN cameras c ON d.camera_id = c.id
JOIN plans p ON c.plan_id = p.id
WHERE p.site_id = ?
ORDER BY d.timestamp DESC LIMIT ?;

-- Users
-- name: ListUsers :many
SELECT id, username, email, is_active, created_at FROM users ORDER BY username;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateUser :exec
UPDATE users SET username = ?, email = ?, is_active = ? WHERE id = ?;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = ? WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- Roles
-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;

-- name: AssignRole :exec
INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?);

-- name: RemoveRole :exec
DELETE FROM user_roles WHERE user_id = ? AND role_id = ?;

-- HSE Rules
-- name: ListHSERules :many
SELECT * FROM hse_rules ORDER BY severity DESC;

-- name: GetHSERule :one
SELECT * FROM hse_rules WHERE id = ?;

-- name: CreateHSERule :one
INSERT INTO hse_rules (name, description, condition_logic, severity, is_active)
VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateHSERule :exec
UPDATE hse_rules SET name = ?, description = ?, condition_logic = ?, severity = ?, is_active = ? WHERE id = ?;

-- name: DeleteHSERule :exec
DELETE FROM hse_rules WHERE id = ?;

-- Models
-- name: ListModels :many
SELECT * FROM models ORDER BY name;

-- name: GetModel :one
SELECT * FROM models WHERE id = ?;

-- name: CreateModel :one
INSERT INTO models (name, type, version, metrics, is_active) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateModel :exec
UPDATE models SET name = ?, type = ?, version = ?, metrics = ?, is_active = ? WHERE id = ?;

-- name: DeleteModel :exec
DELETE FROM models WHERE id = ?;
