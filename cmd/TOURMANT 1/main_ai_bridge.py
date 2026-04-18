import sys
import threading
import time
import warnings

# Chemin vers le projet Python
sys.path.insert(0, r"D:\ETUDEµ\CITE\ETAPE4\PROJET DE FIN DE SESSION\TOURMANT 1")

from db.pool import init_pools
from inference.hse_acquisition_manager import HSEAcquisitionManager
from system_automatic.hse_analysis_system_save_bd import HSEAnalysisSystem

# --- SITE ID : passé par Go via variable d'environnement ou argv ---
import os
SITE_ID = int(os.environ.get("DETECTION_SITE_ID", sys.argv[1] if len(sys.argv) > 1 else "2"))
LOOKBACK  = 1
FREQUENCY = 20

def main():
    warnings.filterwarnings("ignore", category=UserWarning)
    print(f"--- Initialisation Système IA pour le Site ID: {SITE_ID} ---")

    init_pools(read_size=10, write_size=10)

    print("Démarrage de l'acquisition vidéo...")
    acquisition_manager = HSEAcquisitionManager(site_id=SITE_ID)
    acquisition_manager.start_dual_flux_pipeline()

    time.sleep(5)

    print("Démarrage du pipeline d'analyse des risques...")
    hse_system = HSEAnalysisSystem()
    hse_system.run(
        site_id=SITE_ID,
        lookback_minutes=LOOKBACK,
        frequency_seconds=FREQUENCY
    )

if __name__ == "__main__":
    main()
