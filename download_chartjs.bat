@echo off
echo Téléchargement de Chart.js...
if not exist "static\js" mkdir "static\js"
curl -L "https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js" -o "static\js\chart.umd.min.js"
echo Terminé. Fichier enregistré dans static\js\chart.umd.min.js
