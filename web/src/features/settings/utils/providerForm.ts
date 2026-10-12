import type { MessageKey } from "@/i18n"
import type { Provider, ProviderInput, ProviderKind } from "../types/settings"

export type ProviderForm = {
  name: string
  kind: ProviderKind
  model: string
  apiKey: string
  baseUrl: string
  command: string
  args: string
  env: string
}

export const acpPresets: { label: MessageKey; name: string; command: string; args: string }[] = [
  { label: "settings.provider.preset.codex", name: "Codex", command: "codex-acp", args: "" },
  { label: "settings.provider.preset.claude", name: "Claude Code", command: "claude-agent-acp", args: "" },
  { label: "settings.provider.preset.gemini", name: "Gemini", command: "gemini", args: "--experimental-acp" },
]

export const kindLabels: Record<ProviderKind, MessageKey> = {
  anthropic: "settings.provider.kind.anthropic",
  openai: "settings.provider.kind.openai",
  acp: "settings.provider.kind.acp",
}

export function emptyForm(): ProviderForm {
  return { name: "", kind: "acp", model: "", apiKey: "", baseUrl: "", command: "", args: "", env: "" }
}

export function toForm(p: Provider): ProviderForm {
  return {
    name: p.name,
    kind: p.kind,
    model: p.model,
    apiKey: "",
    baseUrl: p.baseUrl,
    command: p.command,
    args: (p.args ?? []).join("\n"),
    env: Object.entries(p.env ?? {})
      .map(([k, v]) => `${k}=${v}`)
      .join("\n"),
  }
}

function lines(text: string) {
  return text
    .split("\n")
    .map((l) => l.trim())
    .filter(Boolean)
}

export function toInput(f: ProviderForm): ProviderInput {
  const env: Record<string, string> = {}
  for (const line of lines(f.env)) {
    const eq = line.indexOf("=")
    if (eq > 0) env[line.slice(0, eq).trim()] = line.slice(eq + 1).trim()
  }
  return {
    name: f.name.trim(),
    kind: f.kind,
    config: {
      model: f.model.trim(),
      apiKey: f.apiKey.trim(),
      baseUrl: f.baseUrl.trim(),
      command: f.command.trim(),
      args: lines(f.args),
      env,
    },
  }
}
