import { useCallback, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useParams } from "react-router"
import { CheckIcon, ChevronLeftIcon, CloudOffIcon, EyeIcon, LifeBuoyIcon, Loader2Icon, PlayIcon, RotateCcwIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { DifficultyBadge } from "@/components/DifficultyBadge"
import { Markdown } from "@/components/Markdown"
import { PageLoader } from "@/components/PageLoader"
import { ErrorAlert } from "@/components/ErrorAlert"
import { useI18n } from "@/i18n"
import { drillApi } from "../api"
import { useCodeAutosave } from "../hooks/useCodeAutosave"
import { useDrillHelp } from "../hooks/useDrillHelp"
import type { Drill, RunResult } from "../types/drill"
import { CodeEditor } from "./CodeEditor"
import { HelpPanel } from "./HelpPanel"
import { RunOutput } from "./RunOutput"
import { SolutionDialog } from "./SolutionDialog"

export function DrillPage() {
  const id = Number(useParams().id)
  const drill = useQuery({ queryKey: ["drill", id], queryFn: () => drillApi.get(id) })

  if (drill.isPending) return <PageLoader />
  if (drill.error) return <ErrorAlert error={drill.error} />
  return <DrillWorkspace key={id} drill={drill.data} />
}

function SaveIndicator({ state }: { state: ReturnType<typeof useCodeAutosave> }) {
  const { t } = useI18n()
  if (state === "saving") return <span className="flex items-center gap-1"><Loader2Icon className="size-3 animate-spin" />{t("drills.saving")}</span>
  if (state === "saved") return <span className="flex items-center gap-1 text-success"><CheckIcon className="size-3" />{t("drills.saved")}</span>
  if (state === "error") return <span className="flex items-center gap-1 text-destructive"><CloudOffIcon className="size-3" />{t("drills.saveFailed")}</span>
  return null
}

function DrillWorkspace({ drill }: { drill: Drill }) {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const [code, setCode] = useState(drill.draft || drill.lastAttempt?.code || drill.starterCode)
  const [result, setResult] = useState<RunResult | null>(
    drill.lastAttempt ? { passed: drill.lastAttempt.passed, buildError: false, timedOut: false, output: drill.lastAttempt.output, durationMs: 0 } : null,
  )
  const [showSolution, setShowSolution] = useState(false)
  const saveState = useCodeAutosave(drill.id, code)
  const help = useDrillHelp(drill.id)

  const run = useMutation({
    mutationFn: (src: string) => drillApi.run(drill.id, src),
    onSuccess: (res) => {
      setResult(res)
      if (res.passed) {
        queryClient.invalidateQueries({ queryKey: ["drill", drill.id] })
        queryClient.invalidateQueries({ queryKey: ["drills", drill.topicSlug] })
        queryClient.invalidateQueries({ queryKey: ["progress"] })
      }
    },
  })

  const runNow = useCallback(() => {
    if (!run.isPending) run.mutate(code)
  }, [run, code])

  return (
    <>
      <Link to={`/topics/${drill.topicSlug}`} className="mb-3 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
        <ChevronLeftIcon className="size-4" />
        {t("common.backToTopic")}
      </Link>
      <div className="mb-4 flex flex-wrap items-center gap-2">
        <h1 className="text-2xl font-semibold tracking-tight">{drill.title}</h1>
        <DifficultyBadge level={drill.difficulty} />
        {drill.solved && <span className="rounded-full bg-success/15 px-2 py-0.5 text-xs font-medium text-success">{t("drills.solved")}</span>}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Markdown className="prose-sm">{drill.description}</Markdown>

        <div className="flex min-w-0 flex-col gap-3">
          <CodeEditor value={code} onChange={setCode} onRun={runNow} />
          <div className="flex flex-wrap gap-2">
            <Button onClick={runNow} disabled={run.isPending}>
              {run.isPending ? <Loader2Icon className="animate-spin" /> : <PlayIcon />}
              {t("drills.run")}
            </Button>
            <Button variant="outline" onClick={() => setCode(drill.starterCode)}>
              <RotateCcwIcon />
              {t("drills.reset")}
            </Button>
            <Button
              variant="outline"
              onClick={() => help.ask(code, result?.output ?? "")}
              disabled={help.streaming}
              className="border-hue-yellow/40 text-hue-yellow hover:bg-hue-yellow/10 hover:text-hue-yellow"
            >
              {help.streaming ? <Loader2Icon className="animate-spin" /> : <LifeBuoyIcon />}
              {t("drills.help")}
            </Button>
            <Button variant="ghost" onClick={() => setShowSolution(true)}>
              <EyeIcon />
              {t("drills.solution")}
            </Button>
          </div>
          <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
            <span>{t("drills.editorHint")}</span>
            <SaveIndicator state={saveState} />
          </div>
          <HelpPanel hints={drill.hints} streamingText={help.streaming ? help.text : null} error={help.error} />
          {result && <RunOutput result={result} />}
        </div>
      </div>

      <SolutionDialog drillId={drill.id} open={showSolution} onOpenChange={setShowSolution} />
    </>
  )
}
