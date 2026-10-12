import type { LucideIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { toneChip, type Tone } from "@/lib/tones"

export function IconChip({ icon: Icon, tone, className }: { icon: LucideIcon; tone: Tone; className?: string }) {
  return (
    <span className={cn("inline-flex size-8 shrink-0 items-center justify-center rounded-lg", toneChip[tone], className)}>
      <Icon className="size-4" />
    </span>
  )
}
