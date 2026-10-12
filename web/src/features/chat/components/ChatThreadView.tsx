import { useEffect, useRef } from "react"
import { useLocation, useNavigate } from "react-router"
import { Loader2Icon } from "lucide-react"
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
    <div className="flex min-h-[calc(100svh-10rem)] flex-col md:min-h-[calc(100svh-3rem)]">
      <h2 className="mb-4 truncate text-lg font-semibold">{thread.data.thread.title || t("chat.newChat")}</h2>
      {thread.data.thread.context && (
        <details className="mb-4 rounded-lg border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
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
      <div className="sticky bottom-16 mt-4 bg-background pt-2 pb-2 md:bottom-0 md:pb-4">
        <Composer disabled={streaming} onSend={send} />
      </div>
    </div>
  )
}
