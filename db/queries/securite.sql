-- Stats
-- name: CountActiveAlerts :one
SELECT COUNT(*) FROM alerts WHERE status = 'new';

-- name: CountAlertsToday :one
SELECT COUNT(*) FROM alerts WHERE date(sent_at) = date('now');

-- name: CountActiveCameras :one
SELECT COUNT(*) FROM cameras;

-- name: CountHighRiskZones :one
SELECT COUNT(*) FROM zones WHERE risk_level = 'HIGH' AND is_active = 1;

-- Sites
-- name: ListSites :many
SELECT * FROM sites ORDER BY name;

-- name: GetSite :one
SELECT * FROM sites WHERE id = ?;

-- name: CreateSite :one
INSERT INTO sites (name, location, description) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateSite :exec
UPDATE sites SET name = ?, location = ?, description = ? WHERE id = ?;

-- name: DeleteSite :exec
DELETE FROM sites WHERE id = ?;

-- Plans
-- name: ListPlans :many
SELECT p.*, s.name as site_name FROM plans p LEFT JOIN sites s ON p.site_id = s.id ORDER BY s.name, p.level;

-- name: ListPlansBySite :many
SELECT * FROM plans WHERE site_id = ? ORDER BY level;

-- name: GetPlan :one
SELECT * FROM plans WHERE id = ?;

-- name: CreatePlan :one
INSERT INTO plans (site_id, level, image_path, scale_factor) VALUES (?, ?, ?, ?) RETURNING *;

-- name: UpdatePlan :exec
UPDATE plans SET site_id = ?, level = ?, image_path = ?, scale_factor = ? WHERE id = ?;

-- name: DeletePlan :exec
DELETE FROM plans WHERE id = ?;

-- Cameras
-- name: ListCameras :many
SELECT c.*, p.level as plan_level, s.name as site_name 
FROM cameras c 
LEFT JOIN plans p ON c.plan_id = p.id 
LEFT JOIN sites s ON p.site_id = s.id 
ORDER BY s.name, p.level, c.name;

-- name: ListCamerasByPlan :many
SELECT * FROM cameras WHERE plan_id = ? ORDER BY name;

-- name: GetCamera :one
SELECT * FROM cameras WHERE id = ?;

-- name: CreateCamera :one
INSERT INTO cameras (plan_id, name, stream_url, x_plan, y_plan, orientation, fov, is_webcam) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateCamera :exec
UPDATE cameras SET plan_id = ?, name = ?, stream_url = ?, x_plan = ?, y_plan = ?, orientation = ?, fov = ?, is_webcam = ? WHERE id = ?;

-- name: DeleteCamera :exec
DELETE FROM cameras WHERE id = ?;

-- Zones
-- name: ListZones :many
SELECT z.*, p.level as plan_level, s.name as site_name 
FROM zones z 
LEFT JOIN plans p ON z.plan_id = p.id 
LEFT JOIN sites s ON p.site_id = s.id 
WHERE z.is_active = 1
ORDER BY s.name, p.level, z.name;

-- name: ListZonesByPlan :many
SELECT * FROM zones WHERE plan_id = ? AND is_active = 1 ORDER BY name;

-- name: GetZone :one
SELECT * FROM zones WHERE id = ?;

-- name: CreateZone :one
INSERT INTO zones (plan_id, name, type, polygon, risk_level, is_active) 
VALUES (?, ?, ?, ?, ?, 1) RETURNING *;

-- name: UpdateZone :exec
UPDATE zones SET plan_id = ?, name = ?, type = ?, polygon = ?, risk_level = ?, is_active = ? WHERE id = ?;

-- name: DeleteZone :exec
UPDATE zones SET is_active = 0 WHERE id = ?;

-- HSE Rules
-- name: ListHSERules :many
SELECT * FROM hse_rules ORDER BY severity DESC, name;

-- name: GetHSERule :one
SELECT * FROM hse_rules WHERE id = ?;

-- name: CreateHSERule :one
INSERT INTO hse_rules (name, description, condition_logic, severity, is_active) 
VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateHSERule :exec
UPDATE hse_rules SET name = ?, description = ?, condition_logic = ?, severity = ?, is_active = ? WHERE id = ?;

-- name: DeleteHSERule :exec
DELETE FROM hse_rules WHERE id = ?;

-- Users
-- name: ListUsers :many
SELECT u.*, GROUP_CONCAT(r.name) as roles
FROM users u 
LEFT JOIN user_roles ur ON u.id = ur.user_id 
LEFT JOIN roles r ON ur.role_id = r.id
GROUP BY u.id
ORDER BY u.username;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, is_active) VALUES (?, ?, ?, 1) RETURNING *;

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

-- Alerts
-- name: ListAlerts :many
SELECT a.*, re.risk_score, re.risk_level as event_risk_level, re.explanation,
       z.name as zone_name, hr.name as rule_name
FROM alerts a
LEFT JOIN risk_events re ON a.risk_event_id = re.id
LEFT JOIN zones z ON re.zone_id = z.id
LEFT JOIN hse_rules hr ON re.rule_id = hr.id
ORDER BY a.sent_at DESC
LIMIT ?;

-- name: ListActiveAlerts :many
SELECT a.*, re.risk_score, re.risk_level as event_risk_level, re.explanation,
       z.name as zone_name, hr.name as rule_name
FROM alerts a
LEFT JOIN risk_events re ON a.risk_event_id = re.id
LEFT JOIN zones z ON re.zone_id = z.id
LEFT JOIN hse_rules hr ON re.rule_id = hr.id
WHERE a.status = 'new'
ORDER BY a.sent_at DESC;

-- name: AcknowledgeAlert :exec
UPDATE alerts SET status = 'acknowledged', acknowledged_at = ? WHERE id = ?;

-- name: CloseAlert :exec
UPDATE alerts SET status = 'closed' WHERE id = ?;

-- Models
-- name: ListModels :many
SELECT * FROM models ORDER BY name, version DESC;

-- name: GetModel :one
SELECT * FROM models WHERE id = ?;

-- name: CreateModel :one
INSERT INTO models (name, type, version, metrics, is_active) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateModel :exec
UPDATE models SET name = ?, type = ?, version = ?, metrics = ?, is_active = ? WHERE id = ?;

-- name: DeleteModel :exec
DELETE FROM models WHERE id = ?;

-- Detections (for stats)
-- name: CountDetectionsToday :one
SELECT COUNT(*) FROM detections WHERE date(timestamp) = date('now');

-- name: GetRecentDetections :many
SELECT d.*, c.name as camera_name 
FROM detections d
LEFT JOIN cameras c ON d.camera_id = c.id
ORDER BY d.timestamp DESC
LIMIT ?;
