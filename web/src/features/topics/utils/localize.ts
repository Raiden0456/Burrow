import type { Locale } from "@/i18n"
import type { CurriculumModule, Module } from "../types/curriculum"

export function localizeModules(modules: CurriculumModule[], locale: Locale): Module[] {
  const ru = locale === "ru"
  return modules.map((m) => ({
    slug: m.slug,
    title: ru ? m.titleRu : m.title,
    topics: m.topics.map((t) => ({
      slug: t.slug,
      title: ru ? t.titleRu : t.title,
      objectives: ru ? t.objectivesRu : t.objectives,
      module: ru ? t.moduleRu : t.module,
    })),
  }))
}
