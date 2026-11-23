(function () {
  const nav = document.getElementById('nav');
  const onScroll = () => nav.classList.toggle('scrolled', window.scrollY > 40);
  onScroll();
  window.addEventListener('scroll', onScroll, { passive: true });

  const items = document.querySelectorAll('.reveal');
  if ('IntersectionObserver' in window) {
    const io = new IntersectionObserver((entries) => {
      entries.forEach((e) => {
        if (e.isIntersecting) {
          e.target.classList.add('in');
          io.unobserve(e.target);
        }
      });
    }, { threshold: 0.14 });
    items.forEach((el, i) => {
      el.style.transitionDelay = (i % 4) * 80 + 'ms';
      io.observe(el);
    });
  } else {
    items.forEach((el) => el.classList.add('in'));
  }
})();