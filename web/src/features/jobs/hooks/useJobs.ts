import { useQuery } from "@tanstack/react-query"
import { jobsApi } from "../api"

export function useJobs() {
  return useQuery({
    queryKey: ["jobs"],
    queryFn: jobsApi.recent,
    refetchInterval: (query) => (query.state.data?.some((j) => j.status === "running") ? 2000 : 15000),
    refetchIntervalInBackground: true,
  })
}

export function useRunningJob(kind: string, topicSlug: string) {
  const jobs = useJobs()
  return jobs.data?.find((j) => j.kind === kind && j.topicSlug === topicSlug && j.status === "running") ?? null
}
