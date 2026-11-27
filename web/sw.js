// Service worker that lets the game start with no network after one online
// visit. It stores the whole bundle on install, then answers every request
// from the network first, so an online visit always gets the newest build,
// and falls back to the stored copy only when the network fails.
'use strict';

const CACHE = 'critter-kingdom-v1';
const CACHE_PREFIX = 'critter-kingdom-';

// Every file of the bundle, relative to this script. Keep one plain string
// per line: pwa_test.go reads this list. `./` stands for the page, because
// wasmserve answers index.html with a page of its own.
const PRECACHE = [
  './',
  'loader.js',
  'wasm_exec.js',
  'main.wasm',
  'manifest.webmanifest',
  'icons/icon-192.png',
  'icons/icon-512.png',
  'icons/icon-maskable-512.png',
  'icons/apple-touch-icon.png',
];

self.addEventListener('install', (event) => {
  // A new worker takes over at once. Network-first means it never serves a
  // build older than the one it just fetched.
  self.skipWaiting();
  event.waitUntil(caches.open(CACHE).then((cache) => cache.addAll(PRECACHE)));
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    (async () => {
      const names = await caches.keys();
      await Promise.all(
        names
          .filter((name) => name.startsWith(CACHE_PREFIX) && name !== CACHE)
          .map((name) => caches.delete(name)),
      );
      await self.clients.claim();
    })(),
  );
});

// fromCache answers a request the network could not. The start URL can carry
// a query (?pad=1) or name index.html, so a page load falls back to `./`.
async function fromCache(request) {
  const cached = await caches.match(request, { ignoreSearch: true });
  if (cached) {
    return cached;
  }
  if (request.mode === 'navigate') {
    const page = await caches.match('./');
    if (page) {
      return page;
    }
  }
  return Response.error();
}

self.addEventListener('fetch', (event) => {
  const request = event.request;
  if (request.method !== 'GET' || new URL(request.url).origin !== self.location.origin) {
    return;
  }

  event.respondWith(
    (async () => {
      let response;
      try {
        response = await fetch(request);
      } catch (err) {
        return fromCache(request);
      }
      if (response.ok) {
        // Store a copy and hand back the original untouched, so loader.js
        // still sees the body stream and headers for its progress bar.
        const copy = response.clone();
        event.waitUntil(
          caches
            .open(CACHE)
            .then((cache) => cache.put(request, copy))
            .catch((err) => console.warn('Could not cache', request.url, err)),
        );
      }
      return response;
    })(),
  );
});
