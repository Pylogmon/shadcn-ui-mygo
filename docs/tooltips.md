# 富内容 Tooltip

`desktop.TooltipWith` 在独立原生 Flyout 中组合文字、图标、Kbd 或分段排版，可以超出父窗口。保留的 `Tooltip` 简写继续接受字符串。

```go
desktop.TooltipWith(c, host, "save-hint", saveButton,
    desktop.TooltipProps{Description: "Save changes with S"},
    func(child *ui.Context) {
        ui.Text(child, "Save Changes")
        display.KbdWith(child, "S", display.KbdProps{InTooltip: true})
    })
```

每个 Host 显式绑定父窗口，并由 `BeginFrame` / `EndFrame` 管理消失的触发器。内容回调使用传入的 child Context，只能构建非交互视觉内容；不能捕获父窗口 Context、Element 或旧帧。`Description` 是必填的无障碍描述，内容的应用状态改变时增加 `Revision`。视图卸载会关闭窗口并清理回调。

Tooltip 不抢焦点，整个窗口穿透鼠标；hover 或触发器获得焦点时显示。Escape 关闭后保持抑制，直到指针与焦点离开，避免立即重开。需要交互内容时使用 HoverCard。默认立即打开，仍支持 OpenDelay / CloseDelay、Disabled、MaxWidth 和 OnOpenChange。

设置 `TooltipProps{HoverOnly: true}` 可仅由悬停触发视觉提示，焦点不会打开或维持提示，也不会跳过 `OpenDelay`；指针离开后按 `CloseDelay` 关闭，Escape 后移出再移入可重新触发。默认 `false` 保留键盘聚焦显示，所有模式都保留触发器的无障碍描述和正常键盘焦点。该选项适用于 `Tooltip`、`TooltipWith` 及对应的 Native 入口；`overlay.InWindowTooltip` 和兼容入口也接受 `HoverOnly`，继续使用 MyGo 的原生 Tooltip 延迟与关闭策略。

`Placement` 保留十二种物理 side/alignment。可用 `Side` 选择 `TooltipTop`、`TooltipBottom`、`TooltipLeft`、`TooltipRight`、`TooltipInlineStart`、`TooltipInlineEnd`；逻辑侧按父内容方向解析，Placement 的 start/end 后缀保留。独立窗口由系统翻转或移动后，箭头根据后端实际几何朝向触发器；未知或歧义几何时省略箭头。10 DIP 的圆角方形旋转45度形成箭头，原生窗口四周保留7 DIP透明空间，Gap 量到外窗口边缘。字体度量及布局继续由 MyGo 完成。

`NativeTooltipWith` 接受相同的富内容与描述，保留独立窗口的平台主题，不绘制 styled 箭头。MyGo 没有 OS Tooltip widget 接口，因此它是平台主题的 Flyout。原有 `NativeTooltip`、同窗兼容 API 与 HoverCard 行为保留；菜单透明预设不会改变 Tooltip 自己的动画。

运行 `go run ./examples/tooltip-composition` 或 `make tooltips-app`，可操作固定源的八种配置（含六个侧向按钮）、状态更新与平台主题切换。Disabled 例使用外层容器承载提示。原文、许可、SHA 与配置映射见 [source archive](../pkg/desktop/testdata/shadcn-tooltip/README.md)。真实平台的完整辅助技术与多显示器验收仍需完成。
