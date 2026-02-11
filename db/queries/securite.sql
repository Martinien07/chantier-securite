-- name: CountActiveAlertes :one
SELECT COUNT(*) FROM alertes WHERE statut = 'active';

-- name: CountAlertesToday :one
SELECT COUNT(*) FROM alertes WHERE date(created_at) = date('now');

-- name: CountActiveCameras :one
SELECT COUNT(*) FROM cameras WHERE statut = 'active';

-- name: CountHighRiskZones :one
SELECT COUNT(*) FROM zones WHERE niveau_risque IN ('eleve', 'critique');

-- name: ListAlertes :many
SELECT 
    a.id,
    a.camera_id,
    c.nom as camera_nom,
    a.zone_id,
    z.nom as zone_nom,
    a.type_risque_id,
    tr.code as type_risque_code,
    tr.nom as type_risque_nom,
    tr.couleur as type_risque_couleur,
    a.severite,
    a.description,
    a.details_ia,
    a.confiance,
    a.image_url,
    a.statut,
    a.acknowledged_at,
    a.resolved_at,
    a.notes,
    a.created_at
FROM alertes a
LEFT JOIN cameras c ON a.camera_id = c.id
LEFT JOIN zones z ON a.zone_id = z.id
LEFT JOIN types_risque tr ON a.type_risque_id = tr.id
ORDER BY a.created_at DESC
LIMIT ?;

-- name: ListActiveAlertes :many
SELECT 
    a.id,
    a.camera_id,
    c.nom as camera_nom,
    a.zone_id,
    z.nom as zone_nom,
    a.type_risque_id,
    tr.code as type_risque_code,
    tr.nom as type_risque_nom,
    tr.couleur as type_risque_couleur,
    a.severite,
    a.description,
    a.details_ia,
    a.confiance,
    a.image_url,
    a.statut,
    a.created_at
FROM alertes a
LEFT JOIN cameras c ON a.camera_id = c.id
LEFT JOIN zones z ON a.zone_id = z.id
LEFT JOIN types_risque tr ON a.type_risque_id = tr.id
WHERE a.statut = 'active'
ORDER BY a.severite DESC, a.created_at DESC;

-- name: AcknowledgeAlerte :exec
UPDATE alertes SET statut = 'acknowledged', acknowledged_at = ? WHERE id = ?;

-- name: ResolveAlerte :exec
UPDATE alertes SET statut = 'resolved', resolved_at = ?, notes = ? WHERE id = ?;

-- name: ListCameras :many
SELECT 
    c.id,
    c.nom,
    c.zone_id,
    z.nom as zone_nom,
    z.niveau_risque as zone_niveau_risque,
    c.emplacement,
    c.flux_url,
    c.statut,
    c.derniere_detection,
    c.created_at,
    (SELECT COUNT(*) FROM alertes WHERE camera_id = c.id AND statut = 'active') as alertes_actives
FROM cameras c
LEFT JOIN zones z ON c.zone_id = z.id
ORDER BY c.zone_id, c.nom;

-- name: ListZones :many
SELECT 
    z.id,
    z.nom,
    z.description,
    z.niveau_risque,
    z.created_at,
    (SELECT COUNT(*) FROM cameras WHERE zone_id = z.id AND statut = 'active') as cameras_actives,
    (SELECT COUNT(*) FROM alertes WHERE zone_id = z.id AND statut = 'active') as alertes_actives
FROM zones z
ORDER BY 
    CASE z.niveau_risque 
        WHEN 'critique' THEN 1 
        WHEN 'eleve' THEN 2 
        WHEN 'moyen' THEN 3 
        ELSE 4 
    END;

-- name: GetTypesRisque :many
SELECT * FROM types_risque ORDER BY severite DESC;

-- name: GetAlertesParJour :many
SELECT 
    date(created_at) as jour,
    COUNT(*) as total,
    SUM(CASE WHEN severite >= 8 THEN 1 ELSE 0 END) as critiques,
    SUM(CASE WHEN severite >= 5 AND severite < 8 THEN 1 ELSE 0 END) as moyennes,
    SUM(CASE WHEN severite < 5 THEN 1 ELSE 0 END) as faibles
FROM alertes
WHERE created_at >= date('now', '-7 days')
GROUP BY date(created_at)
ORDER BY jour;

-- name: GetAlertesParType :many
SELECT 
    tr.code,
    tr.nom,
    tr.couleur,
    COUNT(*) as total
FROM alertes a
JOIN types_risque tr ON a.type_risque_id = tr.id
WHERE a.created_at >= date('now', '-30 days')
GROUP BY tr.id
ORDER BY total DESC;
