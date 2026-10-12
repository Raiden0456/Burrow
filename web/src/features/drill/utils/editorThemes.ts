import { oneDark } from "@codemirror/theme-one-dark"
import { tags as t } from "@lezer/highlight"
import { createTheme } from "@uiw/codemirror-themes"

export const oneLight = createTheme({
  theme: "light",
  settings: {
    background: "#fafafa",
    foreground: "#383a42",
    caret: "#526fff",
    selection: "#e5e5e6",
    selectionMatch: "#e5e5e6",
    lineHighlight: "#f0f0f1",
    gutterBackground: "#fafafa",
    gutterForeground: "#9d9d9f",
    gutterBorder: "transparent",
  },
  styles: [
    { tag: [t.comment, t.lineComment, t.blockComment], color: "#a0a1a7", fontStyle: "italic" },
    { tag: [t.keyword, t.operatorKeyword, t.controlKeyword, t.definitionKeyword, t.moduleKeyword], color: "#a626a4" },
    { tag: [t.string, t.special(t.string)], color: "#50a14f" },
    { tag: [t.number, t.bool, t.null, t.atom], color: "#986801" },
    { tag: [t.function(t.variableName), t.function(t.propertyName)], color: "#4078f2" },
    { tag: [t.typeName, t.className, t.namespace], color: "#c18401" },
    { tag: [t.propertyName, t.attributeName], color: "#e45649" },
    { tag: [t.operator, t.punctuation], color: "#383a42" },
    { tag: [t.escape, t.regexp], color: "#0184bc" },
    { tag: t.invalid, color: "#ca1243" },
  ],
})

export { oneDark }
