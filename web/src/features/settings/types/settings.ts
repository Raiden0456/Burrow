export type ProviderKind = "anthropic" | "openai" | "acp"

export type Provider = {
  id: number
  name: string
  kind: ProviderKind
  model: string
  baseUrl: string
  command: string
  args: string[] | null
  env: Record<string, string> | null
  hasApiKey: boolean
}

export type ProviderConfig = {
  model: string
  apiKey: string
  baseUrl: string
  command: string
  args: string[]
  env: Record<string, string>
}

export type ProviderInput = {
  name: string
  kind: ProviderKind
  config: ProviderConfig
}

export type ProviderTest =
  | { ok: true; reply: string; latencyMs: number }
  | { ok: false; error: string }

export type Routing = {
  chat: number
  generate: number
  grade: number
}

export type Preferences = {
  language: string
  profile: string
  dailyEnabled: boolean
  dailyTime: string
  dailySize: number
}

export type InterfaceSettings = {
  language: string
}

export type PushInfo = {
  publicKey: string
  subscriptions: number
}
