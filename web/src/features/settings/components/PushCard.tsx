import { useEffect, useState, type ComponentProps } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { BellIcon, BellOffIcon, BellRingIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { IconChip } from "@/components/IconChip"
import { useI18n } from "@/i18n"
import { settingsApi } from "../api"
import { currentSubscription, pushSupported, subscribe } from "../utils/push"

export function PushCard(props: ComponentProps<"div">) {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const info = useQuery({ queryKey: ["push"], queryFn: settingsApi.pushInfo })
  const [endpoint, setEndpoint] = useState<string | null>(null)

  useEffect(() => {
    currentSubscription().then((s) => setEndpoint(s?.endpoint ?? null))
  }, [])

  const enable = useMutation({
    mutationFn: async () => {
      const sub = await subscribe(info.data?.publicKey ?? "")
      await settingsApi.subscribe(sub.toJSON())
      return sub.endpoint
    },
    onSuccess: (ep) => {
      setEndpoint(ep)
      queryClient.invalidateQueries({ queryKey: ["push"] })
      toast.success(t("settings.push.enabled"))
    },
  })
  const disable = useMutation({
    mutationFn: async () => {
      const sub = await currentSubscription()
      if (!sub) return
      await settingsApi.unsubscribe(sub.endpoint)
      await sub.unsubscribe()
    },
    onSuccess: () => {
      setEndpoint(null)
      queryClient.invalidateQueries({ queryKey: ["push"] })
    },
  })
  const test = useMutation({
    mutationFn: settingsApi.testPush,
    onSuccess: ({ sent }) => (sent > 0 ? toast.success(t("settings.push.sent", { count: sent })) : toast.warning(t("settings.push.none"))),
  })

  return (
    <Card {...props}>
      <CardHeader className="flex flex-row items-start gap-3">
        <IconChip icon={BellRingIcon} tone="yellow" />
        <div className="flex flex-col gap-1">
          <CardTitle>{t("settings.push.title")}</CardTitle>
          <CardDescription>
            {t("settings.push.description")} {info.data && t("settings.push.devices", { count: info.data.subscriptions })}
            {!pushSupported() && ` ${t("settings.push.unsupported")}`}
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent className="flex flex-wrap gap-2 max-md:flex-col">
        {endpoint ? (
          <Button variant="outline" onClick={() => disable.mutate()} disabled={disable.isPending}>
            <BellOffIcon />
            {t("settings.push.disable")}
          </Button>
        ) : (
          <Button onClick={() => enable.mutate()} disabled={enable.isPending || !info.data}>
            <BellIcon />
            {t("settings.push.enable")}
          </Button>
        )}
        <Button variant="ghost" onClick={() => test.mutate()} disabled={test.isPending}>
          {t("settings.push.test")}
        </Button>
      </CardContent>
    </Card>
  )
}
