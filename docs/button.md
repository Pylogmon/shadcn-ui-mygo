# Button

`control.Button` 提供 Default、Secondary、Outline、Ghost、Soft、Link 六种变体；`control.StyleButton` 可将同样的外观应用到原生菜单或对话框触发器。

```go
// Soft 默认使用当前主题的 Primary 色。
control.Button(c, "查看详情", control.ButtonProps{Variant: control.Soft})

// 五种变体都可通过 Color 设置颜色。
green := ui.RGB(22, 163, 74)
for _, variant := range []control.Variant{
    control.Default, control.Outline, control.Soft, control.Link, control.Ghost,
} {
    control.Button(c, "确认", control.ButtonProps{Variant: variant, Color: &green})
}

// 危险操作通过 Soft + 主题危险色表达。
tokens := theme.TokensOf(c)
control.Button(c, "删除", control.ButtonProps{
    Variant: control.Soft,
    Color: &tokens.Destructive,
})
```

`ButtonProps.Color` 类型为 `*ui.Color`，支持主要（Default）、描边（Outline）、柔和（Soft）、链接（Link）和幽灵（Ghost）变体。nil 保留各变体原有主题外观；Secondary 继续使用主题次要色。可显式指定黑色、透明色或半透明色，`StyleButton` 同样支持。

| 变体 | 自定义颜色的作用 |
| --- | --- |
| Default | 填充背景；悬停／按下为颜色自身透明度的 80%；文字和图标根据填充色与主题背景的合成结果，自动选择对比度较高的黑色或白色 |
| Outline | 文字、图标和描边；悬停／按下使用同色浅背景，亮色主题为 10%，暗色主题为 20% |
| Soft | 文字、图标和浅色背景；亮色主题背景为 10%，悬停／按下为 20%；暗色主题分别为 20%／30% |
| Link | 文字和图标；保留透明背景与悬停下划线 |
| Ghost | 文字和图标；默认透明背景，悬停／按下使用同色浅背景，亮色主题为 10%，暗色主题为 20% |

上述透明度与自定义颜色自身的透明度相乘。五种变体的焦点边框和焦点环均跟随自定义颜色：边框透明度为 40%，焦点环在亮色主题下为 20%、暗色主题下为 40%。`Invalid` 仍优先使用主题危险色提示错误。

原 `control.Destructive` 已替换为 `control.Soft`。旧用法迁移为 `ButtonProps{Variant: control.Soft, Color: &tokens.Destructive}`，可保留原有危险操作外观。Gallery 的 Buttons 展示默认 Soft，Themes 页演示成功、警告、危险色；固定源组合中的 Destructive 示例也通过此方式保留。

尺寸、图标、Loading、Disabled、Invalid 和原生按钮的键盘／可访问性行为沿用其他变体。字号与系统缩放规则见 [控件字号与窗口上下文](control-capabilities.md)，完整源组合见 [Button compositions](../examples/button-composition/README.md)。
