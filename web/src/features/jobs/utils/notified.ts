const storageKey = "gonteractive:notified-jobs"

function read(): number[] {
  try {
    const parsed = JSON.parse(localStorage.getItem(storageKey) ?? "[]")
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export function wasNotified(id: number) {
  return read().includes(id)
}

export function markNotified(id: number) {
  localStorage.setItem(storageKey, JSON.stringify([...read(), id].slice(-100)))
}
