import { NavLink, Outlet } from "react-router"
import { JobsIndicator, JobsWatcher } from "@/features/jobs"
import { useInterface, useInterfaceSync } from "@/features/settings"
import { WallpaperBackdrop, useWallpaper } from "@/features/wallpaper"
import { useFocusMode } from "@/hooks/useFocusMode"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import { RequireAuth } from "./RequireAuth"
import { navItems } from "./navItems"

function Shell() {
  const { t } = useI18n()
  const focus = useFocusMode()
  useInterfaceSync()
  const { dim } = useInterface()
  const { version } = useWallpaper()

  return (
    <div data-wallpaper={version ? "" : undefined} className="flex min-h-svh">
      {version && <WallpaperBackdrop version={version} dim={dim} />}
      <aside data-glass className="sticky top-0 hidden h-svh w-56 shrink-0 flex-col gap-1 border-r border-sidebar-border bg-sidebar p-3 md:flex">
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
        <JobsIndicator className="mt-auto" />
      </aside>
      <JobsWatcher />

      <main className={cn("min-w-0 flex-1 md:pb-0", focus ? "pb-0" : "pb-24")}>
        <div className={cn("mx-auto w-full max-w-5xl px-4 md:px-8 md:py-6", focus ? "pt-2 pb-4" : "py-5")}>
          {!focus && <JobsIndicator className="mb-4 md:hidden" />}
          <Outlet />
        </div>
      </main>

      {!focus && (
        <nav data-glass className="fixed inset-x-0 bottom-0 z-40 grid grid-cols-5 border-t bg-background/95 pb-[env(safe-area-inset-bottom)] backdrop-blur md:hidden">
          {navItems.map(({ to, label, icon: Icon, end }) => (
            <NavLink key={to} to={to} end={end} className="flex flex-col items-center gap-1 pt-2 pb-2 text-[11px] font-medium text-muted-foreground">
              {({ isActive }) => (
                <>
                  <span className={cn("flex h-8 w-14 items-center justify-center rounded-full transition-colors", isActive && "bg-primary/15 text-primary")}>
                    <Icon className="size-[22px]" />
                  </span>
                  <span className={cn(isActive && "text-foreground")}>{t(label)}</span>
                </>
              )}
            </NavLink>
          ))}
        </nav>
      )}
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
