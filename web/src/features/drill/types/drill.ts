export type DrillSummary = {
  id: number
  title: string
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

export type Drill = {
  id: number
  topicSlug: string
  title: string
  description: string
  starterCode: string
  difficulty: number
  solved: boolean
  lastAttempt: DrillAttempt | null
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
}
