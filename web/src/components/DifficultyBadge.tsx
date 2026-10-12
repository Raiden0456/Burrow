import { Badge } from "@/components/ui/badge"
import { useI18n, type MessageKey } from "@/i18n"
import { cn } from "@/lib/utils"
import { toneChip, type Tone } from "@/lib/tones"

const levels: Record<number, { key: MessageKey; tone: Tone }> = {
  1: { key: "common.difficulty.1", tone: "green" },
  2: { key: "common.difficulty.2", tone: "yellow" },
  3: { key: "common.difficulty.3", tone: "red" },
}

export function DifficultyBadge({ level }: { level: number }) {
  const { t } = useI18n()
  const { key, tone } = levels[level] ?? levels[2]
  return <Badge className={cn("border-transparent", toneChip[tone])}>{t(key)}</Badge>
}
