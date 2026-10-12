import type { Question } from "../types/quiz"

export function questionContext(q: Question): string {
  const lines = [`Quiz question (${q.kind.replace("_", " ")}):`, q.prompt]
  if (q.options.length > 0) {
    lines.push("", "Options:")
    q.options.forEach((o, i) => lines.push(`${i + 1}. ${o}`))
  }
  const r = q.result
  if (r) {
    lines.push("")
    if (q.kind === "open") {
      lines.push(`Learner's answer: ${r.response.text ?? ""}`, `Reference answer: ${r.referenceAnswer}`, `Grader feedback: ${r.feedback}`)
    } else {
      const picked = (r.response.selected ?? []).map((i) => i + 1).join(", ") || "nothing"
      lines.push(`Learner selected: ${picked}`, `Correct: ${r.correctOptions.map((i) => i + 1).join(", ")}`)
    }
    lines.push(`Score: ${Math.round(r.score * 100)}%`, "", `Explanation: ${r.explanation}`)
  }
  return lines.join("\n")
}
