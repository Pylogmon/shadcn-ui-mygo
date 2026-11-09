# 原生文字扫光

`display.ShimmerText` 对应固定上游的 `shimmer` 工具样式。它返回真实的 `ui.Text`，使用公开 ColorEffect 对字形覆盖区域着色；原生字体、中文排版、截断和辅助技术语义继续生效。

```go
color := theme.TokensOf(c).MutedForeground
display.ShimmerText(c, "正在生成回复…", display.ShimmerTextProps{
    Base: &color,
    Font: ui.Font{Size: 14, Weight: 500},
}).SingleLine().Ellipsis("…")
```

默认时长两秒、线性循环、倾角20度，扫光带宽为3ch加40 DIP；RTL反向。可设置Duration、Angle、Spread或SpreadPercent、Highlight、Reverse及Once。SpreadPercent以百分数传入，优先于Spread。Once结束后恢复源中的基础背景位置，不保留动画末帧。Static、桌面减少动画或当前Instant帧都显示普通文字并停止请求动画帧。

颜色在OKLCH/OKLab中计算并映射到sRGB。Base包含透明度；浅色高光降低透明度，深色高光提高亮度与透明度。MyGo没有公开计算后的祖先文字颜色或字体读取接口，因此Base省略时使用当前原生Theme.Text。需要继承其他祖先颜色时显式传Base；动画期间在返回值上设置TextColor不会更新渐变。Font设置原生字形属性，MeasureFont只为3ch估值提供已知祖先字体；祖先有其他字号或字体时，可传MeasureFont或显式Spread。

共享实现也供Attachment和Trace使用。固定CSS、文档、Git blob和SHA-256归档在 `internal/shimmer/testdata/shadcn/`。检查覆盖sRGB目标；没有将隐藏的Display P3原始坐标或完整平台/字体矩阵记为完成。
