@echo off
set MYSQL_DSN=admin:admin@tcp(localhost:3306)/make_hse?parseTime=true
set DETECTION_SCRIPT_PATH=D:\ETUDEµ\CITE\ETAPE4\PROJET DE FIN DE SESSION\TOURMANT 1\system_automatic\main_ai_bridge.py
set PYTHON_EXE=py

cd /d "D:\ETUDEµ\CITE\ETAPE4\PROJET DE FIN DE SESSION\interface web\chantier-securite"
go build -o safesite.exe ./cmd/srv/
safesite.exe
