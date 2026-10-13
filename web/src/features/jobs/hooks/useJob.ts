import { useEffect, useLayoutEffect, useRef } from "react"
import { useQuery } from "@tanstack/react-query"
import { jobsApi } from "../api"
import type { Job } from "../types/job"

export function useJob(id: number | null, onFinish?: (job: Job) => void) {
  const query = useQuery({
    queryKey: ["job", id],
    queryFn: () => jobsApi.get(id as number),
    enabled: id !== null,
    refetchInterval: (q) => (q.state.data?.status === "running" || !q.state.data ? 1500 : false),
  })

  const finish = useRef(onFinish)
  useLayoutEffect(() => {
    finish.current = onFinish
  })
  const reported = useRef<number | null>(null)
  const job = query.data
  useEffect(() => {
    if (job && job.status !== "running" && reported.current !== job.id) {
      reported.current = job.id
      finish.current?.(job)
    }
  }, [job])

  return job ?? null
}
