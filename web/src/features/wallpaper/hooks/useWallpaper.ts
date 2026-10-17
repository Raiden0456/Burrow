import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { wallpaperApi, type WallpaperInfo } from "../api"

const key = ["wallpaper"]

export function useWallpaper() {
  const queryClient = useQueryClient()
  const info = useQuery({ queryKey: key, queryFn: wallpaperApi.info, staleTime: Infinity })
  const upload = useMutation({
    mutationFn: wallpaperApi.upload,
    onSuccess: (data) => queryClient.setQueryData(key, data),
  })
  const remove = useMutation({
    mutationFn: wallpaperApi.remove,
    onSuccess: () => queryClient.setQueryData<WallpaperInfo>(key, { version: null }),
  })
  return { info, upload, remove, version: info.data?.version ?? null }
}
