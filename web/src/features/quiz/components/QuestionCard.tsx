import { useMemo, useState } from "react"
import { Loader2Icon } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/card"
import { Textarea } from "@/components/ui/textarea"
import { DifficultyBadge } from "@/components/DifficultyBadge"
import { Markdown } from "@/components/Markdown"
import { useI18n } from "@/i18n"
import { useDraftSaver } from "../hooks/useDraftSaver"
import type { Question, QuizResponse } from "../types/quiz"
import { AnswerFeedback } from "./AnswerFeedback"
import { ChoiceOptions } from "./ChoiceOptions"
import { HintsPanel } from "./HintsPanel"

type Props = {
  quizId: number
  question: Question
  topicTitle?: string
  submitting: boolean
  onSubmit: (response: QuizResponse) => void
  onNext: () => void
  isLast: boolean
}

export function QuestionCard({ quizId, question, topicTitle, submitting, onSubmit, onNext, isLast }: Props) {
  const { t } = useI18n()
  const [selected, setSelected] = useState<number[]>(question.draft?.selected ?? [])
  const [text, setText] = useState(question.draft?.text ?? "")
  const answered = !!question.result

  const response = useMemo<QuizResponse>(() => (question.kind === "open" ? { text } : { selected }), [question.kind, text, selected])
  useDraftSaver(quizId, question.id, response, !answered && !submitting)

  const toggle = (i: number) =>
    setSelected((prev) => (question.kind === "multi_choice" ? (prev.includes(i) ? prev.filter((x) => x !== i) : [...prev, i]) : [i]))

  const canSubmit = question.kind === "open" ? text.trim().length > 0 : selected.length > 0
  const submit = () => canSubmit && onSubmit(response)

  return (
    <Card>
      <CardHeader className="flex flex-wrap gap-2">
        {topicTitle && <Badge variant="outline">{topicTitle}</Badge>}
        <Badge variant="secondary">{t(`quiz.kind.${question.kind}`)}</Badge>
        <DifficultyBadge level={question.difficulty} />
      </CardHeader>
      <CardContent className="flex flex-col gap-5">
        <Markdown>{question.prompt}</Markdown>

        {question.kind === "open" ? (
          answered ? (
            <div className="rounded-lg border p-3 text-sm whitespace-pre-wrap">{question.result?.response.text}</div>
          ) : (
            <Textarea
              rows={6}
              placeholder={t("quiz.placeholder")}
              value={text}
              onChange={(e) => setText(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && (e.metaKey || e.ctrlKey) && submit()}
            />
          )
        ) : (
          <ChoiceOptions question={question} selected={selected} onToggle={toggle} />
        )}

        {answered && <AnswerFeedback question={question} />}
        <HintsPanel quizId={quizId} question={question} />
      </CardContent>
      <CardFooter className="flex justify-end gap-2 border-t pt-4">
        {answered ? (
          <Button onClick={onNext}>{isLast ? t("quiz.seeResults") : t("quiz.next")}</Button>
        ) : (
          <Button onClick={submit} disabled={!canSubmit || submitting}>
            {submitting && <Loader2Icon className="animate-spin" />}
            {submitting && question.kind === "open" ? t("quiz.grading") : t("quiz.submit")}
          </Button>
        )}
      </CardFooter>
    </Card>
  )
}
