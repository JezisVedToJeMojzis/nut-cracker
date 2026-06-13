/// <reference types="@sveltejs/kit" />
/// <reference lib="webworker" />

// SvelteKit service worker: precache the built app shell and static assets,
// serve them cache-first, and fall back to the cache for navigations when
// offline. API requests (/api/*) always go to the network.
import { build, files, version } from '$service-worker';

const sw = self as unknown as ServiceWorkerGlobalScope;

const CACHE = `nutcracker-cache-${version}`;
const ASSETS = [...build, ...files];

sw.addEventListener('install', (event) => {
	event.waitUntil(caches.open(CACHE).then((cache) => cache.addAll(ASSETS)).then(() => sw.skipWaiting()));
});

sw.addEventListener('activate', (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
			.then(() => sw.clients.claim())
	);
});

sw.addEventListener('fetch', (event) => {
	const { request } = event;
	if (request.method !== 'GET') return;

	const url = new URL(request.url);
	// Never cache API calls — always hit the network.
	if (url.pathname.startsWith('/api/')) return;

	event.respondWith(
		(async () => {
			const cache = await caches.open(CACHE);

			// Cache-first for precached build/static assets.
			if (ASSETS.includes(url.pathname)) {
				const cached = await cache.match(url.pathname);
				if (cached) return cached;
			}

			try {
				const res = await fetch(request);
				if (res.ok && url.origin === location.origin) cache.put(request, res.clone());
				return res;
			} catch {
				// Offline: fall back to any cached copy (e.g. the app shell).
				const cached = await cache.match(request);
				if (cached) return cached;
				throw new Error('offline and not cached');
			}
		})()
	);
});
