// Execute the archived official d3-shape functions without a replacement
// implementation. A recording CanvasPathMethods context captures exact doubles.
// node pkg/display/testdata/shadcn-chart/cartesian/generate-oracle.mjs
import fs from "node:fs/promises"
const base = new URL(".", import.meta.url)
const curves = {}
for (const file of ["natural", "linear", "step", "monotone"]) {
  const raw = await fs.readFile(new URL(file + ".js", base))
  const module = await import("data:text/javascript;base64," + raw.toString("base64"))
  curves[file] = module.default || module.monotoneX
}
const datasets = {
  one: [[40, 90]],
  two: [[0, 140], [260, 20]],
  visitors: [[0, 186], [100, 305], [200, 237], [300, 73], [400, 209], [500, 214]],
  nonuniform: [[0, 0], [30, 140], [75, -20], [260, 220]],
}
const cases = []
for (const [curve, factory] of Object.entries(curves)) {
  for (const [name, points] of Object.entries(datasets)) {
    const commands = []
    const context = {
      moveTo: (...v) => commands.push(["M", ...v]),
      lineTo: (...v) => commands.push(["L", ...v]),
      bezierCurveTo: (...v) => commands.push(["C", ...v]),
      closePath: () => commands.push(["Z"]),
    }
    const path = factory(context)
    path.lineStart()
    for (const point of points) path.point(...point)
    path.lineEnd()
    cases.push({name, curve, points, commands})
  }
}
await fs.writeFile(new URL("curve-oracle.json", base), JSON.stringify({
  source: "d3-shape@3.2.0 / 8ec82658454750cfa29efb1e0ea514e3dd9b2297",
  method: "direct official factory execution with recording path context",
  cases,
}, null, 2) + "\n")
