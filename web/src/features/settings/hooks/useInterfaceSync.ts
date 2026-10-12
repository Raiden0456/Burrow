import { useEffect } from "react"
import { useQuery } from "@tanstack/react-query"
import { getLocale, isLocale, setLocale } from "@/i18n"
import { settingsApi } from "../api"

export function useInterfaceSync() {
  const ui = useQuery({ queryKey: ["interface"], queryFn: settingsApi.interface, staleTime: Infinity })

  useEffect(() => {
    if (!ui.data) return
    if (isLocale(ui.data.language)) setLocale(ui.data.language)
    else settingsApi.saveInterface({ language: getLocale() })
  }, [ui.data])
}
