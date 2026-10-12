import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useNavigate } from "react-router"
import { CalendarCheckIcon, Loader2Icon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { IconChip } from "@/components/IconChip"
import { useI18n } from "@/i18n"
import { dashboardApi } from "../api"

export function DailyCard() {
  const { t, locale } = useI18n()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const daily = useQuery({ queryKey: ["daily"], queryFn: dashboardApi.daily })
  const create = useMutation({
    mutationFn: dashboardApi.createDaily,
    onSuccess: ({ id }) => {
      queryClient.invalidateQueries({ queryKey: ["daily"] })
      navigate(`/quiz/${id}`)
    },
  })

  const d = daily.data
  const time = d && new Date(d.scheduledAt * 1000).toLocaleTimeString(locale, { hour: "2-digit", minute: "2-digit" })
  const description = !d ? " " : d.quizId ? t("dashboard.daily.ready") : d.enabled ? t("dashboard.daily.scheduled", { time: time ?? "" }) : t("dashboard.daily.off")

  return (
    <Card className="border-l-4 border-l-hue-blue">
      <CardHeader className="flex flex-row items-start gap-3">
        <IconChip icon={CalendarCheckIcon} tone="blue" />
        <div className="flex flex-col gap-1">
          <CardTitle>{t("dashboard.daily.title")}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </div>
      </CardHeader>
      <CardContent>
        {daily.isPending ? (
          <Skeleton className="h-8 w-32" />
        ) : d?.quizId ? (
          <Button asChild>
            <Link to={`/quiz/${d.quizId}`}>{t("dashboard.daily.open")}</Link>
          </Button>
        ) : (
          <Button onClick={() => create.mutate()} disabled={create.isPending}>
            {create.isPending && <Loader2Icon className="animate-spin" />}
            {create.isPending ? t("dashboard.daily.building") : t("dashboard.daily.start")}
          </Button>
        )}
      </CardContent>
    </Card>
  )
}
