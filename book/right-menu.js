// Inject a right-side outline ("On this page") similar to Khedra book theme.
// Lightweight adaptation: only runs if viewport wide enough and at least two h2s.
(function(){
  const VERSION = 'right-menu-20250820-2';
  console.log('[right-menu]', VERSION, 'script loaded');
  if(typeof window === 'undefined') return;
  window.addEventListener('DOMContentLoaded', function(){
    const existing = document.getElementById('right-menu');
    // mdBook vanilla theme doesn't have this container; create if missing inside #content
    let container = existing;
    if(!container){
      const content = document.getElementById('content');
      if(!content) return;
  container = document.createElement('div');
      container.id = 'right-menu';
      container.className = 'right-menu';
  container.setAttribute('data-version', VERSION);
      content.appendChild(container);
    }
    // Avoid on small screens
    if(window.innerWidth < 1350){ container.style.display='none'; return; }
    const headers = Array.from(document.querySelectorAll('main h2'));
    // Always show; if fewer than 2 headers, display a notice instead of hiding entirely
    container.innerHTML='';
    const title = document.createElement('h3');
    title.textContent = 'On this page';
    container.appendChild(title);
    if(headers.length){
      const list = document.createElement('ul');
      list.className = 'submenu';
      headers.forEach(h => {
        if(!h.id){ h.id = h.textContent.trim().toLowerCase().replace(/[^a-z0-9\s-]/gi,'').replace(/\s+/g,'-'); }
        const li = document.createElement('li');
        const a = document.createElement('a');
        a.href = '#'+h.id;
        a.textContent = h.textContent.trim();
        li.appendChild(a); list.appendChild(li);
      });
      container.appendChild(list);
    } else {
      const p = document.createElement('div');
      p.style.fontSize = '.75rem';
      p.style.opacity = '.6';
      p.textContent = 'No section headings on this page';
      container.appendChild(p);
    }
    // Scrollspy
    const links = Array.from(container.querySelectorAll('a'));
    const linkById = new Map(links.map(l => [l.getAttribute('href').slice(1), l]));
    const observer = new IntersectionObserver(entries => {
      entries.forEach(entry => {
        if(entry.isIntersecting){
          links.forEach(l => l.classList.remove('active'));
          const link = linkById.get(entry.target.id);
            if(link) link.classList.add('active');
        }
      });
    }, {root:null, rootMargin:'0px 0px -70% 0px', threshold:0});
  headers.forEach(h => observer.observe(h));
  });
})();
