export type DrillMode = "implement" | "bug_hunt"

export type DrillSummary = {
  id: number
  title: string
  mode: DrillMode
  difficulty: number
  solved: boolean
  createdAt: number
}

export type DrillAttempt = {
  code: string
  passed: boolean
  output: string
  createdAt: number
}

export type DrillHint = {
  id: number
  content: string
  createdAt: number
}

export type Drill = {
  id: number
  topicSlug: string
  title: string
  description: string
  starterCode: string
  mode: DrillMode
  testCode?: string
  initialOutput?: string
  difficulty: number
  solved: boolean
  draft: string
  lastAttempt: DrillAttempt | null
  hints: DrillHint[]
}

export type RunResult = {
  passed: boolean
  buildError: boolean
  timedOut: boolean
  output: string
  durationMs: number
}

export type DrillSolution = {
  solution: string
  testCode: string
  explanation: string
}
