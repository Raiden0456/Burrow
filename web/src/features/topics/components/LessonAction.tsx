import { BookCheckIcon, Loader2Icon, SparklesIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { useI18n } from "@/i18n"
import { useGenerateLesson } from "../hooks/useGenerateLesson"

type Props = {
  slug: string
  hasLesson: boolean
  generating: boolean
}

export function LessonAction({ slug, hasLesson, generating }: Props) {
  const { t } = useI18n()
  const generate = useGenerateLesson()

  if (generating || generate.isPending) {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-hue-blue/12 px-2 py-0.5 text-[11px] font-medium text-hue-blue">
        <Loader2Icon className="size-3 animate-spin" />
        {t("topics.lesson.generating")}
      </span>
    )
  }
  if (hasLesson) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <span className="inline-flex items-center text-hue-green" aria-label={t("topics.lesson.ready")}>
            <BookCheckIcon className="size-4" />
          </span>
        </TooltipTrigger>
        <TooltipContent>{t("topics.lesson.ready")}</TooltipContent>
      </Tooltip>
    )
  }
  return (
    <Button
      size="xs"
      variant="ghost"
      className="text-muted-foreground hover:text-hue-blue"
      onClick={(e) => {
        e.preventDefault()
        e.stopPropagation()
        generate.mutate(slug)
      }}
    >
      <SparklesIcon />
      {t("topics.lesson.generateShort")}
    </Button>
  )
}
