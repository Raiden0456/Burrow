import { AlertCircleIcon } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { useI18n } from "@/i18n"
import { errorMessage } from "@/lib/api"

export function ErrorAlert({ error, title }: { error: unknown; title?: string }) {
  const { t } = useI18n()
  return (
    <Alert variant="destructive">
      <AlertCircleIcon />
      <AlertTitle>{title ?? t("common.errorTitle")}</AlertTitle>
      <AlertDescription>{errorMessage(error)}</AlertDescription>
    </Alert>
  )
}
