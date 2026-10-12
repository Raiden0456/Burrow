import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import type { DayActivity } from "../types/progress"
import { heatClass, heatLegend } from "../utils/heat"

export function ActivityGrid({ days }: { days: DayActivity[] }) {
  const { t, locale } = useI18n()
  const format = new Intl.DateTimeFormat(locale, { day: "numeric", month: "short", weekday: "short" })

  return (
    <div className="flex flex-col gap-3">
      <div className="grid grid-cols-14 gap-1">
        {days.map((d) => (
          <Tooltip key={d.date}>
            <TooltipTrigger asChild>
              <div className={cn("aspect-square rounded-[4px] transition-transform hover:scale-110", heatClass(d.count))} />
            </TooltipTrigger>
            <TooltipContent>{t("dashboard.activity.tooltip", { date: format.format(new Date(`${d.date}T12:00:00`)), count: d.count })}</TooltipContent>
          </Tooltip>
        ))}
      </div>
      <div className="flex items-center justify-end gap-1 text-xs text-muted-foreground">
        {t("dashboard.activity.less")}
        {heatLegend.map((c) => (
          <span key={c} className={cn("size-3 rounded-[3px]", c)} />
        ))}
        {t("dashboard.activity.more")}
      </div>
    </div>
  )
}
