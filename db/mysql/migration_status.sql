-- ============================================================
-- Migration : ajout du champ status à risk_events
-- À exécuter UNE SEULE FOIS dans phpMyAdmin sur make_hse
-- ============================================================

-- Ajouter la colonne status (pending / validated / rejected)
ALTER TABLE `risk_events`
  ADD COLUMN IF NOT EXISTS `status`
    ENUM('pending', 'validated', 'rejected') NOT NULL DEFAULT 'pending'
    AFTER `explanation`;

-- Index pour que le polling soit rapide
ALTER TABLE `risk_events`
  ADD INDEX IF NOT EXISTS `idx_risk_events_status` (`status`);

-- Tous les risk_events existants qui n'ont pas encore de statut
-- sont considérés comme déjà validés (ils ont leur alerte correspondante)
UPDATE `risk_events` SET `status` = 'validated'
WHERE `status` = 'pending'
  AND `id` IN (SELECT `risk_event_id` FROM `alerts` WHERE `risk_event_id` IS NOT NULL);

-- Vérification
SELECT
  status,
  COUNT(*) AS total
FROM risk_events
GROUP BY status;
