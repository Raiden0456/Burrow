/// <reference lib="webworker" />
import { clientsClaim } from "workbox-core"
import { cleanupOutdatedCaches, createHandlerBoundToURL, precacheAndRoute } from "workbox-precaching"
import { NavigationRoute, registerRoute } from "workbox-routing"

declare let self: ServiceWorkerGlobalScope

type PushPayload = { title?: string; body?: string; url?: string }

self.skipWaiting()
clientsClaim()
cleanupOutdatedCaches()
precacheAndRoute(self.__WB_MANIFEST)
registerRoute(new NavigationRoute(createHandlerBoundToURL("index.html"), { denylist: [/^\/api\//, /^\/healthz/] }))

self.addEventListener("push", (event) => {
  const data: PushPayload = event.data?.json() ?? {}
  event.waitUntil(
    self.registration.showNotification(data.title ?? "Gonteractive", {
      body: data.body,
      icon: "/pwa-192x192.png",
      badge: "/pwa-64x64.png",
      data: { url: data.url ?? "/" },
    }),
  )
})

self.addEventListener("notificationclick", (event) => {
  event.notification.close()
  const url = new URL((event.notification.data as { url: string }).url, self.location.origin).href
  event.waitUntil(
    self.clients.matchAll({ type: "window", includeUncontrolled: true }).then((windows) => {
      const existing = windows.find((w) => w.url.startsWith(self.location.origin))
      if (existing) return existing.focus().then((w) => w.navigate(url))
      return self.clients.openWindow(url)
    }),
  )
})
