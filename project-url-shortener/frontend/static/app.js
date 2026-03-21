// State
let currentUser = localStorage.getItem('snip_username') || '';
let currentView = 'welcome';

// Init
document.addEventListener('DOMContentLoaded', () => {
  document.getElementById('username-form').addEventListener('submit', handleLogin);
  document.getElementById('shorten-form').addEventListener('submit', handleShorten);
  document.getElementById('logout-btn').addEventListener('click', handleLogout);
  document.getElementById('analytics-logout-btn').addEventListener('click', handleLogout);

  if (currentUser) {
    navigateTo('dashboard');
  }
});

// Navigation
function navigateTo(view, data) {
  document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
  document.getElementById(view).classList.add('active');
  currentView = view;

  if (view === 'dashboard') {
    document.getElementById('display-username').textContent = currentUser;
    loadURLs();
  } else if (view === 'analytics' && data) {
    document.getElementById('analytics-username').textContent = currentUser;
    loadAnalytics(data.code);
  }
}

// Tab switching
function switchTab(tab) {
  document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
  document.querySelector(`.tab[data-tab="${tab}"]`).classList.add('active');
  document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
  document.getElementById('tab-' + tab).classList.add('active');

  if (tab === 'top-urls') {
    loadTopURLs();
  }
}

// Auth
function handleLogin(e) {
  e.preventDefault();
  const input = document.getElementById('username-input');
  const username = input.value.trim();
  if (!username) return;

  currentUser = username;
  localStorage.setItem('snip_username', username);
  input.value = '';
  navigateTo('dashboard');
}

function handleLogout() {
  currentUser = '';
  localStorage.removeItem('snip_username');
  navigateTo('welcome');
}

// Shorten
async function handleShorten(e) {
  e.preventDefault();
  const urlInput = document.getElementById('url-input');
  const slugInput = document.getElementById('slug-input');
  const resultDiv = document.getElementById('shorten-result');
  const errorDiv = document.getElementById('shorten-error');

  resultDiv.classList.add('hidden');
  errorDiv.classList.add('hidden');

  const body = {
    url: urlInput.value.trim(),
    username: currentUser,
  };
  const slug = slugInput.value.trim();
  if (slug) body.custom_slug = slug;

  try {
    const resp = await fetch('/api/shorten', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const data = await resp.json();

    if (!resp.ok) {
      errorDiv.textContent = data.error || 'Something went wrong';
      errorDiv.classList.remove('hidden');
      return;
    }

    const shortURL = window.location.origin + '/r/' + data.short_code;
    resultDiv.innerHTML = `
      <div class="result-info">
        <span class="result-label">snipped!</span>
        <a class="short-url" href="/r/${escapeAttr(data.short_code)}" target="_blank">${escapeHtml(shortURL)}</a>
      </div>
      <div class="result-actions">
        <span class="copy-feedback" id="copy-feedback">copied!</span>
        <button type="button" class="btn-icon" onclick="copyURL('${escapeAttr(shortURL)}')" title="Copy">${icons.copy}</button>
        <a class="btn-icon" href="/r/${escapeAttr(data.short_code)}" target="_blank" title="Visit">${icons.visit}</a>
      </div>
    `;
    resultDiv.classList.remove('hidden');
    urlInput.value = '';
    slugInput.value = '';
    loadURLs();
    // Re-fetch after a delay so async metadata (title/favicon) shows up
    setTimeout(loadURLs, 3000);
  } catch (err) {
    errorDiv.textContent = 'Network error — is the server running?';
    errorDiv.classList.remove('hidden');
  }
}

function copyURL(url) {
  navigator.clipboard.writeText(url).then(() => {
    const fb = document.getElementById('copy-feedback');
    if (fb) {
      fb.classList.add('show');
      setTimeout(() => fb.classList.remove('show'), 1500);
    }
  });
}

// Load user's URLs
async function loadURLs() {
  const list = document.getElementById('urls-list');

  try {
    const resp = await fetch(`/api/urls?username=${encodeURIComponent(currentUser)}`);
    const urls = await resp.json();

    if (!urls || urls.length === 0) {
      list.innerHTML = '<p class="empty-state">no URLs yet — create your first one above</p>';
      return;
    }

    list.innerHTML = urls.map((u, i) => renderURLCard(u, i, true)).join('');
  } catch (err) {
    list.innerHTML = '<p class="empty-state">failed to load URLs</p>';
  }
}

// Load top URLs (global)
async function loadTopURLs() {
  const list = document.getElementById('top-urls-list');
  list.innerHTML = '<div class="loading">loading...</div>';

  try {
    const resp = await fetch('/api/analytics/top');
    const data = await resp.json();

    if (!data.urls || data.urls.length === 0) {
      list.innerHTML = '<p class="empty-state">no URLs yet</p>';
      return;
    }

    list.innerHTML = data.urls.map((u, i) => renderURLCard(u, i, false)).join('');
  } catch (err) {
    list.innerHTML = '<p class="empty-state">failed to load top URLs</p>';
  }
}

// Shared URL card renderer
function renderURLCard(u, index, showActions) {
  const favicon = u.favicon_url
    ? `<img class="favicon" src="${escapeAttr(u.favicon_url)}" alt="" onerror="this.style.display='none'">`
    : '<div class="favicon" style="background:var(--border);"></div>';

  const title = u.title
    ? `<div class="url-card-title">${escapeHtml(u.title)}</div>`
    : '';

  const shortPath = '/r/' + escapeHtml(u.short_code);
  const shortURL = window.location.origin + shortPath;

  const ownerBadge = !showActions && u.username
    ? `<span class="owner-badge">${escapeHtml(u.username)}</span>`
    : '';

  const deleteBtn = showActions
    ? `<button type="button" class="btn-icon btn-icon-danger" onclick="deleteURL('${escapeAttr(u.short_code)}')" title="Delete">${icons.delete}</button>`
    : '';

  return `
    <div class="url-card" style="animation-delay: ${index * 0.05}s">
      ${favicon}
      <div class="url-card-info">
        <div class="url-card-original">${escapeHtml(u.original_url)}</div>
        ${title}
        <a class="url-card-code" href="${shortPath}" target="_blank">${shortURL}</a>
      </div>
      <div class="url-card-actions">
        ${ownerBadge}
        <span class="click-count" onclick="navigateTo('analytics', {code: '${escapeAttr(u.short_code)}'})">${u.click_count} clicks</span>
        <button type="button" class="btn-icon" onclick="copyURL('${escapeAttr(shortURL)}')" title="Copy">${icons.copy}</button>
        <a class="btn-icon" href="${shortPath}" target="_blank" title="Visit">${icons.visit}</a>
        ${deleteBtn}
      </div>
    </div>
  `;
}

// Delete URL
async function deleteURL(code) {
  if (!confirm(`Delete /r/${code}?`)) return;
  try {
    await fetch(`/api/urls/${code}`, { method: 'DELETE' });
    loadURLs();
  } catch (err) {
    alert('Failed to delete URL');
  }
}

// Analytics
async function loadAnalytics(code) {
  const content = document.getElementById('analytics-content');
  content.innerHTML = '<div class="loading">loading analytics...</div>';

  try {
    const resp = await fetch(`/api/analytics/${code}`);
    if (!resp.ok) {
      content.innerHTML = '<div class="loading">URL not found</div>';
      return;
    }
    const data = await resp.json();

    const recentHTML = data.recent_clicks && data.recent_clicks.length > 0
      ? data.recent_clicks.map((c, i) => {
          const t = new Date(c.clicked_at);
          return `
            <div class="click-item" style="animation-delay: ${i * 0.04}s">
              <span class="click-time">${t.toLocaleString()}</span>
              <span class="click-ago">${timeAgo(t)}</span>
            </div>
          `;
        }).join('')
      : '<p class="empty-state">no clicks yet</p>';

    const createdAt = new Date(data.created_at);
    const shortPath = '/r/' + escapeHtml(data.short_code);

    content.innerHTML = `
      <div class="analytics-header">
        <h2><a href="${shortPath}" target="_blank" style="color:var(--accent);text-decoration:none">${shortPath}</a></h2>
        ${data.title ? `<div style="font-size:15px;margin:4px 0">${escapeHtml(data.title)}</div>` : ''}
        <div class="original-url">${escapeHtml(data.original_url)}</div>
      </div>
      <div class="analytics-stats">
        <div class="stat-card">
          <div class="stat-value">${data.click_count}</div>
          <div class="stat-label">total clicks</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">${createdAt.toLocaleDateString()}</div>
          <div class="stat-label">created</div>
        </div>
      </div>
      <div class="recent-clicks">
        <h3>recent clicks</h3>
        <div class="click-timeline">${recentHTML}</div>
      </div>
    `;
  } catch (err) {
    content.innerHTML = '<div class="loading">failed to load analytics</div>';
  }
}

const icons = {
  visit: `<svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 3H3v10h10V9"/><path d="M9 1h6v6"/><path d="M15 1L7 9"/></svg>`,
  copy: `<svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="5" width="9" height="9" rx="1"/><path d="M5 11H3a1 1 0 0 1-1-1V3a1 1 0 0 1 1-1h7a1 1 0 0 1 1 1v2"/></svg>`,
  delete: `<svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 4h12"/><path d="M5 4V2h6v2"/><path d="M6 7v5"/><path d="M10 7v5"/><path d="M3 4l1 10h8l1-10"/></svg>`,
};

// Helpers
function escapeHtml(str) {
  if (!str) return '';
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

function escapeAttr(str) {
  if (!str) return '';
  return str.replace(/&/g, '&amp;').replace(/'/g, '&#39;').replace(/"/g, '&quot;');
}

function timeAgo(date) {
  const seconds = Math.floor((new Date() - date) / 1000);
  if (seconds < 60) return 'just now';
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}
