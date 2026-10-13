import { hasMessage, t } from "@/i18n"

export class ApiError extends Error {
  readonly status: number
  readonly code?: string

  constructor(status: number, message: string, code?: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

const codeMessages: Record<string, string> = {
  provider_not_configured: "errors.providerNotConfigured",
  wrong_password: "errors.wrongPassword",
  nothing_due: "errors.nothingDue",
  already_answered: "errors.alreadyAnswered",
  no_toolchain: "errors.noToolchain",
  hint_limit: "errors.hintLimit",
}

type Method = "GET" | "POST" | "PUT" | "DELETE"

async function toApiError(res: Response): Promise<ApiError> {
  const data = await res.json().catch(() => null)
  return new ApiError(res.status, data?.error ?? res.statusText, data?.code)
}

export async function request<T>(method: Method, path: string, body?: unknown, signal?: AbortSignal): Promise<T> {
  const res = await fetch(path, {
    method,
    signal,
    headers: body === undefined ? undefined : { "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  if (!res.ok) throw await toApiError(res)
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export const api = {
  get: <T>(path: string) => request<T>("GET", path),
  post: <T>(path: string, body: unknown = {}) => request<T>("POST", path, body),
  put: <T>(path: string, body: unknown) => request<T>("PUT", path, body),
  delete: <T = void>(path: string, body?: unknown) => request<T>("DELETE", path, body),
}

export type StreamHandlers = Record<string, (data: never) => void>

export async function stream(path: string, body: unknown, handlers: StreamHandlers, signal?: AbortSignal) {
  const res = await fetch(path, {
    method: "POST",
    signal,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw await toApiError(res)
  if (!res.body) throw new ApiError(500, "empty response stream")

  const reader = res.body.pipeThrough(new TextDecoderStream()).getReader()
  let buffer = ""
  for (;;) {
    const { value, done } = await reader.read()
    if (done) return
    buffer += value
    let boundary = buffer.indexOf("\n\n")
    while (boundary >= 0) {
      dispatch(buffer.slice(0, boundary), handlers)
      buffer = buffer.slice(boundary + 2)
      boundary = buffer.indexOf("\n\n")
    }
  }
}

function dispatch(raw: string, handlers: StreamHandlers) {
  let event = "message"
  let data = ""
  for (const line of raw.split("\n")) {
    if (line.startsWith("event: ")) event = line.slice(7)
    else if (line.startsWith("data: ")) data += line.slice(6)
  }
  const payload = data ? JSON.parse(data) : null
  if (event === "error") throw new ApiError(500, payload?.error ?? "stream failed", payload?.code)
  handlers[event]?.(payload as never)
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError && err.code) {
    const key = codeMessages[err.code]
    if (key && hasMessage(key)) return t(key)
  }
  return err instanceof Error ? err.message : String(err)
}
