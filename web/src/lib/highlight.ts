import { createHighlighterCore, type HighlighterCore } from "shiki/core"
import { createJavaScriptRegexEngine } from "shiki/engine/javascript"

let highlighter: Promise<HighlighterCore> | undefined

export function getHighlighter() {
  highlighter ??= createHighlighterCore({
    themes: [import("@shikijs/themes/one-light"), import("@shikijs/themes/one-dark-pro")],
    langs: [
      import("@shikijs/langs/go"),
      import("@shikijs/langs/bash"),
      import("@shikijs/langs/json"),
      import("@shikijs/langs/yaml"),
      import("@shikijs/langs/sql"),
      import("@shikijs/langs/diff"),
    ],
    engine: createJavaScriptRegexEngine(),
  })
  return highlighter
}

export async function highlight(code: string, lang: string) {
  const h = await getHighlighter()
  const loaded = h.getLoadedLanguages().includes(lang) ? lang : "text"
  return h.codeToHtml(code, {
    lang: loaded,
    themes: { light: "one-light", dark: "one-dark-pro" },
    defaultColor: false,
  })
}
