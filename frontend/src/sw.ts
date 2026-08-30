/// <reference lib="webworker" />
import { clientsClaim } from 'workbox-core'
import { cleanupOutdatedCaches, matchPrecache, precacheAndRoute } from 'workbox-precaching'
import { registerRoute, setCatchHandler } from 'workbox-routing'
import { NetworkFirst } from 'workbox-strategies'

declare let self: ServiceWorkerGlobalScope

self.skipWaiting()
clientsClaim()
precacheAndRoute(self.__WB_MANIFEST)
cleanupOutdatedCaches()

// HTML и API — сеть первая; иначе SW отдаёт старый index после деплоя. SSE не перехватываем.
registerRoute(
  ({ request }) => request.mode === 'navigate',
  new NetworkFirst({ cacheName: 'pages', networkTimeoutSeconds: 3 }),
)

registerRoute(
  ({ url, request }) =>
    request.method === 'GET' && url.pathname.startsWith('/api/') && !url.pathname.startsWith('/api/v1/events'),
  new NetworkFirst({ cacheName: 'api', networkTimeoutSeconds: 5 }),
)

setCatchHandler(async ({ request }) => {
  if (request.mode === 'navigate' || request.destination === 'document') {
    const offline = (await matchPrecache('/offline.html')) ?? (await caches.match('/offline.html'))
    if (offline) {
      return offline
    }
  }
  return Response.error()
})

type PushPayload = {
  id?: string
  item_id?: string
  to_status?: string
  title?: string
}

self.addEventListener('push', (event) => {
  let data: PushPayload = {}
  try {
    data = event.data ? (event.data.json() as PushPayload) : {}
  } catch {
    data = { title: event.data?.text() }
  }
  const title = data.title ?? 'Duekeep'
  const body = data.to_status ? `Статус: ${data.to_status}` : 'Новое уведомление'
  event.waitUntil(
    self.registration.showNotification(title, {
      body,
      icon: '/icons/icon-192.png',
      data: { url: data.item_id ? `/items/${data.item_id}` : '/notifications' },
    }),
  )
})

self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  const url = (event.notification.data as { url?: string } | undefined)?.url ?? '/'
  event.waitUntil(
    self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windows) => {
      const existing = windows.find((c) => 'focus' in c)
      if (existing) {
        return existing.navigate(url).then((client) => client?.focus())
      }
      return self.clients.openWindow(url)
    }),
  )
})
