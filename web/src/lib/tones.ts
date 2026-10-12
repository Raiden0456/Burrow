export type Tone = "blue" | "green" | "yellow" | "orange" | "red" | "purple" | "cyan"

export const toneChip: Record<Tone, string> = {
  blue: "bg-hue-blue/15 text-hue-blue",
  green: "bg-hue-green/15 text-hue-green",
  yellow: "bg-hue-yellow/15 text-hue-yellow",
  orange: "bg-hue-orange/15 text-hue-orange",
  red: "bg-hue-red/15 text-hue-red",
  purple: "bg-hue-purple/15 text-hue-purple",
  cyan: "bg-hue-cyan/15 text-hue-cyan",
}
