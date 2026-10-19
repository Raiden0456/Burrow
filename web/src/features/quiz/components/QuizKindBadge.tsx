import { Badge } from "@/components/ui/badge"
import { useI18n, type MessageKey } from "@/i18n"
import { cn } from "@/lib/utils"
import { toneChip, type Tone } from "@/lib/tones"
import type { QuizKind } from "../types/quiz"

const kinds: Record<QuizKind, { label: MessageKey; tone: Tone }> = {
  topic: { label: "quizKind.topic", tone: "blue" },
  review: { label: "quizKind.review", tone: "cyan" },
  daily: { label: "quizKind.daily", tone: "purple" },
  checkpoint: { label: "quizKind.checkpoint", tone: "green" },
}

export function QuizKindBadge({ kind }: { kind: QuizKind }) {
  const { t } = useI18n()
  return <Badge className={cn("border-transparent", toneChip[kinds[kind].tone])}>{t(kinds[kind].label)}</Badge>
}
