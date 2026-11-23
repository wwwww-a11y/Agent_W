const root = document.documentElement;
const themeBtn = document.getElementById('themeBtn');
const saved = localStorage.getItem('theme');
if (saved === 'dark') root.setAttribute('data-theme', 'dark');

themeBtn.addEventListener('click', () => {
  const dark = root.getAttribute('data-theme') === 'dark';
  if (dark) {
    root.removeAttribute('data-theme');
    localStorage.setItem('theme', 'light');
  } else {
    root.setAttribute('data-theme', 'dark');
    localStorage.setItem('theme', 'dark');
  }
});

const bars = document.querySelectorAll('.bar i');
const barObserver = new IntersectionObserver((entries) => {
  entries.forEach(e => {
    if (e.isIntersecting) {
      e.target.style.width = e.target.dataset.w || e.target.style.width;
      barObserver.unobserve(e.target);
    }
  });
}, { threshold: 0.4 });
bars.forEach(b => { b.dataset.w = b.style.width; b.style.width = '0'; barObserver.observe(b); });

const revealEls = document.querySelectorAll('.section');
revealEls.forEach(el => el.classList.add('reveal'));
const revealObserver = new IntersectionObserver((entries) => {
  entries.forEach(e => { if (e.isIntersecting) { e.target.classList.add('show'); revealObserver.unobserve(e.target); } });
}, { threshold: 0.12 });
revealEls.forEach(el => revealObserver.observe(el));

document.querySelectorAll('.card, .work-card').forEach(el => {
  el.addEventListener('mousemove', (e) => {
    const r = el.getBoundingClientRect();
    const x = (e.clientX - r.left) / r.width - 0.5;
    const y = (e.clientY - r.top) / r.height - 0.5;
    el.style.transform = `translateY(-4px) rotateX(${-y * 4}deg) rotateY(${x * 4}deg)`;
  });
  el.addEventListener('mouseleave', () => { el.style.transform = ''; });
});

const copyBtn = document.getElementById('copyBtn');
const copyTip = document.getElementById('copyTip');
copyBtn.addEventListener('click', async () => {
  const mail = 'hello@example.com';
  try {
    await navigator.clipboard.writeText(mail);
    copyTip.textContent = '邮箱已复制 ✓';
  } catch {
    copyTip.textContent = mail;
  }
  setTimeout(() => { copyTip.textContent = ''; }, 2000);
});
