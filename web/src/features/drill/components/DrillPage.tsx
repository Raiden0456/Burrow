import { useCallback, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useParams } from "react-router"
import { BugIcon, CheckIcon, ChevronLeftIcon, CloudOffIcon, EllipsisIcon, EyeIcon, LifeBuoyIcon, Loader2Icon, PlayIcon, RotateCcwIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { ActionBar } from "@/components/ActionBar"
import { useIsMobile } from "@/hooks/useIsMobile"
import { cn } from "@/lib/utils"
import { DifficultyBadge } from "@/components/DifficultyBadge"
import { CodeBlock } from "@/components/CodeBlock"
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

type Pane = "task" | "code" | "result"

const helpTone = "border-hue-yellow/40 text-hue-yellow hover:bg-hue-yellow/10 hover:text-hue-yellow"

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
    drill.lastAttempt
      ? { passed: drill.lastAttempt.passed, buildError: false, timedOut: false, output: drill.lastAttempt.output, durationMs: 0 }
      : drill.initialOutput
        ? { passed: false, buildError: false, timedOut: false, output: drill.initialOutput, durationMs: 0 }
        : null,
  )
  const bugHunt = drill.mode === "bug_hunt"
  const solved = drill.solved || result?.passed === true
  const solutionLocked = bugHunt && !solved
  const solutionLabel = t(bugHunt ? "drills.bugHunt.debrief" : "drills.solution")
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

  const isMobile = useIsMobile()
  const [pane, setPane] = useState<Pane>("code")

  const runNow = useCallback(() => {
    if (run.isPending) return
    run.mutate(code)
    if (isMobile) setPane("result")
  }, [run, code, isMobile])

  const askHelp = () => {
    help.ask(code, result?.output ?? "")
    if (isMobile) setPane("result")
  }

  const panes: { value: Pane; label: string }[] = [
    { value: "task", label: t("drills.pane.task") },
    { value: "code", label: t("drills.pane.code") },
    { value: "result", label: t("drills.pane.result") },
  ]
  const hidden = (p: Pane) => pane !== p && "max-lg:hidden"

  return (
    <div className="max-lg:pb-28">
      <div className="mb-3 flex items-center gap-2 max-md:-ml-2">
        <Button asChild variant="ghost" size="icon" className="md:hidden" aria-label={t("common.backToTopic")}>
          <Link to={`/topics/${drill.topicSlug}`}>
            <ChevronLeftIcon />
          </Link>
        </Button>
        <Link to={`/topics/${drill.topicSlug}`} className="hidden items-center gap-1 text-sm text-muted-foreground hover:text-foreground md:inline-flex wallpaper:text-legible">
          <ChevronLeftIcon className="size-4" />
          {t("common.backToTopic")}
        </Link>
        <div className="ml-auto md:hidden">
          <SaveIndicator state={saveState} />
        </div>
      </div>
      <div className="mb-4 flex flex-wrap items-center gap-2 wallpaper:text-legible">
        <h1 className="text-xl font-semibold tracking-tight md:text-2xl">{drill.title}</h1>
        {bugHunt && (
          <span className="flex items-center gap-1 rounded-full bg-hue-yellow/10 px-2 py-0.5 text-xs font-medium text-hue-yellow">
            <BugIcon className="size-3" />
            {t("drills.bugHunt.title")}
          </span>
        )}
        <DifficultyBadge level={drill.difficulty} />
        {solved && <span className="rounded-full bg-success/15 px-2 py-0.5 text-xs font-medium text-success">{t("drills.solved")}</span>}
      </div>

      <div data-glass className="sticky top-0 z-30 -mx-4 mb-4 bg-background/95 px-4 py-2 backdrop-blur lg:hidden">
        <div className="grid grid-cols-3 gap-1 rounded-xl bg-muted p-1">
          {panes.map((p) => (
            <button
              key={p.value}
              type="button"
              onClick={() => setPane(p.value)}
              className={cn(
                "relative h-10 rounded-lg text-sm font-medium text-muted-foreground transition-colors",
                pane === p.value && "bg-background text-foreground shadow-sm",
              )}
            >
              {p.label}
              {p.value === "result" && result && (
                <span className={cn("absolute top-2 right-3 size-2 rounded-full", result.passed ? "bg-success-solid" : "bg-destructive")} />
              )}
            </button>
          ))}
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div data-glass className={cn("min-w-0 wallpaper:rounded-xl wallpaper:p-5", hidden("task"))}>
          <Markdown className="prose-sm">{drill.description}</Markdown>
          {bugHunt && (
            <div className="mt-4 flex flex-col gap-3">
              <p className="text-sm text-muted-foreground">{t(solutionLocked ? "drills.bugHunt.locked" : "drills.bugHunt.unlocked")}</p>
              {!solutionLocked && (
                <Button variant="outline" className="self-start" onClick={() => setShowSolution(true)}>
                  <EyeIcon />
                  {solutionLabel}
                </Button>
              )}
              <details className="min-w-0 rounded-lg border p-3">
                <summary className="cursor-pointer text-sm font-medium">{t("drills.bugHunt.tests")}</summary>
                <div className="mt-3"><CodeBlock code={drill.testCode ?? ""} lang="go" /></div>
              </details>
            </div>
          )}
        </div>

        <div className="flex min-w-0 flex-col gap-3">
          <div className={cn("flex flex-col gap-3", hidden("code"))}>
            <CodeEditor value={code} onChange={setCode} onRun={runNow} minHeight={isMobile ? "60svh" : "320px"} />
            <div className="flex flex-wrap gap-2 max-lg:hidden">
              <Button onClick={runNow} disabled={run.isPending}>
                {run.isPending ? <Loader2Icon className="animate-spin" /> : <PlayIcon />}
                {t("drills.run")}
              </Button>
              <Button variant="outline" onClick={() => setCode(drill.starterCode)}>
                <RotateCcwIcon />
                {t("drills.reset")}
              </Button>
              <Button variant="outline" onClick={askHelp} disabled={help.streaming} className={helpTone}>
                {help.streaming ? <Loader2Icon className="animate-spin" /> : <LifeBuoyIcon />}
                {t("drills.help")}
              </Button>
              <Button variant="ghost" onClick={() => setShowSolution(true)} disabled={solutionLocked}>
                <EyeIcon />
                {solutionLabel}
              </Button>
            </div>
            <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground max-lg:hidden">
              <span>{t("drills.editorHint")}</span>
              <SaveIndicator state={saveState} />
            </div>
          </div>
          <div className={cn("flex flex-col gap-3", hidden("result"))}>
            {run.error && <ErrorAlert error={run.error} />}
            {!result && !help.streaming && drill.hints.length === 0 && (
              <p className="text-sm text-muted-foreground lg:hidden">{t("drills.resultEmpty")}</p>
            )}
            {run.isPending && (
              <div className="flex items-center gap-2 text-sm text-muted-foreground lg:hidden">
                <Loader2Icon className="size-4 animate-spin" />
                {t("drills.running")}
              </div>
            )}
            {result && <RunOutput result={result} />}
            {bugHunt && solved && (
              <Button variant="outline" onClick={() => setShowSolution(true)}>
                <EyeIcon />
                {solutionLabel}
              </Button>
            )}
            <HelpPanel hints={drill.hints} streamingText={help.streaming ? help.text : null} error={help.error} />
          </div>
        </div>
      </div>

      <ActionBar className="lg:hidden md:flex">
        <Button size="lg" className="flex-1" onClick={runNow} disabled={run.isPending}>
          {run.isPending ? <Loader2Icon className="animate-spin" /> : <PlayIcon />}
          {t("drills.run")}
        </Button>
        <Button size="icon-lg" variant="outline" onClick={askHelp} disabled={help.streaming} className={helpTone} aria-label={t("drills.help")}>
          {help.streaming ? <Loader2Icon className="animate-spin" /> : <LifeBuoyIcon />}
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button size="icon-lg" variant="outline" aria-label={t("drills.more")}>
              <EllipsisIcon />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" side="top" className="min-w-48">
            <DropdownMenuItem className="py-3" onSelect={() => setCode(drill.starterCode)}>
              <RotateCcwIcon />
              {t("drills.reset")}
            </DropdownMenuItem>
            <DropdownMenuItem className="py-3" onSelect={() => setShowSolution(true)} disabled={solutionLocked}>
              <EyeIcon />
              {solutionLabel}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </ActionBar>

      <SolutionDialog drillId={drill.id} bugHunt={bugHunt} open={showSolution} onOpenChange={setShowSolution} />
    </div>
  )
}
