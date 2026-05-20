# 前端风格指南（Style Guide）

---

## 1. 技术栈

| 层面 | 选型 | 说明 |
|------|------|------|
| 页面结构 | HTML5 | 单文件 `index.html`，语义化标签 |
| 样式 | 纯 CSS（CSS Variables） | 无预处理器，无框架 |
| 脚本 | 原生 ES6 JavaScript | 无框架、无构建、无 TypeScript |
| 路由 | Hash SPA (`#/page`) | 监听 `hashchange`，手动派发渲染 |
| 认证 | Session Cookie | `fetch` 带 `credentials: 'include'` |
| 依赖 | **零外部依赖** | 不引入 jQuery、React、Vue 等 |

**核心理念：** 极简、零依赖、服务端 API + 客户端 SPA，适合管理后台类应用。

---

## 2. 目录结构

```
web/
├── index.html          # 入口（仅容器 DOM + 资源引用）
├── css/
│   └── style.css       # 全局样式（~420 行）
├── js/
│   └── app.js          # 全部逻辑（~790 行）
└── favicon.ico
```

---

## 3. HTML 结构模式

### 3.1 入口页面（index.html）

只负责提供 **容器 DOM** 和资源引用，所有内容由 JS 动态渲染：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>App</title>
    <link rel="stylesheet" href="css/style.css">
</head>
<body>
    <header id="header">
        <div class="header-inner">
            <nav id="nav-main"></nav>       <!-- 桌面端导航 -->
            <div id="user-area"></div>       <!-- 用户信息/操作 -->
            <button class="menu-toggle">     <!-- 移动端汉堡菜单 -->
                <svg><!-- inline SVG icon --></svg>
            </button>
        </div>
    </header>

    <div class="drawer-overlay" id="drawer-overlay"></div>  <!-- 遮罩 -->
    <nav class="drawer" id="drawer"></nav>                  <!-- 侧滑菜单 -->

    <div id="toast" class="toast"></div>    <!-- 全局提示 -->

    <main id="main"></main>                 <!-- 页面内容区 -->

    <script src="js/app.js"></script>
</body>
</html>
```

### 3.2 动态页面渲染

每个页面由 JS 函数负责生成 HTML 字符串并注入 `#main`：

```js
function renderPage() {
    $main.innerHTML = `<div class="card">...</div>`;
    // 绑定事件监听
}
```

---

## 4. CSS 架构

### 4.1 设计令牌（Design Tokens）

使用 `:root` CSS 变量统一管理颜色、圆角、字体等：

```css
:root {
    /* 背景色 */
    --c-bg:         #ffffff;      /* 卡片/模态框背景 */
    --c-bg-alt:     #f7f8fa;      /* 页面底色、表头 */
    --c-bg-hover:   #f0f1f3;      /* 表格行悬停 */

    /* 边框 */
    --c-border:     #e5e6eb;

    /* 文字 */
    --c-text:       #1d2129;      /* 主文字 */
    --c-text-sub:   #86909c;      /* 辅助文字 */
    --c-text-light: #c9cdd4;      /* 占位/禁用文字 */

    /* 主题色 */
    --c-primary:    #1d2129;      /* 主色（深灰） */
    --c-primary-h:  #000000;      /* 主色 hover */

    /* 功能色 */
    --c-success:    #00b42a;
    --c-danger:     #f53f3f;
    --c-danger-h:   #cb2634;
    --c-warning:    #ff7d00;

    /* 其他 */
    --c-shadow:     rgba(0,0,0,.06);
    --c-overlay:    rgba(0,0,0,.3);
    --radius:       8px;
    --radius-sm:    4px;

    /* 字体栈 */
    --font:      -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
                 "Helvetica Neue", Arial, "Noto Sans SC", sans-serif;
    --font-mono: "SF Mono", "Fira Code", "Fira Mono", "Roboto Mono", monospace;
}
```

### 4.2 全局 Reset

```css
*, *::before, *::after {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}
body {
    font-family: var(--font);
    font-size: 14px;
    color: var(--c-text);
    background: var(--c-bg-alt);
    line-height: 1.6;
    min-height: 100vh;
}
a { color: var(--c-primary); text-decoration: none; }
a:hover { color: var(--c-primary-h); }
```

### 4.3 命名约定

- **ID 选择器**：用于页面级容器（`#header`, `#main`, `#toast`）
- **类选择器**：用于可复用组件，使用短横线命名（`form-group`, `btn-primary`, `card-header`）
- **无 BEM**：不使用 `__` / `--` 等冗长修饰符，直接用嵌套的短类名

### 4.4 布局系统

采用 **max-width 居中** 布局，不使用 Grid/Flexbox 栅格系统：

```css
.header-inner,
#main {
    max-width: 1120px;
    margin: 0 auto;
    padding: 0 24px;
}
```

---

## 5. 组件库

### 5.1 按钮（Button）

```css
.btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 9px 20px;
    font-size: 14px;
    font-family: var(--font);
    font-weight: 500;
    border: 1px solid transparent;
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: .15s;
    outline: none;
    white-space: nowrap;
    line-height: 1.4;
}
.btn:disabled { opacity: .5; cursor: not-allowed; }
```

**变体：**

| 类名 | 用途 |
|------|------|
| `.btn-primary` | 主要操作（深色背景 + 白色文字） |
| `.btn-outline` | 次要操作（白底 + 边框） |
| `.btn-danger` | 危险操作（红色） |
| `.btn-link` | 链接风格按钮（无边框背景） |
| `.btn-sm` | 小尺寸 |
| `.btn-block` | 通栏宽度 |

### 5.2 卡片（Card）

```css
.card {
    background: var(--c-bg);
    border: 1px solid var(--c-border);
    border-radius: var(--radius);
    padding: 24px;
    box-shadow: 0 1px 4px var(--c-shadow);
}
.card + .card { margin-top: 16px; }
.card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
}
.card-header h2 { font-size: 18px; font-weight: 600; }
```

### 5.3 表单（Form）

```css
.form-group { margin-bottom: 18px; }
.form-group label {
    display: block;
    font-size: 13px;
    font-weight: 500;
    margin-bottom: 6px;
    color: var(--c-text);
}
.form-group input,
.form-group textarea,
.form-group select {
    width: 100%;
    padding: 9px 12px;
    border: 1px solid var(--c-border);
    border-radius: var(--radius-sm);
    font-size: 14px;
    font-family: var(--font);
    color: var(--c-text);
    background: var(--c-bg);
    transition: border-color .15s;
    outline: none;
}
.form-group input:focus,
.form-group textarea:focus,
.form-group select:focus {
    border-color: var(--c-primary);
    box-shadow: 0 0 0 2px rgba(0,0,0,.15);
}
.form-group .hint  { font-size: 12px; color: var(--c-text-sub); margin-top: 4px; }
.form-group .error { font-size: 12px; color: var(--c-danger); margin-top: 4px; }

/* 移动端防止 iOS 缩放 */
@media (max-width: 768px) {
    .form-group input,
    .form-group textarea,
    .form-group select { font-size: 16px; }
}
```

### 5.4 登录/注册居中布局

```css
.page-auth {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: calc(100vh - 56px);
    padding: 24px;
}
.auth-card {
    background: var(--c-bg);
    border: 1px solid var(--c-border);
    border-radius: var(--radius);
    padding: 40px 36px;
    width: 100%;
    max-width: 400px;
    box-shadow: 0 2px 12px var(--c-shadow);
}
.auth-card h1 { font-size: 22px; margin-bottom: 4px; }
.auth-card .sub { font-size: 13px; color: var(--c-text-sub); margin-bottom: 28px; }
```

### 5.5 表格（Table）

```css
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
table th, table td {
    padding: 11px 14px;
    text-align: left;
    font-size: 13px;
    border-bottom: 1px solid var(--c-border);
}
table th {
    font-weight: 600;
    color: var(--c-text-sub);
    background: var(--c-bg-alt);
    white-space: nowrap;
}
table tbody tr:hover { background: var(--c-bg-hover); }
table .actions { display: flex; gap: 8px; align-items: center; }
```

**移动端表格 → 卡片模式**（768px 以下）：隐藏 `<thead>`，每行变成卡片，用 `data-label` 伪元素替代表头：

```css
@media (max-width: 768px) {
    table thead { display: none; }
    table, table tbody, table tr, table td { display: block; }
    table tr {
        background: var(--c-bg);
        border: 1px solid var(--c-border);
        border-radius: var(--radius-sm);
        padding: 12px;
        margin-bottom: 10px;
        box-shadow: 0 1px 3px var(--c-shadow);
    }
    table td {
        padding: 6px 0;
        border-bottom: none;
        display: flex;
        gap: 8px;
    }
    table td::before {
        content: attr(data-label);
        flex-shrink: 0;
        width: 72px;
        font-weight: 600;
        color: var(--c-text-sub);
        font-size: 12px;
    }
}
```

> 每个 `<td>` 必须设置 `data-label` 属性，如 `<td data-label="用户名">张三</td>`

### 5.6 模态框（Modal）

```css
.modal-overlay {
    position: fixed;
    inset: 0;
    background: var(--c-overlay);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    padding: 24px;
}
.modal {
    background: var(--c-bg);
    border-radius: var(--radius);
    width: 100%;
    max-width: 520px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 8px 32px rgba(0,0,0,.12);
    padding: 28px;
}
.modal h3 { font-size: 18px; margin-bottom: 20px; }
.modal .modal-actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
    margin-top: 24px;
}

/* 移动端底部弹出 */
@media (max-width: 768px) {
    .modal-overlay { padding: 0; align-items: flex-end; }
    .modal {
        max-width: 100%;
        border-radius: var(--radius) var(--radius) 0 0;
        padding: 24px 20px;
        max-height: 85vh;
    }
    .modal .modal-actions { flex-direction: column-reverse; gap: 8px; }
    .modal .modal-actions .btn { width: 100%; }
}
```

点击遮罩关闭模态框：

```js
function showModal() {
    document.getElementById('modal-container').innerHTML = `
    <div class="modal-overlay" onclick="if(event.target===this)closeModal()">
        <div class="modal">...</div>
    </div>`;
}
```

### 5.7 Toast 提示

```css
.toast {
    position: fixed;
    top: 72px;
    left: 50%;
    transform: translateX(-50%);
    padding: 10px 24px;
    border-radius: var(--radius-sm);
    font-size: 14px;
    z-index: 9999;
    opacity: 0;
    pointer-events: none;
    transition: opacity .25s, top .25s;
    box-shadow: 0 4px 16px var(--c-shadow);
}
.toast.show { opacity: 1; top: 80px; }
.toast.success { background: #e8ffea; color: var(--c-success); border: 1px solid #b7eb8f; }
.toast.error   { background: #fff1f0; color: var(--c-danger);  border: 1px solid #ffa39e; }
.toast.info    { background: #f0f1f3; color: var(--c-primary); border: 1px solid #c9cdd4; }

/* 移动端全宽 */
@media (max-width: 768px) {
    .toast { left: 12px; right: 12px; transform: none; top: 56px; text-align: center; }
    .toast.show { top: 60px; }
}
```

JS 用法：

```js
function showToast(msg, type) {
    type = type || 'info';
    $toast.textContent = msg;
    $toast.className = 'toast ' + type + ' show';
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => { $toast.classList.remove('show'); }, 2800);
}
```

### 5.8 侧滑菜单（Drawer）

手机端的抽屉式导航：

```css
.drawer-overlay {
    position: fixed; inset: 0;
    background: rgba(0,0,0,.35);
    z-index: 150;
    opacity: 0; pointer-events: none;
    transition: opacity .25s;
}
.drawer-overlay.open { opacity: 1; pointer-events: auto; }

.drawer {
    position: fixed; top: 0; right: 0; bottom: 0;
    width: 270px; max-width: 80vw;
    background: var(--c-bg);
    z-index: 151;
    transform: translateX(100%);
    transition: transform .25s ease;
    display: flex; flex-direction: column;
    padding: 24px 20px;
    box-shadow: -4px 0 20px rgba(0,0,0,.1);
}
.drawer.open { transform: translateX(0); }

.drawer .btn-logout {
    margin-top: auto;     /* 退出按钮推到底部 */
}
```

### 5.9 Toggle Switch

纯 CSS 实现：

```css
.toggle-switch {
    position: relative; display: inline-block;
    width: 44px; height: 24px; flex-shrink: 0;
}
.toggle-switch input { opacity: 0; width: 0; height: 0; }
.toggle-slider {
    position: absolute; cursor: pointer;
    top: 0; left: 0; right: 0; bottom: 0;
    background: #c9cdd4; border-radius: 24px;
    transition: .2s;
}
.toggle-slider::before {
    content: "";
    position: absolute;
    height: 18px; width: 18px;
    left: 3px; bottom: 3px;
    background: #fff; border-radius: 50%;
    transition: .2s;
}
.toggle-switch input:checked + .toggle-slider { background: var(--c-primary); }
.toggle-switch input:checked + .toggle-slider::before { transform: translateX(20px); }
```

HTML：

```html
<label class="toggle-switch">
    <input type="checkbox" onchange="toggle()">
    <span class="toggle-slider"></span>
</label>
```

### 5.10 其他小组件

**标签/Badge:**

```css
.role-badge {
    font-size: 12px; padding: 2px 8px;
    border-radius: 10px;
    background: #f0f1f3; color: var(--c-primary);
}
.role-badge.admin { background: #ffece8; color: var(--c-danger); }
```

**URI 标签列表:**

```css
.tag-uris { display: flex; flex-wrap: wrap; gap: 4px; }
.tag-uri {
    font-size: 11px; font-family: var(--font-mono);
    background: var(--c-bg-alt); color: var(--c-text-sub);
    padding: 2px 8px; border-radius: 3px;
    border: 1px solid var(--c-border);
    word-break: break-all; max-width: 260px;
    overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
```

**Secret 展示区:**

```css
.secret-display {
    background: var(--c-bg-alt);
    border: 1px solid var(--c-border);
    border-radius: var(--radius-sm);
    padding: 12px 16px; margin: 12px 0;
    display: flex; align-items: center; gap: 12px;
    font-family: var(--font-mono);
    font-size: 13px; word-break: break-all;
}
```

**空状态:**

```css
.empty { text-align: center; padding: 48px 24px; color: var(--c-text-sub); }
.empty .icon { font-size: 40px; margin-bottom: 12px; opacity: .4; }
.empty p { font-size: 14px; }
```

**分页:**

```css
.pagination {
    display: flex; align-items: center; justify-content: center;
    gap: 8px; margin-top: 24px;
}
.pagination span { font-size: 13px; color: var(--c-text-sub); }
```

---

## 6. JavaScript 架构

### 6.1 全局状态

```js
let currentUser = null;   // { id, username, email, role }
let currentPage = null;
```

### 6.2 API 客户端

使用 `fetch` 封装，统一处理 JSON 解析、认证过期、错误抛出：

```js
async function api(url, opts = {}) {
    const headers = opts.headers || {};
    if (!(opts.body instanceof FormData)) {
        headers['Content-Type'] = headers['Content-Type'] || 'application/json';
    }
    const res = await fetch(url, { ...opts, headers, credentials: 'include' });

    let data;
    try { data = await res.json(); } catch (_) {
        throw { code: -1, msg: '响应解析失败' };
    }

    if (data.code !== 0) {
        // 未登录 → 强制跳转
        if (data.code === 2001 && !url.includes('/auth/')) {
            logout(false);
            navigate('#/login');
            throw data;
        }
        throw data;
    }
    return data;
}

function apiGet(url)             { return api(url); }
function apiPost(url, body)      { return api(url, { method: 'POST', body: JSON.stringify(body) }); }
function apiPut(url, body)       { return api(url, { method: 'PUT',  body: JSON.stringify(body) }); }
function apiDel(url)             { return api(url, { method: 'DELETE' }); }
```

### 6.3 路由系统

基于 `location.hash` 的手动路由：

```js
const routes = {
    'login':       { render: renderLogin,      auth: false },
    'register':    { render: renderRegister,   auth: false },
    'dashboard':   { render: renderDashboard,  auth: true  },
    'admin-users': { render: renderAdminUsers, auth: true, admin: true },
    'admin-apps':  { render: renderAdminApps,  auth: true, admin: true },
};

function parseHash() {
    const h = location.hash.replace(/^#\/?/, '');
    // 映射 hash 到路由 key
}

async function doRoute() {
    const page = parseHash();
    const def  = routes[page];
    // 认证守卫 + 管理员守卫
    currentPage = page;
    renderHeader();
    await def.render();
}

window.addEventListener('hashchange', () => { closeDrawer(); doRoute(); });
```

### 6.4 页面渲染模式

每个页面一个 `render*()` 函数，内联 HTML 模板字符串 + 事件绑定：

```js
async function renderDashboard() {
    $main.className = '';         // 清除 page-auth 等类
    $main.innerHTML = `
        <div class="card">
            <div class="card-header"><h2>账户信息</h2></div>
            <div id="dashboard-info">...</div>
        </div>`;
    // ... 填充数据
}
```

### 6.5 XSS 防护

三个转义函数针对不同上下文：

```js
function escHtml(s) {
    // 用于 HTML 文本节点
    return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;')
                    .replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
function escAttr(s) {
    // 用于 HTML 属性值
    return String(s).replace(/&/g,'&amp;').replace(/"/g,'&quot;')
                    .replace(/</g,'&lt;').replace(/>/g,'&gt;');
}
function escJs(s) {
    // 用于内联 JS 字符串（如 onclick 参数）
    return String(s).replace(/\\/g,'\\\\').replace(/'/g,"\\'");
}
```

### 6.6 工具函数

```js
function formatTime(iso) {
    // ISO 字符串 → "YYYY-MM-DD HH:mm:ss"
    const d = new Date(iso);
    const pad = (n) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}
```

---

## 7. 响应式设计

### 7.1 断点

| 断点 | 目标 |
|------|------|
| `max-width: 768px` | 平板/手机（主要断点） |
| `max-width: 400px` | 超小屏微调 |

### 7.2 响应式策略

| 组件 | 桌面端 | 移动端 |
|------|--------|--------|
| Header 导航 | `display: flex` | 隐藏，改用汉堡菜单 + Drawer |
| 表格 | 标准 `<table>` | 隐藏 thead，每行变卡片，`data-label` 伪元素显示标签 |
| 模态框 | 居中 | 底部弹出（bottom sheet） |
| 表单输入 | `font-size: 14px` | `font-size: 16px`（防 iOS 缩放） |
| Toast | 居中 | 全宽 |
| 卡片 | `padding: 24px` | `padding: 16px` |

---

## 8. 动效（Animation）

全程使用 **150ms ~ 250ms** 的 `transition`，不使用 `@keyframes` 动画：

| 位置 | 时长 | 效果 |
|------|------|------|
| 按钮 hover | `.15s` | 颜色过渡 |
| 表单 focus | `.15s` | 边框色过渡 |
| Toast 弹出 | `.25s` | opacity + top |
| Drawer 滑入 | `.25s ease` | transform translateX |
| Toggle Switch | `.2s` | 背景色 + 滑块位移 |
| 遮罩层 | `.25s` | opacity |

---

## 9. 颜色体系速查

```
主色：       #1d2129（深灰）→ hover: #000000
页面底色：   #f7f8fa
卡片白底：   #ffffff
边框：       #e5e6eb
主文字：     #1d2129
辅助文字：   #86909c
浅文字：     #c9cdd4
成功：       #00b42a   → 浅底 #e8ffea
危险：       #f53f3f   → 浅底 #fff1f0
警告：       #ff7d00
阴影：       rgba(0,0,0,.06)
遮罩：       rgba(0,0,0,.3)
```

---

## 10. 设计原则总结

1. **零依赖** — 不引入任何第三方库，浏览器原生能力足够
2. **CSS Variables 统一管理** — 所有颜色/圆角/字体通过 `:root` 变量，易于换肤
3. **Spacing 靠组件自身** — 通过 `margin-bottom`、`gap`、`margin-top` 控制间距，无全局 spacing 类
4. **移动端友好** — 表格→卡片、模态框→底部弹出、汉堡菜单+侧滑抽屉
5. **动画克制** — 只用 150~250ms transition，无复杂 keyframe 动画
6. **转义必须** — 所有用户数据插入 DOM 前必须经过 `escHtml` / `escAttr` / `escJs`
7. **内联事件** — `onclick` 直接写在 HTML 模板字符串中，无需 `addEventListener` 委派（简单场景足够）
8. **CSS 文件 ≤ 500 行，JS 文件 ≤ 1000 行** — 单文件可控，不拆分模块
