# Cartesian 源例配置

`go run ./examples/chart-cartesian` 提供固定 shadcn 源码的 30 个 Area / Bar / Line 配置；`make cartesian-app` 构建 macOS 应用。日期范围与 Desktop / Mobile 控件修改实际数据，图形检查继续显示原值。完整数据、颜色与来源映射见 [示例](../examples/chart-cartesian/README.md)。

`ChartProps.Cartesian` 增加内边距、坐标轴预算与刻度配置、曲线、分类间距、柱组间距及光标。`ChartInsets` 使用物理 top/right/bottom/left DIP，可使用负值。`ChartCurveMonotone` 保留旧默认；`ChartCurveNatural`、`ChartCurveLinear`、`ChartCurveStep` 对应源例显式配置。Natural 使用原生 Path 实现自然三次样条，四类曲线均有官方 D3 函数输出对照。

每个 `ChartSeries.Style` 可设置独立曲线、填充透明度、渐变、描边、柱形四角半径、普通/活动点、活动柱形与数据标签。指针选项区分未提供与显式零值；渐变停止点在构建时复制。已有半透明颜色与 FillOpacity 相乘一次。

`ChartDotProps.Content` 和 `BarContent` 在当前构建中配置实际原生 mark 元素。`ChartDatum` 提供原始 Value、系列/分类、颜色、标准化坐标及活动状态。标签使用 `ChartLabelProps`，支持源例所需的 top / inside-left / right、offset、字体、格式化与原生内容。标签在数据动画结束后出现；检查值不随中间几何插值改变。

所有内容回调仅在 Chart 构建中调用，不能保留 Context、Element 或帧对象。延迟 Draw 仅捕获复制后的几何与颜色；应用可保留纯数据。主题、减少动态效果与原有七家族动画继续使用既有接口，参见 [Chart](chart.md)、[动画](chart-animation.md) 与 [组合](chart-composition.md)。

此阶段交付 Cartesian 能力和 30 个可运行配置，尚未把完整目录合并进 Gallery。Polar 31 例、Tooltip 9 例及其 docs / RTL / Create 流程继续实施。原生字体度量、刻度避让、图标与渐变栅格化是平台适配；配置构建检查不等于 30 例独立视觉或真实辅助技术验收。
