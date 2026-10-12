import { Loader2Icon } from "lucide-react"

export function PageLoader({ label }: { label?: string }) {
  return (
    <div className="flex min-h-[40vh] flex-col items-center justify-center gap-3 text-sm text-muted-foreground">
      <Loader2Icon className="size-6 animate-spin" />
      {label}
    </div>
  )
}
