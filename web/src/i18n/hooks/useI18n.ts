import { useSyncExternalStore } from "react"
import { getLocale, setLocale, subscribe, t } from "../store"

export function useI18n() {
  const locale = useSyncExternalStore(subscribe, getLocale)
  return { locale, t, setLocale }
}
