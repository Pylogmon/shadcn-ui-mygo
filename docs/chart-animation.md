# Chart 数据动画

高层`display.Chart`默认对实际矩形、路径和极坐标角度做数据动画。默认时序来自固定shadcn依赖的Recharts 3.8.0：

| 家族 | 开始延迟 | 时长 |
| --- | ---: | ---: |
| Bar（横/纵） | 0 | 400ms |
| Line、Area、Radar、Radial | 0 | 1500ms |
| Pie、Donut | 400ms | 1500ms |

默认曲线为CSS ease，即cubic-bezier(.25,.1,.25,1)。通过`ChartProps.Animation`配置：

```go
duration := 800 * time.Millisecond
begin := time.Duration(0)
props.Animation = display.ChartAnimationProps{
    Duration: &duration,
    Begin: &begin,
    Ease: func(progress float32) float32 { return progress },
}
display.Chart(c, props)
```

Duration/Begin为nil时使用家族默认值；显式非正Duration立即完成，负Begin按0处理。Ease的结果限定为0..1，非有限结果回退为时间进度。Disabled关闭本图动画。系统减少动态效果、Instant帧及Gallery的Reduce motion优先立即显示最终数据，持续中的动画也会收敛并停止请求帧。静态PNG渲染与静态几何检查应明确使用Instant或ReduceMotion；需要观察动画时使用原生窗口或可推进时钟的Tester。

Bar从零数值尺寸进入，更新插值实际矩形；正负变化保留带符号的数值方向。Line首次按路径总长度显现，更新插值点坐标并从已显现的绝对长度继续。缺失观察仍断开，每个连续段使用相同的绝对显现长度并行绘制，符合SVG对每个子路径重置dash模式的规则；短段会先完成。Area首次沿X裁剪最终曲线与填充，后续插值上边界及基线。Pie/Donut插值各扇区角宽并依次累加，Radial插值起止角，Radar从各点中心进入并插值实际多边形。它们均保持颜色和尺寸的作用域，不以整体淡入或根节点缩放代替路径动画。

数据快速更新或反转时从当前显示的姿态继续；仅主题、检查值或其它不改变几何的重建不会重播。Tooltip、原生无障碍值和数据回调始终使用最新原始测量值，独立于绘图中间值。隐藏窗口或离屏图表直接收敛至最新目标，重新显示未改变的数据不自动重播；可见首帧会等待原生首次布局信号，避免误判为离屏。

Line已经绘制过后，暂时全部缺失的数据仍保留数值X，重新加载时缺失Y直接采用新目标Y，X继续插值，同时接续已显现的绝对长度。若分类数组曾完全清空，后续无匹配的新增点沿用源的animateNewValues规则，从绘图区右侧两倍宽度、垂直中点进入；它与从未存在可绘制路径的首次进入不同。

原生Line通过公开Path.Dash保存最长绝对显现长度，并按当前曲线总长度重新计算不会重复的间隔。固定Recharts源使用渲染时测得的SVG长度作为dash周期；旧短线或空路径更新成长曲线时，该生命周期可能保留旧周期，产生重复绘制缺口或零长度周期。原生实现保持各连续段内的连续显现，不复制这一缺口，因此不声称与React/SVG每一中间帧逐像素相同。原生曲线长度由独立数值计算取得，不读取Path私有字段；子路径规则见[SVG 2 Stroke Dashing](https://www.w3.org/TR/SVG2/painting.html#StrokeDashing)。完整字体、平台、视觉和其它Recharts配置仍按[差异清单](parity.md)跟踪。

固定npm包的SHA-512与shadcn lockfile相符；官方gitHead为`c04a79039ebc4d5057de163298d2e9917a5e5a66`。六家族与动画引擎TS/TSX、发布包中的引擎/命中辅助、包信息、官方metadata和MIT许可归档于`pkg/display/testdata/recharts-animation`。manifest记录Git tree blob或发布包成员及SHA-256，原生应用随包附带对应Recharts许可。
