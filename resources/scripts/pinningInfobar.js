const pinObserver = new IntersectionObserver(
  ([e]) => e.target.classList.toggle("is-pinned",  e.intersectionRatio < 1),
  { threshold: [1] }
);

pinObserver.observe(document.getElementById("infobar"));
