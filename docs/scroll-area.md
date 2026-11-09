# ScrollArea 与 AspectRatio

ScrollAreaWith 把布局 Root 与真实 MyGo 滚动 Viewport 分开。默认滚动条按 shadcn 样式绘制，控制的就是该原生 Viewport 的 ScrollState；文字、输入、键盘、滚轮和边界滚动链仍由 MyGo 处理。

```go
var position ui.ScrollState

parts := display.ScrollAreaWith(c, display.ScrollAreaProps{
    Label: "Project notes", State: &position,
    Orientation: ui.Horizontal | ui.Vertical,
    Scrollbars: display.ScrollbarsAuto,
    OnScroll: func(p display.ScrollAreaPosition) { savePlace(p.X, p.Y) },
}, func(child *ui.Context) {
    buildNotes(child)
})
parts.Root.Size(360, 240).Radius(8)
```

零 Orientation 是竖向，支持横向与双轴。零 State 使用实例本地状态；提供指针可在离页后保留位置。Root 用于 Size/Border/Radius，Viewport 用于实际原生滚动配置。兼容 ScrollArea 保持返回真实竖向 viewport 的旧契约。

Always 在对应轴溢出时显示，Auto 在 hover/focus/捕获期间显示，Hidden 隐藏条但保留原生滚轮、键盘和 AX 滚动。默认 10 DIP 轨道与 7 DIP Border token 拇指，最小拇指 16 DIP；双轴 corner 会缩短实际轨道，比例仍对应可用行程。RTL 的横向 offset 从右侧起点计算。

默认条按固定 Base UI 1.6.0 的组合不增加 Tab 目标，原生 viewport 处理键盘与可访问滚动。FocusableScrollbars 是原生扩展，使用框架确实支持的 RoleSlider 与 Range；没有伪造 RoleScrollbar。可用 ScrollBar 为实际 TrackScroll viewport 单独组合。

拇指/轨道点击、原生捕获、越界拖动和取消会改变或恢复真实 offset。开始拖动会停止旧 wheel/reveal/ScrollTo 运动；外部位置/约束、条样式、状态、禁用或卸载变化不会让旧手势继续提交。父字体、当前颜色、完整自定义 token 和内部编辑器身份均保留。

NativeScrollbars 选择 MyGo 标准的 GPU overlay renderer，使用其 theme 参数。它不是 OS 的 NSScroller 等系统控件，布局与 shadcn 自绘条有区别。

OnScroll 在本帧原生 layout/clamp 后的 Draw 中报告接受的位置，初始放置静默，应用更改 offset 也会通知。Root.Changed 在一次修正构建中报告。回调可以改位置，下一帧重新报告最终接受值；稳定位置不持续重绘。

单独组合ScrollBar时，OnScrollIntent可在有效主键按下、滚轮及已处理的聚焦滚动键改变offset前报告用户意图，例如解除会话的自动跟随。程序ScrollTo保持静默，禁用、非滚动键和零轴滚轮不会触发。回调只操作应用模型；它的更新不重置已开始的原生拖拽。

当前 MyGo 的已启动 ScrollState.ScrollTo 在中途开启 ReduceMotion 时仍沿原动画前进；公开 API 无法读出其目标，因此本组件没有丢弃目标或假装完成收敛。减少动画下新发起的原生滚动会立即定位。该运行时缺口仍待修复，不计为完整减少动画验收通过。

```go
display.AspectRatioWith(c, display.AspectRatioProps{
    Ratio: 16.0 / 9, MaxWidth: 640, Clip: true,
}, func(child *ui.Context) {
    ui.Image(child, artwork).Fill().Fit(ui.Cover)
})
```

AspectRatioWith 默认允许内容溢出，与源 primitive 一致；旧 AspectRatio 保留剪裁契约。支持 Width/MaxWidth/MaxHeight、命名与描述。非正、非有限或无法表示高度的比值回退有限正方形并标记 Invalid；尺寸限制不伪造图片缩放。原生 Image 决定 Cover/Contain。

51 项正式检查与独审覆盖滚动实际内容、像素/主题/字体、双轴/RTL、拖拽/取消、native scrolling 中断、回调、稳定编辑器、嵌套链、AX 扩展及四种图片比例。真实触屏、全部平台/辅助技术/视觉与上述运行时动画缺口继续验收。
