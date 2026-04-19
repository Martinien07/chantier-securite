import mysql.connector
import json
import numpy as np
import cv2

# pts_image: les 4 coins de l'image de la caméra (en 1920x1080)
# Ordre: Haut-Gauche, Haut-Droite, Bas-Droite, Bas-Gauche
pts_image = np.array([
    [0, 0],            # Haut-Gauche (fond du couloir, mur gauche)
    [1920, 0],         # Haut-Droite (fond du couloir, mur droit)
    [1920, 1080],      # Bas-Droite (proche caméra, mur droit)
    [0, 1080]          # Bas-Gauche (proche caméra, mur gauche)
], dtype="float32")

# pts_plan: les 4 coins correspondants sur le plan (le couloir bleu)
# Ordre correspondant:
# Haut-Gauche image = Fond du couloir, mur du haut sur le plan (X=950, Y=30)
# Haut-Droite image = Fond du couloir, mur du bas sur le plan (X=950, Y=110)
# Bas-Droite image = Proche caméra, mur du bas sur le plan (X=150, Y=110)
# Bas-Gauche image = Proche caméra, mur du haut sur le plan (X=150, Y=30)
pts_plan = np.array([
    [950, 30],
    [950, 110],
    [150, 110],
    [150, 30]
], dtype="float32")

# Calcul de l'homographie
H, status = cv2.findHomography(pts_image, pts_plan)

# Mise à jour BDD
conn = mysql.connector.connect(
    host="localhost",
    user="admin",
    password="admin",
    database="make_hse"
)
cursor = conn.cursor()

cursor.execute("SELECT id FROM cameras WHERE name='CAM01'")
cam_id = cursor.fetchone()[0]

update_query = """
UPDATE camera_calibrations 
SET pts_image = %s, pts_plan = %s, homography = %s 
WHERE camera_id = %s
"""
cursor.execute(update_query, (
    json.dumps(pts_image.tolist()), 
    json.dumps(pts_plan.tolist()), 
    json.dumps(H.tolist()), 
    cam_id
))
conn.commit()

print("Homographie FULL (1920x1080) mise à jour avec succès!")

cursor.close()
conn.close()
