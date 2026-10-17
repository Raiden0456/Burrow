import { useState } from "react"
import { cn } from "@/lib/utils"
import { wallpaperUrl } from "../api"

export function WallpaperBackdrop({ version, dim }: { version: number; dim: number }) {
  const [loaded, setLoaded] = useState(false)
  return (
    <div aria-hidden className="pointer-events-none fixed inset-0 -z-10">
      <img
        key={version}
        src={wallpaperUrl(version)}
        alt=""
        onLoad={() => setLoaded(true)}
        className={cn("size-full object-cover transition-opacity duration-1000", loaded ? "opacity-100" : "opacity-0")}
      />
      <div className="absolute inset-0 bg-background transition-opacity duration-150" style={{ opacity: dim / 100 }} />
    </div>
  )
}
