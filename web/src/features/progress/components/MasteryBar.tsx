import { Progress } from "@/components/ui/progress"
import { cn } from "@/lib/utils"

export function MasteryBar({ value, className }: { value: number; className?: string }) {
  const pct = Math.round(value * 100)
  return (
    <div className={cn("flex items-center gap-2", className)}>
      <Progress value={pct} className="h-2" />
      <span className="w-9 shrink-0 text-right text-xs tabular-nums text-muted-foreground">{pct}%</span>
    </div>
  )
}
