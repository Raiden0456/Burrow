import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { WaypointsIcon } from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { IconChip } from "@/components/IconChip"
import { useI18n, type MessageKey } from "@/i18n"
import { settingsApi } from "../api"
import type { Routing } from "../types/settings"

const tasks: { key: keyof Routing; label: MessageKey; hint: MessageKey }[] = [
  { key: "chat", label: "settings.routing.chat", hint: "settings.routing.chatHint" },
  { key: "generate", label: "settings.routing.generate", hint: "settings.routing.generateHint" },
  { key: "grade", label: "settings.routing.grade", hint: "settings.routing.gradeHint" },
]

export function RoutingCard() {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const providers = useQuery({ queryKey: ["providers"], queryFn: settingsApi.providers })
  const routing = useQuery({ queryKey: ["routing"], queryFn: settingsApi.routing })
  const save = useMutation({
    mutationFn: settingsApi.saveRouting,
    onSuccess: (r) => queryClient.setQueryData(["routing"], r),
  })

  if (!providers.data?.length || !routing.data) return null
  const current = routing.data

  return (
    <Card>
      <CardHeader className="flex flex-row items-start gap-3">
        <IconChip icon={WaypointsIcon} tone="cyan" />
        <div className="flex flex-col gap-1">
          <CardTitle>{t("settings.routing.title")}</CardTitle>
          <CardDescription>{t("settings.routing.description")}</CardDescription>
        </div>
      </CardHeader>
      <CardContent className="grid gap-4 sm:grid-cols-3">
        {tasks.map(({ key, label, hint }) => (
          <div key={key} className="flex flex-col gap-1.5">
            <Label>{t(label)}</Label>
            <Select value={String(current[key])} onValueChange={(v) => save.mutate({ ...current, [key]: Number(v) })}>
              <SelectTrigger className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="0">{t("settings.routing.first")}</SelectItem>
                {providers.data.map((p) => (
                  <SelectItem key={p.id} value={String(p.id)}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">{t(hint)}</p>
          </div>
        ))}
      </CardContent>
    </Card>
  )
}
