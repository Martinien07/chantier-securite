-- Seed data for SafeSite AI demo

-- Types de risques
INSERT OR REPLACE INTO types_risque (id, code, nom, description, severite, couleur, icone) VALUES
    (1, 'EPI_CASQUE', 'Absence de casque', 'Personne detectee sans casque de securite', 7, '#e74c3c', 'hard-hat'),
    (2, 'EPI_GILET', 'Absence de gilet', 'Personne detectee sans gilet haute visibilite', 5, '#f39c12', 'vest'),
    (3, 'ZONE_INTERDITE', 'Intrusion zone interdite', 'Presence non autorisee dans une zone dangereuse', 9, '#9b59b6', 'ban'),
    (4, 'PROXIMITE_ENGIN', 'Proximite engin dangereux', 'Personne trop proche d''un engin en mouvement', 10, '#e74c3c', 'truck'),
    (5, 'CHUTE_HAUTEUR', 'Risque de chute de hauteur', 'Personne sans harnais pres d''un bord', 10, '#c0392b', 'arrow-down'),
    (6, 'COACTIVITE', 'Coactivite dangereuse', 'Activites incompatibles simultanees detectees', 8, '#e67e22', 'users'),
    (7, 'ENCOMBREMENT', 'Voie obstruee', 'Passage ou issue de secours obstrue', 6, '#3498db', 'road-barrier'),
    (8, 'CHUTE_OBJET', 'Risque chute objet', 'Materiel mal arrime en hauteur', 8, '#8e44ad', 'box');

-- Zones
INSERT OR REPLACE INTO zones (id, nom, description, niveau_risque) VALUES
    (1, 'Zone A - Gros oeuvre', 'Travaux de structure beton', 'eleve'),
    (2, 'Zone B - Echafaudages', 'Travaux en hauteur facade nord', 'critique'),
    (3, 'Zone C - Stockage', 'Zone de stockage materiaux', 'moyen'),
    (4, 'Zone D - Circulation engins', 'Voie de circulation poids lourds', 'eleve'),
    (5, 'Zone E - Base vie', 'Vestiaires et refectoire', 'faible');

-- Cameras
INSERT OR REPLACE INTO cameras (id, nom, zone_id, emplacement, statut) VALUES
    (1, 'CAM-A1', 1, 'Entree zone gros oeuvre', 'active'),
    (2, 'CAM-A2', 1, 'Banches coffrage', 'active'),
    (3, 'CAM-B1', 2, 'Pied echafaudage nord', 'active'),
    (4, 'CAM-B2', 2, 'Plateforme niveau 3', 'active'),
    (5, 'CAM-C1', 3, 'Aire stockage principale', 'active'),
    (6, 'CAM-D1', 4, 'Entree chantier', 'active'),
    (7, 'CAM-D2', 4, 'Croisement voies', 'maintenance'),
    (8, 'CAM-E1', 5, 'Parking base vie', 'active');

-- Alertes de demonstration
INSERT OR REPLACE INTO alertes (id, camera_id, zone_id, type_risque_id, severite, description, details_ia, confiance, statut, created_at) VALUES
    (1, 1, 1, 1, 7, 'Absence de casque detectee - 2 personnes', 
     '{"objets_detectes":["personne","personne"],"epi_manquants":["casque","casque"],"contexte":"Zone de travaux actifs avec risque de chute d''objets","recommandation":"Intervention immediate requise pour rappel des regles EPI"}', 
     0.94, 'active', datetime('now', '-15 minutes')),
    (2, 4, 2, 5, 10, 'Risque chute de hauteur - Travailleur sans harnais', 
     '{"objets_detectes":["personne","echafaudage","garde-corps"],"analyse":"Personne detectee a moins de 1m du bord sans ligne de vie visible","facteurs_risque":["hauteur > 3m","absence harnais","vent modere"],"recommandation":"Arret immediat des travaux et securisation"}', 
     0.89, 'active', datetime('now', '-8 minutes')),
    (3, 6, 4, 4, 10, 'Proximite dangereuse engin/pieton', 
     '{"objets_detectes":["chargeuse","personne"],"distance_estimee":"2.3m","vitesse_engin":"lent","angle_mort":true,"recommandation":"Alerte conducteur et pieton immediate"}', 
     0.97, 'active', datetime('now', '-3 minutes')),
    (4, 3, 2, 1, 7, 'Absence de casque en zone echafaudage', 
     '{"objets_detectes":["personne"],"epi_manquants":["casque"],"contexte":"Pied echafaudage - risque chute objets","recommandation":"Port du casque obligatoire dans cette zone"}', 
     0.91, 'acknowledged', datetime('now', '-45 minutes')),
    (5, 2, 1, 6, 8, 'Coactivite dangereuse detectee', 
     '{"activites":["coulage beton","passage pieton"],"risques":["projection","glissade"],"recommandation":"Etablir perimetre de securite"}', 
     0.85, 'resolved', datetime('now', '-2 hours'));
