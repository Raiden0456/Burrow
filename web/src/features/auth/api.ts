import { api } from "@/lib/api"

export const authApi = {
  me: () => api.get<void>("/api/auth/me"),
  login: (password: string) => api.post<void>("/api/auth/login", { password }),
  logout: () => api.post<void>("/api/auth/logout"),
}
