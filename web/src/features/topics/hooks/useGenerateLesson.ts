import { useMutation, useQueryClient } from "@tanstack/react-query"
import { topicsApi } from "../api"

export function useGenerateLesson() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: topicsApi.generateLesson,
    onSuccess: (_, slug) => {
      queryClient.invalidateQueries({ queryKey: ["jobs"] })
      queryClient.invalidateQueries({ queryKey: ["lesson", slug] })
    },
  })
}
