import { useEffect, useRef, useState } from "react"
import { drillApi } from "../api"

export type SaveState = "idle" | "saving" | "saved" | "error"

export function useCodeAutosave(id: number, code: string) {
  const [state, setState] = useState<SaveState>("idle")
  const last = useRef(code)

  useEffect(() => {
    if (code === last.current) return
    const timer = setTimeout(async () => {
      setState("saving")
      try {
        await drillApi.saveDraft(id, code)
        last.current = code
        setState("saved")
      } catch {
        setState("error")
      }
    }, 800)
    return () => clearTimeout(timer)
  }, [id, code])

  return state
}
