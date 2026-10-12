import { useQuery } from "@tanstack/react-query"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { CodeBlock } from "@/components/CodeBlock"
import { PageLoader } from "@/components/PageLoader"
import { useI18n } from "@/i18n"
import { drillApi } from "../api"

export function SolutionDialog({ drillId, open, onOpenChange }: { drillId: number; open: boolean; onOpenChange: (open: boolean) => void }) {
  const { t } = useI18n()
  const solution = useQuery({ queryKey: ["drill-solution", drillId], queryFn: () => drillApi.solution(drillId), enabled: open })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90svh] overflow-y-auto sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{t("drills.solutionTitle")}</DialogTitle>
        </DialogHeader>
        {solution.data ? (
          <Tabs defaultValue="solution">
            <TabsList>
              <TabsTrigger value="solution">{t("drills.tab.solution")}</TabsTrigger>
              <TabsTrigger value="tests">{t("drills.tab.tests")}</TabsTrigger>
            </TabsList>
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
