# 可调整大小的 Textarea

原 `Textarea` 保持返回真实原生编辑器的契约。需要用户调整高度时用 `TextareaWith`，它返回独立的 Root、Input 和 Handle：Root 参与布局，Input 保留原生编辑、选择和撤销，Handle 是位于逻辑末端的原生 splitter。

```go
// 保存到应用状态，而不是每帧重新创建。
var height control.TextareaResizeState

control.Field(c, control.FieldProps{Label: "Notes"}, func() ui.Element {
    parts := control.TextareaWith(c, &notes, control.TextareaProps{},
        control.TextareaResizeProps{
            State: &height, Label: "Resize notes", MaxHeight: 360,
            OnResize: func(dips float32) { persistHeight(dips) },
        })
    return parts.Root
})
```

Height 为零时使用原生内容高度，非零高度覆盖行数布局并在范围内夹取。State 可用 JSON 保存；不提供指针时使用该实例的本地状态。直接设置合法外部高度不触发用户变化回调。

默认最小高度 64 DIP，默认最大高度为整个窗口高度减 32，且不低于最小高度；这不是按控件当前位置计算的剩余窗口空间。可显式指定更大 MaxHeight。MinHeight 会为当前字体与系统文字缩放保留至少一行光标空间；非法参数禁用并标记手柄 Invalid，原生文字仍可编辑，调用方高度不被错误覆盖。

拖拽采用原生指针捕获，移出手柄也可继续调整。取消、失焦或卸载会恢复开始时的高度，包括原来的零值自适应状态。外部替换状态、范围、行数、方向或禁用状态会中止旧手势。单击手柄不会无故把自适应改为固定高度。

Up/Down 每次调整默认 8 DIP，Shift 调整十步，Home/End 跳到上下限。Left/Right 同样可用，以兼容原生 splitter 的可访问性增减桥接。RTL 将手柄移至左下角，调高/调低方向保持一致。手柄公开方向、范围、当前高度与 Controls 关联。

Root 和 Handle 的 Changed 只报告用户调整，OnResize 在当前构建中得到最终接受的高度；同一批原生事件可能合并成一次通知。回调自行改高度会在后续帧采用，并取消旧的捕获。ResizeProps.Disabled 只禁用调整，TextareaProps.Disabled 禁用整个控件；只读文本仍可调整。

Field 和 InputGroup 会解析 Root 内真实 Input，标签、错误、addon 聚焦和去除内部边框都作用于原生编辑器；调整手柄保持独立可操作。不要将交互手柄做成原生 Input 的子内容，MyGo 文本编辑器是布局和可访问性叶子。

原生捕获/取消、连续批次按键、状态恢复、稳定编辑器/undo/caret、无空闲重绘及组合目标均有检查。macOS 实窗已确认原生编辑后 Tab 到手柄，再按两次 Down 将高度 64 改为 80。真实指针/触屏、辅助技术和完整跨平台视觉矩阵继续验收。
