import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { ApiError, errorMessage } from "@/lib/api"

function handleUnauthorized(err: unknown) {
  if (err instanceof ApiError && err.status === 401 && window.location.pathname !== "/login") {
    window.location.assign("/login")
    return true
  }
  return false
}

export const queryClient = new QueryClient({
  queryCache: new QueryCache({ onError: handleUnauthorized }),
  mutationCache: new MutationCache({
    onError: (err, _vars, _ctx, mutation) => {
      if (handleUnauthorized(err) || mutation.meta?.silent) return
      toast.error(errorMessage(err))
    },
  }),
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      refetchOnWindowFocus: false,
      retry: (count, err) => !(err instanceof ApiError && err.status < 500) && count < 2,
    },
  },
})
