// Modifié manuellement pour MySQL/MariaDB (RETURNING * non supporté)
// versions:
//   sqlc v1.30.0
// source: securite.sql

package dbgen

import (
	"context"
	"time"
)

const acknowledgeAlert = `UPDATE alerts SET status = 'acknowledged', acknowledged_at = ? WHERE id = ?`

type AcknowledgeAlertParams struct {
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	ID             int64      `json:"id"`
}

func (q *Queries) AcknowledgeAlert(ctx context.Context, arg AcknowledgeAlertParams) error {
	_, err := q.db.ExecContext(ctx, acknowledgeAlert, arg.AcknowledgedAt, arg.ID)
	return err
}

const assignRole = `INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)`

type AssignRoleParams struct {
	UserID int64 `json:"user_id"`
	RoleID int64 `json:"role_id"`
}

func (q *Queries) AssignRole(ctx context.Context, arg AssignRoleParams) error {
	_, err := q.db.ExecContext(ctx, assignRole, arg.UserID, arg.RoleID)
	return err
}

const closeAlert = `UPDATE alerts SET status = 'closed' WHERE id = ?`

func (q *Queries) CloseAlert(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, closeAlert, id)
	return err
}

// Stats
func (q *Queries) CountActiveAlerts(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE status = 'new'`)
	var count int64
	return count, row.Scan(&count)
}

func (q *Queries) CountActiveAlertsBySite(ctx context.Context, siteID *int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts a
JOIN cameras c ON a.camera_id = c.id
JOIN plans p ON c.plan_id = p.id
WHERE a.status = 'new' AND p.site_id = ?`, siteID)
	var count int64
	return count, row.Scan(&count)
}

func (q *Queries) CountActiveCameras(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM cameras`)
	var count int64
	return count, row.Scan(&count)
}

func (q *Queries) CountAlertsToday(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE DATE(sent_at) = CURDATE()`)
	var count int64
	return count, row.Scan(&count)
}

func (q *Queries) CountCamerasBySite(ctx context.Context, siteID *int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM cameras c
JOIN plans p ON c.plan_id = p.id
WHERE p.site_id = ?`, siteID)
	var count int64
	return count, row.Scan(&count)
}

func (q *Queries) CountDetectionsToday(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM detections WHERE DATE(timestamp) = CURDATE()`)
	var count int64
	return count, row.Scan(&count)
}

func (q *Queries) CountDetectionsTodayBySite(ctx context.Context, siteID *int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM detections d
JOIN cameras c ON d.camera_id = c.id
JOIN plans p ON c.plan_id = p.id
WHERE DATE(d.timestamp) = CURDATE() AND p.site_id = ?`, siteID)
	var count int64
	return count, row.Scan(&count)
}

func (q *Queries) CountHighRiskZones(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM zones WHERE risk_level = 'HIGH' AND is_active = 1`)
	var count int64
	return count, row.Scan(&count)
}

func (q *Queries) CountHighRiskZonesBySite(ctx context.Context, siteID *int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM zones z
JOIN plans p ON z.plan_id = p.id
WHERE z.risk_level = 'HIGH' AND z.is_active = 1 AND p.site_id = ?`, siteID)
	var count int64
	return count, row.Scan(&count)
}

// ── Camera ────────────────────────────────────────────────────────────────────

type CreateCameraParams struct {
	PlanID      *int64   `json:"plan_id"`
	Name        *string  `json:"name"`
	StreamUrl   *string  `json:"stream_url"`
	XPlan       *float64 `json:"x_plan"`
	YPlan       *float64 `json:"y_plan"`
	Orientation *float64 `json:"orientation"`
	Fov         *float64 `json:"fov"`
	IsWebcam    *int64   `json:"is_webcam"`
}

func (q *Queries) CreateCamera(ctx context.Context, arg CreateCameraParams) (Camera, error) {
	res, err := q.db.ExecContext(ctx,
		`INSERT INTO cameras (plan_id, name, stream_url, x_plan, y_plan, orientation, fov, is_webcam)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		arg.PlanID, arg.Name, arg.StreamUrl, arg.XPlan, arg.YPlan, arg.Orientation, arg.Fov, arg.IsWebcam,
	)
	if err != nil {
		return Camera{}, err
	}
	id, _ := res.LastInsertId()
	return q.GetCamera(ctx, id)
}

func (q *Queries) GetCamera(ctx context.Context, id int64) (Camera, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, plan_id, name, stream_url, x_plan, y_plan, orientation, fov, calibration_matrix, confidence_config, is_webcam, created_at
		 FROM cameras WHERE id = ?`, id)
	var i Camera
	err := row.Scan(&i.ID, &i.PlanID, &i.Name, &i.StreamUrl, &i.XPlan, &i.YPlan, &i.Orientation, &i.Fov,
		&i.CalibrationMatrix, &i.ConfidenceConfig, &i.IsWebcam, &i.CreatedAt)
	return i, err
}

type UpdateCameraParams struct {
	Name        *string  `json:"name"`
	StreamUrl   *string  `json:"stream_url"`
	XPlan       *float64 `json:"x_plan"`
	YPlan       *float64 `json:"y_plan"`
	Orientation *float64 `json:"orientation"`
	Fov         *float64 `json:"fov"`
	IsWebcam    *int64   `json:"is_webcam"`
	ID          int64    `json:"id"`
}

func (q *Queries) UpdateCamera(ctx context.Context, arg UpdateCameraParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE cameras SET name=?, stream_url=?, x_plan=?, y_plan=?, orientation=?, fov=?, is_webcam=? WHERE id=?`,
		arg.Name, arg.StreamUrl, arg.XPlan, arg.YPlan, arg.Orientation, arg.Fov, arg.IsWebcam, arg.ID)
	return err
}

func (q *Queries) DeleteCamera(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM cameras WHERE id = ?`, id)
	return err
}

func (q *Queries) ListCameras(ctx context.Context) ([]ListCamerasRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT c.id, c.plan_id, c.name, c.stream_url, c.x_plan, c.y_plan,
		c.orientation, c.fov, c.calibration_matrix, c.confidence_config, c.is_webcam, c.created_at,
		p.level as plan_level, s.name as site_name
		FROM cameras c
		LEFT JOIN plans p ON c.plan_id = p.id
		LEFT JOIN sites s ON p.site_id = s.id
		ORDER BY s.name, c.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListCamerasRow
	for rows.Next() {
		var i ListCamerasRow
		if err := rows.Scan(&i.ID, &i.PlanID, &i.Name, &i.StreamUrl, &i.XPlan, &i.YPlan,
			&i.Orientation, &i.Fov, &i.CalibrationMatrix, &i.ConfidenceConfig, &i.IsWebcam, &i.CreatedAt,
			&i.PlanLevel, &i.SiteName); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) ListCamerasBySite(ctx context.Context, siteID *int64) ([]ListCamerasBySiteRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT c.id, c.plan_id, c.name, c.stream_url, c.x_plan, c.y_plan,
		c.orientation, c.fov, c.calibration_matrix, c.confidence_config, c.is_webcam, c.created_at,
		p.level as plan_level
		FROM cameras c
		JOIN plans p ON c.plan_id = p.id
		WHERE p.site_id = ?
		ORDER BY c.name`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListCamerasBySiteRow
	for rows.Next() {
		var i ListCamerasBySiteRow
		if err := rows.Scan(&i.ID, &i.PlanID, &i.Name, &i.StreamUrl, &i.XPlan, &i.YPlan,
			&i.Orientation, &i.Fov, &i.CalibrationMatrix, &i.ConfidenceConfig, &i.IsWebcam, &i.CreatedAt,
			&i.PlanLevel); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) ListCamerasByPlan(ctx context.Context, planID *int64) ([]Camera, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT id, plan_id, name, stream_url, x_plan, y_plan,
		orientation, fov, calibration_matrix, confidence_config, is_webcam, created_at
		FROM cameras WHERE plan_id = ?`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Camera
	for rows.Next() {
		var i Camera
		if err := rows.Scan(&i.ID, &i.PlanID, &i.Name, &i.StreamUrl, &i.XPlan, &i.YPlan,
			&i.Orientation, &i.Fov, &i.CalibrationMatrix, &i.ConfidenceConfig, &i.IsWebcam, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ── Plan ──────────────────────────────────────────────────────────────────────

type CreatePlanParams struct {
	SiteID      *int64   `json:"site_id"`
	Level       *string  `json:"level"`
	ImagePath   *string  `json:"image_path"`
	ScaleFactor *float64 `json:"scale_factor"`
}

func (q *Queries) CreatePlan(ctx context.Context, arg CreatePlanParams) (Plan, error) {
	res, err := q.db.ExecContext(ctx,
		`INSERT INTO plans (site_id, level, image_path, scale_factor) VALUES (?, ?, ?, ?)`,
		arg.SiteID, arg.Level, arg.ImagePath, arg.ScaleFactor,
	)
	if err != nil {
		return Plan{}, err
	}
	id, _ := res.LastInsertId()
	return q.GetPlan(ctx, id)
}

func (q *Queries) GetPlan(ctx context.Context, id int64) (Plan, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, site_id, level, image_path, scale_factor, created_at FROM plans WHERE id = ?`, id)
	var i Plan
	err := row.Scan(&i.ID, &i.SiteID, &i.Level, &i.ImagePath, &i.ScaleFactor, &i.CreatedAt)
	return i, err
}

type UpdatePlanParams struct {
	Level       *string  `json:"level"`
	ImagePath   *string  `json:"image_path"`
	ScaleFactor *float64 `json:"scale_factor"`
	ID          int64    `json:"id"`
}

func (q *Queries) UpdatePlan(ctx context.Context, arg UpdatePlanParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE plans SET level=?, image_path=?, scale_factor=? WHERE id=?`,
		arg.Level, arg.ImagePath, arg.ScaleFactor, arg.ID)
	return err
}

func (q *Queries) DeletePlan(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM plans WHERE id = ?`, id)
	return err
}

func (q *Queries) ListPlans(ctx context.Context) ([]ListPlansRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT p.id, p.site_id, p.level, p.image_path, p.scale_factor, p.created_at,
		s.name as site_name FROM plans p
		JOIN sites s ON p.site_id = s.id
		ORDER BY s.name, p.level`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListPlansRow
	for rows.Next() {
		var i ListPlansRow
		if err := rows.Scan(&i.ID, &i.SiteID, &i.Level, &i.ImagePath, &i.ScaleFactor, &i.CreatedAt, &i.SiteName); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) ListPlansBySite(ctx context.Context, siteID *int64) ([]Plan, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, site_id, level, image_path, scale_factor, created_at FROM plans WHERE site_id = ? ORDER BY level`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Plan
	for rows.Next() {
		var i Plan
		if err := rows.Scan(&i.ID, &i.SiteID, &i.Level, &i.ImagePath, &i.ScaleFactor, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ── Site ──────────────────────────────────────────────────────────────────────

type CreateSiteParams struct {
	Name        string  `json:"name"`
	Location    *string `json:"location"`
	Description *string `json:"description"`
}

func (q *Queries) CreateSite(ctx context.Context, arg CreateSiteParams) (Site, error) {
	res, err := q.db.ExecContext(ctx,
		`INSERT INTO sites (name, location, description) VALUES (?, ?, ?)`,
		arg.Name, arg.Location, arg.Description,
	)
	if err != nil {
		return Site{}, err
	}
	id, _ := res.LastInsertId()
	return q.GetSite(ctx, id)
}

func (q *Queries) GetSite(ctx context.Context, id int64) (Site, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, name, location, description, confidence_config, created_at FROM sites WHERE id = ?`, id)
	var i Site
	err := row.Scan(&i.ID, &i.Name, &i.Location, &i.Description, &i.ConfidenceConfig, &i.CreatedAt)
	return i, err
}

type UpdateSiteParams struct {
	Name        string  `json:"name"`
	Location    *string `json:"location"`
	Description *string `json:"description"`
	ID          int64   `json:"id"`
}

func (q *Queries) UpdateSite(ctx context.Context, arg UpdateSiteParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE sites SET name=?, location=?, description=? WHERE id=?`,
		arg.Name, arg.Location, arg.Description, arg.ID)
	return err
}

func (q *Queries) DeleteSite(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM sites WHERE id = ?`, id)
	return err
}

func (q *Queries) ListSites(ctx context.Context) ([]Site, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, name, location, description, confidence_config, created_at FROM sites ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Site
	for rows.Next() {
		var i Site
		if err := rows.Scan(&i.ID, &i.Name, &i.Location, &i.Description, &i.ConfidenceConfig, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

type SearchSitesParams struct {
	Name     string  `json:"name"`
	Location *string `json:"location"`
}

func (q *Queries) SearchSites(ctx context.Context, arg SearchSitesParams) ([]Site, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, name, location, description, confidence_config, created_at FROM sites WHERE name LIKE ? OR location LIKE ? ORDER BY name LIMIT 20`,
		arg.Name, arg.Location)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Site
	for rows.Next() {
		var i Site
		if err := rows.Scan(&i.ID, &i.Name, &i.Location, &i.Description, &i.ConfidenceConfig, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ── Zone ──────────────────────────────────────────────────────────────────────

type CreateZoneParams struct {
	PlanID    *int64  `json:"plan_id"`
	Name      *string `json:"name"`
	Type      *string `json:"type"`
	Polygon   string  `json:"polygon"`
	RiskLevel *string `json:"risk_level"`
}

func (q *Queries) CreateZone(ctx context.Context, arg CreateZoneParams) (Zone, error) {
	res, err := q.db.ExecContext(ctx,
		`INSERT INTO zones (plan_id, name, type, polygon, risk_level, is_active) VALUES (?, ?, ?, ?, ?, 1)`,
		arg.PlanID, arg.Name, arg.Type, arg.Polygon, arg.RiskLevel,
	)
	if err != nil {
		return Zone{}, err
	}
	id, _ := res.LastInsertId()
	return q.GetZone(ctx, id)
}

func (q *Queries) GetZone(ctx context.Context, id int64) (Zone, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, plan_id, name, type, polygon, risk_level, is_active, created_at FROM zones WHERE id = ?`, id)
	var i Zone
	err := row.Scan(&i.ID, &i.PlanID, &i.Name, &i.Type, &i.Polygon, &i.RiskLevel, &i.IsActive, &i.CreatedAt)
	return i, err
}

type UpdateZoneParams struct {
	Name      *string `json:"name"`
	Type      *string `json:"type"`
	Polygon   string  `json:"polygon"`
	RiskLevel *string `json:"risk_level"`
	IsActive  *int64  `json:"is_active"`
	ID        int64   `json:"id"`
}

func (q *Queries) UpdateZone(ctx context.Context, arg UpdateZoneParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE zones SET name=?, type=?, polygon=?, risk_level=?, is_active=? WHERE id=?`,
		arg.Name, arg.Type, arg.Polygon, arg.RiskLevel, arg.IsActive, arg.ID)
	return err
}

func (q *Queries) DeleteZone(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM zones WHERE id = ?`, id)
	return err
}

func (q *Queries) ListZones(ctx context.Context) ([]ListZonesRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT z.id, z.plan_id, z.name, z.type, z.polygon, z.risk_level, z.is_active, z.created_at,
		p.level as plan_level, s.name as site_name FROM zones z
		LEFT JOIN plans p ON z.plan_id = p.id
		LEFT JOIN sites s ON p.site_id = s.id
		ORDER BY s.name, z.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListZonesRow
	for rows.Next() {
		var i ListZonesRow
		if err := rows.Scan(&i.ID, &i.PlanID, &i.Name, &i.Type, &i.Polygon, &i.RiskLevel, &i.IsActive, &i.CreatedAt, &i.PlanLevel, &i.SiteName); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) ListZonesBySite(ctx context.Context, siteID *int64) ([]ListZonesBySiteRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT z.id, z.plan_id, z.name, z.type, z.polygon, z.risk_level, z.is_active, z.created_at,
		p.level as plan_level FROM zones z
		JOIN plans p ON z.plan_id = p.id
		WHERE p.site_id = ?
		ORDER BY z.name`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListZonesBySiteRow
	for rows.Next() {
		var i ListZonesBySiteRow
		if err := rows.Scan(&i.ID, &i.PlanID, &i.Name, &i.Type, &i.Polygon, &i.RiskLevel, &i.IsActive, &i.CreatedAt, &i.PlanLevel); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) ListZonesByPlan(ctx context.Context, planID *int64) ([]Zone, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, plan_id, name, type, polygon, risk_level, is_active, created_at FROM zones WHERE plan_id = ?`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Zone
	for rows.Next() {
		var i Zone
		if err := rows.Scan(&i.ID, &i.PlanID, &i.Name, &i.Type, &i.Polygon, &i.RiskLevel, &i.IsActive, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ── Alert ─────────────────────────────────────────────────────────────────────

func (q *Queries) ListAlerts(ctx context.Context, limit int32) ([]ListAlertsRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT a.id, a.risk_event_id, a.alert_level, a.status, a.sent_at, a.acknowledged_at, a.camera_id,
		re.explanation, z.name as zone_name FROM alerts a
		LEFT JOIN risk_events re ON a.risk_event_id = re.id
		LEFT JOIN zones z ON re.zone_id = z.id
		ORDER BY a.sent_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAlertsRow
	for rows.Next() {
		var i ListAlertsRow
		if err := rows.Scan(&i.ID, &i.RiskEventID, &i.AlertLevel, &i.Status, &i.SentAt, &i.AcknowledgedAt, &i.CameraID,
			&i.Explanation, &i.ZoneName); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) ListAlertsBySite(ctx context.Context, siteID *int64) ([]ListAlertsBySiteRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT a.id, a.risk_event_id, a.alert_level, a.status, a.sent_at, a.acknowledged_at, a.camera_id,
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
		ORDER BY a.sent_at DESC LIMIT 50`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAlertsBySiteRow
	for rows.Next() {
		var i ListAlertsBySiteRow
		if err := rows.Scan(&i.ID, &i.RiskEventID, &i.AlertLevel, &i.Status, &i.SentAt, &i.AcknowledgedAt, &i.CameraID,
			&i.Explanation, &i.ZoneName, &i.CameraName, &i.CameraX, &i.CameraY,
			&i.CameraOrientation, &i.CameraFov, &i.CameraPlanID, &i.CameraIsWebcam); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) ListActiveAlerts(ctx context.Context) ([]ListActiveAlertsRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT a.id, a.risk_event_id, a.alert_level, a.status, a.sent_at, a.acknowledged_at, a.camera_id,
		re.explanation, z.name as zone_name FROM alerts a
		LEFT JOIN risk_events re ON a.risk_event_id = re.id
		LEFT JOIN zones z ON re.zone_id = z.id
		WHERE a.status = 'new'
		ORDER BY a.sent_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListActiveAlertsRow
	for rows.Next() {
		var i ListActiveAlertsRow
		if err := rows.Scan(&i.ID, &i.RiskEventID, &i.AlertLevel, &i.Status, &i.SentAt, &i.AcknowledgedAt, &i.CameraID,
			&i.Explanation, &i.ZoneName); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ── Detection ─────────────────────────────────────────────────────────────────

type GetRecentDetectionsBySiteParams struct {
	SiteID *int64 `json:"site_id"`
	Limit  int32  `json:"limit"`
}

func (q *Queries) GetRecentDetections(ctx context.Context, limit int32) ([]Detection, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, camera_id, timestamp, object_class, confidence, bbox_x, bbox_y, bbox_w, bbox_h, track_id
		 FROM detections ORDER BY timestamp DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Detection
	for rows.Next() {
		var i Detection
		if err := rows.Scan(&i.ID, &i.CameraID, &i.Timestamp, &i.ObjectClass, &i.Confidence,
			&i.BboxX, &i.BboxY, &i.BboxW, &i.BboxH, &i.TrackID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) GetRecentDetectionsBySite(ctx context.Context, arg GetRecentDetectionsBySiteParams) ([]Detection, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT d.id, d.camera_id, d.timestamp, d.object_class, d.confidence,
		d.bbox_x, d.bbox_y, d.bbox_w, d.bbox_h, d.track_id
		FROM detections d
		JOIN cameras c ON d.camera_id = c.id
		JOIN plans p ON c.plan_id = p.id
		WHERE p.site_id = ?
		ORDER BY d.timestamp DESC LIMIT ?`, arg.SiteID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Detection
	for rows.Next() {
		var i Detection
		if err := rows.Scan(&i.ID, &i.CameraID, &i.Timestamp, &i.ObjectClass, &i.Confidence,
			&i.BboxX, &i.BboxY, &i.BboxW, &i.BboxH, &i.TrackID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ── User ──────────────────────────────────────────────────────────────────────

type CreateUserParams struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	res, err := q.db.ExecContext(ctx,
		`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`,
		arg.Username, arg.Email, arg.PasswordHash,
	)
	if err != nil {
		return User{}, err
	}
	id, _ := res.LastInsertId()
	return q.GetUser(ctx, id)
}

func (q *Queries) GetUser(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, is_active, created_at, last_login FROM users WHERE id = ?`, id)
	var i User
	err := row.Scan(&i.ID, &i.Username, &i.Email, &i.PasswordHash, &i.IsActive, &i.CreatedAt, &i.LastLogin)
	return i, err
}

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, is_active, created_at, last_login FROM users WHERE username = ?`, username)
	var i User
	err := row.Scan(&i.ID, &i.Username, &i.Email, &i.PasswordHash, &i.IsActive, &i.CreatedAt, &i.LastLogin)
	return i, err
}

type UpdateUserParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive *int64 `json:"is_active"`
	ID       int64  `json:"id"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE users SET username=?, email=?, is_active=? WHERE id=?`,
		arg.Username, arg.Email, arg.IsActive, arg.ID)
	return err
}

type UpdateUserPasswordParams struct {
	PasswordHash string `json:"password_hash"`
	ID           int64  `json:"id"`
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, arg.PasswordHash, arg.ID)
	return err
}

func (q *Queries) DeleteUser(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

func (q *Queries) ListUsers(ctx context.Context) ([]ListUsersRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, username, email, is_active, created_at FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListUsersRow
	for rows.Next() {
		var i ListUsersRow
		if err := rows.Scan(&i.ID, &i.Username, &i.Email, &i.IsActive, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ── Role ──────────────────────────────────────────────────────────────────────

func (q *Queries) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, name, description, created_at FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Role
	for rows.Next() {
		var i Role
		if err := rows.Scan(&i.ID, &i.Name, &i.Description, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) RemoveRole(ctx context.Context, arg RemoveRoleParams) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = ? AND role_id = ?`, arg.UserID, arg.RoleID)
	return err
}

// ── HSE Rule ──────────────────────────────────────────────────────────────────

type CreateHSERuleParams struct {
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	ConditionLogic *string `json:"condition_logic"`
	Severity       *int64  `json:"severity"`
	IsActive       *int64  `json:"is_active"`
}

func (q *Queries) CreateHSERule(ctx context.Context, arg CreateHSERuleParams) (HseRule, error) {
	res, err := q.db.ExecContext(ctx,
		`INSERT INTO hse_rules (name, description, condition_logic, severity, is_active) VALUES (?, ?, ?, ?, ?)`,
		arg.Name, arg.Description, arg.ConditionLogic, arg.Severity, arg.IsActive,
	)
	if err != nil {
		return HseRule{}, err
	}
	id, _ := res.LastInsertId()
	return q.GetHSERule(ctx, id)
}

func (q *Queries) GetHSERule(ctx context.Context, id int64) (HseRule, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, name, description, condition_logic, severity, is_active FROM hse_rules WHERE id = ?`, id)
	var i HseRule
	err := row.Scan(&i.ID, &i.Name, &i.Description, &i.ConditionLogic, &i.Severity, &i.IsActive)
	return i, err
}

type UpdateHSERuleParams struct {
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	ConditionLogic *string `json:"condition_logic"`
	Severity       *int64  `json:"severity"`
	IsActive       *int64  `json:"is_active"`
	ID             int64   `json:"id"`
}

func (q *Queries) UpdateHSERule(ctx context.Context, arg UpdateHSERuleParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE hse_rules SET name=?, description=?, condition_logic=?, severity=?, is_active=? WHERE id=?`,
		arg.Name, arg.Description, arg.ConditionLogic, arg.Severity, arg.IsActive, arg.ID)
	return err
}

func (q *Queries) DeleteHSERule(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM hse_rules WHERE id = ?`, id)
	return err
}

func (q *Queries) ListHSERules(ctx context.Context) ([]HseRule, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, name, description, condition_logic, severity, is_active FROM hse_rules ORDER BY severity DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []HseRule
	for rows.Next() {
		var i HseRule
		if err := rows.Scan(&i.ID, &i.Name, &i.Description, &i.ConditionLogic, &i.Severity, &i.IsActive); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ── Model ─────────────────────────────────────────────────────────────────────

type CreateModelParams struct {
	Name     *string `json:"name"`
	Type     *string `json:"type"`
	Version  *string `json:"version"`
	Metrics  *string `json:"metrics"`
	IsActive *int64  `json:"is_active"`
}

func (q *Queries) CreateModel(ctx context.Context, arg CreateModelParams) (Model, error) {
	res, err := q.db.ExecContext(ctx,
		`INSERT INTO models (name, type, version, metrics, is_active) VALUES (?, ?, ?, ?, ?)`,
		arg.Name, arg.Type, arg.Version, arg.Metrics, arg.IsActive,
	)
	if err != nil {
		return Model{}, err
	}
	id, _ := res.LastInsertId()
	return q.GetModel(ctx, id)
}

func (q *Queries) GetModel(ctx context.Context, id int64) (Model, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, name, type, version, trained_at, metrics, is_active FROM models WHERE id = ?`, id)
	var i Model
	err := row.Scan(&i.ID, &i.Name, &i.Type, &i.Version, &i.TrainedAt, &i.Metrics, &i.IsActive)
	return i, err
}

type UpdateModelParams struct {
	Name     *string `json:"name"`
	Type     *string `json:"type"`
	Version  *string `json:"version"`
	Metrics  *string `json:"metrics"`
	IsActive *int64  `json:"is_active"`
	ID       int64   `json:"id"`
}

func (q *Queries) UpdateModel(ctx context.Context, arg UpdateModelParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE models SET name=?, type=?, version=?, metrics=?, is_active=? WHERE id=?`,
		arg.Name, arg.Type, arg.Version, arg.Metrics, arg.IsActive, arg.ID)
	return err
}

func (q *Queries) DeleteModel(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM models WHERE id = ?`, id)
	return err
}

func (q *Queries) ListModels(ctx context.Context) ([]Model, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, name, type, version, trained_at, metrics, is_active FROM models ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Model
	for rows.Next() {
		var i Model
		if err := rows.Scan(&i.ID, &i.Name, &i.Type, &i.Version, &i.TrainedAt, &i.Metrics, &i.IsActive); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ── Row types ─────────────────────────────────────────────────────────────────

type ListCamerasRow struct {
	Camera
	PlanLevel *string `json:"plan_level"`
	SiteName  *string `json:"site_name"`
}

type ListCamerasBySiteRow struct {
	Camera
	PlanLevel *string `json:"plan_level"`
}

type ListPlansRow struct {
	Plan
	SiteName string `json:"site_name"`
}

type ListZonesRow struct {
	Zone
	PlanLevel *string `json:"plan_level"`
	SiteName  *string `json:"site_name"`
}

type ListZonesBySiteRow struct {
	Zone
	PlanLevel *string `json:"plan_level"`
}

type ListAlertsRow struct {
	ID             int64      `json:"id"`
	RiskEventID    *int64     `json:"risk_event_id"`
	AlertLevel     *string    `json:"alert_level"`
	Status         *string    `json:"status"`
	SentAt         time.Time  `json:"sent_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	CameraID       *int64     `json:"camera_id"`
	Explanation    *string    `json:"explanation"`
	ZoneName       *string    `json:"zone_name"`
}

type ListAlertsBySiteRow struct {
	ID                int64      `json:"id"`
	RiskEventID       *int64     `json:"risk_event_id"`
	AlertLevel        *string    `json:"alert_level"`
	Status            *string    `json:"status"`
	SentAt            time.Time  `json:"sent_at"`
	AcknowledgedAt    *time.Time `json:"acknowledged_at"`
	CameraID          *int64     `json:"camera_id"`
	Explanation       *string    `json:"explanation"`
	ZoneName          *string    `json:"zone_name"`
	CameraName        *string    `json:"camera_name"`
	CameraX           *float64   `json:"camera_x"`
	CameraY           *float64   `json:"camera_y"`
	CameraOrientation *float64   `json:"camera_orientation"`
	CameraFov         *float64   `json:"camera_fov"`
	CameraPlanID      *int64     `json:"camera_plan_id"`
	CameraIsWebcam    *int64     `json:"camera_is_webcam"`
}

type ListActiveAlertsRow = ListAlertsRow

type ListUsersRow struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	IsActive  *int64     `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

type RemoveRoleParams struct {
	UserID int64 `json:"user_id"`
	RoleID int64 `json:"role_id"`
}
