# Chart

`display.Chart`使用MyGo原生Plot、坐标尺度、矩形和路径绘图，支持Bar、Line、Area、Pie、Donut、Radar与Radial。数据来自应用的Labels和Series，颜色默认取当前主题的五个Chart token；也可为系列显式设置Color。

```go
radius := float32(5)
display.Chart(c, display.ChartProps{
    Kind: display.BarChart,
    Orientation: display.ChartHorizontal,
    Label: "过去六个月的访问量", Height: 240,
    Labels: []string{"January", "February", "March", "April", "May", "June"},
    Series: []display.ChartSeries{
        {Name: "Desktop", Values: []float64{186, 305, 237, 73, 209, 214}},
    },
    Cartesian: display.ChartCartesianProps{
        HideValueAxis: true, HideGrid: true, HideCursor: true,
        BarRadius: &radius,
        FormatCategory: func(label string) string {
            letters := []rune(label)
            return string(letters[:min(3, len(letters))])
        },
    },
})
```

FormatCategory只改变坐标轴显示，Tooltip、检查值和数据索引仍保留原始分类。

Orientation的零值ChartVertical保持现有纵向入口；ChartHorizontal仅作用于Bar。它让数值沿物理X轴增长，正值向右、负值向左，分类从上到下排列。RTL保留这一数据坐标方向，分类文本本身仍按应用方向排版。Line/Area及极坐标家族不随该属性转置。

Cartesian配置提供隐藏数值轴、分类轴、网格和检查线的独立选项，以及分类/数值轴格式化和Bar圆角。BarRadius为nil时沿用原来的3DIP，可显式传0。横向分类轴默认最多占60DIP，长文字省略以保留绘图区；CategoryAxisMaxWidth可以设置其它有限正值。分类轴隐藏时不预留对应轨道。

单系列Bar可使用CategoryColors逐分类着色。它同时作用于条形、Tooltip指标和可选分类Legend；多系列Bar继续按系列着色与生成Legend。应用也可直接组合ChartLegend和ChartTooltip，或用TooltipContent自组原生内容，例如省略标题、指标或添加应用信息。回调只在当前构建调用，不保存Context或Element。

Stacked使用正负分开的堆叠，Normalized沿MyGo的StackNormalized契约，按每个分类中全部有限值的绝对值总量计算比例；混合正负值分别保留方向，两侧不会各自变成100%。数值轴以实际几何单位格式化，Tooltip仍接收原始数据值；Cartesian.FormatValue可独立格式化数值轴，顶层FormatValue继续作为通用值格式入口。缺失、NaN和Infinity不伪造为测量值；极大或极小的有限值只缩放绘图计算，不改变检查回调中的原始值。

Selected是应用持有的分类索引，-1表示未检查；省略时使用组件局部状态。每个图表只有一个原生键盘停靠点，纵向用Left/Right、横向用Up/Down检查有效分类，Home/End到首尾，Escape退出检查。指针移动按对应分类维度选择；键盘操作不会被静止指针立即覆盖。无障碍Value包含分类、系列及格式化值。真实读屏与平台输入矩阵仍须验证。

ChartContainer在网页源中负责上下文和样式，Recharts负责几何。此Go入口是原生高层图表 API；源中任意Recharts组合、Brush、Reference line、柱顶LabelList、自由绘图区margin及Bar几何/路径数据插值动画还未全部具备。当前横向轴、颜色与圆角配置不代表完整图表目录已验收。对应固定示例、文档、MIT许可及文件校验归档于`pkg/display/testdata/shadcn-chart`，完整差异见[parity.md](parity.md)。
