import { useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import { useI18n } from "@/i18n"
import { topicsApi } from "../api"
import type { Topic } from "../types/curriculum"
import { localizeModules } from "../utils/localize"

export function useCurriculum() {
  const { locale } = useI18n()
  const query = useQuery({ queryKey: ["curriculum"], queryFn: topicsApi.curriculum, staleTime: Infinity })

  const modules = useMemo(() => localizeModules(query.data ?? [], locale), [query.data, locale])
  const topics = useMemo(() => new Map<string, Topic>(modules.flatMap((m) => m.topics.map((t) => [t.slug, t]))), [modules])

  return { isPending: query.isPending, error: query.error, modules, topics }
}
