import { useRef } from "react"
import { ImageIcon, ImageOffIcon, ImageUpIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { useWallpaper } from "@/features/wallpaper"
import { useI18n } from "@/i18n"
import { errorMessage } from "@/lib/api"
import { useInterface } from "../hooks/useInterface"

export function WallpaperField() {
  const { t } = useI18n()
  const input = useRef<HTMLInputElement>(null)
  const { version, upload, remove } = useWallpaper()
  const { dim, previewDim, commitDim } = useInterface()

  return (
    <div className="flex flex-col gap-2">
      <Label className="flex items-center gap-1.5">
        <ImageIcon className="size-4" />
        {t("settings.interface.wallpaper")}
      </Label>
      <input
        ref={input}
        type="file"
        accept="image/*"
        hidden
        onChange={(e) => {
          const file = e.target.files?.[0]
          if (file) upload.mutate(file, { onError: (err) => toast.error(errorMessage(err)) })
          e.target.value = ""
        }}
      />
      <div className="flex flex-wrap gap-2">
        <Button variant="outline" size="sm" disabled={upload.isPending} onClick={() => input.current?.click()}>
          <ImageUpIcon />
          {version ? t("settings.interface.wallpaperChange") : t("settings.interface.wallpaperChoose")}
        </Button>
        {version && (
          <Button variant="ghost" size="sm" disabled={remove.isPending} onClick={() => remove.mutate()}>
            <ImageOffIcon />
            {t("settings.interface.wallpaperRemove")}
          </Button>
        )}
      </div>
      {version && (
        <div className="flex max-w-sm items-center gap-3 text-sm">
          <span className="text-muted-foreground">{t("settings.interface.wallpaperDim")}</span>
          <input
            type="range"
            min={0}
            max={90}
            step={5}
            value={dim}
            aria-label={t("settings.interface.wallpaperDim")}
            className="flex-1 accent-primary"
            onChange={(e) => previewDim(Number(e.target.value))}
            onPointerUp={commitDim}
            onKeyUp={commitDim}
          />
          <span className="w-10 text-right tabular-nums text-muted-foreground">{dim}%</span>
        </div>
      )}
    </div>
  )
}
