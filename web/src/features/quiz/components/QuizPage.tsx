import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useParams } from "react-router"
import { XIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { PageHeader } from "@/components/PageHeader"
import { PageLoader } from "@/components/PageLoader"
import { ErrorAlert } from "@/components/ErrorAlert"
import { useCurriculum } from "@/features/topics"
import { useI18n } from "@/i18n"
import { quizApi } from "../api"
import type { Quiz, QuizResponse } from "../types/quiz"
import { QuestionCard } from "./QuestionCard"
import { QuizStepper } from "./QuizStepper"
import { QuizSummary } from "./QuizSummary"


function firstUnanswered(quiz: Quiz) {
  const i = quiz.questions.findIndex((q) => !q.result)
  return i === -1 ? null : i
}

export function QuizPage() {
  const { t } = useI18n()
  const id = Number(useParams().id)
  const queryClient = useQueryClient()
  const { topics } = useCurriculum()
  const quiz = useQuery({ queryKey: ["quiz", id], queryFn: () => quizApi.get(id) })
  const [current, setCurrent] = useState<number | null | undefined>(undefined)

  const answer = useMutation({
    mutationFn: ({ questionId, response }: { questionId: number; response: QuizResponse }) => quizApi.answer(id, questionId, response),
    onSuccess: (outcome, { questionId }) => {
      queryClient.setQueryData<Quiz>(["quiz", id], (prev) =>
        prev && {
          ...prev,
          completedAt: outcome.quizCompleted ? Date.now() / 1000 : prev.completedAt,
          questions: prev.questions.map((q) => (q.id === questionId ? { ...q, result: outcome.result } : q)),
        },
      )
      queryClient.invalidateQueries({ queryKey: ["progress"] })
      queryClient.invalidateQueries({ queryKey: ["quizzes"] })
    },
  })

  if (quiz.isPending) return <PageLoader />
  if (quiz.error) return <ErrorAlert error={quiz.error} />

  const data = quiz.data
  const index = current === undefined ? firstUnanswered(data) : current
  const answeredCount = data.questions.filter((q) => q.result).length
  const topicTitle = data.topicSlug ? topics.get(data.topicSlug)?.title : undefined

  const next = () => {
    const after = data.questions.findIndex((q, i) => i > (index ?? -1) && !q.result)
    setCurrent(after === -1 ? firstUnanswered(data) : after)
  }

  const exit = data.topicSlug ? `/topics/${data.topicSlug}` : "/quizzes"

  return (
    <div className="max-md:pb-28">
      <div className="sticky top-0 z-30 -mx-4 mb-4 flex items-center gap-2 bg-background/95 px-2 py-2 backdrop-blur md:hidden">
        <Button asChild variant="ghost" size="icon" aria-label={t("quiz.exit")}>
          <Link to={exit}>
            <XIcon />
          </Link>
        </Button>
        <QuizStepper questions={data.questions} current={index ?? -1} onSelect={setCurrent} className="flex-1" />
        <span className="w-12 shrink-0 text-center text-sm font-medium tabular-nums text-muted-foreground">
          {answeredCount}/{data.questions.length}
        </span>
      </div>
      <div className="hidden md:block">
        <PageHeader
          title={topicTitle ?? t(`quiz.title.${data.kind}`)}
          description={t("quiz.progress", { answered: answeredCount, total: data.questions.length })}
        />
        <div className="mb-4">
          <QuizStepper questions={data.questions} current={index ?? -1} onSelect={setCurrent} />
        </div>
      </div>
      <h1 className="mb-3 text-lg font-semibold md:hidden">{topicTitle ?? t(`quiz.title.${data.kind}`)}</h1>
      {index === null ? (
        <QuizSummary quiz={data} onReview={setCurrent} />
      ) : (
        <QuestionCard
          key={data.questions[index].id}
          quizId={data.id}
          question={data.questions[index]}
          topicTitle={data.kind === "topic" ? undefined : topics.get(data.questions[index].topicSlug)?.title}
          submitting={answer.isPending}
          isLast={answeredCount === data.questions.length || data.questions.every((q, i) => i === index || q.result)}
          onSubmit={(response) => answer.mutate({ questionId: data.questions[index].id, response })}
          onNext={next}
        />
      )}
    </div>
  )
}
