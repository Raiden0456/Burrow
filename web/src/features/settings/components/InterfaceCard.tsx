import type { ComponentProps } from "react"
import { useMutation } from "@tanstack/react-query"
import { LanguagesIcon, LogOutIcon, MonitorIcon, MoonIcon, PaletteIcon, SunIcon } from "lucide-react"
import { useTheme } from "next-themes"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { IconChip } from "@/components/IconChip"
import { authApi } from "@/features/auth"
import { locales, useI18n, type MessageKey } from "@/i18n"
import { useInterface } from "../hooks/useInterface"
import { WallpaperField } from "./WallpaperField"

const themes: { value: string; label: MessageKey; icon: typeof SunIcon }[] = [
  { value: "light", label: "settings.interface.light", icon: SunIcon },
  { value: "dark", label: "settings.interface.dark", icon: MoonIcon },
  { value: "system", label: "settings.interface.system", icon: MonitorIcon },
]

const languageNames = { en: "English", ru: "Русский" }

export function InterfaceCard(props: ComponentProps<"div">) {
  const { t, locale } = useI18n()
  const { theme, setTheme } = useTheme()
  const { setLanguage } = useInterface()
  const logout = useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => window.location.assign("/login"),
  })

  return (
    <Card {...props}>
      <CardHeader className="flex flex-row items-center gap-3">
        <IconChip icon={PaletteIcon} tone="purple" />
        <CardTitle>{t("settings.interface.title")}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-5">
        <div className="flex flex-col gap-2">
          <Label>{t("settings.interface.theme")}</Label>
          <div className="flex w-fit flex-wrap gap-1 rounded-lg border p-1 max-md:grid max-md:w-full max-md:grid-cols-3">
            {themes.map(({ value, label, icon: Icon }) => (
              <Button key={value} size="sm" variant={theme === value ? "secondary" : "ghost"} onClick={() => setTheme(value)}>
                <Icon />
                {t(label)}
              </Button>
            ))}
          </div>
        </div>
        <div className="flex flex-col gap-2">
          <Label className="flex items-center gap-1.5">
            <LanguagesIcon className="size-4" />
            {t("settings.interface.language")}
          </Label>
          <div className="flex w-fit gap-1 rounded-lg border p-1 max-md:grid max-md:w-full max-md:grid-cols-2">
            {locales.map((l) => (
              <Button key={l} size="sm" variant={locale === l ? "secondary" : "ghost"} onClick={() => setLanguage(l)}>
                {languageNames[l]}
              </Button>
            ))}
          </div>
        </div>
        <WallpaperField />
        <div>
          <Button variant="outline" onClick={() => logout.mutate()} className="max-md:w-full">
            <LogOutIcon />
            {t("settings.interface.signOut")}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
