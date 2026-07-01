/* ================================================================
   Account - Service Worker (PWA)
   ================================================================ */

const CACHE_NAME = "account-v1";
const STATIC_ASSETS = [
  "/",
  "/css/style.css",
  "/js/app.js",
  "/manifest.json",
  "/icon.svg",
];

/* ---- Install: 预缓存静态资源 ---- */
self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(STATIC_ASSETS).catch((err) => {
        console.warn("[SW] Pre-cache failed (some assets may be unavailable):", err);
      });
    })
  );
  self.skipWaiting();
});

/* ---- Activate: 清理旧缓存 ---- */
self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) => {
      return Promise.all(
        keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key))
      );
    })
  );
  self.clients.claim();
});

/* ---- Fetch: 策略分发 ---- */
self.addEventListener("fetch", (event) => {
  const url = new URL(event.request.url);

  // API 请求：网络优先，失败时返回离线提示
  if (url.pathname.startsWith("/api/") || url.pathname.startsWith("/oauth/")) {
    event.respondWith(networkFirstWithOffline(event.request));
    return;
  }

  // 头像 / 静态文件：缓存优先（后台更新）
  if (
    url.pathname.startsWith("/avatar/") ||
    url.pathname.startsWith("/css/") ||
    url.pathname.startsWith("/js/") ||
    url.pathname === "/manifest.json" ||
    url.pathname === "/icon.svg"
  ) {
    event.respondWith(cacheFirstWithRefresh(event.request));
    return;
  }

  // SPA 页面 (HTML)：网络优先，回退到缓存 index.html
  event.respondWith(networkFirstWithCacheFallback(event.request));
});

/* ================================================================
   策略函数
   ================================================================ */

// 缓存优先 + 后台更新（适合不常变的静态资源）
async function cacheFirstWithRefresh(request) {
  const cached = await caches.match(request);
  const fetchPromise = fetch(request).then((response) => {
    if (response && response.ok) {
      caches.open(CACHE_NAME).then((cache) => cache.put(request, response.clone()));
    }
    return response;
  });
  return cached || fetchPromise;
}

// 网络优先 + 缓存回退（适合 SPA 页面）
async function networkFirstWithCacheFallback(request) {
  try {
    const response = await fetch(request);
    if (response && response.ok) {
      caches.open(CACHE_NAME).then((cache) => cache.put(request, response.clone()));
    }
    return response;
  } catch (_) {
    const cached = await caches.match(request);
    return cached || caches.match("/");
  }
}

// 网络优先 + 离线 JSON 响应（适合 API）
async function networkFirstWithOffline(request) {
  try {
    const response = await fetch(request);
    return response;
  } catch (_) {
    return new Response(
      JSON.stringify({ code: -1, msg: "当前处于离线状态，请检查网络连接", data: null }),
      { status: 200, headers: { "Content-Type": "application/json" } }
    );
  }
}
