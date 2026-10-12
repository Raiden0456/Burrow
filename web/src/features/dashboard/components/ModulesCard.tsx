import { Link } from "react-router"
import { LayersIcon } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { IconChip } from "@/components/IconChip"
import { MasteryBar, type TopicStat } from "@/features/progress"
import type { Module } from "@/features/topics"
import { useI18n } from "@/i18n"

export function ModulesCard({ modules, stats }: { modules: Module[]; stats: TopicStat[] }) {
  const { t } = useI18n()
  const mastery = new Map(stats.map((s) => [s.slug, s.mastery]))

  return (
    <Card>
      <CardHeader className="flex flex-row items-center gap-3">
        <IconChip icon={LayersIcon} tone="orange" />
        <CardTitle>{t("dashboard.modules.title")}</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-x-8 gap-y-3 md:grid-cols-2">
        {modules.map((m, i) => {
          const value = m.topics.reduce((sum, topic) => sum + (mastery.get(topic.slug) ?? 0), 0) / m.topics.length
          return (
            <Link key={m.slug} to="/topics" className="flex flex-col gap-1.5 rounded-md py-1 hover:opacity-80">
              <span className="flex items-center gap-2 text-sm">
                <span className="flex size-5 items-center justify-center rounded-md bg-primary/12 text-[11px] font-semibold text-primary">{i + 1}</span>
                {m.title}
              </span>
              <MasteryBar value={value} />
            </Link>
          )
        })}
      </CardContent>
    </Card>
  )
}
