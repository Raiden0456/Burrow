import type { ComponentProps } from "react"
import { FlameIcon } from "lucide-react"
import { useI18n, type MessageKey } from "@/i18n"
import { cn } from "@/lib/utils"

function greetingKey(hour: number): MessageKey {
  if (hour < 5) return "dashboard.greeting.night"
  if (hour < 12) return "dashboard.greeting.morning"
  if (hour < 18) return "dashboard.greeting.afternoon"
  return "dashboard.greeting.evening"
}

export function HeroCard({ streak, activeToday, className, ...props }: { streak: number; activeToday: boolean } & ComponentProps<"div">) {
  const { t } = useI18n()

  return (
    <div {...props} className={cn("flex flex-wrap items-center justify-between gap-4", className)}>
      <div className="text-legible">
        <h1 className="text-2xl font-semibold tracking-tight md:text-3xl">{t(greetingKey(new Date().getHours()))}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{activeToday ? t("dashboard.subtitle.active") : t("dashboard.subtitle.idle")}</p>
      </div>
      {streak > 0 && (
        <div className="flex items-center gap-2 rounded-full bg-hue-orange/15 px-3 py-1.5 text-sm font-medium text-hue-orange">
          <FlameIcon className="size-4" />
          {t("dashboard.streakBadge", { count: streak })}
        </div>
      )}
    </div>
  )
}
