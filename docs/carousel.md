# Carousel

`CarouselContainer`、`CarouselContent`、`CarouselItem`和独立的Previous/Next控件
提供原生多项布局。`Carousel`和`CarouselWith`继续保留原有无间距分页入口。

```go
// selectedSnap和slides由应用持有。
var snapshot display.CarouselSnapshot
display.CarouselContainer(c, &selectedSnap, display.CarouselContainerProps{
    Key: "projects",
    Label: "Projects",
    Count: len(slides),
    Height: 160,
    Align: display.CarouselAlignStart,
}, func(c *ui.Context) {
    display.CarouselContent(c, display.CarouselContentProps{}, func(c *ui.Context) {
        for i, slide := range slides {
            display.CarouselItem(c, i, display.CarouselItemProps{
                Key: slide.ID,
                Basis: display.CarouselFraction(.5),
            }, func(c *ui.Context) {
                ui.Text(c, slide.Title).Padding(24)
            })
        }
    })
    display.CarouselPrevious(c, display.CarouselControlProps{})
    display.CarouselNext(c, display.CarouselControlProps{})
    snapshot = display.CarouselSnapshotOf(c)
})
```

调用方需为外侧按钮保留至少48DIP空间。默认不附加状态文字、圆点或按钮；
可以省略控件，也可通过返回的原生Element和`CarouselControlProps`定制位置、内容、按钮样式。
默认使用按需Material Symbols chevron、Outline/IconSmall和圆形外观。

## 布局与吸附位置

`selectedSnap`与`OnSelect`使用吸附位置索引，不是slide索引。
例如五项在start对齐、trimSnaps下，full、half、third分别产生5、4、3个吸附位置。
`CarouselSnapshot`提供独立的SlideCount、Snaps、Visible、选中位置和前后可滚动状态；
返回的slice是值快照，应用不能通过修改它改变组件内部状态。
Content前后都可读取快照：前置使用最近解析的item/spacing布局并采样当前动画，
Content收集新规格和原生拖动值后更新本帧结果；布局改变会请求后续帧收敛。
它是当前同步构建的值快照，不是可在任意时刻调用的长寿命controller。

- Basis零值填满一屏；`CarouselFraction(.5)`为半屏，`CarouselFixed(180)`为固定DIP。
- 默认Spacing为16DIP，包含补偿的前置padding；指向0的`Spacing`显式取消间距。
  fraction依据实测viewport加前置补偿计算，Basis包含该项的前置padding。
  实际outer尺寸至少为前置Spacing，与原生padding的最小尺寸一致。
- Align默认Center，另有Start、End；多个源示例显式选择Start。
- Contain默认TrimSnaps，另有KeepSnaps和Uncontained；Loop空间不足时按源规则回退为非循环。
- Vertical使用Height作为viewport高度，未指定时为240DIP；RTL使用逻辑方向。
- 正在进入或离开的部分可见项仍保留原生输入和可访问性；完全离屏项排除输入/AX。

Content内必须按索引顺序构建Count个Item，每项一次。
可比较、唯一且自反的Item.Key保留本地状态、原生编辑器和选区；非法或重复key回退到独立位置身份。
改变Count、Basis、Spacing、方向、尺寸或Loop会重新解析几何并校正选中位置，不触发用户OnSelect。

## 原生输入与阶段边界

组合函数在Container的同步children回调内使用；不要替换返回Element的Children来重建作用域。
可嵌套或并列，不保留跨帧Context、Element或应用回调。
编辑器内的方向键继续编辑文字，Carousel根或控件取得焦点时才导航；
垂直方向使用Up/Down，水平RTL反转左右键。这是对网页capture处理的原生适配。

当前组合入口沿用原生分页的250ms EaseOut与拖动阈值，`Duration`、`DragThreshold`是该兼容动画参数，
不是Embla同名选项；负Duration、减少动态效果和布局重配立即收敛。
窗口隐藏期间停止调度，恢复后收敛；若隐藏状态下仍执行构建，则立即跳至目标。
固定源的物理运动、速度投影、完整controller/options/events、Autoplay及插件接口仍需后续阶段实现。
独立窗口、真实触摸/平台读屏和完整视觉矩阵也尚未验收。

固定shadcn、Embla8.5.2、许可和可重现的源函数几何数据见
[`testdata/shadcn-carousel`](../pkg/display/testdata/shadcn-carousel/README.md)。

## Gallery

Data & display页提供可操作的Carousel示例：单项、双项、三项、响应式和固定宽度，
可切换间距、循环、集合长度及项目顺序。卡片内的真实原生编辑器把备注保留在对应项目上，
导航、重排和切换预览页后仍保留。另有270DIP垂直半屏布局及独立RTL示例，
导航按钮位于viewport外侧，小窗口通过页面滚动到达所有示例。
搜索Carousel、Slides或Slideshow可进入该页。
当前示例使用已交付的组合API；控制器、源物理和Autoplay示例会随其实现补齐。
