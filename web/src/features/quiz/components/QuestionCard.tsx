import { useState } from "react"
import { Loader2Icon, MessageSquareIcon } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { DifficultyBadge } from "@/components/DifficultyBadge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/card"
import { Textarea } from "@/components/ui/textarea"
import { Markdown } from "@/components/Markdown"
import { useStartChat } from "@/features/chat"
import { useI18n } from "@/i18n"
import { questionContext } from "../utils/chatContext"
import type { Question, QuizResponse } from "../types/quiz"
import { AnswerFeedback } from "./AnswerFeedback"
import { ChoiceOptions } from "./ChoiceOptions"


type Props = {
  question: Question
  topicTitle?: string
  submitting: boolean
  onSubmit: (response: QuizResponse) => void
  onNext: () => void
  isLast: boolean
}

export function QuestionCard({ question, topicTitle, submitting, onSubmit, onNext, isLast }: Props) {
  const { t } = useI18n()
  const [selected, setSelected] = useState<number[]>([])
  const [text, setText] = useState("")
  const startChat = useStartChat()
  const answered = !!question.result

  const toggle = (i: number) =>
    setSelected((prev) => (question.kind === "multi_choice" ? (prev.includes(i) ? prev.filter((x) => x !== i) : [...prev, i]) : [i]))

  const canSubmit = question.kind === "open" ? text.trim().length > 0 : selected.length > 0
  const submit = () => canSubmit && onSubmit(question.kind === "open" ? { text } : { selected })

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
      </CardContent>
      <CardFooter className="flex flex-wrap justify-between gap-2 border-t pt-4">
        <Button
          variant="ghost"
          disabled={startChat.isPending}
          onClick={() =>
            startChat.mutate({
              title: question.prompt.slice(0, 60),
              context: questionContext(question),
              message: answered ? t("quiz.explainMessage") : t("quiz.hintMessage"),
            })
          }
        >
          <MessageSquareIcon />
          {answered ? t("quiz.explainMore") : t("quiz.hint")}
        </Button>
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
