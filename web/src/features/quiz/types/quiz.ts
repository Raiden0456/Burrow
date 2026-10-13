export type QuestionKind = "single_choice" | "multi_choice" | "open"

export type QuizResponse = {
  selected?: number[]
  text?: string
}

export type AnswerResult = {
  response: QuizResponse
  score: number
  correct: boolean
  feedback: string
  explanation: string
  correctOptions: number[]
  referenceAnswer: string
}

export type HintKind = "hint" | "explain"

export type Hint = {
  id: number
  kind: HintKind
  content: string
  createdAt: number
}

export type Question = {
  id: number
  topicSlug: string
  kind: QuestionKind
  prompt: string
  options: string[]
  difficulty: number
  result: AnswerResult | null
  draft: QuizResponse | null
  hints: Hint[]
}

export type QuizKind = "topic" | "review" | "daily"

export type Quiz = {
  id: number
  kind: QuizKind
  topicSlug: string | null
  dailyDate: string | null
  createdAt: number
  completedAt: number | null
  score: number | null
  questions: Question[]
}

export type QuizSummary = Omit<Quiz, "questions"> & {
  total: number
  answered: number
}

export type AnswerOutcome = {
  result: AnswerResult
  quizCompleted: boolean
}
