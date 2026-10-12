import { useMutation, useQueryClient } from "@tanstack/react-query"
import { setLocale, type Locale } from "@/i18n"
import { settingsApi } from "../api"

export function useChangeLanguage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (language: Locale) => {
      setLocale(language)
      return settingsApi.saveInterface({ language })
    },
    onSuccess: (saved) => queryClient.setQueryData(["interface"], saved),
  })
}
