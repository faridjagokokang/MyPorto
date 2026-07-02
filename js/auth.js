document.addEventListener('DOMContentLoaded', function () {
    const authContainer = document.getElementById('auth-container');
    const mainApp = document.getElementById('main-app');
    const authForm = document.getElementById('auth-form');
    const authTitle = document.getElementById('auth-title');
    const authSubmitBtn = document.getElementById('auth-submit-btn');
    const toggleAuthBtn = document.getElementById('toggle-auth');
    const authMsg = document.getElementById('auth-msg');
    const logoutBtn = document.getElementById('logout-btn');

    let isLoginMode = true;

    checkAuthStatus();

    function checkAuthStatus() {
        const currentUser = sessionStorage.getItem('porto_current_user');
        if (currentUser) {
            authContainer.style.display = 'none';
            mainApp.style.display = 'block';

            setTimeout(() => {
                window.dispatchEvent(new Event('start-typing'));
                window.dispatchEvent(new Event('app-ready'));
            }, 500);

            const greetingElement = document.getElementById('greeting');
            if (greetingElement) {
                const hour = new Date().getHours();
                let timeGreeting = '';
                if (hour >= 5 && hour < 12) {
                    timeGreeting = 'Good morning';
                } else if (hour >= 12 && hour < 17) {
                    timeGreeting = 'Good afternoon';
                } else if (hour >= 17 && hour < 21) {
                    timeGreeting = 'Good evening';
                } else {
                    timeGreeting = 'Good night';
                }
                greetingElement.textContent = timeGreeting + ', ' + currentUser;
            }
        } else {
            authContainer.style.display = 'block';
            mainApp.style.display = 'none';
        }
    }

    const emailGroup = document.getElementById('auth-email-group');
    const emailInput = document.getElementById('auth-email');

    if (toggleAuthBtn) {
        toggleAuthBtn.addEventListener('click', function (e) {
            e.preventDefault();
            isLoginMode = !isLoginMode;
            authMsg.textContent = '';

            if (isLoginMode) {
                authTitle.textContent = 'SYSTEM_LOGIN';
                authSubmitBtn.textContent = 'EXECUTE_LOGIN';
                toggleAuthBtn.innerHTML = 'DON\'T HAVE ACCOUNT? <span>CREATE_ACCOUNT</span>';
                if (emailGroup) emailGroup.style.display = 'none';
                if (emailInput) emailInput.removeAttribute('required');
            } else {
                authTitle.textContent = 'SYSTEM_REGISTER';
                authSubmitBtn.textContent = 'EXECUTE_REGISTER';
                toggleAuthBtn.innerHTML = 'ALREADY HAVE ACCOUNT? <span>LOGIN</span>';
                if (emailGroup) emailGroup.style.display = 'block';
                if (emailInput) emailInput.setAttribute('required', 'required');
            }
        });
    }

    if (authForm) {
        authForm.addEventListener('submit', function (e) {
            e.preventDefault();
            const usernameInput = document.getElementById('auth-username').value.trim();
            const passwordInput = document.getElementById('auth-password').value.trim();
            const emailValue = emailInput ? emailInput.value.trim() : '';

            if (!usernameInput || !passwordInput || (!isLoginMode && !emailValue)) {
                authMsg.style.color = '#ef4444';
                authMsg.textContent = 'Harap isi semua field.';
                return;
            }

            if (isLoginMode) {
                fetch('/api/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username: usernameInput, password: passwordInput })
                })
                    .then(res => {
                        if (res.ok) return res.json();
                        throw new Error('Login failed');
                    })
                    .then(data => {
                        window.dispatchEvent(new Event('access-granted'));
                        authMsg.style.color = '#22c55e';
                        authMsg.textContent = 'Login berhasil! Memuat profil...';

                        sessionStorage.setItem('porto_current_user', data.username);
                        sessionStorage.setItem('porto_token', data.token);
                        sessionStorage.setItem('porto_role', data.role);

                        setTimeout(() => {
                            authForm.reset();
                            authMsg.textContent = '';
                            checkAuthStatus();
                        }, 1000);
                    })
                    .catch(err => {
                        authMsg.style.color = '#ef4444';
                        authMsg.textContent = 'Username atau password salah.';
                    });
            } else {
                fetch('/api/register', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        username: usernameInput,
                        email: emailValue,
                        password: passwordInput
                    })
                })
                    .then(res => {
                        if (res.ok) {
                            authMsg.style.color = '#22c55e';
                            authMsg.textContent = 'Pendaftaran berhasil! Silakan login.';

                            setTimeout(() => {
                                authForm.reset();
                                authMsg.textContent = '';
                                toggleAuthBtn.click();
                            }, 1500);
                        } else if (res.status === 409) {
                            authMsg.style.color = '#ef4444';
                            authMsg.textContent = 'Username sudah digunakan. Silakan pilih yang lain.';
                        } else {
                            throw new Error('Registration failed');
                        }
                    })
                    .catch(err => {
                        authMsg.style.color = '#ef4444';
                        authMsg.textContent = 'Gagal melakukan pendaftaran.';
                    });
            }
        });
    }

    if (logoutBtn) {
        logoutBtn.addEventListener('click', function (e) {
            e.preventDefault();
            if (confirm('Yakin ingin logout?')) {
                sessionStorage.removeItem('porto_current_user');
                sessionStorage.removeItem('porto_token');
                sessionStorage.removeItem('porto_role');
                window.dispatchEvent(new Event('app-logout'));
                checkAuthStatus();
            }
        });
    }
});
