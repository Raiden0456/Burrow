import { useState, type FormEvent } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { GraduationCapIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import { Skeleton } from "@/components/ui/skeleton"
import { IconChip } from "@/components/IconChip"
import { useI18n } from "@/i18n"
import { settingsApi } from "../api"
import type { Preferences } from "../types/settings"

export function PreferencesCard() {
  const prefs = useQuery({ queryKey: ["preferences"], queryFn: settingsApi.preferences })
  if (!prefs.data) return <Skeleton className="h-64 w-full" />
  return <PreferencesForm initial={prefs.data} />
}

function PreferencesForm({ initial }: { initial: Preferences }) {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const [form, setForm] = useState(initial)
  const set = <K extends keyof Preferences>(key: K, value: Preferences[K]) => setForm((f) => ({ ...f, [key]: value }))
  const save = useMutation({
    mutationFn: settingsApi.savePreferences,
    onSuccess: (p) => {
      queryClient.setQueryData(["preferences"], p)
      queryClient.invalidateQueries({ queryKey: ["daily"] })
      toast.success(t("settings.prefs.saved"))
    },
  })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    save.mutate(form)
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-start gap-3">
        <IconChip icon={GraduationCapIcon} tone="green" />
        <div className="flex flex-col gap-1">
          <CardTitle>{t("settings.prefs.title")}</CardTitle>
          <CardDescription>{t("settings.prefs.description")}</CardDescription>
        </div>
      </CardHeader>
      <CardContent>
        <form onSubmit={submit} className="flex flex-col gap-5">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="language">{t("settings.prefs.language")}</Label>
            <Input id="language" value={form.language} onChange={(e) => set("language", e.target.value)} placeholder="Russian" />
            <p className="text-xs text-muted-foreground">{t("settings.prefs.languageHint")}</p>
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="profile">{t("settings.prefs.about")}</Label>
            <Textarea id="profile" rows={3} value={form.profile} onChange={(e) => set("profile", e.target.value)} placeholder={t("settings.prefs.aboutPlaceholder")} />
          </div>
          <div className="flex items-center justify-between gap-4 rounded-lg border p-3">
            <div>
              <Label htmlFor="daily">{t("settings.prefs.daily")}</Label>
              <p className="text-xs text-muted-foreground">{t("settings.prefs.dailyHint")}</p>
            </div>
            <Switch id="daily" checked={form.dailyEnabled} onCheckedChange={(v) => set("dailyEnabled", v)} />
          </div>
          {form.dailyEnabled && (
            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="dailyTime">{t("settings.prefs.time")}</Label>
                <Input id="dailyTime" type="time" value={form.dailyTime} onChange={(e) => set("dailyTime", e.target.value)} />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="dailySize">{t("settings.prefs.questions")}</Label>
                <Input id="dailySize" type="number" min={3} max={30} value={form.dailySize} onChange={(e) => set("dailySize", Number(e.target.value))} />
              </div>
            </div>
          )}
          <div>
            <Button type="submit" disabled={save.isPending} className="max-md:w-full">
              {t("common.save")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
