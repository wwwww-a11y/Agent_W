const greetings = [
  '你好呀！👋',
  '很高兴见到你 😊',
  '你好，世界！🌍',
  '今天也要开心哦 ✨',
  '欢迎来到 site_20261123115213 🎉'
];

const btn = document.getElementById('helloBtn');
const message = document.getElementById('message');
let index = 0;

btn.addEventListener('click', () => {
  message.textContent = greetings[index % greetings.length];
  index++;
});

const links = document.querySelectorAll('.nav a');
links.forEach(link => {
  link.addEventListener('click', () => {
    links.forEach(l => l.classList.remove('active'));
    link.classList.add('active');
  });
});
