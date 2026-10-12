import type { LucideIcon } from "lucide-react"
import { Card, CardContent } from "@/components/ui/card"
import { IconChip } from "@/components/IconChip"
import type { Tone } from "@/lib/tones"

type Props = {
  icon: LucideIcon
  tone: Tone
  label: string
  value: string | number
  hint?: string
}

export function StatTile({ icon, tone, label, value, hint }: Props) {
  return (
    <Card size="sm">
      <CardContent className="flex items-start gap-3">
        <IconChip icon={icon} tone={tone} className="size-9" />
        <div className="min-w-0">
          <div className="truncate text-xs text-muted-foreground">{label}</div>
          <div className="flex items-baseline gap-1.5">
            <span className="text-2xl font-semibold tabular-nums">{value}</span>
            {hint && <span className="truncate text-xs text-muted-foreground">{hint}</span>}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
