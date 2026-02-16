-- SafeSite AI - MySQL Queries

-- Sites
-- name: ListSites :many
SELECT * FROM sites ORDER BY name;

-- name: GetSite :one
SELECT * FROM sites WHERE id = ?;

-- name: SearchSites :many
SELECT * FROM sites WHERE name LIKE ? OR location LIKE ? ORDER BY name;

-- name: CreateSite :execresult
INSERT INTO sites (name, location, description) VALUES (?, ?, ?);

-- name: UpdateSite :exec
UPDATE sites SET name = ?, location = ?, description = ? WHERE id = ?;

-- name: DeleteSite :exec
DELETE FROM sites WHERE id = ?;

-- Plans
-- name: ListPlansBySite :many
SELECT * FROM plans WHERE site_id = ? ORDER BY level;

-- name: GetPlan :one
SELECT * FROM plans WHERE id = ?;

-- name: CreatePlan :execresult
INSERT INTO plans (site_id, level, image_path, scale_factor) VALUES (?, ?, ?, ?);

-- name: UpdatePlan :exec
UPDATE plans SET level = ?, image_path = ?, scale_factor = ? WHERE id = ?;

-- name: DeletePlan :exec
DELETE FROM plans WHERE id = ?;

-- Cameras
-- name: ListCamerasBySite :many
SELECT c.*, p.level as plan_level 
FROM cameras c
JOIN plans p ON c.plan_id = p.id
WHERE p.site_id = ?
ORDER BY c.name;

-- name: ListCamerasByPlan :many
SELECT * FROM cameras WHERE plan_id = ?;

-- name: GetCamera :one
SELECT * FROM cameras WHERE id = ?;

-- name: CreateCamera :execresult
INSERT INTO cameras (plan_id, name, stream_url, x_plan, y_plan, orientation, fov, is_webcam)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateCamera :exec
UPDATE cameras SET name = ?, stream_url = ?, x_plan = ?, y_plan = ?, orientation = ?, fov = ?, is_webcam = ? WHERE id = ?;

-- name: DeleteCamera :exec
DELETE FROM cameras WHERE id = ?;

-- Zones
-- name: ListZonesBySite :many
SELECT z.*, p.level as plan_level 
FROM zones z
JOIN plans p ON z.plan_id = p.id
WHERE p.site_id = ?
ORDER BY z.name;

-- name: ListZonesByPlan :many
SELECT * FROM zones WHERE plan_id = ?;

-- name: GetZone :one
SELECT * FROM zones WHERE id = ?;

-- name: CreateZone :execresult
INSERT INTO zones (plan_id, name, type, polygon, risk_level) VALUES (?, ?, ?, ?, ?);

-- name: UpdateZone :exec
UPDATE zones SET name = ?, type = ?, polygon = ?, risk_level = ?, is_active = ? WHERE id = ?;

-- name: DeleteZone :exec
DELETE FROM zones WHERE id = ?;

-- Alerts
-- name: ListAlertsBySite :many
SELECT a.*, a.camera_id, re.explanation, z.name as zone_name,
       c.name as camera_name, c.x_plan as camera_x, c.y_plan as camera_y,
       c.orientation as camera_orientation, c.fov as camera_fov, c.plan_id as camera_plan_id
FROM alerts a
LEFT JOIN risk_events re ON a.risk_event_id = re.id
LEFT JOIN zones z ON re.zone_id = z.id
LEFT JOIN cameras c ON a.camera_id = c.id
LEFT JOIN plans p ON c.plan_id = p.id
WHERE p.site_id = ?
ORDER BY a.sent_at DESC;

-- name: ListActiveAlertsBySite :many
SELECT a.id, a.alert_level, a.status, a.camera_id, c.name as camera_name, re.explanation
FROM alerts a
LEFT JOIN cameras c ON a.camera_id = c.id
LEFT JOIN plans p ON c.plan_id = p.id
LEFT JOIN risk_events re ON a.risk_event_id = re.id
WHERE a.status = 'new' AND p.site_id = ?;

-- name: GetAlert :one
SELECT * FROM alerts WHERE id = ?;

-- name: UpdateAlert :exec
UPDATE alerts SET alert_level = ?, status = ? WHERE id = ?;

-- name: AcknowledgeAlert :exec
UPDATE alerts SET status = 'acknowledged', acknowledged_at = NOW() WHERE id = ?;

-- name: CloseAlert :exec
UPDATE alerts SET status = 'closed' WHERE id = ?;

-- Stats
-- name: CountActiveAlertsBySite :one
SELECT COUNT(*) FROM alerts a
JOIN cameras c ON a.camera_id = c.id
JOIN plans p ON c.plan_id = p.id
WHERE a.status = 'new' AND p.site_id = ?;

-- name: CountCamerasBySite :one
SELECT COUNT(*) FROM cameras c
JOIN plans p ON c.plan_id = p.id
WHERE p.site_id = ?;

-- name: CountHighRiskZonesBySite :one
SELECT COUNT(*) FROM zones z
JOIN plans p ON z.plan_id = p.id
WHERE z.risk_level = 'HIGH' AND p.site_id = ?;

-- name: CountDetectionsTodayBySite :one
SELECT COUNT(*) FROM detections d
JOIN cameras c ON d.camera_id = c.id
JOIN plans p ON c.plan_id = p.id
WHERE DATE(d.timestamp) = CURDATE() AND p.site_id = ?;

-- HSE Rules
-- name: ListHSERules :many
SELECT * FROM hse_rules ORDER BY severity DESC;

-- name: GetHSERule :one
SELECT * FROM hse_rules WHERE id = ?;

-- name: CreateHSERule :execresult
INSERT INTO hse_rules (name, description, condition_logic, severity, is_active)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateHSERule :exec
UPDATE hse_rules SET name = ?, description = ?, condition_logic = ?, severity = ?, is_active = ? WHERE id = ?;

-- name: DeleteHSERule :exec
DELETE FROM hse_rules WHERE id = ?;

-- Models
-- name: ListModels :many
SELECT * FROM models ORDER BY trained_at DESC;

-- name: GetModel :one
SELECT * FROM models WHERE id = ?;

-- Users
-- name: ListUsers :many
SELECT * FROM users ORDER BY username;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- Roles
-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;
