import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "react-router"
import { HistoryIcon, ListChecksIcon, Loader2Icon, MessageSquareIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { ErrorAlert } from "@/components/ErrorAlert"
import { IconChip } from "@/components/IconChip"
import { useStartChat } from "@/features/chat"
import { DrillList } from "@/features/drill"
import { useJob, useRunningJob } from "@/features/jobs"
import { QuizList, quizApi } from "@/features/quiz"
import { useI18n } from "@/i18n"
import type { Topic } from "../types/curriculum"

export function PracticePanel({ topic }: { topic: Topic }) {
  const { t } = useI18n()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [size, setSize] = useState("5")
  const [jobId, setJobId] = useState<number | null>(null)
  const [failure, setFailure] = useState<string | null>(null)
  const startChat = useStartChat()
  const background = useRunningJob("quiz", topic.slug)

  const job = useJob(jobId ?? background?.id ?? null, (finished) => {
    queryClient.invalidateQueries({ queryKey: ["quizzes"] })
    queryClient.invalidateQueries({ queryKey: ["jobs"] })
    if (finished.status === "done" && finished.resultId && jobId === finished.id) navigate(`/quiz/${finished.resultId}`)
    if (finished.status === "failed" && jobId === finished.id) setFailure(finished.error)
    setJobId(null)
  })

  const createQuiz = useMutation({
    mutationFn: () => quizApi.createForTopic(topic.slug, Number(size)),
    onSuccess: (started) => {
      setFailure(null)
      setJobId(started.id)
      queryClient.invalidateQueries({ queryKey: ["jobs"] })
    },
  })
  const generating = createQuiz.isPending || job?.status === "running"

  return (
    <div className="flex flex-col gap-6">
      <Card className="border-l-4 border-l-hue-blue">
        <CardHeader className="flex flex-row items-start gap-3">
          <IconChip icon={ListChecksIcon} tone="blue" />
          <div className="flex flex-col gap-1">
            <CardTitle>{t("topics.quiz.title")}</CardTitle>
            <CardDescription>{t("topics.quiz.description")}</CardDescription>
          </div>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <div className="flex flex-wrap items-center gap-2 max-md:flex-col max-md:items-stretch">
            <Select value={size} onValueChange={setSize} disabled={generating}>
              <SelectTrigger className="w-40 max-md:w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {["3", "5", "8", "10"].map((n) => (
                  <SelectItem key={n} value={n}>
                    {t("topics.quiz.size", { count: Number(n) })}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button onClick={() => createQuiz.mutate()} disabled={generating}>
              {generating && <Loader2Icon className="animate-spin" />}
              {generating ? t("topics.quiz.generating") : t("topics.quiz.start")}
            </Button>
            <Button
              variant="ghost"
              disabled={startChat.isPending}
              onClick={() =>
                startChat.mutate({
                  title: topic.title,
                  context: `The learner is studying the topic "${topic.title}".\nObjectives:\n${topic.objectives.map((o) => `- ${o}`).join("\n")}`,
                })
              }
            >
              <MessageSquareIcon />
              {t("topics.askAi")}
            </Button>
          </div>
          {generating && <p className="text-xs text-muted-foreground">{t("jobs.leaveHint")}</p>}
          {failure && <ErrorAlert error={failure} />}
        </CardContent>
      </Card>

      <div className="flex flex-col gap-3">
        <h3 className="flex items-center gap-3 font-medium">
          <IconChip icon={HistoryIcon} tone="cyan" />
          {t("topics.quiz.history")}
        </h3>
        <QuizList filter={{ topic: topic.slug, limit: 10 }} empty={t("topics.quiz.historyEmpty")} />
      </div>

      <DrillList slug={topic.slug} />
    </div>
  )
}
