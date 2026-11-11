# Button

`control.Button` 提供 Default、Secondary、Outline、Ghost、Soft、Link 六种变体；`control.StyleButton` 可将同样的外观应用到原生菜单或对话框触发器。

```go
// Soft 默认使用当前主题的 Primary 色。
control.Button(c, "查看详情", control.ButtonProps{Variant: control.Soft})

// 自定义颜色同时用于文字、图标、浅色背景、悬停及焦点反馈。
green := ui.RGB(22, 163, 74)
control.Button(c, "确认", control.ButtonProps{
    Variant: control.Soft,
    Color: &green,
})

// 危险操作通过 Soft + 主题危险色表达。
tokens := theme.TokensOf(c)
control.Button(c, "删除", control.ButtonProps{
    Variant: control.Soft,
    Color: &tokens.Destructive,
})
```

`ButtonProps.Color` 类型为 `*ui.Color`，仅作用于 Soft；nil 跟随当前作用域的主题主色，也可显式指定黑色、透明色或半透明色。背景透明度在亮色主题下为 10%，悬停／按下为 20%；暗色主题下分别为 20%／30%，与自定义颜色自身的透明度相乘。焦点边框和焦点环使用相同颜色；`Invalid` 仍优先使用主题危险色提示错误。

原 `control.Destructive` 已替换为 `control.Soft`。旧用法迁移为 `ButtonProps{Variant: control.Soft, Color: &tokens.Destructive}`，可保留原有危险操作外观。Gallery 的 Buttons 展示默认 Soft，Themes 页演示成功、警告、危险色；固定源组合中的 Destructive 示例也通过此方式保留。

尺寸、图标、Loading、Disabled、Invalid 和原生按钮的键盘／可访问性行为沿用其他变体。字号与系统缩放规则见 [控件字号与窗口上下文](control-capabilities.md)，完整源组合见 [Button compositions](../examples/button-composition/README.md)。
