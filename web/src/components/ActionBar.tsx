import type { ReactNode } from "react"
import { createPortal } from "react-dom"
import { cn } from "@/lib/utils"

export function ActionBar({ children, className }: { children: ReactNode; className?: string }) {
  return createPortal(
    <div
      data-glass
      className={cn(
        "fixed inset-x-0 bottom-0 z-40 flex gap-2 border-t bg-background/95 px-4 pt-3 pb-[max(env(safe-area-inset-bottom),0.75rem)] backdrop-blur md:hidden",
        className,
      )}
    >
      {children}
    </div>,
    document.body,
  )
}
