import type { ComponentProps } from "react"
import { useQuery } from "@tanstack/react-query"
import { Link } from "react-router"
import { PlayCircleIcon } from "lucide-react"
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { IconChip } from "@/components/IconChip"
import { QuizList, quizApi } from "@/features/quiz"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"

const filter = { status: "open" as const, limit: 3 }

export function UnfinishedQuizzes(props: ComponentProps<"div">) {
  const { t } = useI18n()
  const open = useQuery({ queryKey: ["quizzes", filter], queryFn: () => quizApi.list(filter) })
  if (!open.data?.length) return null

  return (
    <Card {...props} className={cn("border-l-4 border-l-hue-orange", props.className)}>
      <CardHeader className="flex flex-row items-center gap-3">
        <IconChip icon={PlayCircleIcon} tone="orange" />
        <CardTitle>{t("dashboard.unfinished.title")}</CardTitle>
        <CardAction>
          <Link to="/quizzes" className="text-sm text-primary hover:underline">
            {t("dashboard.unfinished.all")}
          </Link>
        </CardAction>
      </CardHeader>
      <CardContent>
        <QuizList filter={filter} />
      </CardContent>
    </Card>
  )
}
