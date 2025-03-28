const CACHE_NAME = 'nutrition-scanner-v1';
const ASSETS_TO_CACHE = [
    '/',
    '/index.html',
    '/css/styles.css',
    '/js/app.js',
    '/manifest.json',
    '/favicon.ico'
];

// Install service worker and cache static assets
self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then((cache) => {
                // Use Promise.allSettled instead of Promise.all to continue even if some assets fail
                return Promise.allSettled(
                    ASSETS_TO_CACHE.map(url => 
                        cache.add(url).catch(error => {
                            console.warn(`Failed to cache asset: ${url}`, error);
                            return null; // Continue despite the error
                        })
                    )
                );
            })
    );
});

// Activate service worker and clean up old caches
self.addEventListener('activate', (event) => {
    event.waitUntil(
        caches.keys().then((cacheNames) => {
            return Promise.all(
                cacheNames
                    .filter((name) => name !== CACHE_NAME)
                    .map((name) => caches.delete(name))
            );
        })
    );
});

// Fetch event handler
self.addEventListener('fetch', (event) => {
    // Don't cache WebSocket connections
    if (event.request.url.includes('/ws')) {
        return;
    }

    event.respondWith(
        caches.match(event.request)
            .then((response) => {
                // Return cached response if found
                if (response) {
                    return response;
                }

                // Clone the request because it can only be used once
                const fetchRequest = event.request.clone();

                return fetch(fetchRequest)
                    .then((response) => {
                        // Check if response is valid
                        if (!response || response.status !== 200 || response.type !== 'basic') {
                            return response;
                        }

                        // Clone the response because it can only be used once
                        const responseToCache = response.clone();

                        caches.open(CACHE_NAME)
                            .then((cache) => {
                                cache.put(event.request, responseToCache);
                            });

                        return response;
                    })
                    .catch(error => {
                        console.warn(`Failed to fetch: ${event.request.url}`, error);
                        // Return a fallback response or just let the error propagate
                        return new Response('Network error occurred', {
                            status: 503,
                            statusText: 'Service Unavailable'
                        });
                    });
            })
    );
}); 
