import { useMemo } from "react"
import CodeMirror from "@uiw/react-codemirror"
import { go } from "@codemirror/lang-go"
import { keymap } from "@codemirror/view"
import { useTheme } from "next-themes"
import { oneDark, oneLight } from "../utils/editorThemes"

type Props = {
  value: string
  onChange?: (value: string) => void
  onRun?: () => void
  readOnly?: boolean
  minHeight?: string
}

export function CodeEditor({ value, onChange, onRun, readOnly, minHeight = "320px" }: Props) {
  const { resolvedTheme } = useTheme()
  const extensions = useMemo(
    () => [
      go(),
      keymap.of([
        {
          key: "Mod-Enter",
          run: () => {
            onRun?.()
            return true
          },
        },
      ]),
    ],
    [onRun],
  )

  return (
    <CodeMirror
      value={value}
      onChange={onChange}
      readOnly={readOnly}
      editable={!readOnly}
      theme={resolvedTheme === "dark" ? oneDark : oneLight}
      extensions={extensions}
      minHeight={minHeight}
      basicSetup={{ tabSize: 4, foldGutter: false }}
      className="overflow-hidden rounded-lg border text-sm"
    />
  )
}
