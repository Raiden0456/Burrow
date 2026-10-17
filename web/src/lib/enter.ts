import type { CSSProperties } from "react"

export const enter = (i: number) => ({ className: "enter", style: { "--i": i } as CSSProperties })
