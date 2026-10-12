import { Link, useParams } from "react-router"
import { ChevronLeftIcon, MessageSquareIcon } from "lucide-react"
import { useI18n } from "@/i18n"
import { ChatThreadView } from "./ChatThreadView"
import { ThreadList } from "./ThreadList"

export function ChatPage() {
  const { t } = useI18n()
  const { id } = useParams()
  const threadId = id ? Number(id) : undefined

  return (
    <div className="grid gap-6 md:grid-cols-[14rem_1fr]">
      <aside className={threadId ? "hidden md:block" : "block"}>
        <ThreadList activeId={threadId} />
      </aside>
      <section className={threadId ? "block min-w-0" : "hidden md:block"}>
        {threadId ? (
          <>
            <Link to="/chat" className="mb-3 inline-flex items-center gap-1 text-sm text-muted-foreground md:hidden">
              <ChevronLeftIcon className="size-4" />
              {t("chat.all")}
            </Link>
            <ChatThreadView key={threadId} id={threadId} />
          </>
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
