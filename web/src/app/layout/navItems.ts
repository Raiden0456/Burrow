import { BookOpenIcon, HomeIcon, ListChecksIcon, MessageSquareIcon, SettingsIcon, type LucideIcon } from "lucide-react"
import type { MessageKey } from "@/i18n"

export type NavItem = {
  to: string
  label: MessageKey
  icon: LucideIcon
  end?: boolean
}

export const navItems: NavItem[] = [
  { to: "/", label: "nav.home", icon: HomeIcon, end: true },
  { to: "/topics", label: "nav.topics", icon: BookOpenIcon },
  { to: "/quizzes", label: "nav.quizzes", icon: ListChecksIcon },
  { to: "/chat", label: "nav.chat", icon: MessageSquareIcon },
  { to: "/settings", label: "nav.settings", icon: SettingsIcon },
]
