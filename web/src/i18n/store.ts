import { en, type MessageKey, type Messages, type Plural } from "./locales/en"
import { ru } from "./locales/ru"

export type Locale = "en" | "ru"
export type Vars = Record<string, string | number>

export const locales: Locale[] = ["en", "ru"]

const storageKey = "gonteractive:locale"
const dictionaries: Record<Locale, Messages> = { en, ru }
const listeners = new Set<() => void>()

function detect(): Locale {
  const saved = localStorage.getItem(storageKey)
  if (saved === "en" || saved === "ru") return saved
  return navigator.language.toLowerCase().startsWith("ru") ? "ru" : "en"
}

let current = detect()
document.documentElement.lang = current

export function isLocale(value: unknown): value is Locale {
  return value === "en" || value === "ru"
}

export function getLocale() {
  return current
}

export function setLocale(locale: Locale) {
  if (locale === current) return
  current = locale
  localStorage.setItem(storageKey, locale)
  document.documentElement.lang = locale
  listeners.forEach((notify) => notify())
}

export function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

export function hasMessage(key: string): key is MessageKey {
  return key in en
}

export function t(key: MessageKey, vars?: Vars): string {
  const message = dictionaries[current][key]
  const text =
    typeof message === "string" ? message : (message[new Intl.PluralRules(current).select(Number(vars?.count ?? 0)) as keyof Plural] ?? message.other)
  return vars ? text.replace(/\{(\w+)\}/g, (_, name: string) => String(vars[name] ?? "")) : text
}
