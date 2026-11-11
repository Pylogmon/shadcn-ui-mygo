# ScrollArea 与 AspectRatio

ScrollAreaWith 把布局 Root 与 MyGo 滚动 Viewport 分开。竖向、横向和双轴分别使用 `ui.Scroll`、`ui.ScrollHorizontal`、`ui.ScrollBoth`，滚动条的样式、尺寸、显示、拖拽和输入均由 MyGo 统一处理。组件库不单独绘制滚动条。

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

默认 `ScrollbarsAuto` 使用 MyGo 的悬停与拖动显示策略；`ScrollbarsAlways` 映射到 `KeepScrollbars(true)`，在对应轴溢出时保持显示；`ScrollbarsHidden` 映射到 `Scrollbars(false)`，隐藏滚动条但保留原生滚轮、键盘和 AX 滚动。双轴角落、RTL 位置及横向逻辑坐标均遵循 MyGo。

滚动条使用当前 MyGo 主题的 `Scrollbar` 与 `ScrollbarWidth`，与应用的其他原生滚动区域一致。需要调整时在应用主题中统一设置。滚动条不增加独立 Tab 目标，键盘与无障碍滚动由真实 Viewport 处理。

父字体、当前颜色、完整自定义 token 和内部编辑器身份均保留。`ScrollAreaParts` 只提供 `Root`、`Viewport` 和 `State`；原来的独立 `ScrollBar` / `ScrollBarProps`、自绘条部件以及 `NativeScrollbars`、`Thickness`、`MinThumbSize`、`FocusableScrollbars` 参数已移除。原有调用应删除这些参数，直接使用 MyGo Viewport 与应用主题。

OnScroll 在本帧原生 layout/clamp 后的 Draw 中报告接受的位置，初始放置静默，应用更改 offset 也会通知。Root.Changed 在一次修正构建中报告。回调可以改位置，下一帧重新报告最终接受值；稳定位置不持续重绘。

当前 MyGo 的已启动 ScrollState.ScrollTo 在中途开启 ReduceMotion 时仍沿原动画前进；公开 API 无法读出其目标，因此本组件没有丢弃目标或假装完成收敛。减少动画下新发起的原生滚动会立即定位。该运行时缺口仍待修复，不计为完整减少动画验收通过。

```go
display.AspectRatioWith(c, display.AspectRatioProps{
    Ratio: 16.0 / 9, MaxWidth: 640, Clip: true,
}, func(child *ui.Context) {
    ui.Image(child, artwork).Fill().Fit(ui.Cover)
})
```

AspectRatioWith 默认允许内容溢出，与源 primitive 一致；旧 AspectRatio 保留剪裁契约。支持 Width/MaxWidth/MaxHeight、命名与描述。非正、非有限或无法表示高度的比值回退有限正方形并标记 Invalid；尺寸限制不伪造图片缩放。原生 Image 决定 Cover/Contain。

回归检查将 ScrollArea 与直接使用 MyGo 的滚动区域逐像素、逐位置比较，覆盖明暗主题、自定义原生滚动条颜色与宽度、三种轴向、RTL、显示策略、滚轮与真实拇指拖动，并验证回调、编辑器身份和嵌套滚动。真实触屏、全部平台/辅助技术/视觉与上述运行时动画缺口继续验收。
