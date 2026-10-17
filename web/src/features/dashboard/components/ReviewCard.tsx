import type { ComponentProps } from "react"
import { useMutation } from "@tanstack/react-query"
import { useNavigate } from "react-router"
import { RepeatIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { IconChip } from "@/components/IconChip"
import { quizApi } from "@/features/quiz"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"

export function ReviewCard({ due, ...props }: { due: number } & ComponentProps<"div">) {
  const { t } = useI18n()
  const navigate = useNavigate()
  const size = Math.min(due, 15)
  const create = useMutation({
    mutationFn: () => quizApi.createReview(size),
    onSuccess: ({ id }) => navigate(`/quiz/${id}`),
  })

  return (
    <Card {...props} className={cn("border-l-4 border-l-hue-cyan", props.className)}>
      <CardHeader className="flex flex-row items-start gap-3">
        <IconChip icon={RepeatIcon} tone="cyan" />
        <div className="flex flex-col gap-1">
          <CardTitle>{t("dashboard.review.title")}</CardTitle>
          <CardDescription>{due > 0 ? t("dashboard.review.due", { count: due }) : t("dashboard.review.none")}</CardDescription>
        </div>
      </CardHeader>
      <CardContent>
        <Button variant="outline" disabled={due === 0 || create.isPending} onClick={() => create.mutate()} className="max-md:w-full">
          {t("dashboard.review.button", { count: due > 0 ? size : "" })}
        </Button>
      </CardContent>
    </Card>
  )
}
