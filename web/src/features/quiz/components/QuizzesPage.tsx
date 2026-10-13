import { PageHeader } from "@/components/PageHeader"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useI18n } from "@/i18n"
import { QuizList } from "./QuizList"

export function QuizzesPage() {
  const { t } = useI18n()
  return (
    <>
      <PageHeader title={t("quizzes.title")} description={t("quizzes.description")} />
      <Tabs defaultValue="open">
        <TabsList>
          <TabsTrigger value="open">{t("quizzes.open")}</TabsTrigger>
          <TabsTrigger value="done">{t("quizzes.done")}</TabsTrigger>
        </TabsList>
        <TabsContent value="open" className="pt-4">
          <QuizList filter={{ status: "open", limit: 100 }} empty={t("quizzes.emptyOpen")} />
        </TabsContent>
        <TabsContent value="done" className="pt-4">
          <QuizList filter={{ status: "done", limit: 100 }} empty={t("quizzes.emptyDone")} />
        </TabsContent>
      </Tabs>
    </>
  )
}
