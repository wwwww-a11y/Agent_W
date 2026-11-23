(function () {
  var nav = document.getElementById('nav');
  var glow = document.getElementById('cursorGlow');

  window.addEventListener('scroll', function () {
    nav.classList.toggle('scrolled', window.scrollY > 40);
  }, { passive: true });

  document.addEventListener('mousemove', function (e) {
    glow.style.left = e.clientX + 'px';
    glow.style.top = e.clientY + 'px';
  });

  var io = new IntersectionObserver(function (entries) {
    entries.forEach(function (en) {
      if (en.isIntersecting) {
        en.target.classList.add('in');
        io.unobserve(en.target);
      }
    });
  }, { threshold: 0.15 });

  document.querySelectorAll('.reveal').forEach(function (el, i) {
    el.style.transitionDelay = (i % 4) * 80 + 'ms';
    io.observe(el);
  });

  var stats = document.querySelectorAll('[data-count]');
  var counted = false;
  function runCount() {
    if (counted || !stats.length) return;
    var box = stats[0].closest('.stats');
    if (!box) return;
    var rect = box.getBoundingClientRect();
    if (rect.top < window.innerHeight * 0.85) {
      counted = true;
      stats.forEach(function (el) {
        var target = +el.dataset.count, cur = 0;
        var step = Math.max(1, Math.round(target / 45));
        var t = setInterval(function () {
          cur += step;
          if (cur >= target) { cur = target; clearInterval(t); }
          el.textContent = cur;
        }, 26);
      });
    }
  }
  window.addEventListener('scroll', runCount, { passive: true });
  runCount();

  var form = document.getElementById('helloForm');
  var tip = document.getElementById('formTip');
  if (form) {
    form.addEventListener('submit', function (e) {
      e.preventDefault();
      var name = form.name.value.trim();
      if (!name) { tip.textContent = '请先告诉我你的名字 🙂'; return; }
      tip.textContent = '你好，' + name + '！问候已收到，我们会尽快回复。';
      form.reset();
    });
  }
})();
