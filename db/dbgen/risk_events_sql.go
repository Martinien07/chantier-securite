// db/dbgen/risk_events_sql.go
package dbgen

import (
	"context"
	"time"
)

type RiskEventPending struct {
	ID           int64      `json:"id"`
	CameraID     *int64     `json:"camera_id"`
	RiskLevel    *string    `json:"risk_level"`
	RiskScore    *float64   `json:"risk_score"`
	RiskMessages *string    `json:"risk_messages"`
	Explanation  *string    `json:"explanation"`
	WindowStart  *time.Time `json:"window_start"`
	WindowEnd    *time.Time `json:"window_end"`
	CreatedAt    time.Time  `json:"created_at"`
	CameraName   *string    `json:"camera_name"`
	SiteName     *string    `json:"site_name"`
}

func (q *Queries) ListPendingRiskEventsBySite(ctx context.Context, siteID *int64) ([]RiskEventPending, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT re.id, re.camera_id, re.risk_level, re.risk_score,
		       re.risk_messages, re.explanation,
		       re.window_start, re.window_end, re.created_at,
		       c.name as camera_name, s.name as site_name
		FROM risk_events re
		INNER JOIN cameras c ON re.camera_id = c.id
		INNER JOIN plans p   ON c.plan_id = p.id
		INNER JOIN sites s   ON p.site_id = s.id
		WHERE re.status = 'pending'
		  AND p.site_id = ?
		ORDER BY re.created_at DESC
		LIMIT 100
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RiskEventPending
	for rows.Next() {
		var i RiskEventPending
		if err := rows.Scan(
			&i.ID, &i.CameraID, &i.RiskLevel, &i.RiskScore,
			&i.RiskMessages, &i.Explanation,
			&i.WindowStart, &i.WindowEnd, &i.CreatedAt,
			&i.CameraName, &i.SiteName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) AcceptRiskEvent(ctx context.Context, riskEventID int64, comment string, action string) error {
	// 1. Passer le risk_event en accepted
	// Déterminer le statut final selon l'action
	newStatus := "accepted"
	if action == "correcting" {
		newStatus = "accepted"
	}

	_, err := q.db.ExecContext(ctx,
		`UPDATE risk_events SET status = ?, operator_comment = ?, resolved_at = NOW() WHERE id = ? AND status = 'pending'`,
		newStatus, comment, riskEventID,
	)
	if err != nil {
		return err
	}

	// 2. Récupérer les infos du risk_event
	row := q.db.QueryRowContext(ctx,
		`SELECT camera_id, risk_level, activity_pred FROM risk_events WHERE id = ?`,
		riskEventID,
	)
	var cameraID *int64
	var riskLevel *string
	var activityPred *string
	if err := row.Scan(&cameraID, &riskLevel, &activityPred); err != nil {
		return err
	}

	// 3. Déduplication : vérifier qu'une alerte active existe déjà
	//    pour la même caméra + même niveau (évite les doublons si l'opérateur
	//    valide plusieurs risk_events du même type)
	checkRow := q.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM alerts a
		JOIN risk_events re ON a.risk_event_id = re.id
		WHERE re.camera_id     = ?
		  AND re.activity_pred = ?
		  AND re.risk_level    = ?
		  AND a.status         IN ('new', 'acknowledged')
	`, cameraID, activityPred, riskLevel)
	var existing int64
	if err := checkRow.Scan(&existing); err == nil && existing > 0 {
		// Alerte déjà active pour ce contexte — pas de doublon
		return nil
	}

	// 4. Créer l'alerte
	_, err = q.db.ExecContext(ctx,
		`INSERT INTO alerts (risk_event_id, alert_level, status, sent_at, camera_id) VALUES (?, ?, 'new', NOW(), ?)`,
		riskEventID, riskLevel, cameraID,
	)
	return err
}

func (q *Queries) RejectRiskEvent(ctx context.Context, riskEventID int64, comment string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE risk_events SET status = 'rejected', operator_comment = ?, resolved_at = NOW() WHERE id = ? AND status = 'pending'`,
		comment, riskEventID,
	)
	return err
}

func (q *Queries) CountPendingRiskEventsBySite(ctx context.Context, siteID *int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM risk_events re
		LEFT JOIN cameras c ON re.camera_id = c.id
		LEFT JOIN plans p   ON c.plan_id = p.id
		WHERE re.status = 'pending' AND p.site_id = ?
	`, siteID)
	var count int64
	return count, row.Scan(&count)
}
