import type { ReactNode } from "react"
import { useQuery } from "@tanstack/react-query"
import { Navigate } from "react-router"
import { authApi } from "@/features/auth/api"
import { ApiError } from "@/lib/api"
import { PageLoader } from "@/components/PageLoader"

export function RequireAuth({ children }: { children: ReactNode }) {
  const me = useQuery({ queryKey: ["me"], queryFn: authApi.me, retry: false, staleTime: Infinity })

  if (me.isPending) return <PageLoader />
  if (me.error instanceof ApiError && me.error.status === 401) return <Navigate to="/login" replace />
  return children
}
