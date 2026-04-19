import mysql.connector
import json
import numpy as np
import cv2

# Points from the camera image (trapezoid of the corridor view)
pts_image = np.array([[260, 360], [410, 350], [460, 430], [240, 450]], dtype="float32")

# Points for the blue zone (corridor on the 2D plan)
# Estimated based on the image size and the blue loop drawn
pts_plan = np.array([[150, 30], [950, 30], [950, 110], [150, 110]], dtype="float32")

# Compute Homography
H, status = cv2.findHomography(pts_image, pts_plan)

# Connect to database
conn = mysql.connector.connect(
    host="localhost",
    user="admin",
    password="admin",
    database="make_hse"
)
cursor = conn.cursor()

# Get CAM01 ID
cursor.execute("SELECT id FROM cameras WHERE name='CAM01'")
result = cursor.fetchone()
if not result:
    print("CAM01 not found!")
    exit(1)
cam_id = result[0]

# Update query
update_query = """
UPDATE camera_calibrations 
SET pts_plan = %s, homography = %s 
WHERE camera_id = %s
"""

cursor.execute(update_query, (json.dumps(pts_plan.tolist()), json.dumps(H.tolist()), cam_id))
conn.commit()

print("Homography updated successfully for CAM01 to map strictly inside the blue zone!")

cursor.close()
conn.close()
