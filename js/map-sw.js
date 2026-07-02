document.addEventListener('DOMContentLoaded', () => {
    initCyberMap();

    initServiceWorkerUI();

    setupAppEventNotifications();
});

let cyberMap = null;

function initCyberMap() {
    const mapElement = document.getElementById('map');
    if (!mapElement) return;

    if (typeof L === 'undefined') {
        console.warn('[Map] Leaflet library is not loaded. Displaying offline fallback.');
        mapElement.innerHTML = `
            <div style="display: flex; flex-direction: column; justify-content: center; align-items: center; height: 100%; color: var(--error-color); font-family: 'Fira Code', monospace; padding: 20px; text-align: center; border: 1px dashed var(--error-color); background: rgba(165, 13, 13, 0.56);">
                <div style="font-weight: bold; margin-bottom: 8px; font-size: 1.1rem;">[!] ERROR: MAP_UPLINK_OFFLINE</div>
                <div style="font-size: 0.8rem; color: var(--text-secondary); max-width: 400px; line-height: 1.5;">
                    Gagal menghubungi modul satelit peta (Leaflet CDN tidak terjangkau). 
                    Pastikan perangkat terhubung ke internet.
                </div>
            </div>
        `;
        return;
    }
    const targetCoords = [-7.6186, 109.0834];
    cyberMap = L.map('map', {
        zoomControl: true,
        scrollWheelZoom: false
    }).setView(targetCoords, 14);
    L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright" style="color:var(--text-secondary)">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions" style="color:var(--text-secondary)">CARTO</a>',
        subdomains: 'abcd',
        maxZoom: 20
    }).addTo(cyberMap);

    const cyberIcon = L.divIcon({
        className: 'cyber-map-marker',
        html: '<div class="marker-pulse"></div><div class="marker-core"></div>',
        iconSize: [40, 40],
        iconAnchor: [20, 20]
    });
    const marker = L.marker(targetCoords, { icon: cyberIcon }).addTo(cyberMap);
    const popupContent = `
        <div class="map-popup-cyber">
            <h4 style="color: var(--accent-color); border-bottom: 1px dashed var(--text-primary); padding-bottom: 5px; margin-bottom: 5px; font-family: 'Fira Code', monospace; text-transform: uppercase;">SYS_TARGET: HOME_BASE</h4>
            <p style="color: var(--text-primary); margin: 0; font-size: 0.85rem; font-family: 'Fira Code', monospace;">Muhammad Farid Donovant</p>
            <p style="color: var(--text-secondary); margin: 0; font-size: 0.8rem; font-family: 'Fira Code', monospace;">Kesugihan, Cilacap, Jateng</p>
        </div>
    `;

    marker.bindPopup(popupContent).openPopup();

    cyberMap.on('click', () => {
        cyberMap.scrollWheelZoom.enable();
    });
    window.addEventListener('app-ready', () => {
        setTimeout(() => {
            if (cyberMap) {
                cyberMap.invalidateSize();
                console.log('[Map] Size invalidated - App Ready');
            }
        }, 300);
    });

    window.addEventListener('access-granted', () => {
        setTimeout(() => {
            if (cyberMap) {
                cyberMap.invalidateSize();
                console.log('[Map] Size invalidated - Access Granted');
            }
        }, 800);
    });
}

function initServiceWorkerUI() {
    const swStatusVal = document.getElementById('sw-status-val');
    const notifyStatusVal = document.getElementById('notify-status-val');
    const btnRegisterSw = document.getElementById('btn-register-sw');
    const btnUnregisterSw = document.getElementById('btn-unregister-sw');
    const btnRequestNotify = document.getElementById('btn-request-notify');
    const btnTestNotify = document.getElementById('btn-test-notify');

    if (!swStatusVal) return;

    if (!('serviceWorker' in navigator)) {
        swStatusVal.textContent = 'UNSUPPORTED_BROWSER';
        swStatusVal.style.color = 'var(--error-color)';
        if (btnRegisterSw) btnRegisterSw.disabled = true;
        if (btnUnregisterSw) btnUnregisterSw.disabled = true;
        return;
    }

    function updateSWStatus() {
        navigator.serviceWorker.getRegistrations().then(registrations => {
            const hasActiveSW = registrations.some(reg => reg.active);
            if (hasActiveSW) {
                swStatusVal.textContent = 'ACTIVE_SECURE';
                swStatusVal.style.color = 'var(--text-primary)';
                if (btnRegisterSw) btnRegisterSw.style.display = 'none';
                if (btnUnregisterSw) btnUnregisterSw.style.display = 'inline-block';
            } else {
                swStatusVal.textContent = 'INACTIVE_DISCONNECTED';
                swStatusVal.style.color = 'var(--error-color)';
                if (btnRegisterSw) btnRegisterSw.style.display = 'inline-block';
                if (btnUnregisterSw) btnUnregisterSw.style.display = 'none';
            }
        });
    }

    function updateNotificationStatus() {
        if (!('Notification' in window)) {
            notifyStatusVal.textContent = 'UNSUPPORTED';
            notifyStatusVal.style.color = 'var(--error-color)';
            if (btnRequestNotify) btnRequestNotify.disabled = true;
            if (btnTestNotify) btnTestNotify.disabled = true;
            return;
        }

        const permission = Notification.permission;
        if (permission === 'granted') {
            notifyStatusVal.textContent = 'GRANTED_AUTHORIZED';
            notifyStatusVal.style.color = 'var(--text-primary)';
            if (btnRequestNotify) btnRequestNotify.style.display = 'none';
            if (btnTestNotify) btnTestNotify.style.display = 'inline-block';
        } else if (permission === 'denied') {
            notifyStatusVal.textContent = 'DENIED_BLOCKED';
            notifyStatusVal.style.color = 'var(--error-color)';
            if (btnRequestNotify) {
                btnRequestNotify.style.display = 'inline-block';
                btnRequestNotify.textContent = 'RESET_IN_BROWSER_SETTINGS';
                btnRequestNotify.disabled = true;
            }
            if (btnTestNotify) btnTestNotify.style.display = 'none';
        } else {
            notifyStatusVal.textContent = 'DEFAULT_AWAITING';
            notifyStatusVal.style.color = 'var(--warning-color)';
            if (btnRequestNotify) btnRequestNotify.style.display = 'inline-block';
            if (btnTestNotify) btnTestNotify.style.display = 'none';
        }
    }

    updateSWStatus();
    updateNotificationStatus();
    navigator.serviceWorker.register('/sw.js')
        .then(reg => {
            console.log('[Service Worker] Auto-registered successfully:', reg.scope);
            updateSWStatus();
        })
        .catch(err => {
            console.error('[Service Worker] Auto-registration failed:', err);
            updateSWStatus();
        });

    if (btnRegisterSw) {
        btnRegisterSw.addEventListener('click', () => {
            navigator.serviceWorker.register('/sw.js')
                .then(reg => {
                    console.log('[Service Worker] Registered:', reg);
                    if (window.showCyberToast) window.showCyberToast("SERVICE_WORKER REGISTERED");
                    updateSWStatus();
                })
                .catch(err => {
                    console.error('[Service Worker] Registration failed:', err);
                    if (window.showCyberToast) window.showCyberToast("REGISTRATION FAILED", "error");
                    updateSWStatus();
                });
        });
    }
    if (btnUnregisterSw) {
        btnUnregisterSw.addEventListener('click', () => {
            navigator.serviceWorker.getRegistrations().then(registrations => {
                for (let reg of registrations) {
                    reg.unregister().then(success => {
                        if (success) {
                            console.log('[Service Worker] Unregistered successfully');
                            if (window.showCyberToast) window.showCyberToast("SERVICE_WORKER UNREGISTERED", "error");
                            updateSWStatus();
                        }
                    });
                }
            });
        });
    }

    // Request Notification Permission Button Click
    if (btnRequestNotify) {
        btnRequestNotify.addEventListener('click', () => {
            Notification.requestPermission().then(permission => {
                console.log('[Notification] Permission state:', permission);
                if (window.showCyberToast) {
                    if (permission === 'granted') {
                        window.showCyberToast("NOTIFICATIONS PERMITTED");
                    } else {
                        window.showCyberToast("NOTIFICATIONS BLOCKED", "error");
                    }
                }
                updateNotificationStatus();
            });
        });
    }

    // Test Notification Button Click
    if (btnTestNotify) {
        btnTestNotify.addEventListener('click', () => {
            triggerNotification(
                'CYBERNEXUS SECURE UPLINK',
                'Secure telemetry connection verified. Push notifications active.'
            );
        });
    }
}

// Global helper to trigger notifications via Service Worker
function triggerNotification(title, body) {
    if (!('Notification' in window) || Notification.permission !== 'granted') {
        console.warn('Notifications not permitted or unsupported.');
        return;
    }

    // Send postMessage to Service Worker to trigger notification showNotification
    if (navigator.serviceWorker.controller) {
        navigator.serviceWorker.controller.postMessage({
            type: 'SHOW_NOTIFICATION',
            title: title,
            body: body
        });
    } else {
        // Fallback if Service Worker is not fully controlling page yet
        navigator.serviceWorker.ready.then(reg => {
            if (reg.active) {
                reg.active.postMessage({
                    type: 'SHOW_NOTIFICATION',
                    title: title,
                    body: body
                });
            } else {
                new Notification(title, { body: body, icon: '/Farid.jpg' });
            }
        });
    }
}

// Helper to hook into existing application events and trigger notifications
function setupAppEventNotifications() {

    const contactForm = document.getElementById('contact-form');
    if (contactForm) {
        contactForm.addEventListener('submit', () => {
            // Wait a brief moment to check if success text or toast appeared
            setTimeout(() => {
                const successMsg = document.getElementById('form-success');
                if (successMsg && successMsg.textContent.includes('berhasil')) {
                    triggerNotification(
                        'SECURE DATA PACKET SENT',
                        'Your communication packet has been logged in the server mainframe database.'
                    );
                }
            }, 1200);
        });
    }

    // Article Form success handler hook
    const articleForm = document.getElementById('articleForm');
    if (articleForm) {
        articleForm.addEventListener('submit', () => {
            setTimeout(() => {

                triggerNotification(
                    'BROADCAST PROTOCOL ACTIVE',
                    'A new database record has been written to system articles.'
                );
            }, 1000);
        });
    }

    // Login Form Success Notification Hook
    const authForm = document.getElementById('auth-form');
    if (authForm) {
        authForm.addEventListener('submit', () => {
            setTimeout(() => {
                const token = sessionStorage.getItem('porto_token');
                if (token) {
                    triggerNotification(
                        'ACCESS GRANTED (SYS_ADMIN)',
                        `Session initialized for operator: ${sessionStorage.getItem('username') || 'Unknown'}`
                    );
                }
            }, 1500);
        });
    }

    // Logout Notification Hook
    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', () => {
            triggerNotification(
                'SESSION TERMINATED',
                'Decryption keys cleared. Connection locked.'
            );
        });
    }
}
