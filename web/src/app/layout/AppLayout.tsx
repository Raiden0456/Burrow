import { NavLink, Outlet } from "react-router"
import { useInterfaceSync } from "@/features/settings"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import { RequireAuth } from "./RequireAuth"
import { navItems } from "./navItems"

function Shell() {
  const { t } = useI18n()
  useInterfaceSync()

  return (
    <div className="flex min-h-svh">
      <aside className="sticky top-0 hidden h-svh w-56 shrink-0 flex-col gap-1 border-r border-sidebar-border bg-sidebar p-3 md:flex">
        <div className="mb-4 flex items-center gap-2 px-2 pt-2 font-semibold">
          <img src="/logo.svg" alt="" className="size-7" />
          <span className="bg-gradient-to-r from-hue-blue to-hue-purple bg-clip-text text-transparent">Gonteractive</span>
        </div>
        {navItems.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              cn(
                "flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground",
                isActive && "bg-primary/12 font-medium text-primary hover:bg-primary/15 hover:text-primary",
              )
            }
          >
            <Icon className="size-4" />
            {t(label)}
          </NavLink>
        ))}
      </aside>

      <main className="min-w-0 flex-1 pb-20 md:pb-0">
        <div className="mx-auto w-full max-w-5xl px-4 py-6 md:px-8">
          <Outlet />
        </div>
      </main>

      <nav className="fixed inset-x-0 bottom-0 z-40 grid grid-cols-4 border-t bg-background/95 pb-[env(safe-area-inset-bottom)] backdrop-blur md:hidden">
        {navItems.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) => cn("flex flex-col items-center gap-0.5 py-2 text-[11px] text-muted-foreground", isActive && "text-primary")}
          >
            <Icon className="size-5" />
            {t(label)}
          </NavLink>
        ))}
      </nav>
    </div>
  )
}

export function AppLayout() {
  return (
    <RequireAuth>
      <Shell />
    </RequireAuth>
  )
}
