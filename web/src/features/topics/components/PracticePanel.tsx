import { useState } from "react"
import { useMutation } from "@tanstack/react-query"
import { useNavigate } from "react-router"
import { ListChecksIcon, Loader2Icon, MessageSquareIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { IconChip } from "@/components/IconChip"
import { useStartChat } from "@/features/chat"
import { DrillList } from "@/features/drill"
import { quizApi } from "@/features/quiz"
import { useI18n } from "@/i18n"
import type { Topic } from "../types/curriculum"

export function PracticePanel({ topic }: { topic: Topic }) {
  const { t } = useI18n()
  const navigate = useNavigate()
  const [size, setSize] = useState("5")
  const startChat = useStartChat()
  const createQuiz = useMutation({
    mutationFn: () => quizApi.createForTopic(topic.slug, Number(size)),
    onSuccess: ({ id }) => navigate(`/quiz/${id}`),
  })

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
        <CardContent className="flex flex-wrap items-center gap-2">
          <Select value={size} onValueChange={setSize}>
            <SelectTrigger className="w-40">
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
          <Button onClick={() => createQuiz.mutate()} disabled={createQuiz.isPending}>
            {createQuiz.isPending && <Loader2Icon className="animate-spin" />}
            {createQuiz.isPending ? t("topics.quiz.generating") : t("topics.quiz.start")}
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
        </CardContent>
      </Card>

      <DrillList slug={topic.slug} />
    </div>
  )
}
