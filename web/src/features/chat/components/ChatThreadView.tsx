import { useEffect, useRef } from "react"
import { Link, useLocation, useNavigate } from "react-router"
import { ChevronLeftIcon, Loader2Icon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ErrorAlert } from "@/components/ErrorAlert"
import { PageLoader } from "@/components/PageLoader"
import { useI18n } from "@/i18n"
import { useChatThread } from "../hooks/useChatThread"
import { Composer } from "./Composer"
import { MessageBubble } from "./MessageBubble"

export function ChatThreadView({ id }: { id: number }) {
  const { t } = useI18n()
  const { thread, send, pending, reply, error, streaming } = useChatThread(id)
  const location = useLocation()
  const navigate = useNavigate()
  const bottom = useRef<HTMLDivElement>(null)
  const autoSent = useRef(false)

  const autoSend = (location.state as { autoSend?: string } | null)?.autoSend
  useEffect(() => {
    if (!autoSend || autoSent.current || !thread.data) return
    autoSent.current = true
    navigate(location.pathname, { replace: true, state: null })
    send(autoSend)
  }, [autoSend, thread.data, send, navigate, location.pathname])

  const count = thread.data?.messages.length ?? 0
  useEffect(() => {
    bottom.current?.scrollIntoView({ block: "end" })
  }, [count, reply, pending])

  if (thread.isPending) return <PageLoader />
  if (thread.error) return <ErrorAlert error={thread.error} />

  return (
    <div className="flex min-h-[calc(100svh-6rem)] flex-col max-md:pb-24 md:min-h-[calc(100svh-3rem)]">
      <div className="sticky top-0 z-30 -mx-4 mb-4 flex items-center gap-1 bg-background/95 px-2 py-2 backdrop-blur md:static md:mx-0 md:bg-transparent md:p-0 md:backdrop-blur-none">
        <Button asChild variant="ghost" size="icon" className="md:hidden" aria-label={t("chat.all")}>
          <Link to="/chat">
            <ChevronLeftIcon />
          </Link>
        </Button>
        <h2 className="min-w-0 truncate text-lg font-semibold wallpaper:text-legible">{thread.data.thread.title || t("chat.newChat")}</h2>
      </div>
      {thread.data.thread.context && (
        <details data-glass className="mb-4 rounded-lg border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
          <summary className="cursor-pointer select-none">{t("chat.context")}</summary>
          <pre className="mt-2 whitespace-pre-wrap font-sans">{thread.data.thread.context}</pre>
        </details>
      )}
      <div className="flex flex-1 flex-col gap-5">
        {thread.data.messages.map((m) => (
          <MessageBubble key={m.id} role={m.role} content={m.content} />
        ))}
        {pending && <MessageBubble role="user" content={pending} />}
        {reply && <MessageBubble role="assistant" content={reply} />}
        {streaming && !reply && (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2Icon className="size-4 animate-spin" />
            {t("chat.thinking")}
          </div>
        )}
        {error && <ErrorAlert error={error} />}
        {count === 0 && !streaming && <p className="text-sm text-muted-foreground">{t("chat.empty")}</p>}
        <div ref={bottom} />
      </div>
      <div data-glass className="fixed inset-x-0 bottom-0 z-40 border-t bg-background/95 px-4 pt-3 pb-[max(env(safe-area-inset-bottom),0.75rem)] backdrop-blur md:sticky md:mt-4 md:border-0 md:bg-background md:backdrop-blur-none md:not-wallpaper:px-0 md:not-wallpaper:pt-2 md:not-wallpaper:pb-4 md:wallpaper:bottom-4 md:wallpaper:rounded-xl md:wallpaper:p-3">
        <Composer disabled={streaming} onSend={send} />
      </div>
    </div>
  )
}
