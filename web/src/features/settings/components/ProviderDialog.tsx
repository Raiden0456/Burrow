import { useState, type FormEvent, type ReactNode } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { useI18n } from "@/i18n"
import { settingsApi } from "../api"
import type { Provider, ProviderKind } from "../types/settings"
import { acpPresets, emptyForm, kindLabels, toForm, toInput, type ProviderForm } from "../utils/providerForm"

type Props = {
  provider: Provider | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function Field({ id, label, hint, children }: { id: string; label: string; hint?: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor={id}>{label}</Label>
      {children}
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  )
}

function LoginHint() {
  const { t } = useI18n()
  const [before, after] = t("settings.provider.loginClaude").split("{setup}")
  const [middle, end] = (after ?? "").split("{env}")
  return (
    <div className="flex flex-col gap-1 rounded-lg border bg-muted/40 p-3 text-xs text-muted-foreground">
      <span>{t("settings.provider.loginIntro")}</span>
      <span>
        {t("settings.provider.loginCodex")} <code className="text-foreground">docker compose exec -u app app codex login --device-auth</code>
      </span>
      <span>
        {before}
        <code className="text-foreground">claude setup-token</code>
        {middle}
        <code className="text-foreground">CLAUDE_CODE_OAUTH_TOKEN</code>
        {end}
      </span>
    </div>
  )
}

export function ProviderDialog({ provider, open, onOpenChange }: Props) {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const [form, setForm] = useState<ProviderForm>(() => (provider ? toForm(provider) : emptyForm()))
  const set = <K extends keyof ProviderForm>(key: K, value: ProviderForm[K]) => setForm((f) => ({ ...f, [key]: value }))

  const save = useMutation({
    mutationFn: () => (provider ? settingsApi.updateProvider(provider.id, toInput(form)) : settingsApi.createProvider(toInput(form))),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["providers"] })
      onOpenChange(false)
    },
  })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    save.mutate()
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90svh] overflow-y-auto max-sm:pb-0">
        <form onSubmit={submit} className="flex flex-col gap-4">
          <DialogHeader>
            <DialogTitle>{provider ? t("settings.provider.editTitle") : t("settings.provider.addTitle")}</DialogTitle>
            <DialogDescription>{t("settings.provider.description")}</DialogDescription>
          </DialogHeader>

          <Field id="kind" label={t("settings.provider.type")}>
            <Select value={form.kind} onValueChange={(v) => set("kind", v as ProviderKind)}>
              <SelectTrigger id="kind" className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {(Object.keys(kindLabels) as ProviderKind[]).map((k) => (
                  <SelectItem key={k} value={k}>
                    {t(kindLabels[k])}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </Field>

          {form.kind === "acp" && !provider && (
            <div className="flex flex-wrap gap-2">
              {acpPresets.map((p) => (
                <Button key={p.command} type="button" size="sm" variant="outline" onClick={() => setForm((f) => ({ ...f, name: f.name || p.name, command: p.command, args: p.args }))}>
                  {t(p.label)}
                </Button>
              ))}
            </div>
          )}

          <Field id="name" label={t("settings.provider.name")}>
            <Input id="name" required value={form.name} onChange={(e) => set("name", e.target.value)} />
          </Field>

          {form.kind === "acp" ? (
            <>
              <LoginHint />
              <Field id="command" label={t("settings.provider.command")} hint={t("settings.provider.commandHint")}>
                <Input id="command" required value={form.command} onChange={(e) => set("command", e.target.value)} placeholder="codex-acp" />
              </Field>
              <Field id="args" label={t("settings.provider.args")} hint={t("settings.provider.argsHint")}>
                <Textarea id="args" rows={2} value={form.args} onChange={(e) => set("args", e.target.value)} />
              </Field>
              <Field id="env" label={t("settings.provider.env")} hint={t("settings.provider.envHint")}>
                <Textarea id="env" rows={2} value={form.env} onChange={(e) => set("env", e.target.value)} />
              </Field>
            </>
          ) : (
            <>
              <Field id="apiKey" label={t("settings.provider.apiKey")} hint={provider?.hasApiKey ? t("settings.provider.apiKeyKeep") : undefined}>
                <Input
                  id="apiKey"
                  type="password"
                  autoComplete="off"
                  required={form.kind === "anthropic" && !provider?.hasApiKey}
                  value={form.apiKey}
                  onChange={(e) => set("apiKey", e.target.value)}
                />
              </Field>
              <Field
                id="model"
                label={t("settings.provider.model")}
                hint={form.kind === "anthropic" ? t("settings.provider.modelHint.anthropic") : t("settings.provider.modelHint.openai")}
              >
                <Input
                  id="model"
                  required={form.kind === "openai"}
                  value={form.model}
                  onChange={(e) => set("model", e.target.value)}
                  placeholder={form.kind === "anthropic" ? "claude-opus-5-5" : ""}
                />
              </Field>
              <Field
                id="baseUrl"
                label={t("settings.provider.baseUrl")}
                hint={form.kind === "openai" ? t("settings.provider.baseUrlHint.openai") : t("settings.provider.baseUrlHint.anthropic")}
              >
                <Input id="baseUrl" value={form.baseUrl} onChange={(e) => set("baseUrl", e.target.value)} />
              </Field>
            </>
          )}

          <DialogFooter className="sticky bottom-0 z-10 bg-popover max-sm:mb-0 max-sm:pb-[max(env(safe-area-inset-bottom),1rem)]">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t("common.cancel")}
            </Button>
            <Button type="submit" disabled={save.isPending}>
              {t("common.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
