export type DailyStatus = {
  date: string
  quizId: number | null
  quizDate: string | null
  enabled: boolean
  pausedUntil: string | null
  scheduledAt: number
}
