const form = document.getElementById('articleForm');
const titleInput = document.getElementById('title');
const authorInput = document.getElementById('author');
const contentInput = document.getElementById('content');
const articleIdInput = document.getElementById('articleId');
const formTitle = document.getElementById('formTitle');
const btnSubmit = document.getElementById('btnSubmit');
const btnCancel = document.getElementById('btnCancel');
const articleList = document.getElementById('articleList');

window.showCyberToast = function(message, type = 'success') {
    const container = document.getElementById('cyber-toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `cyber-toast ${type === 'error' ? 'toast-error' : ''}`;
    
    const icon = type === 'error' ? '[!]' : '[+]';
    toast.innerHTML = `<strong>${icon} SYSTEM:</strong> ${message}`;
    
    container.appendChild(toast);
    
    // Trigger reflow to animate
    setTimeout(() => {
        toast.classList.add('show');
    }, 10);
    
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => {
            toast.remove();
        }, 300);
    }, 3000);
};

let articles = [];
let isEditing = false;

async function fetchArticles() {
    try {
        const response = await fetch('/api/articles');
        if (response.ok) {
            articles = await response.json();
            if (!articles) articles = [];
            renderArticles();
        }
    } catch (err) {
        console.error('Error fetching articles:', err);
    }
}

fetchArticles();


form.addEventListener('submit', async function (e) {
    e.preventDefault();
    if (!titleInput.value || !authorInput.value || !contentInput.value) {
        showCyberToast("Harap isi semua field!", "error");
        return;
    }

    const articleData = {
        id: isEditing ? articleIdInput.value : Date.now().toString(),
        title: titleInput.value,
        author: authorInput.value,
        content: contentInput.value
    };

    try {
        const token = sessionStorage.getItem('porto_token');
        const res = await fetch('/api/articles', {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(articleData)
        });
        
        if (res.ok) {
            if (isEditing) {
                showCyberToast("Artikel berhasil diubah!");
            } else {
                showCyberToast("Artikel berhasil disimpan!");
            }
            await fetchArticles();
            resetForm();
        } else {
            showCyberToast("Gagal menyimpan artikel", "error");
        }
    } catch (err) {
        showCyberToast("Terjadi kesalahan server", "error");
    }
});

btnCancel.addEventListener('click', resetForm);

// saveData is no longer needed locally

function renderArticles() {
    articleList.innerHTML = '';

    if (articles.length === 0) {
        articleList.innerHTML = '<p>Belum ada artikel yang ditambahkan.</p>';
        return;
    }

    articles.forEach(article => {
        const div = document.createElement('div');
        div.className = 'article-item';
        const mdContent = window.marked ? window.marked.parse(article.content) : article.content;
        div.innerHTML = `
            <h3>${article.title}</h3>
            <p><strong>Penulis:</strong> ${article.author}</p>
            <div class="markdown-body" style="margin-top: 10px; margin-bottom: 15px;">${mdContent}</div>
            <div class="action-btn">
                <button class="btn-edit" onclick="editArticle('${article.id}')">Edit</button>
                <button class="btn-delete" onclick="deleteArticle('${article.id}')">Hapus</button>
            </div>
        `;
        articleList.appendChild(div);
    });
}

window.deleteArticle = async function (id) {
    if (confirm("Yakin ingin menghapus artikel ini?")) {
        try {
            const token = sessionStorage.getItem('porto_token');
            const res = await fetch(`/api/articles?id=${id}`, { 
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            if (res.ok) {
                showCyberToast("Artikel berhasil dihapus!");
                fetchArticles();
            }
        } catch (err) {
            showCyberToast("Gagal menghapus artikel", "error");
        }
    }
};

window.editArticle = function (id) {
    const article = articles.find(a => a.id === id);
    if (article) {
        isEditing = true;
        articleIdInput.value = article.id;
        titleInput.value = article.title;
        authorInput.value = article.author;
        contentInput.value = article.content;

        formTitle.textContent = "Edit Artikel";
        btnSubmit.textContent = "Update";
        btnCancel.style.display = 'inline-block';
    }
};

function resetForm() {
    form.reset();
    isEditing = false;
    articleIdInput.value = '';
    formTitle.textContent = "Tambah Artikel";
    btnSubmit.textContent = "Simpan";
    btnCancel.style.display = 'none';
}

const btnDownloadCv = document.getElementById('btn-download-cv');
if (btnDownloadCv) {
    btnDownloadCv.addEventListener('click', () => {
        window.showCyberToast('INITIALIZING PRINT PROTOCOL...', 'success');
        setTimeout(() => {
            window.print();
        }, 1000);
    });
}

// --- Phase 3: Analytics & Leaderboard Dashboard ---
const btnLoadDashboard = document.getElementById('btn-load-dashboard');
if (btnLoadDashboard) {
    btnLoadDashboard.addEventListener('click', async () => {
        const container = document.getElementById('dashboard-container');
        const token = sessionStorage.getItem('porto_token');
        if (!token) {
            showCyberToast('Unauthorized: Please login first', 'error');
            return;
        }

        try {
            // Load Analytics
            const resAnal = await fetch('/api/analytics', { headers: { 'Authorization': `Bearer ${token}` } });
            if (resAnal.ok) {
                const analytics = await resAnal.json();
                const list = document.getElementById('analytics-list');
                list.innerHTML = '';
                analytics.forEach(a => {
                    list.innerHTML += `<li><span style="color:var(--accent-color)">${a.event}</span>: ${a.count} hits</li>`;
                });
            } else {
                showCyberToast('Access Denied for Analytics', 'error');
            }

            // Load Leaderboard
            const resLead = await fetch('/api/leaderboard', { headers: { 'Authorization': `Bearer ${token}` } });
            if (resLead.ok) {
                const leaderboard = await resLead.json();
                const list = document.getElementById('leaderboard-list');
                list.innerHTML = '';
                leaderboard.forEach(l => {
                    list.innerHTML += `<li><span style="color:var(--text-primary)">${l.username}</span> - ${l.score} PTS</li>`;
                });
            }
            container.style.display = 'block';
        } catch (err) {
            showCyberToast('Error loading dashboard', 'error');
        }
    });
}

// Analytics Trigger on Page Load
fetch('/api/analytics', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ event: 'page_view_home' })
}).catch(e => console.log('Analytics err', e));

// Red Alert Trigger
window.triggerRedAlert = function() {
    document.body.classList.add('glitch-active');
    const overlay = document.getElementById('self-destruct-overlay');
    if (overlay) {
        overlay.classList.remove('hidden');
        setTimeout(() => {
            overlay.classList.add('hidden');
            document.body.classList.remove('glitch-active');
        }, 3000);
    }
};