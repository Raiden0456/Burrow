import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { BotIcon, Loader2Icon, PencilIcon, PlugZapIcon, PlusIcon, Trash2Icon } from "lucide-react"
import { toast } from "sonner"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { IconChip } from "@/components/IconChip"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import { toneChip } from "@/lib/tones"
import { settingsApi } from "../api"
import type { Provider, ProviderKind } from "../types/settings"
import { kindLabels } from "../utils/providerForm"
import { ProviderDialog } from "./ProviderDialog"

const kindTone = { anthropic: "orange", openai: "green", acp: "blue" } as const satisfies Record<ProviderKind, string>

export function ProvidersCard() {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const providers = useQuery({ queryKey: ["providers"], queryFn: settingsApi.providers })
  const [editing, setEditing] = useState<Provider | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)

  const remove = useMutation({
    mutationFn: settingsApi.deleteProvider,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["providers"] })
      queryClient.invalidateQueries({ queryKey: ["routing"] })
    },
  })
  const test = useMutation({
    mutationFn: settingsApi.testProvider,
    onSuccess: (res) =>
      res.ok
        ? toast.success(t("settings.providers.works", { seconds: (res.latencyMs / 1000).toFixed(1), reply: res.reply }))
        : toast.error(res.error, { duration: 15_000 }),
  })

  const openDialog = (p: Provider | null) => {
    setEditing(p)
    setDialogOpen(true)
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-start gap-3">
        <IconChip icon={BotIcon} tone="blue" />
        <div className="flex flex-1 flex-col gap-1">
          <CardTitle>{t("settings.providers.title")}</CardTitle>
          <CardDescription>{t("settings.providers.description")}</CardDescription>
        </div>
        <CardAction>
          <Button size="sm" onClick={() => openDialog(null)}>
            <PlusIcon />
            {t("settings.providers.add")}
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="flex flex-col gap-2">
        {providers.isPending && <Skeleton className="h-12 w-full" />}
        {providers.data?.length === 0 && <p className="text-sm text-muted-foreground">{t("settings.providers.empty")}</p>}
        {providers.data?.map((p) => (
          <div key={p.id} className="flex flex-col gap-3 rounded-lg border px-3 py-3 sm:flex-row sm:items-center sm:gap-2 sm:py-2">
            <div className="min-w-0 flex-1">
              <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 font-medium">
                <span className="min-w-0 truncate">{p.name}</span>
                <Badge className={cn("max-w-full shrink border-transparent", toneChip[kindTone[p.kind]])}>
                  <span className="truncate">{t(kindLabels[p.kind])}</span>
                </Badge>
              </div>
              <div className="mt-1 truncate font-mono text-xs text-muted-foreground">
                {p.kind === "acp" ? [p.command, ...(p.args ?? [])].join(" ") : p.model || t("settings.providers.defaultModel")}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button
                size="sm"
                variant="outline"
                className="max-sm:flex-1"
                onClick={() => test.mutate(p.id)}
                disabled={test.isPending && test.variables === p.id}
              >
                {test.isPending && test.variables === p.id ? <Loader2Icon className="animate-spin" /> : <PlugZapIcon />}
                {t("settings.providers.test")}
              </Button>
              <Button size="icon-sm" variant="outline" aria-label={t("settings.providers.edit")} onClick={() => openDialog(p)}>
                <PencilIcon />
              </Button>
              <Button
                size="icon-sm"
                variant="outline"
                aria-label={t("settings.providers.delete")}
                onClick={() => confirm(t("settings.providers.deleteConfirm", { name: p.name })) && remove.mutate(p.id)}
              >
                <Trash2Icon />
              </Button>
            </div>
          </div>
        ))}
      </CardContent>
      {dialogOpen && <ProviderDialog key={editing?.id ?? "new"} provider={editing} open={dialogOpen} onOpenChange={setDialogOpen} />}
    </Card>
  )
}
