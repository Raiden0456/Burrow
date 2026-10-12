import { useState, type FormEvent, type KeyboardEvent } from "react"
import { SendIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { useI18n } from "@/i18n"

export function Composer({ disabled, onSend }: { disabled: boolean; onSend: (text: string) => void }) {
  const { t } = useI18n()
  const [text, setText] = useState("")

  const submit = (e?: FormEvent) => {
    e?.preventDefault()
    const value = text.trim()
    if (!value || disabled) return
    onSend(value)
    setText("")
  }

  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey && !e.nativeEvent.isComposing && window.matchMedia("(pointer: fine)").matches) {
      e.preventDefault()
      submit()
    }
  }

  return (
    <form onSubmit={submit} className="flex items-end gap-2">
      <Textarea
        rows={1}
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={onKeyDown}
        placeholder={t("chat.placeholder")}
        className="max-h-48 min-h-10 resize-none"
      />
      <Button type="submit" size="icon-lg" disabled={disabled || !text.trim()} aria-label={t("chat.send")}>
        <SendIcon />
      </Button>
    </form>
  )
}
