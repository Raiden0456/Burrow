import { Link } from "react-router"
import { ArrowRightIcon, CompassIcon } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { IconChip } from "@/components/IconChip"
import { MasteryBar, type TopicStat } from "@/features/progress"
import type { Module } from "@/features/topics"
import { useI18n } from "@/i18n"

export function ContinueCard({ modules, stats }: { modules: Module[]; stats: TopicStat[] }) {
  const { t } = useI18n()
  const all = modules.flatMap((m) => m.topics)
  const titles = new Map(all.map((topic) => [topic.slug, topic.title]))
  const recent = [...stats].sort((a, b) => b.lastActivity - a.lastActivity).slice(0, 3)
  const seen = new Set(stats.map((s) => s.slug))
  const next = all.find((topic) => !seen.has(topic.slug))

  return (
    <Card>
      <CardHeader className="flex flex-row items-center gap-3">
        <IconChip icon={CompassIcon} tone="purple" />
        <CardTitle>{t("dashboard.continue.title")}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-1">
        {recent.map((s) => (
          <Link key={s.slug} to={`/topics/${s.slug}`} className="-mx-2 flex flex-col gap-1.5 rounded-md px-2 py-2 hover:bg-muted/60">
            <span className="text-sm font-medium">{titles.get(s.slug) ?? s.slug}</span>
            <MasteryBar value={s.mastery} />
          </Link>
        ))}
        {next && (
          <Link
            to={`/topics/${next.slug}`}
            className="-mx-2 mt-1 flex items-center justify-between gap-2 rounded-md bg-hue-purple/10 px-3 py-2.5 text-sm transition-colors hover:bg-hue-purple/15"
          >
            <span>
              <span className="text-muted-foreground">{t("dashboard.continue.next")}: </span>
              <span className="font-medium">{next.title}</span>
            </span>
            <ArrowRightIcon className="size-4 text-hue-purple" />
          </Link>
        )}
      </CardContent>
    </Card>
  )
}
