// 🦕 Diplodocs Client Runtime

// 1. Theme Management
(function initTheme() {
  const saved = localStorage.getItem('diplodocs-theme');
  if (saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }
})();

function diplodocsToggleTheme() {
  const isDark = document.documentElement.classList.toggle('dark');
  localStorage.setItem('diplodocs-theme', isDark ? 'dark' : 'light');
}

// 2. Mobile Menu Toggle
function diplodocsToggleMenu() {
  const sidebar = document.querySelector('.sidebar-left');
  if (sidebar) {
    sidebar.classList.toggle('open');
  }
}

// 3. Code Copy Button
function diplodocsCopyCode(btn) {
  const codeBlock = btn.closest('.code-block');
  if (!codeBlock) return;
  const codeEl = codeBlock.querySelector('code') || codeBlock.querySelector('pre');
  if (!codeEl) return;

  const text = codeEl.innerText;
  navigator.clipboard.writeText(text).then(() => {
    const textSpan = btn.querySelector('.copy-text');
    const originalText = textSpan ? textSpan.innerText : 'Copy';
    if (textSpan) textSpan.innerText = 'Copied!';
    btn.classList.add('copied');
    setTimeout(() => {
      if (textSpan) textSpan.innerText = originalText;
      btn.classList.remove('copied');
    }, 2000);
  });
}

// 4. Code Tabs
function diplodocsSwitchTab(btn, index) {
  const tabsContainer = btn.closest('.code-tabs');
  if (!tabsContainer) return;

  const buttons = tabsContainer.querySelectorAll('.tab-button');
  const panels = tabsContainer.querySelectorAll('.tab-panel');

  buttons.forEach((b, i) => {
    b.classList.toggle('active', i === index);
  });

  panels.forEach((p, i) => {
    p.classList.toggle('active', i === index);
  });
}

// 5. Client Search
let searchIndex = null;
let activeSearchResultIndex = 0;

async function loadSearchIndex() {
  if (searchIndex !== null) return searchIndex;
  try {
    // Relative to base or root
    const res = await fetch('search-index.json');
    if (res.ok) {
      searchIndex = await res.json();
    } else {
      // try root path
      const resRoot = await fetch('/search-index.json');
      if (resRoot.ok) {
        searchIndex = await resRoot.json();
      }
    }
  } catch (e) {
    console.error('Failed to load search index:', e);
    searchIndex = [];
  }
  return searchIndex || [];
}

function openSearchModal() {
  const modal = document.getElementById('search-modal');
  if (!modal) return;
  modal.classList.add('open');
  loadSearchIndex();
  const input = document.getElementById('search-input');
  if (input) {
    input.value = '';
    input.focus();
    renderSearchResults('');
  }
}

function closeSearchModal() {
  const modal = document.getElementById('search-modal');
  if (modal) {
    modal.classList.remove('open');
  }
}

function renderSearchResults(query) {
  const list = document.getElementById('search-results');
  if (!list) return;
  list.innerHTML = '';

  const q = query.trim().toLowerCase();
  if (!q || !searchIndex) {
    list.innerHTML = '<li class="search-result-item" style="color:var(--text-muted);cursor:default;">Type to search docs...</li>';
    return;
  }

  const results = [];
  for (const doc of searchIndex) {
    const titleMatch = doc.title.toLowerCase().indexOf(q);
    const contentMatch = doc.content ? doc.content.toLowerCase().indexOf(q) : -1;

    if (titleMatch !== -1 || contentMatch !== -1) {
      let score = 0;
      if (titleMatch !== -1) score += 50 - titleMatch;
      if (contentMatch !== -1) score += 10;

      let snippet = '';
      if (contentMatch !== -1) {
        const start = Math.max(0, contentMatch - 40);
        const end = Math.min(doc.content.length, contentMatch + 80);
        snippet = (start > 0 ? '...' : '') + doc.content.substring(start, end) + (end < doc.content.length ? '...' : '');
      } else {
        snippet = doc.content ? doc.content.substring(0, 100) + '...' : '';
      }

      results.push({
        title: doc.title,
        url: doc.url,
        snippet: snippet,
        score: score,
      });
    }
  }

  results.sort((a, b) => b.score - a.score);

  if (results.length === 0) {
    list.innerHTML = '<li class="search-result-item" style="color:var(--text-muted);cursor:default;">No matches found.</li>';
    return;
  }

  activeSearchResultIndex = 0;
  results.slice(0, 8).forEach((item, idx) => {
    const li = document.createElement('li');
    li.className = 'search-result-item' + (idx === 0 ? ' selected' : '');
    li.innerHTML = `
      <div class="search-result-title">${escapeHTML(item.title)}</div>
      <div class="search-result-snippet">${escapeHTML(item.snippet)}</div>
    `;
    li.onclick = () => {
      window.location.href = item.url;
    };
    list.appendChild(li);
  });
}

function escapeHTML(str) {
  return str.replace(/[&<>'"]/g, 
    tag => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[tag] || tag)
  );
}

// Keyboard shortcuts for search
document.addEventListener('keydown', (e) => {
  // Ctrl+K, Cmd+K, or Slash (when not in input)
  if ((e.key === 'k' && (e.metaKey || e.ctrlKey)) || (e.key === '/' && document.activeElement.tagName !== 'INPUT' && document.activeElement.tagName !== 'TEXTAREA')) {
    e.preventDefault();
    openSearchModal();
  }

  const modal = document.getElementById('search-modal');
  if (modal && modal.classList.contains('open')) {
    if (e.key === 'Escape') {
      e.preventDefault();
      closeSearchModal();
    } else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
      e.preventDefault();
      const items = modal.querySelectorAll('.search-result-item');
      if (items.length > 0) {
        items[activeSearchResultIndex]?.classList.remove('selected');
        if (e.key === 'ArrowDown') {
          activeSearchResultIndex = (activeSearchResultIndex + 1) % items.length;
        } else {
          activeSearchResultIndex = (activeSearchResultIndex - 1 + items.length) % items.length;
        }
        items[activeSearchResultIndex]?.classList.add('selected');
        items[activeSearchResultIndex]?.scrollIntoView({ block: 'nearest' });
      }
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const items = modal.querySelectorAll('.search-result-item');
      if (items[activeSearchResultIndex]) {
        items[activeSearchResultIndex].click();
      }
    }
  }
});

// 6. Mermaid Diagrams Initialization
document.addEventListener('DOMContentLoaded', () => {
  const mermaids = document.querySelectorAll('.mermaid');
  if (mermaids.length > 0) {
    const script = document.createElement('script');
    script.src = 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js';
    script.onload = () => {
      window.mermaid.initialize({
        startOnLoad: true,
        theme: document.documentElement.classList.contains('dark') ? 'dark' : 'default'
      });
    };
    document.head.appendChild(script);
  }

  diplodocsFetchRepoStats();
});

// 7. GitHub Repository Stats (Stars & Forks with caching)
function diplodocsFetchRepoStats() {
  const card = document.querySelector('.github-repo-card');
  if (!card) return;
  const repo = card.getAttribute('data-repo');
  if (!repo || repo === 'GitHub') return;

  const cacheKey = 'diplodocs-repo-' + repo;
  const cached = localStorage.getItem(cacheKey);
  const now = Date.now();

  function applyStats(stars, forks) {
    const starsRow = card.querySelector('.repo-stats-row');
    const starsVal = card.querySelector('.stars-val');
    const forksVal = card.querySelector('.forks-val');
    if (!starsRow || !starsVal || !forksVal) return;

    function formatNum(n) {
      if (n >= 1000) return (n / 1000).toFixed(1) + 'k';
      return n.toString();
    }

    starsVal.innerText = formatNum(stars);
    forksVal.innerText = formatNum(forks);
    starsRow.style.display = 'inline-flex';
  }

  if (cached) {
    try {
      const data = JSON.parse(cached);
      if (now - data.time < 3600000) { // 1 hour cache
        applyStats(data.stars, data.forks);
        return;
      }
    } catch (e) {}
  }

  fetch('https://api.github.com/repos/' + repo)
    .then(r => r.ok ? r.json() : null)
    .then(data => {
      if (!data) return;
      const stars = data.stargazers_count || 0;
      const forks = data.forks_count || 0;
      localStorage.setItem(cacheKey, JSON.stringify({ stars, forks, time: now }));
      applyStats(stars, forks);
    })
    .catch(() => {});
}
