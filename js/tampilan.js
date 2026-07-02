document.addEventListener('DOMContentLoaded', () => {

    // 0. Draggable Windows Logic
    function makeDraggable(elmnt, handle) {
        let pos1 = 0, pos2 = 0, pos3 = 0, pos4 = 0;
        if (handle) {
            handle.onmousedown = dragMouseDown;
            handle.ontouchstart = dragTouchStart;
        } else {
            elmnt.onmousedown = dragMouseDown;
            elmnt.ontouchstart = dragTouchStart;
        }

        function dragMouseDown(e) {
            e = e || window.event;
            // Prevent drag if clicking input/buttons
            if (e.target.tagName === 'INPUT' || e.target.tagName === 'BUTTON' || e.target.tagName === 'TEXTAREA') return;
            e.preventDefault();
            pos3 = e.clientX;
            pos4 = e.clientY;
            document.onmouseup = closeDragElement;
            document.onmousemove = elementDrag;
        }

        function dragTouchStart(e) {
            if (e.target.tagName === 'INPUT' || e.target.tagName === 'BUTTON' || e.target.tagName === 'TEXTAREA') return;
            pos3 = e.touches[0].clientX;
            pos4 = e.touches[0].clientY;
            document.ontouchend = closeDragElement;
            document.ontouchmove = elementTouchDrag;
        }

        function elementDrag(e) {
            e = e || window.event;
            e.preventDefault();
            pos1 = pos3 - e.clientX;
            pos2 = pos4 - e.clientY;
            pos3 = e.clientX;
            pos4 = e.clientY;

            // Fix transform offset issue by removing it and using straight top/left
            if (window.getComputedStyle(elmnt).transform !== 'none') {
                const rect = elmnt.getBoundingClientRect();
                elmnt.style.transform = 'none';
                elmnt.style.top = rect.top + 'px';
                elmnt.style.left = rect.left + 'px';
            }

            elmnt.style.top = (elmnt.offsetTop - pos2) + "px";
            elmnt.style.left = (elmnt.offsetLeft - pos1) + "px";
        }

        function elementTouchDrag(e) {
            pos1 = pos3 - e.touches[0].clientX;
            pos2 = pos4 - e.touches[0].clientY;
            pos3 = e.touches[0].clientX;
            pos4 = e.touches[0].clientY;

            if (window.getComputedStyle(elmnt).transform !== 'none') {
                const rect = elmnt.getBoundingClientRect();
                elmnt.style.transform = 'none';
                elmnt.style.top = rect.top + 'px';
                elmnt.style.left = rect.left + 'px';
            }

            elmnt.style.top = (elmnt.offsetTop - pos2) + "px";
            elmnt.style.left = (elmnt.offsetLeft - pos1) + "px";
        }

        function closeDragElement() {
            document.onmouseup = null;
            document.onmousemove = null;
            document.ontouchend = null;
            document.ontouchmove = null;
        }
    }

    const terminalEl = document.getElementById('cmd-terminal');
    const terminalHeader = document.getElementById('terminal-header');
    if (terminalEl && terminalHeader) {
        terminalHeader.style.cursor = 'grab';
        makeDraggable(terminalEl, terminalHeader);
    }

    const authCard = document.querySelector('.auth-card');
    const authHeader = document.querySelector('.auth-header');
    if (authCard && authHeader) {
        // authCard is inside flex, need to make it absolute or fixed when dragged
        authHeader.style.cursor = 'grab';
        authHeader.addEventListener('mousedown', () => {
            if (authCard.style.position !== 'fixed') {
                const rect = authCard.getBoundingClientRect();
                authCard.style.position = 'fixed';
                authCard.style.top = rect.top + 'px';
                authCard.style.left = rect.left + 'px';
                authCard.style.margin = '0';
            }
        });
        makeDraggable(authCard, authHeader);
    }
    const AudioContext = window.AudioContext || window.webkitAudioContext;
    let audioCtx = null;

    let isDronePlaying = false;
    function startDrone() {
        if (isDronePlaying || !audioCtx) return;
        isDronePlaying = true;
        const osc1 = audioCtx.createOscillator();
        const osc2 = audioCtx.createOscillator();
        const filter = audioCtx.createBiquadFilter();
        const gainNode = audioCtx.createGain();
        osc1.type = 'sawtooth'; osc1.frequency.value = 55;
        osc2.type = 'square'; osc2.frequency.value = 54.5;
        filter.type = 'lowpass'; filter.frequency.value = 400;
        gainNode.gain.value = 0.05;
        osc1.connect(filter); osc2.connect(filter); filter.connect(gainNode); gainNode.connect(audioCtx.destination);
        osc1.start(); osc2.start();
        setInterval(() => {
            if (audioCtx.state === 'running') {
                const newFreq = 200 + Math.random() * 400;
                filter.frequency.linearRampToValueAtTime(newFreq, audioCtx.currentTime + 2);
            }
        }, 2000);
    }

    function initAudio() {
        if (!audioCtx) {
            audioCtx = new AudioContext();
        }
        if (audioCtx.state === 'suspended') {
            audioCtx.resume();
        }
        startDrone();
    }

    function playTone(freq = 440, type = 'sine', duration = 0.1, vol = 0.05) {
        if (!audioCtx) return;
        const osc = audioCtx.createOscillator();
        const gainNode = audioCtx.createGain();
        osc.type = type;
        osc.frequency.setValueAtTime(freq, audioCtx.currentTime);
        gainNode.gain.setValueAtTime(vol, audioCtx.currentTime);
        gainNode.gain.exponentialRampToValueAtTime(0.001, audioCtx.currentTime + duration);
        osc.connect(gainNode);
        gainNode.connect(audioCtx.destination);
        osc.start();
        osc.stop(audioCtx.currentTime + duration);
    }

    document.body.addEventListener('click', initAudio, { once: true });
    document.body.addEventListener('keydown', initAudio, { once: true });

    document.addEventListener('click', (e) => {
        if (e.target.closest('button') || e.target.closest('a') || e.target.closest('.submit-btn')) {
            playTone(800, 'square', 0.1, 0.05);
        }
    });

    document.addEventListener('mouseover', (e) => {
        if (e.target.closest('button') || e.target.closest('a') || e.target.closest('.tab-btn') || e.target.closest('.project-card')) {
            playTone(400, 'sine', 0.05, 0.01);
        }
    });

    document.addEventListener('keydown', (e) => {
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
            playTone(Math.random() * 200 + 600, 'triangle', 0.05, 0.02);
        }
    });

    window.addEventListener('access-granted', () => {
        initAudio();
        setTimeout(() => playTone(1200, 'square', 0.4, 0.1), 100);
        setTimeout(() => playTone(1600, 'square', 0.6, 0.1), 300);
    });

    const hud = document.createElement('div');
    hud.id = 'cyber-hud';
    document.body.appendChild(hud);

    setInterval(() => {
        const cpu = Math.floor(Math.random() * 20 + 80);
        const mem = '0x' + Math.floor(Math.random() * 0xFFFFFF).toString(16).toUpperCase().padStart(6, '0');
        const net = Math.floor(Math.random() * 900 + 100);
        hud.innerHTML = `CPU: ${cpu}% | MEM: ${mem} | NET: ${net}Mb/s<br>SYS_STATUS: SECURE`;
    }, 500);

    const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$%^&*()";

    function runDecrypt(target) {
        let iteration = 0;
        const original = target.dataset.value;

        clearInterval(target.interval);
        target.interval = setInterval(() => {
            target.innerText = original.split("").map((letter, index) => {
                if (index < iteration) {
                    return original[index];
                }
                return letters[Math.floor(Math.random() * letters.length)]
            }).join("");

            if (iteration >= original.length) {
                clearInterval(target.interval);
            }
            iteration += 1 / 3;
        }, 30);
    }

    const headingObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                if (!entry.target.dataset.value) {
                    entry.target.dataset.value = entry.target.innerText;
                }
                runDecrypt(entry.target);
                headingObserver.unobserve(entry.target);
            }
        });
    }, { threshold: 0.5 });

    document.querySelectorAll('h2').forEach(h2 => {
        h2.dataset.value = h2.innerText;
        headingObserver.observe(h2);
    });

    window.addEventListener('app-ready', () => {
        document.querySelectorAll('#main-app h2').forEach(h2 => {
            headingObserver.observe(h2);
        });
    });

    const pElements = document.querySelectorAll('p');
    pElements.forEach(p => {
        if (p.id === 'typing-text' || p.id === 'age-display' || p.closest('#cheat-hud') || p.closest('#cmd-terminal') || p.closest('.auth-wrapper') || p.querySelector('a')) return;

        p.addEventListener('mouseenter', () => {
            if (p.isScrambling || !p.innerText) return;
            p.isScrambling = true;
            if (!p.dataset.original) p.dataset.original = p.innerText;

            let iteration = 0;
            const original = p.dataset.original;
            clearInterval(p.interval);

            p.interval = setInterval(() => {
                p.innerText = original.split("").map((letter, index) => {
                    if (index < iteration) return original[index];
                    if (letter === ' ') return ' ';
                    return letters[Math.floor(Math.random() * letters.length)];
                }).join("");

                if (iteration >= original.length) {
                    clearInterval(p.interval);
                    p.isScrambling = false;
                    p.innerText = original;
                }
                iteration += 1;
            }, 10);
        });
    });

    document.querySelectorAll('.card').forEach(card => {
        card.addEventListener('mousemove', (e) => {
            const rect = card.getBoundingClientRect();
            const x = e.clientX - rect.left;
            const y = e.clientY - rect.top;

            const centerX = rect.width / 2;
            const centerY = rect.height / 2;

            const rotateX = ((y - centerY) / centerY) * -6;
            const rotateY = ((x - centerX) / centerX) * 6;

            card.style.transform = `perspective(1000px) rotateX(${rotateX}deg) rotateY(${rotateY}deg)`;
        });

        card.addEventListener('mouseleave', () => {
            card.style.transform = `perspective(1000px) rotateX(0) rotateY(0)`;
            card.style.transition = `transform 0.5s ease`;
        });

        card.addEventListener('mouseenter', () => {
            card.style.transition = `none`;
        });
    });

    const cursor = document.getElementById('cyber-cursor');
    const trailContainer = document.getElementById('cursor-trail-container');
    if (cursor && trailContainer && window.innerWidth > 768) {
        let mouseX = window.innerWidth / 2, mouseY = window.innerHeight / 2;
        let cursorX = mouseX, cursorY = mouseY;

        document.addEventListener('mousemove', (e) => {
            mouseX = e.clientX;
            mouseY = e.clientY;

            if (Math.random() > 0.3) {
                const trail = document.createElement('div');
                trail.className = 'cursor-trail';
                trail.style.left = mouseX + 'px';
                trail.style.top = mouseY + 'px';
                trailContainer.appendChild(trail);

                setTimeout(() => {
                    trail.style.opacity = '0';
                    trail.style.transition = 'opacity 0.4s';
                    setTimeout(() => trail.remove(), 400);
                }, 20);
            }
        });

        function animateCursor() {
            cursorX += (mouseX - cursorX) * 0.4;
            cursorY += (mouseY - cursorY) * 0.4;
            cursor.style.left = cursorX + 'px';
            cursor.style.top = cursorY + 'px';
            requestAnimationFrame(animateCursor);
        }
        animateCursor();

        document.addEventListener('mousedown', () => cursor.style.transform = 'translate(-50%, -50%) scale(0.8)');
        document.addEventListener('mouseup', () => cursor.style.transform = 'translate(-50%, -50%) scale(1)');

        const interactables = document.querySelectorAll('a, button, input, textarea, .tab-btn, .project-card, .skill-item');
        interactables.forEach(el => {
            el.addEventListener('mouseenter', () => {
                cursor.style.width = '40px';
                cursor.style.height = '40px';
                cursor.style.borderColor = 'var(--text-primary)';
            });
            el.addEventListener('mouseleave', () => {
                cursor.style.width = '24px';
                cursor.style.height = '24px';
                cursor.style.borderColor = 'var(--accent-color)';
            });
        });
    }

    const holoCanvas = document.getElementById('hologram-canvas');
    if (holoCanvas) {
        const hCtx = holoCanvas.getContext('2d');
        let width = 200, height = 200;
        holoCanvas.width = width;
        holoCanvas.height = height;

        const phi = (1 + Math.sqrt(5)) / 2;
        let vertices = [
            [-1, phi, 0], [1, phi, 0], [-1, -phi, 0], [1, -phi, 0],
            [0, -1, phi], [0, 1, phi], [0, -1, -phi], [0, 1, -phi],
            [phi, 0, -1], [phi, 0, 1], [-phi, 0, -1], [-phi, 0, 1]
        ];

        vertices = vertices.map(v => {
            const len = Math.sqrt(v[0] * v[0] + v[1] * v[1] + v[2] * v[2]);
            return [v[0] / len * 60, v[1] / len * 60, v[2] / len * 60];
        });

        const edges = [
            [0, 11], [0, 5], [0, 1], [0, 7], [0, 10], [1, 5], [1, 9], [1, 8], [1, 7], [5, 11], [5, 4], [5, 9],
            [11, 10], [11, 2], [11, 4], [10, 7], [10, 6], [10, 2], [7, 8], [7, 6], [9, 8], [9, 3], [9, 4],
            [8, 6], [8, 3], [6, 2], [6, 3], [2, 4], [2, 3], [4, 3]
        ];

        let angleX = 0, angleY = 0;
        let targetAngleX = 0, targetAngleY = 0;

        document.addEventListener('mousemove', (e) => {
            const x = (e.clientX / window.innerWidth) - 0.5;
            const y = (e.clientY / window.innerHeight) - 0.5;
            targetAngleY = x * Math.PI;
            targetAngleX = y * Math.PI;
        });

        function hexToRgb(hex) {
            if (hex.startsWith('#')) {
                let r = 0, g = 0, b = 0;
                if (hex.length === 4) { r = parseInt(hex[1] + hex[1], 16); g = parseInt(hex[2] + hex[2], 16); b = parseInt(hex[3] + hex[3], 16); }
                else if (hex.length === 7) { r = parseInt(hex.substring(1, 3), 16); g = parseInt(hex.substring(3, 5), 16); b = parseInt(hex.substring(5, 7), 16); }
                return `${r},${g},${b}`;
            }
            return '0,255,255';
        }

        function drawHologram() {
            hCtx.clearRect(0, 0, width, height);

            angleX += (targetAngleX - angleX) * 0.05;
            angleY += (targetAngleY - angleY) * 0.05 + 0.01;

            const cosX = Math.cos(angleX), sinX = Math.sin(angleX);
            const cosY = Math.cos(angleY), sinY = Math.sin(angleY);

            const projected = vertices.map(v => {
                let x1 = v[0] * cosY - v[2] * sinY;
                let z1 = v[0] * sinY + v[2] * cosY;
                let y1 = v[1];
                let y2 = y1 * cosX - z1 * sinX;
                let z2 = y1 * sinX + z1 * cosX;
                let x2 = x1;

                const fov = 150;
                const scale = fov / (fov + z2);
                return [x2 * scale + width / 2, y2 * scale + height / 2, z2];
            });

            const rawAccent = getComputedStyle(document.body).getPropertyValue('--accent-color').trim() || '#00ffff';
            const rgbAccent = hexToRgb(rawAccent);

            edges.forEach(edge => {
                const p1 = projected[edge[0]];
                const p2 = projected[edge[1]];

                hCtx.beginPath();
                hCtx.moveTo(p1[0], p1[1]);
                hCtx.lineTo(p2[0], p2[1]);

                const zAvg = (p1[2] + p2[2]) / 2;
                const alpha = Math.max(0.05, 1 - (zAvg + 60) / 120);

                hCtx.strokeStyle = `rgba(${rgbAccent}, ${alpha})`;
                hCtx.lineWidth = 1.5;
                hCtx.stroke();
            });

            projected.forEach(p => {
                hCtx.beginPath();
                hCtx.arc(p[0], p[1], 2, 0, Math.PI * 2);
                hCtx.fillStyle = `rgb(${rgbAccent})`;
                hCtx.fill();
            });

            requestAnimationFrame(drawHologram);
        }
        drawHologram();
    }

    // 7. Cyber Chat Logic
    const chatBox = document.getElementById('cyber-chat');
    const closeChatBtn = document.getElementById('close-chat');
    const chatInput = document.getElementById('chat-input');
    const chatMessages = document.getElementById('chat-messages');

    if (chatBox && closeChatBtn && chatInput && chatMessages) {
        // Setup initial hacker chat message silently (no auto-popup)
        window.addEventListener('app-ready', () => {
            setTimeout(() => {
                chatInput.disabled = false;
                chatInput.placeholder = "Enter message...";
                addChatMessage("UNKNOWN_HACKER", "I see you've accessed the system. Are you ready?", false);
            }, 5000);
        });

        closeChatBtn.addEventListener('click', () => {
            chatBox.classList.add('hidden');
            const navChatBtn = document.getElementById('nav-chat-btn');
            if (navChatBtn) navChatBtn.classList.remove('active');
            const mobileChatBtn = document.getElementById('mobile-chat-btn');
            if (mobileChatBtn) mobileChatBtn.classList.remove('active');
        });

        function addChatMessage(sender, text, isUser) {
            const msgDiv = document.createElement('div');
            msgDiv.className = 'chat-msg' + (isUser ? ' user-msg' : '');
            msgDiv.innerHTML = `<strong>${sender}:</strong> ${text}`;
            chatMessages.appendChild(msgDiv);
            chatMessages.scrollTop = chatMessages.scrollHeight;
        }

        chatInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                const text = chatInput.value.trim();
                if (text) {
                    addChatMessage("YOU", text, true);
                    chatInput.value = '';

                    // AI Assistant Logic
                    setTimeout(() => {
                        const lowerText = text.toLowerCase();
                        let reply = "";

                        if (lowerText.match(/gaji|salary|bayaran/)) {
                            reply = "Ekspektasi gaji bisa dinegosiasikan. Saya lebih tertarik pada seberapa menantang tech stack yang digunakan dan lingkungan yang kolaboratif.";
                        } else if (lowerText.match(/react|vue|framework|bisa apa|skill/)) {
                            reply = "Tech stack utama saya berkisar di ekosistem web modern. Saya memiliki fondasi Vanilla JS yang sangat kuat, sehingga beradaptasi dengan React/Vue/Next.js adalah proses yang mudah bagi saya.";
                        } else if (lowerText.match(/lokasi|alamat|tinggal|dimana/)) {
                            reply = "Saya berlokasi di Kesugihan, Cilacap. Terbuka penuh untuk pekerjaan remote atau bersedia relokasi jika tawarannya tepat.";
                        } else if (lowerText.match(/kuliah|kampus|pendidikan|nim/)) {
                            reply = "Saat ini saya sedang menempuh pendidikan Informatika (NIM: 24EO10021). Passion saya lebih ke arah praktek dan implementasi langsung di lapangan.";
                        } else if (lowerText.match(/halo|hi|hai|hello/)) {
                            reply = "Halo. Saya adalah FOS Assistant. Ada yang bisa saya bantu terkait Profil Farid Donovant?";
                        } else if (lowerText.match(/siapa kamu|about/)) {
                            reply = "Halo. Saya adalah FOS Assistant. Ada yang bisa saya bantu terkait Profil Farid Donovant?";
                        } else if (lowerText.match(/help|tolong|bantuan/)) {
                            reply = "Tentu, list pertanyaan yang bisa diajukan : 'Gaji' atau 'Salary' untuk mengetahui ekspektasi gaji, 'Skill' atau 'Bisa apa' untuk mengetahui skill yang dimiliki, 'Lokasi' atau 'Alamat' untuk mengetahui lokasi, 'Kuliah' atau 'Kampus' atau 'Pendidikan' atau 'Nim' untuk mengetahui pendidikan, 'Halo' atau 'Hi' atau 'Hai' atau 'Hello' untuk memulai percakapan, 'Help' atau 'Tolong' atau 'Bantuan' untuk mengetahui bantuan.";
                        } else {
                            const defaultReplies = [
                                "Menarik. Sistem sedang memproses input Anda...",
                                "Pertanyaan yang bagus. Namun data tersebut terenkripsi.",
                                "Silakan cek langsung ke author sistem ini.",
                                "Saya mendeteksi anomali pada query tersebut. Coba pertanyaan lain terkait 'skill' atau 'lokasi'."
                            ];
                            reply = defaultReplies[Math.floor(Math.random() * defaultReplies.length)];
                        }

                        addChatMessage("FOS_AI", reply, false);
                        playTone(1000, 'square', 0.1, 0.05);
                    }, 800 + Math.random() * 1000);
                }
            }
        });
    }

    // 8. Sound Toggle Logic
    const soundToggle = document.getElementById('sound-toggle-btn');
    window.isSoundMuted = false;
    if (soundToggle) {
        soundToggle.addEventListener('click', () => {
            window.isSoundMuted = !window.isSoundMuted;
            if (window.isSoundMuted) {
                soundToggle.innerText = '[SOUND: OFF]';
                if (audioCtx && audioCtx.state === 'running') {
                    audioCtx.suspend();
                }
            } else {
                soundToggle.innerText = '[SOUND: ON]';
                if (audioCtx && audioCtx.state === 'suspended') {
                    audioCtx.resume();
                }
                // Initialize if never started
                if (!audioCtx) initAudio();
            }
        });
    }

    // 9. Easter Eggs (Konami Code)
    const konamiCode = ['h', 'e', 's', 'o', 'y', 'a', 'm'];
    let konamiIndex = 0;

    document.addEventListener('keydown', (e) => {
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

        if (e.key === konamiCode[konamiIndex] || e.key === konamiCode[konamiIndex].toLowerCase()) {
            konamiIndex++;
            if (konamiIndex === konamiCode.length) {
                // Konami Code Triggered!
                konamiIndex = 0;
                window.matrixColorOverride = '#ffd700'; // Gold Color
                window.matrixSpeedMultiplier = 5;
                setTimeout(() => window.matrixSpeedMultiplier = 1, 2000);

                if (window.showCyberToast) {
                    window.showCyberToast('[ACHIEVEMENT UNLOCKED: Hacker Master]', 'success');
                    setTimeout(() => window.showCyberToast('[GOD_MODE: ENABLED] SYSTEM OVERRIDE ACCEPTED', 'success'), 1500);
                } else {
                    alert('GOD_MODE ENABLED');
                }

                // Play awesome sound
                setTimeout(() => playTone(800, 'square', 0.2, 0.1), 0);
                setTimeout(() => playTone(1200, 'square', 0.2, 0.1), 200);
                setTimeout(() => playTone(1600, 'square', 0.2, 0.1), 400);

                // GTA HESOYAM Visual Effect
                const moneyDrop = document.createElement('div');
                moneyDrop.innerHTML = '+ $250,000<br><span style="font-size: 0.5em; color: white;">FULL HEALTH & ARMOR</span>';
                moneyDrop.style.position = 'fixed';
                moneyDrop.style.top = '50%';
                moneyDrop.style.left = '50%';
                moneyDrop.style.transform = 'translate(-50%, -50%)';
                moneyDrop.style.color = '#32cd32'; // GTA green
                moneyDrop.style.fontFamily = 'Impact, sans-serif';
                moneyDrop.style.fontSize = '4rem';
                moneyDrop.style.fontWeight = 'bold';
                moneyDrop.style.textShadow = '2px 2px 0 #000, -2px -2px 0 #000, 2px -2px 0 #000, -2px 2px 0 #000';
                moneyDrop.style.zIndex = '9999999';
                moneyDrop.style.textAlign = 'center';
                moneyDrop.style.transition = 'all 2s ease-out';
                moneyDrop.style.pointerEvents = 'none';
                document.body.appendChild(moneyDrop);

                setTimeout(() => {
                    moneyDrop.style.top = '20%';
                    moneyDrop.style.opacity = '0';
                }, 100);

                setTimeout(() => moneyDrop.remove(), 2100);

                // Auto login to CMS if not logged in
                if (!sessionStorage.getItem('porto_current_user')) {
                    sessionStorage.setItem('porto_current_user', 'GOD_MODE');
                    const authContainer = document.getElementById('auth-container');
                    const mainApp = document.getElementById('main-app');
                    if (authContainer) authContainer.style.display = 'none';
                    if (mainApp) mainApp.style.display = 'block';
                    window.dispatchEvent(new Event('app-ready'));
                }
            }
        } else {
            konamiIndex = 0;
        }
    });
    // --- Desktop Header Toggle Navigation Buttons ---
    const navManualBtn = document.getElementById('nav-manual-btn');
    const navChatBtn = document.getElementById('nav-chat-btn');
    const navLogsBtn = document.getElementById('nav-logs-btn');

    if (navManualBtn) {
        navManualBtn.addEventListener('click', () => {
            const cheatHud = document.getElementById('cheat-hud');
            if (cheatHud) {
                cheatHud.classList.toggle('show-desktop');
                navManualBtn.classList.toggle('active');
            }
        });
    }

    if (navChatBtn) {
        navChatBtn.addEventListener('click', () => {
            const chatBox = document.getElementById('cyber-chat');
            if (chatBox) {
                chatBox.classList.toggle('hidden');
                const chatInput = document.getElementById('chat-input');
                if (chatInput && !chatBox.classList.contains('hidden')) {
                    chatInput.disabled = false;
                    chatInput.placeholder = "Enter message...";
                }
                navChatBtn.classList.toggle('active');
            }
        });
    }

    if (navLogsBtn) {
        navLogsBtn.addEventListener('click', () => {
            const logsContainer = document.getElementById('live-logs-container');
            if (logsContainer) {
                logsContainer.classList.toggle('show-desktop');
                navLogsBtn.classList.toggle('active');
            }
        });
    }

});
