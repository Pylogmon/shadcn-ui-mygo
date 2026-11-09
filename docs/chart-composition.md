# Chart 自由组合

`ChartContainer`为调用方的原生绘图提供独立配置作用域。它不选择图形、数据或交互方式；直接组合MyGo的`chart.Plot`、Axes、Rect、Point、Rule、Anchor及原生控件即可。原来的`display.Chart`仍是较简便的高层入口。

```go
colors := theme.TokensOf(c).Charts()
config := display.ChartConfig{
    "desktop": {Label: "Desktop", Color: &colors[0]},
}
display.ChartContainer(c, display.ChartContainerProps{
    Label: "Desktop visitors", Config: config,
}, func(child *ui.Context) {
    desktop, _ := display.ChartColor(child, "desktop")
    x := chart.NewBand("January", "February").Padding(.3, .15)
    y := chart.NewLinear(0, 400)
    chart.Plot(child, chart.Axes{Left: y.Ticks(4), Bottom: x.Ticks()}, func(area ui.Element) {
        for i, value := range []float64{186, 305} {
            chart.Rect(child, x.Band(i), y.Span(0, value)).Background(desktop).Radius(5)
        }
    }).Fill()
})
```

容器默认16:9、水平居中及12DIP文字，返回的Element可继续设置布局。原生布局首次测量即可得到尺寸，不使用网页ResponsiveContainer的SSR初始尺寸占位。任意原生绘图的轴、网格、参考线和输入组件仍由调用方设置样式；网页的后代CSS选择器不会自动修改这些节点。容器也不强制整个子树为LTR，避免改变编辑器或其它原生内容的方向。

`ChartConfigItem`包含Label、可替代文字的LabelContent、Icon，以及Color或Theme。Theme分别指定Light/Dark颜色；Color与Theme同时设置会panic。颜色使用指针，显式透明色与未提供颜色可以区分。`HasLabel`保留显式空标签；非空Label或LabelContent自动表示已提供标签。

`ChartConfigOf`返回独立副本；`ChartColor`返回已解析颜色及是否存在，缺少条目或颜色时不生成替代色。它们与内容组件只能在ChartContainer的同步Children构建中调用，缺少作用域会panic，包含inactive Tooltip的情况。嵌套、并列容器会恢复各自的配置。需要延迟Draw时，在构建期间取得颜色并捕获该值；不要保存Context，或在Draw里重新查配置。之后替换返回Element的Children不会重新安装该作用域，晚设置的fluent样式也不会重跑内容回调。

`ChartTooltipContent`接收应用提供的Active、Payload和可选Label，不管理悬停或浮层位置。支持dot/line/dashed指示器、隐藏标题/指示器、配置图标、NameKey/LabelKey、指示色覆盖及原生Formatter/LabelFormatter。单个payload且非dot时，标题嵌在条目内；有图标时图标优先，HideIndicator只隐藏默认色块。NameKey/LabelKey非空时自动生效；显式使用空字符串键时设置HasNameKey/HasLabelKey，Legend同样支持HasNameKey。Formatter替换整行，可以组合原生文字或控件。Label中的Content非nil时沿用源的非字符串label路径，由首条payload的配置决定标题；要直接提供原生标题，可用Config.LabelContent或LabelFormatter。标题或指示器的显示切换保留仍挂载的原生内容身份。

`ChartPayloadItem`的Name与DataKey保留原始字符串或数字，nil表示未提供。Value可以是数值、文字或数组；非nil值自动表示存在，nil配合HasValue表示显式null，供Formatter区分undefined。默认数值文字采用原生地区/语言格式；默认逗号分隔数组接受`[]any`、`[]string`、`[]float64`及`[]int`，其它Go值使用`fmt.Sprint`，也可由Formatter定制。Type为`"none"`的条目不显示，Formatter收到过滤后的索引和原始datum。Fields表示额外的顶层字段，Payload表示原始datum；`ChartPayloadConfig`按顶层字符串字段、datum字符串字段、原键的顺序查配置，并在间接键不存在时回退原键。

`ChartLegendContent`使用同一份配置的标签和图标，颜色来自payload。HideIcon回退为8DIP色块；顶部位置在下方留12DIP，底部位置在上方留12DIP。窄容器保留整项换行以维持可读性。

这些入口实现的是shadcn的配置与内容组合层。原生Rule/Anchor和应用持有的范围选择可以组成参考线、标签与筛选，但不等于已实现Recharts的全部Brush、LabelList、堆叠或动画API。完整几何/路径动画、平台和视觉矩阵继续在[差异清单](parity.md)中跟踪。

完整可运行示例见[examples/chart-composition](../examples/chart-composition/README.md)。它直接使用公开包，支持原生窗口、明暗PNG、月份范围、数据编辑与笔记；没有调用高层Chart。
