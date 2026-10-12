export type TopicStat = {
  slug: string
  answered: number
  mastery: number
  drillsSolved: number
  lastActivity: number
}

export type DayActivity = {
  date: string
  count: number
}

export type ProgressSummary = {
  streak: number
  activeToday: boolean
  totalAnswers: number
  answersToday: number
  accuracy7d: number | null
  dueCount: number
  drillsSolved: number
  activity: DayActivity[]
  topics: TopicStat[]
}
