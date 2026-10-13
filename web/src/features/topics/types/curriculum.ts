import type { Job } from "@/features/jobs"

export type Topic = {
  slug: string
  title: string
  objectives: string[]
  module: string
}

export type Module = {
  slug: string
  title: string
  topics: Topic[]
}

export type CurriculumTopic = Topic & {
  titleRu: string
  objectivesRu: string[]
  moduleRu: string
}

export type CurriculumModule = Omit<Module, "topics"> & {
  titleRu: string
  topics: CurriculumTopic[]
}

export type Lesson = {
  content: string | null
  createdAt: number | null
  job: Job | null
}

export type LessonSummary = {
  topicSlug: string
  createdAt: number
}
