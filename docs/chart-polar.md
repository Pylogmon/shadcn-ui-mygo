# Polar 源例配置

`go run ./examples/chart-polar` 提供固定源的31个 Pie、Radar、Radial 配置，`make polars-app` 构建 macOS 应用。月份选择实际更新中心数值与活动扇区，键盘和指针检查使用原始数据。

`ChartProps.Polar` 显式启用源配置；nil 保留原有极坐标默认与动画状态。角度为度数，0朝右，正向逆时针；Pie/Radial默认0→360，Radar默认90→−270。可配置起止角、物理margin、内外半径、边线/圆角/背景轨道、Grid、AngleAxis、RadiusAxis和Center。

```go
inner := display.ChartRadiusDIP(60)
outer := display.ChartRadiusPercent(80)
polar := display.ChartPolarProps{InnerRadius: &inner, OuterRadius: &outer}
// 将 &polar 传入 ChartProps.Polar；每个系列仍使用原有 Values/Style。
```

半径由百分比与DIP偏移组成，`radius.Add(10)` 表达随布局变化的外半径+10。每个 `ChartSeries.Style.Polar` 可覆盖半径、圆角、数据标签与纯值 Shape；Shape 返回一个或多个 `ChartPolarSector`，用于活动外环和中间空隙。它可能在同一构建的命中更新前后调用，必须保持纯函数。FillOpacity、Stroke、StrokeWidth、Dots和ActiveDot沿用系列样式。

Grid支持圆形/多边形、径向线、显式半径/角度与分层填充。Radar默认半径域使用源nice ticks；标签可替换为原生内容。Center收到最新原始数值和总量，标签支持外部引线、扇区内部与inside-start沿弧原生glyph。所有内容回调只在当前构建调用，Draw和动画仅保存复制后的值与公开glyph。

Pie按当前实际扇区及活动层命中，洞和层间空隙保持空白。多环Pie仅报告当前ring，PageUp/PageDown切换ring；Radar按角轴、Radial按完整半径轨道检查，未填充部分也可选中对应类别。`SelectedSeries` 可显式持有ring状态。箭头/Home/End/Escape仍通过一个原生焦点目标操作。

现有数据动画继续工作：Radar普通点跟随动画位置，活动点使用目标位置；扇区标签在动画结束后出现，Tooltip/AX/Center始终使用原始值。源配置与legacy几何单位使用独立动画key。

固定源31例、官方Recharts函数执行结果、Sector SVG轮廓、许可及元数据在 [source archive](../pkg/display/testdata/shadcn-chart/polar/manifest.json)。SVG与原生Path的边缘栅格允许一设备像素差异；字体、图标与原生选择器呈现属于平台适配。Tooltip9、Gallery完整目录及实际平台/辅助技术验收继续实施，配置构建不等于31例独立视觉验收。
