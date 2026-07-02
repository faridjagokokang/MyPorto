const canvas = document.getElementById('matrix-rain');
const ctx = canvas.getContext('2d');

function resizeCanvas() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
}

resizeCanvas();

const letters = "@#$%^&*()_+{}|:<>?-=[];',./";
const fontSize = 16;
let columns = canvas.width / fontSize;
let drops = [];
window.matrixSpeedMultiplier = 1;

function initDrops() {
    columns = canvas.width / fontSize;
    drops = [];
    for (let x = 0; x < columns; x++) {
        drops[x] = Math.random() * -100;
    }
}

initDrops();

function draw() {
    ctx.shadowBlur = 0;
    ctx.fillStyle = 'rgba(0, 0, 0, 0.05)';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    const themeColor = window.matrixColorOverride || getComputedStyle(document.documentElement).getPropertyValue('--text-primary').trim() || '#00ff00';

    ctx.shadowBlur = 3;
    ctx.shadowColor = themeColor;
    ctx.fillStyle = themeColor;
    ctx.font = "bold " + fontSize + 'px "Fira Code", monospace';

    for (let i = 0; i < drops.length; i++) {
        const text = letters.charAt(Math.floor(Math.random() * letters.length));

        ctx.fillText(text, i * fontSize, Math.floor(drops[i]) * fontSize);

        if (drops[i] * fontSize > canvas.height && Math.random() > 0.975) {
            drops[i] = 0;
        }
        drops[i] += window.matrixSpeedMultiplier;
    }
}

setInterval(draw, 33);

window.addEventListener('resize', () => {
    resizeCanvas();
    initDrops();
});
