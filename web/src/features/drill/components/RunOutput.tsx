import { CheckCircle2Icon, ClockIcon, XCircleIcon } from "lucide-react"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import type { RunResult } from "../types/drill"

export function RunOutput({ result }: { result: RunResult }) {
  const { t } = useI18n()
  const status = result.passed
    ? t("drills.result.passed")
    : result.timedOut
      ? t("drills.result.timedOut")
      : result.buildError
        ? t("drills.result.buildFailed")
        : t("drills.result.failed")
  const Icon = result.passed ? CheckCircle2Icon : result.timedOut ? ClockIcon : XCircleIcon

  return (
    <div className="flex flex-col gap-2">
      <div className={cn("flex items-center gap-2 text-sm font-medium", result.passed ? "text-success" : "text-destructive")}>
        <Icon className="size-4" />
        {status}
        {result.durationMs > 0 && <span className="font-normal text-muted-foreground">· {(result.durationMs / 1000).toFixed(1)}s</span>}
      </div>
      <pre
        data-glass
        className={cn(
          "max-h-80 overflow-auto rounded-lg border bg-muted/40 p-3 font-mono text-xs whitespace-pre-wrap",
          result.passed ? "border-success/40" : "border-destructive/40",
        )}
      >
        {result.output.trim() || t("drills.result.noOutput")}
      </pre>
    </div>
  )
}
