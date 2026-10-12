const levels = ["bg-heat-0", "bg-heat-1", "bg-heat-2", "bg-heat-3", "bg-heat-4"]

export function heatClass(count: number) {
  if (count === 0) return levels[0]
  if (count <= 2) return levels[1]
  if (count <= 5) return levels[2]
  if (count <= 10) return levels[3]
  return levels[4]
}

export const heatLegend = levels
