# Dialog、Sheet 与 Drawer

这些组件默认是当前窗口中的原生模态区域，使用 MyGo 的焦点作用域、键盘输入和可访问性；Drawer可选择NonModal。跨窗口菜单/Popover 的组件对见 [native-adaptation.md](native-adaptation.md)。

`Dialog` 支持 `Height` / `MaxHeight`、固定页脚、可滚动正文、独立禁止外部点击/Escape关闭和 `OnOpenChange`。只改变调用方的 `open` 值不会被当作用户关闭事件。`NoScroll` 供已自行约束高度的虚拟列表等使用。

```go
var open bool
var opener, nameInput ui.Handle

if control.Button(c, "Edit profile", control.ButtonProps{}).Bind(&opener).Clicked() {
    open = true
}
overlay.Dialog(c, &open, overlay.DialogProps{
    Title: "Edit profile", Description: "Update your public name.",
    InitialFocus: &nameInput, ReturnFocus: &opener,
    Footer: func() {
        if control.Button(c, "Done", control.ButtonProps{}).Clicked() { open = false }
    },
}, func() {
    control.Input(c, &name, control.InputProps{Label: "Name"}).Bind(&nameInput)
})
```

使用稳定 `ui.Handle` 指定初始焦点和返回焦点；不要保留跨帧 `ui.Element`。嵌套模态建议明确传 `ReturnFocus`：当前 MyGo 在某些嵌套关闭次序下会先聚焦父模态首控件，显式 Handle 可确保返回指定的触发器。退出期间保留动画像素，内容立即禁止输入并移出无障碍树；重开会反转退出。

`AlertDialog` 的确认返回值只在可用的动作激活时为true，调用方决定何时关闭。异步工作中设置 `Pending` 防止重复确认；默认同时锁定取消和Escape，`AllowCancelWhilePending` 可允许取消。`ActionDisabled`、`CancelDisabled` 可独立控制按钮。`Small` 提供小号布局，`Media` / `Content` 可补充图标和正文。

`Sheet` 支持四边，`Left` / `Right` 默认是物理边；`LogicalSide=true` 时根据RTL解释为逻辑起止边。正文真实滚动，页脚保持可操作。`ContentPadding` / `ContentGap` 的nil使用默认值，传指针可明确指定0。

`Drawer` 支持相同的边/焦点/页脚约定、原生指针捕获、取消回滚与速度投影。`SnapPoints` 中的 `Size` 是DIP，`Fraction` 是当前窗口沿滑动轴的比例，每个点只设一种单位；解析后的点受MaxSize（默认窗口轴长减16）限制，必须严格递增，裁剪后重复也视为无效。`ActiveSnap` 是从0开始的调用方索引，有效点列下越界会夹到合法范围，用户改变吸附点时触发 `OnSnapChange`；无效点列显示Invalid并禁用拖动，不会默默排序或覆盖错误索引。

捕获期间改变尺寸、边、滑动模式、吸附点或受控索引会取消旧手势并回到目标，保留正文编辑器身份。

`HandleOnly` 仅允许从可见拖动柄滑动，拖动柄也支持方向键/Home/End；`ShowSwipeHandle` 在整面板模式显示拖动柄。`DisableSwipe` 禁用滑动，`PreventSwipeDismiss` 保留吸附但禁止滑动关闭。`NonModal` 允许背景继续接收输入。

当前 MyGo Draggable 不把子按钮、编辑器或滚动视图的按压交给父级。整面板模式可从未被子控件占用的背景/拖动柄开始滑动，子控件保留自己的交互；滚动边界上的触摸接力仍是未完成的运行时能力。真实触屏、全平台及完整上游视觉矩阵仍待验收。
