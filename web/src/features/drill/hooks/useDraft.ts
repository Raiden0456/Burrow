import { useEffect, useState } from "react"

export function useDraft(key: string, initial: string) {
  const storageKey = `gonteractive:draft:${key}`
  const [code, setCode] = useState(() => localStorage.getItem(storageKey) ?? initial)

  useEffect(() => {
    const t = setTimeout(() => localStorage.setItem(storageKey, code), 400)
    return () => clearTimeout(t)
  }, [code, storageKey])

  return [code, setCode] as const
}
