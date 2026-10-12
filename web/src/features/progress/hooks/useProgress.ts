import { useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import { progressApi } from "../api"
import type { TopicStat } from "../types/progress"

export function useProgress() {
  const query = useQuery({ queryKey: ["progress"], queryFn: progressApi.summary })
  const byTopic = useMemo(() => new Map<string, TopicStat>((query.data?.topics ?? []).map((t) => [t.slug, t])), [query.data])
  return { ...query, byTopic }
}
