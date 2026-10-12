import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { NavLink, useNavigate } from "react-router"
import { PlusIcon, Trash2Icon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import { chatApi } from "../api"
import { useStartChat } from "../hooks/useStartChat"

export function ThreadList({ activeId }: { activeId?: number }) {
  const { t } = useI18n()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const startChat = useStartChat()
  const threads = useQuery({ queryKey: ["chat", "threads"], queryFn: chatApi.threads })
  const remove = useMutation({
    mutationFn: chatApi.remove,
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ["chat", "threads"] })
      if (id === activeId) navigate("/chat")
    },
  })

  return (
    <div className="flex flex-col gap-2">
      <Button onClick={() => startChat.mutate({})} disabled={startChat.isPending}>
        <PlusIcon />
        {t("chat.newChat")}
      </Button>
      {threads.isPending && <Skeleton className="h-24 w-full" />}
      <div className="flex flex-col gap-0.5">
        {threads.data?.map((thread) => (
          <div key={thread.id} className="group flex items-center">
            <NavLink
              to={`/chat/${thread.id}`}
              className={({ isActive }) =>
                cn(
                  "min-w-0 flex-1 truncate rounded-md px-2 py-1.5 text-sm text-muted-foreground hover:bg-muted hover:text-foreground",
                  isActive && "bg-primary/12 font-medium text-primary hover:bg-primary/15 hover:text-primary",
                )
              }
            >
              {thread.title || t("chat.newChat")}
            </NavLink>
            <Button
              size="icon-xs"
              variant="ghost"
              aria-label={t("chat.delete")}
              className="opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
              onClick={() => remove.mutate(thread.id)}
            >
              <Trash2Icon />
            </Button>
          </div>
        ))}
      </div>
    </div>
  )
}
