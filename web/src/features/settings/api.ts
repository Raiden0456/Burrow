import { api } from "@/lib/api"
import type { InterfaceSettings, Preferences, Provider, ProviderInput, ProviderTest, PushInfo, Routing } from "./types/settings"

export const settingsApi = {
  providers: () => api.get<Provider[]>("/api/providers"),
  createProvider: (input: ProviderInput) => api.post<Provider>("/api/providers", input),
  updateProvider: (id: number, input: ProviderInput) => api.put<Provider>(`/api/providers/${id}`, input),
  deleteProvider: (id: number) => api.delete(`/api/providers/${id}`),
  testProvider: (id: number) => api.post<ProviderTest>(`/api/providers/${id}/test`),
  routing: () => api.get<Routing>("/api/routing"),
  saveRouting: (r: Routing) => api.put<Routing>("/api/routing", r),
  interface: () => api.get<InterfaceSettings>("/api/settings/interface"),
  saveInterface: (s: InterfaceSettings) => api.put<InterfaceSettings>("/api/settings/interface", s),
  preferences: () => api.get<Preferences>("/api/settings/preferences"),
  savePreferences: (p: Preferences) => api.put<Preferences>("/api/settings/preferences", p),
  pushInfo: () => api.get<PushInfo>("/api/push/public-key"),
  subscribe: (sub: PushSubscriptionJSON) => api.post<void>("/api/push/subscriptions", sub),
  unsubscribe: (endpoint: string) => api.delete("/api/push/subscriptions", { endpoint }),
  testPush: () => api.post<{ sent: number }>("/api/push/test"),
}
