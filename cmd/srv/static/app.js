// SafeSite AI - Main Application JS

document.addEventListener('DOMContentLoaded', function() {
    // Update clock
    updateClock();
    setInterval(updateClock, 1000);
    
    // Sidebar toggle
    const sidebarToggle = document.getElementById('sidebarCollapse');
    if (sidebarToggle) {
        sidebarToggle.addEventListener('click', function() {
            document.getElementById('sidebar').classList.toggle('active');
        });
    }
    
    // Emergency button
    const emergencyBtn = document.getElementById('emergency-btn');
    if (emergencyBtn) {
        emergencyBtn.addEventListener('click', function() {
            if (confirm('Confirmer l\'appel d\'urgence ?')) {
                alert('Num\u00e9ro d\'urgence: 01 23 45 67 89\nPoste de s\u00e9curit\u00e9 alert\u00e9.');
            }
        });
    }
    
    // Update alertes badge periodically
    updateAlertesBadge();
    setInterval(updateAlertesBadge, 30000);
});

function updateClock() {
    const clockEl = document.getElementById('current-time');
    if (clockEl) {
        const now = new Date();
        clockEl.textContent = now.toLocaleString('fr-FR', {
            weekday: 'short',
            day: '2-digit',
            month: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    }
}

async function updateAlertesBadge() {
    try {
        const resp = await fetch('/api/stats');
        const stats = await resp.json();
        const badge = document.getElementById('alertes-badge');
        if (badge) {
            badge.textContent = stats.alertes_actives || 0;
            badge.style.display = stats.alertes_actives > 0 ? 'inline' : 'none';
        }
    } catch (e) {
        console.warn('Could not update alertes badge:', e);
    }
}

// Chart.js default configuration
if (typeof Chart !== 'undefined') {
    Chart.defaults.font.family = "'Segoe UI', system-ui, sans-serif";
    Chart.defaults.color = '#6c757d';
}
