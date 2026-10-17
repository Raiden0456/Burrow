import { api } from "@/lib/api"

export type WallpaperInfo = { version: number | null }

export const wallpaperApi = {
  info: () => api.get<WallpaperInfo>("/api/wallpaper"),
  upload: (file: File) => api.put<WallpaperInfo>("/api/wallpaper/image", file),
  remove: () => api.delete("/api/wallpaper/image"),
}

export const wallpaperUrl = (version: number) => `/api/wallpaper/image?v=${version}`
