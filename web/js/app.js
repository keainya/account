/* ================================================================
   Account Management - Frontend SPA
   ================================================================ */

/* ---------- Config ---------- */
const API_BASE = '';  // 同源，无需写完整 URL

/* ---------- State ---------- */
let currentUser = null;   // { id, username, email, role, ... }
let currentPage = null;

/* ---------- DOM refs ---------- */
const $header     = document.getElementById('header');
const $nav        = document.getElementById('nav-main');
const $userArea   = document.getElementById('user-area');
const $main       = document.getElementById('main');
const $toast      = document.getElementById('toast');
const $menuToggle = document.getElementById('menu-toggle');
const $drawerOverlay = document.getElementById('drawer-overlay');
const $drawer     = document.getElementById('drawer');

/* ================================================================
   Toast
   ================================================================ */
let toastTimer = null;

function showToast(msg, type) {
	type = type || 'info';
	$toast.textContent = msg;
	$toast.className = 'toast ' + type + ' show';
	clearTimeout(toastTimer);
	toastTimer = setTimeout(() => { $toast.classList.remove('show'); }, 2800);
}

/* ================================================================
   API Client
   ================================================================ */
async function api(url, opts) {
	opts = opts || {};
	const headers = opts.headers || {};
	if (!(opts.body instanceof FormData)) {
		headers['Content-Type'] = headers['Content-Type'] || 'application/json';
	}
	const res = await fetch(API_BASE + url, {
		...opts,
		headers,
		credentials: 'include',  // carry session cookie
	});
	let data;
	try { data = await res.json(); } catch (_) {
		throw { code: -1, msg: '响应解析失败' };
	}
	if (data.code !== 0) {
		// not logged in -> force to login
		if (data.code === 2001 && !url.includes('/auth/')) {
			logout(false);
			navigate('#/login');
			throw data;
		}
		throw data;
	}
	return data;
}

function apiGet(url)    { return api(url); }
function apiPost(url, body) { return api(url, { method: 'POST', body: JSON.stringify(body) }); }
function apiPut(url, body)  { return api(url, { method: 'PUT',  body: JSON.stringify(body) }); }
function apiDel(url)        { return api(url, { method: 'DELETE' }); }

/* ================================================================
   Auth Helpers
   ================================================================ */
async function fetchMe() {
	try {
		const d = await apiGet('/api/auth/me');
		currentUser = d.data;
	} catch (_) {
		currentUser = null;
	}
}

function logout(doApi) {
	currentUser = null;
	if (doApi !== false) {
		apiPost('/api/auth/logout').catch(() => {});
	}
}

function isAdmin() { return currentUser && currentUser.role === 'admin'; }

/* ================================================================
   Render Header
   ================================================================ */
function renderHeader() {
	if (!currentUser) {
		$header.style.display = 'none';
		return;
	}
	$header.style.display = '';

	// desktop nav links
	let links = '<a href="#/" class="' + (currentPage === 'dashboard' ? 'active' : '') + '">概览</a>';
	if (isAdmin()) {
		links += '<a href="#/admin/users" class="' + (currentPage === 'admin-users' ? 'active' : '') + '">用户管理</a>';
		links += '<a href="#/admin/apps" class="' + (currentPage === 'admin-apps' ? 'active' : '') + '">应用管理</a>';
	}
	$nav.innerHTML = links;

	// desktop user area
	let badgeCls = isAdmin() ? 'admin' : '';
	$userArea.innerHTML =
		'<span class="username">' + escHtml(currentUser.username) + '</span>' +
		'<span class="role-badge ' + badgeCls + '">' + (isAdmin() ? '管理员' : '用户') + '</span>' +
		'<button class="btn btn-outline btn-sm" onclick="handleLogout()">退出</button>';

	// mobile drawer
	renderDrawer();
}

function renderDrawer() {
	let content = '<div class="drawer-section drawer-user">' +
		'<div><div class="name">' + escHtml(currentUser.username) + '</div>' +
		'<span class="role-badge ' + (isAdmin() ? 'admin' : '') + '">' + (isAdmin() ? '管理员' : '用户') + '</span></div>' +
		'</div>';

	content += '<div class="drawer-section"><div class="drawer-label">导航</div>';
	content += '<a href="#/" class="' + (currentPage === 'dashboard' ? 'active' : '') + '" onclick="closeDrawer()">概览</a>';
	if (isAdmin()) {
		content += '<a href="#/admin/users" class="' + (currentPage === 'admin-users' ? 'active' : '') + '" onclick="closeDrawer()">用户管理</a>';
		content += '<a href="#/admin/apps" class="' + (currentPage === 'admin-apps' ? 'active' : '') + '" onclick="closeDrawer()">应用管理</a>';
	}
	content += '</div>';

	content += '<button class="btn-logout" onclick="handleLogout()">退出登录</button>';
	$drawer.innerHTML = content;
}

function openDrawer() {
	$drawer.classList.add('open');
	$drawerOverlay.classList.add('open');
	document.body.style.overflow = 'hidden';
}

function closeDrawer() {
	$drawer.classList.remove('open');
	$drawerOverlay.classList.remove('open');
	document.body.style.overflow = '';
}

// Menu toggle
$menuToggle.addEventListener('click', openDrawer);
$drawerOverlay.addEventListener('click', closeDrawer);

function handleLogout() {
	logout(true);
	navigate('#/login');
	showToast('已退出登录', 'info');
}

/* ================================================================
   Router
   ================================================================ */
const routes = {
	'login':        { render: renderLogin,        auth: false },
	'register':     { render: renderRegister,     auth: false },
	'dashboard':    { render: renderDashboard,    auth: true  },
	'admin-users':  { render: renderAdminUsers,   auth: true,  admin: true },
	'admin-apps':   { render: renderAdminApps,    auth: true,  admin: true },
};

function parseHash() {
	const h = location.hash.replace(/^#\/?/, '');
	if (!h) return 'dashboard';
	if (h === 'login')      return 'login';
	if (h === 'register')   return 'register';
	if (h === 'admin/users') return 'admin-users';
	if (h === 'admin/apps')  return 'admin-apps';
	return 'dashboard';
}

async function navigate(hash) {
	location.hash = hash;
	await doRoute();
}

async function doRoute() {
	const page = parseHash();
	const def   = routes[page];

	if (!def) {
		$main.innerHTML = '<div class="empty"><p>页面不存在</p></div>';
		return;
	}

	// auth guard
	if (def.auth && !currentUser) {
		await fetchMe();
		if (!currentUser) {
			location.hash = '#/login';
			return doRoute();
		}
	}
	if (def.admin && !isAdmin()) {
		$main.innerHTML = '<div class="empty"><p>无权访问此页面</p></div>';
		currentPage = page;
		renderHeader();
		return;
	}

	currentPage = page;
	renderHeader();
	await def.render();
}

/* ================================================================
   Login Page
   ================================================================ */
async function renderLogin() {
	if (currentUser) { navigate('#/'); return; }

	$main.className = 'page-auth';
	$main.innerHTML = `
	<div class="auth-card">
		<h1>登录</h1>
		<p class="sub">使用您的账户登录管理系统</p>
		<form id="login-form">
			<div class="form-group">
				<label>用户名</label>
				<input type="text" name="username" placeholder="请输入用户名" autocomplete="username" required>
			</div>
			<div class="form-group">
				<label>密码</label>
				<input type="password" name="password" placeholder="请输入密码" autocomplete="current-password" required>
			</div>
			<div class="form-group" style="margin-top: 24px">
				<button type="submit" class="btn btn-primary btn-block" id="login-btn">登 录</button>
			</div>
			<div class="form-group" style="text-align:center; margin-bottom:0">
				<span style="color:var(--c-text-sub);font-size:13px">还没有账户？</span>
				<a href="#/register">立即注册</a>
			</div>
		</form>
	</div>`;

	document.getElementById('login-form').addEventListener('submit', async (e) => {
		e.preventDefault();
		const btn = document.getElementById('login-btn');
		btn.disabled = true;
		btn.textContent = '登录中…';
		const fd = new FormData(e.target);
		try {
			await apiPost('/api/auth/login', {
				username: fd.get('username'),
				password: fd.get('password'),
			});
			await fetchMe();
			showToast('登录成功', 'success');
			navigate('#/');
		} catch (err) {
			showToast(err.msg || '登录失败', 'error');
		} finally {
			btn.disabled = false;
			btn.textContent = '登 录';
		}
	});
}

/* ================================================================
   Register Page
   ================================================================ */
async function renderRegister() {
	if (currentUser) { navigate('#/'); return; }

	$main.className = 'page-auth';
	$main.innerHTML = `
	<div class="auth-card">
		<h1>注册</h1>
		<p class="sub">创建新账户，首个注册用户将成为管理员</p>
		<form id="register-form">
			<div class="form-group">
				<label>用户名</label>
				<input type="text" name="username" placeholder="3-32字符，字母开头" required>
				<p class="hint">仅允许字母、数字、下划线</p>
			</div>
			<div class="form-group">
				<label>邮箱（可选）</label>
				<input type="email" name="email" placeholder="example@email.com">
			</div>
			<div class="form-group">
				<label>密码</label>
				<input type="password" name="password" placeholder="6-128 字符" required minlength="6">
			</div>
			<div class="form-group" style="margin-top: 24px">
				<button type="submit" class="btn btn-primary btn-block" id="register-btn">注 册</button>
			</div>
			<div class="form-group" style="text-align:center; margin-bottom:0">
				<span style="color:var(--c-text-sub);font-size:13px">已有账户？</span>
				<a href="#/login">立即登录</a>
			</div>
		</form>
	</div>`;

	document.getElementById('register-form').addEventListener('submit', async (e) => {
		e.preventDefault();
		const btn = document.getElementById('register-btn');
		btn.disabled = true;
		btn.textContent = '注册中…';
		const fd = new FormData(e.target);
		const body = { username: fd.get('username'), password: fd.get('password') };
		const email = fd.get('email');
		if (email) body.email = email;

		try {
			await apiPost('/api/auth/register', body);
			showToast('注册成功，请登录', 'success');
			navigate('#/login');
		} catch (err) {
			showToast(err.msg || '注册失败', 'error');
		} finally {
			btn.disabled = false;
			btn.textContent = '注 册';
		}
	});
}

/* ================================================================
   Dashboard Page
   ================================================================ */
async function renderDashboard() {
	$main.className = '';
	$main.innerHTML = `
	<div class="card">
		<div class="card-header"><h2>账户信息</h2></div>
		<div id="dashboard-info"></div>
	</div>`;

	const u = currentUser;
	document.getElementById('dashboard-info').innerHTML = `
	<table>
		<tr><th>用户 ID</th><td data-label="用户 ID"><code style="font-size:12px">${escHtml(u.id)}</code></td></tr>
		<tr><th>用户名</th><td data-label="用户名">${escHtml(u.username)}</td></tr>
		<tr><th>邮箱</th><td data-label="邮箱">${escHtml(u.email || '—')}</td></tr>
		<tr><th>角色</th><td data-label="角色">${isAdmin() ? '管理员' : '普通用户'}</td></tr>
		<tr><th>注册时间</th><td data-label="注册时间">${formatTime(u.created_at)}</td></tr>
		<tr><th>更新时间</th><td data-label="更新时间">${formatTime(u.updated_at)}</td></tr>
	</table>`;
}

/* ================================================================
   Admin - Users
   ================================================================ */
async function renderAdminUsers() {
	$main.className = '';
	$main.innerHTML = `
	<div class="card">
		<div class="card-header"><h2>用户管理</h2></div>
		<div id="admin-users-table"></div>
		<div id="admin-users-pager"></div>
	</div>`;
	await loadUsers(1);
}

async function loadUsers(page) {
	try {
		const d = await apiGet('/api/admin/users?page=' + page + '&page_size=15');
		const items = d.data.items;
		const total = d.data.total;
		const ps    = d.data.page_size;

		if (items.length === 0) {
			document.getElementById('admin-users-table').innerHTML =
				'<div class="empty"><p>暂无用户</p></div>';
			document.getElementById('admin-users-pager').innerHTML = '';
			return;
		}

		let html = '<div class="table-wrap"><table><thead><tr>' +
			'<th>用户名</th><th>邮箱</th><th>角色</th><th>注册时间</th><th>操作</th>' +
			'</tr></thead><tbody>';
		for (const u of items) {
			const isSelf = u.id === currentUser.id;
			html += '<tr>' +
				'<td data-label="用户名"><strong>' + escHtml(u.username) + '</strong></td>' +
				'<td data-label="邮箱">' + escHtml(u.email || '—') + '</td>' +
				'<td data-label="角色">' + (u.role === 'admin' ? '<span class="role-badge admin">管理员</span>' : '<span class="role-badge">用户</span>') + '</td>' +
				'<td data-label="注册时间">' + formatTime(u.created_at) + '</td>' +
				'<td data-label="操作" class="actions">' + renderUserActions(u, isSelf) + '</td>' +
				'</tr>';
		}
		html += '</tbody></table></div>';
		document.getElementById('admin-users-table').innerHTML = html;

		// pagination
		const totalPages = Math.ceil(total / ps);
		let pagerHtml = '<div class="pagination">';
		if (page > 1) {
			pagerHtml += '<button class="btn btn-outline btn-sm" onclick="loadUsers(' + (page - 1) + ')">上一页</button>';
		}
		pagerHtml += '<span>第 ' + page + ' / ' + totalPages + ' 页（共 ' + total + ' 条）</span>';
		if (page < totalPages) {
			pagerHtml += '<button class="btn btn-outline btn-sm" onclick="loadUsers(' + (page + 1) + ')">下一页</button>';
		}
		pagerHtml += '</div>';
		document.getElementById('admin-users-pager').innerHTML = pagerHtml;
	} catch (err) {
		showToast(err.msg || '获取用户列表失败', 'error');
	}
}

function renderUserActions(u, isSelf) {
	if (isSelf) return '<span style="color:var(--c-text-sub);font-size:12px">当前用户</span>';
	if (u.role === 'user') {
		return '<button class="btn btn-outline btn-sm" onclick="promoteUser(\'' + u.id + '\')">提升为管理员</button>';
	}
	return '<button class="btn btn-outline btn-sm" onclick="demoteUser(\'' + u.id + '\')">降级</button>';
}

async function promoteUser(userId) {
	if (!confirm('确定将该用户提升为管理员？')) return;
	try {
		await apiPut('/api/admin/users/' + userId + '/promote');
		showToast('已提升为管理员', 'success');
		await loadUsers(1);
	} catch (err) {
		showToast(err.msg || '操作失败', 'error');
	}
}

async function demoteUser(userId) {
	if (!confirm('确定将该管理员降级为普通用户？')) return;
	try {
		await apiPut('/api/admin/users/' + userId + '/demote');
		showToast('已降级为普通用户', 'success');
		await loadUsers(1);
	} catch (err) {
		showToast(err.msg || '操作失败', 'error');
	}
}

/* ================================================================
   Admin - Apps
   ================================================================ */
async function renderAdminApps() {
	$main.className = '';
	$main.innerHTML = `
	<div class="card">
		<div class="card-header">
			<h2>应用管理</h2>
			<button class="btn btn-primary btn-sm" onclick="showCreateAppModal()">+ 创建应用</button>
		</div>
		<div id="admin-apps-table"></div>
		<div id="admin-apps-pager"></div>
	</div>
	<div id="modal-container"></div>`;
	await loadApps(1);
}

let appPage = 1;
async function loadApps(page) {
	appPage = page;
	try {
		const d = await apiGet('/api/admin/apps?page=' + page + '&page_size=15');
		const items = d.data.items;
		const total = d.data.total;
		const ps    = d.data.page_size;

		if (items.length === 0) {
			document.getElementById('admin-apps-table').innerHTML =
				'<div class="empty"><p>暂无应用，点击上方按钮创建</p></div>';
			document.getElementById('admin-apps-pager').innerHTML = '';
			return;
		}

		let html = '<div class="table-wrap"><table><thead><tr>' +
			'<th>名称</th><th>描述</th><th>Client ID</th><th>回调地址</th><th>创建时间</th><th>操作</th>' +
			'</tr></thead><tbody>';
		for (const a of items) {
			html += '<tr>' +
				'<td data-label="名称"><strong>' + escHtml(a.name) + '</strong></td>' +
				'<td data-label="描述"><span style="color:var(--c-text-sub);font-size:12px">' + escHtml(a.description || '—') + '</span></td>' +
				'<td data-label="Client ID"><code>' + escHtml(a.client_id) + '</code></td>' +
				'<td data-label="回调地址"><div class="tag-uris">' + (a.redirect_uris || []).map(u => '<span class="tag-uri" title="' + escHtml(u) + '">' + escHtml(u) + '</span>').join('') + '</div></td>' +
				'<td data-label="创建时间">' + formatTime(a.created_at) + '</td>' +
				'<td data-label="操作" class="actions">' +
					'<button class="btn btn-outline btn-sm" onclick="showEditAppModal(\'' + a.id + '\')">编辑</button>' +
					'<button class="btn btn-outline btn-sm" onclick="showSecretModal(\'' + a.id + '\')">密钥</button>' +
					'<button class="btn btn-outline btn-sm" onclick="resetSecret(\'' + a.id + '\')">重置密钥</button>' +
					'<button class="btn btn-danger btn-sm" onclick="deleteApp(\'' + a.id + '\',\'' + escJs(a.name) + '\')">删除</button>' +
				'</td>' +
				'</tr>';
		}
		html += '</tbody></table></div>';
		document.getElementById('admin-apps-table').innerHTML = html;

		const totalPages = Math.ceil(total / ps);
		let pagerHtml = '<div class="pagination">';
		if (page > 1) pagerHtml += '<button class="btn btn-outline btn-sm" onclick="loadApps(' + (page - 1) + ')">上一页</button>';
		pagerHtml += '<span>第 ' + page + ' / ' + totalPages + ' 页（共 ' + total + ' 条）</span>';
		if (page < totalPages) pagerHtml += '<button class="btn btn-outline btn-sm" onclick="loadApps(' + (page + 1) + ')">下一页</button>';
		pagerHtml += '</div>';
		document.getElementById('admin-apps-pager').innerHTML = pagerHtml;
	} catch (err) {
		showToast(err.msg || '获取应用列表失败', 'error');
	}
}

/* ----- Create App Modal ----- */
function showCreateAppModal() {
	document.getElementById('modal-container').innerHTML = `
	<div class="modal-overlay" onclick="if(event.target===this)closeModal()">
		<div class="modal">
			<h3>创建应用</h3>
			<form id="create-app-form">
				<div class="form-group">
					<label>应用名称 <span style="color:var(--c-danger)">*</span></label>
					<input type="text" name="name" placeholder="1-64字符" required maxlength="64">
				</div>
				<div class="form-group">
					<label>描述</label>
					<textarea name="description" placeholder="可选的应用描述"></textarea>
				</div>
				<div class="form-group">
					<label>回调地址 <span style="color:var(--c-danger)">*</span></label>
					<textarea name="redirect_uris" placeholder="每行一个 URI，例如：&#10;https://example.com/callback&#10;http://localhost:3000/callback" rows="3" required></textarea>
					<p class="hint">每行输入一个 OAuth 回调地址</p>
				</div>
				<div class="modal-actions">
					<button type="button" class="btn btn-outline" onclick="closeModal()">取消</button>
					<button type="submit" class="btn btn-primary" id="create-app-btn">创建</button>
				</div>
			</form>
		</div>
	</div>`;

	document.getElementById('create-app-form').addEventListener('submit', async (e) => {
		e.preventDefault();
		const btn = document.getElementById('create-app-btn');
		btn.disabled = true; btn.textContent = '创建中…';
		const fd = new FormData(e.target);
		const uris = fd.get('redirect_uris').split('\n').map(s => s.trim()).filter(Boolean);
		try {
			const d = await apiPost('/api/admin/apps', {
				name: fd.get('name'),
				description: fd.get('description') || '',
				redirect_uris: uris,
			});
			showSecretCreated(d.data); // show client_secret immediately
			closeModal();
			await loadApps(appPage);
		} catch (err) {
			showToast(err.msg || '创建失败', 'error');
		} finally {
			btn.disabled = false; btn.textContent = '创建';
		}
	});
}

function showSecretCreated(app) {
	document.getElementById('modal-container').innerHTML = `
	<div class="modal-overlay" onclick="if(event.target===this)closeModal()">
		<div class="modal">
			<h3>应用创建成功</h3>
			<p style="margin-bottom:12px;color:var(--c-text-sub);font-size:13px">请保存以下密钥，<strong>此密钥仅显示一次</strong>。</p>
			<div class="form-group">
				<label>Client ID</label>
				<div class="secret-display"><span class="value">${escHtml(app.client_id)}</span></div>
			</div>
			<div class="form-group">
				<label>Client Secret</label>
				<div class="secret-display"><span class="value">${escHtml(app.client_secret)}</span></div>
			</div>
			<div class="modal-actions">
				<button class="btn btn-primary" onclick="closeModal()">我已保存</button>
			</div>
		</div>
	</div>`;
}

/* ----- Edit App Modal ----- */
async function showEditAppModal(appId) {
	try {
		const d = await apiGet('/api/admin/apps/' + appId);
		const a = d.data;

		document.getElementById('modal-container').innerHTML = `
		<div class="modal-overlay" onclick="if(event.target===this)closeModal()">
			<div class="modal">
				<h3>编辑应用</h3>
				<form id="edit-app-form">
					<div class="form-group">
						<label>应用名称</label>
						<input type="text" name="name" value="${escAttr(a.name)}" maxlength="64">
					</div>
					<div class="form-group">
						<label>描述</label>
						<textarea name="description">${escHtml(a.description || '')}</textarea>
					</div>
					<div class="form-group">
						<label>回调地址</label>
						<textarea name="redirect_uris" rows="3">${(a.redirect_uris || []).join('\n')}</textarea>
						<p class="hint">每行一个 URI</p>
					</div>
					<div class="modal-actions">
						<button type="button" class="btn btn-outline" onclick="closeModal()">取消</button>
						<button type="submit" class="btn btn-primary" id="edit-app-btn">保存</button>
					</div>
				</form>
			</div>
		</div>`;

		document.getElementById('edit-app-form').addEventListener('submit', async (e) => {
			e.preventDefault();
			const btn = document.getElementById('edit-app-btn');
			btn.disabled = true; btn.textContent = '保存中…';
			const fd = new FormData(e.target);
			const name = fd.get('name') || undefined;
			const desc = fd.get('description') || undefined;
			const rawUris = fd.get('redirect_uris');
			const uris = rawUris ? rawUris.split('\n').map(s => s.trim()).filter(Boolean) : undefined;

			const body = {};
			if (name !== undefined) body.name = name;
			if (desc !== undefined) body.description = desc;
			if (uris !== undefined) body.redirect_uris = uris;

			try {
				await apiPut('/api/admin/apps/' + appId, body);
				showToast('应用已更新', 'success');
				closeModal();
				await loadApps(appPage);
			} catch (err) {
				showToast(err.msg || '更新失败', 'error');
			} finally {
				btn.disabled = false; btn.textContent = '保存';
			}
		});
	} catch (err) {
		showToast(err.msg || '获取应用信息失败', 'error');
	}
}

/* ----- Secret Modal ----- */
async function showSecretModal(appId) {
	try {
		const d = await apiGet('/api/admin/apps/' + appId);
		const a = d.data;

		document.getElementById('modal-container').innerHTML = `
		<div class="modal-overlay" onclick="if(event.target===this)closeModal()">
			<div class="modal">
				<h3>应用密钥 — ${escHtml(a.name)}</h3>
				<div class="form-group">
					<label>Client ID</label>
					<div class="secret-display"><span class="value">${escHtml(a.client_id)}</span></div>
				</div>
				<div class="form-group">
					<label>Client Secret</label>
					<div class="secret-display"><span class="value">${escHtml(a.client_secret)}</span></div>
				</div>
				<div class="modal-actions">
					<button class="btn btn-outline" onclick="closeModal()">关闭</button>
				</div>
			</div>
		</div>`;
	} catch (err) {
		showToast(err.msg || '获取密钥失败', 'error');
	}
}

/* ----- Reset Secret ----- */
async function resetSecret(appId) {
	if (!confirm('重置后旧密钥将立即失效，确定继续？')) return;
	try {
		const d = await apiPost('/api/admin/apps/' + appId + '/reset-secret');
		const secret = d.data.client_secret;

		// Show new secret modal
		document.getElementById('modal-container').innerHTML = `
		<div class="modal-overlay" onclick="if(event.target===this)closeModal()">
			<div class="modal">
				<h3>密钥已重置</h3>
				<p style="margin-bottom:12px;color:var(--c-text-sub);font-size:13px">新密钥如下，<strong>此密钥仅显示一次</strong>。</p>
				<div class="form-group">
					<label>新 Client Secret</label>
					<div class="secret-display"><span class="value">${escHtml(secret)}</span></div>
				</div>
				<div class="modal-actions">
					<button class="btn btn-primary" onclick="closeModal(); loadApps(appPage);">我已保存</button>
				</div>
			</div>
		</div>`;
	} catch (err) {
		showToast(err.msg || '重置失败', 'error');
	}
}

/* ----- Delete App ----- */
async function deleteApp(appId, name) {
	if (!confirm('确定删除应用 "' + name + '"？此操作不可撤销，将同时清理该应用下的所有元数据。')) return;
	try {
		await apiDel('/api/admin/apps/' + appId);
		showToast('应用已删除', 'success');
		await loadApps(appPage);
	} catch (err) {
		showToast(err.msg || '删除失败', 'error');
	}
}

/* ----- Close Modal ----- */
function closeModal() {
	document.getElementById('modal-container').innerHTML = '';
}

/* ================================================================
   Utility
   ================================================================ */
function escHtml(s) {
	if (!s) return '';
	return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function escAttr(s) {
	if (!s) return '';
	return String(s).replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function escJs(s) {
	if (!s) return '';
	return String(s).replace(/\\/g, '\\\\').replace(/'/g, "\\'");
}

function formatTime(iso) {
	if (!iso) return '—';
	try {
		const d = new Date(iso);
		const pad = (n) => String(n).padStart(2, '0');
		return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
			' ' + pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
	} catch (_) {
		return iso;
	}
}

/* ================================================================
   Init
   ================================================================ */
window.addEventListener('hashchange', () => { closeDrawer(); doRoute(); });

// ESC key to close drawer
document.addEventListener('keydown', (e) => { if (e.key === 'Escape') closeDrawer(); });

(async function init() {
	// Try auto-login via session cookie
	await fetchMe();
	doRoute();
})();
