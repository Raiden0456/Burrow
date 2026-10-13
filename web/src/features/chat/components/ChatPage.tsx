import { useParams } from "react-router"
import { MessageSquareIcon } from "lucide-react"
import { useI18n } from "@/i18n"
import { ChatThreadView } from "./ChatThreadView"
import { ThreadList } from "./ThreadList"

export function ChatPage() {
  const { t } = useI18n()
  const { id } = useParams()
  const threadId = id ? Number(id) : undefined

  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-[14rem_minmax(0,1fr)]">
      <aside className={threadId ? "hidden min-w-0 md:block" : "block min-w-0"}>
        <ThreadList activeId={threadId} />
      </aside>
      <section className={threadId ? "block min-w-0" : "hidden md:block"}>
        {threadId ? (
          <ChatThreadView key={threadId} id={threadId} />
        ) : (
          <div className="flex min-h-[50vh] flex-col items-center justify-center gap-2 text-center text-sm text-muted-foreground">
            <span className="flex size-14 items-center justify-center rounded-2xl bg-hue-blue/15 text-hue-blue">
              <MessageSquareIcon className="size-7" />
            </span>
            {t("chat.pick")}
          </div>
        )}
      </section>
    </div>
  )
}
