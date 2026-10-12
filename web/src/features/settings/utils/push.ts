import { t } from "@/i18n"

function decodeKey(base64: string) {
  const padded = (base64 + "=".repeat((4 - (base64.length % 4)) % 4)).replace(/-/g, "+").replace(/_/g, "/")
  const raw = atob(padded)
  const bytes = new Uint8Array(new ArrayBuffer(raw.length))
  for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i)
  return bytes
}

export function pushSupported() {
  return "serviceWorker" in navigator && "PushManager" in window && "Notification" in window
}

export async function currentSubscription() {
  if (!pushSupported()) return null
  const reg = await navigator.serviceWorker.getRegistration()
  return reg ? reg.pushManager.getSubscription() : null
}

export async function subscribe(publicKey: string) {
  if (!pushSupported()) throw new Error(t("settings.push.unsupported"))
  const reg = await navigator.serviceWorker.getRegistration()
  if (!reg) throw new Error(t("settings.push.noWorker"))
  const permission = await Notification.requestPermission()
  if (permission !== "granted") throw new Error(t("settings.push.denied"))
  return (await reg.pushManager.getSubscription()) ?? (await reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: decodeKey(publicKey) }))
}
