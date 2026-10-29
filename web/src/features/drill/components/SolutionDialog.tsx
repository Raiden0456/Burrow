import { useQuery } from "@tanstack/react-query"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { CodeBlock } from "@/components/CodeBlock"
import { ErrorAlert } from "@/components/ErrorAlert"
import { Markdown } from "@/components/Markdown"
import { PageLoader } from "@/components/PageLoader"
import { useI18n } from "@/i18n"
import { drillApi } from "../api"

export function SolutionDialog({ drillId, bugHunt, open, onOpenChange }: { drillId: number; bugHunt: boolean; open: boolean; onOpenChange: (open: boolean) => void }) {
  const { t } = useI18n()
  const solution = useQuery({ queryKey: ["drill-solution", drillId], queryFn: () => drillApi.solution(drillId), enabled: open })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90svh] overflow-y-auto sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{t(bugHunt ? "drills.bugHunt.debrief" : "drills.solutionTitle")}</DialogTitle>
        </DialogHeader>
        {solution.error ? <ErrorAlert error={solution.error} /> : solution.data ? (
          <Tabs defaultValue={bugHunt ? "explanation" : "solution"}>
            <TabsList>
              {solution.data.explanation && <TabsTrigger value="explanation">{t("drills.tab.explanation")}</TabsTrigger>}
              <TabsTrigger value="solution">{t("drills.tab.solution")}</TabsTrigger>
              <TabsTrigger value="tests">{t("drills.tab.tests")}</TabsTrigger>
            </TabsList>
            {solution.data.explanation && (
              <TabsContent value="explanation">
                <Markdown>{solution.data.explanation}</Markdown>
              </TabsContent>
            )}
            <TabsContent value="solution">
              <CodeBlock code={solution.data.solution} lang="go" />
            </TabsContent>
            <TabsContent value="tests">
              <CodeBlock code={solution.data.testCode} lang="go" />
            </TabsContent>
          </Tabs>
        ) : (
          <PageLoader />
        )}
      </DialogContent>
    </Dialog>
  )
}
