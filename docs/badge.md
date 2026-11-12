# Badge

`Badge(c,text,variant)` 是内容自适应的静态徽章；`BadgeWith` 增加图标、原生动作和自定义视觉槽位。六种变体为 Default、Secondary、Outline、Soft、Ghost、Link。

```go
display.Badge(c, "New", display.BadgeSecondary)
// Soft 默认使用当前主题的 Primary 色。
display.Badge(c, "Ready", display.BadgeSoft)
// 五种变体都可通过 Color 设置颜色。
green := ui.RGB(22, 163, 74)
for _, variant := range []display.BadgeVariant{
    display.BadgeDefault, display.BadgeOutline, display.BadgeSoft,
    display.BadgeLink, display.BadgeGhost,
} {
    display.BadgeWith(c, "Success", display.BadgeProps{Variant: variant, Color: &green})
}
// 危险状态通过 Soft + 主题危险色表达。
tokens := theme.TokensOf(c)
display.BadgeWith(c, "Deleting", display.BadgeProps{
    Variant: display.BadgeSoft, Color: &tokens.Destructive,
})
display.BadgeWith(c, "Remove filter", display.BadgeProps{
    Variant: display.BadgeOutline, EndIcon: closeIcon, Button: true,
    OnActivate: func() { removeFilter() },
})
display.BadgeWith(c, "Documentation", display.BadgeProps{
    Variant: display.BadgeLink, URL: "https://ui.shadcn.com/docs/components/badge",
})
```

`BadgeProps.Color` 类型为 `*ui.Color`，支持主要（BadgeDefault）、描边（BadgeOutline）、柔和（BadgeSoft）、链接（BadgeLink）和幽灵（BadgeGhost）变体。nil 使用各变体主题配色，彩色文字仍参与可读性校正；BadgeSecondary 继续使用主题次要色。可显式指定黑色、透明色或半透明色，静态、按钮和 URL 徽章均支持。

| 变体 | 自定义颜色的作用 |
| --- | --- |
| BadgeDefault | 填充背景；交互徽章悬停／按下为颜色自身透明度的 80%；文字和图标根据填充色与主题背景的合成结果，自动选择对比度较高的黑色或白色 |
| BadgeOutline | 派生可读文字、图标及原始色描边；交互徽章悬停／按下使用同色浅背景，亮色主题为 10%，暗色主题为 20% |
| BadgeSoft | 派生可读文字、图标及原始色浅背景；亮色主题背景为 10%，交互徽章悬停／按下为 20%；暗色主题分别为 20%／30% |
| BadgeLink | 派生可读文字和图标；保留透明背景与悬停下划线 |
| BadgeGhost | 派生可读文字和图标；默认透明背景，悬停使用同色浅背景，亮色主题为 10%，暗色主题为 20% |

上述透明度与自定义颜色自身的透明度相乘。交互徽章的焦点边框和焦点环均跟随自定义颜色：边框透明度为 40%，焦点环在亮色主题下为 20%、暗色主题下为 40%。`Invalid` 优先使用主题危险色。

原 `display.BadgeDestructive` 已替换为 `display.BadgeSoft`，枚举位置不变。旧用法迁移为 `BadgeWith(c, text, BadgeProps{Variant: display.BadgeSoft, Color: &tokens.Destructive})`，可保留危险状态外观。Gallery 的 Badges 展示默认 Soft，Themes 页演示成功、警告、危险色；固定源组合中的 Destructive 示例也通过此方式保留。

`BadgeLink` 单独只选择视觉样式；提供URL才使用实际MyGo Link，保留系统URL/Router和Enter-only激活。Button或OnActivate请求原生按钮，URL优先。返回Element是整个徽章容器，可继续设置样式；`Handle` 绑定真实按钮/Link。`OnActivate` 与Element.Changed是两条可选通知路径。

`BadgeProps.AlignSelf` 作用于整个徽章，包括组合链接的外层容器；零值保留 `ui.Start`。设置 `ui.Center` 可在横排中垂直居中、纵排中水平居中，并保留内容自适应宽度。原 `Badge` 入口与返回元素的链式对齐继续可用。

高度固定20 DIP、文字12 DIP，图标槽12 DIP；逻辑起止图标使对应内边距缩为6 DIP，RTL自动换边。`Leading` / `Trailing` 覆盖图标，`Content` 覆盖文字，只应构建非交互视觉内容。没有文字或含义不同的自定义内容应提供Label。Disabled/Busy禁止动作，Invalid与键盘可见焦点使用语义边框/光环。

Link变体悬停为默认和自定义文字提供继承下划线。原生字体采用1 DIP基线偏移；照搬CSS的4 DIP偏移会让下划线被20 DIP容器裁掉，此处按原生绘制适配。

Gallery 的 Content & feedback 页提供[完整源组合](../examples/badge-composition/README.md)，包括7份文档和7个Create分区。14个配置保留双侧图标、加载、明暗颜色、原始长文本横滚与RTL。示例原链接#／#link通过局部原生Router激活本地状态；当前Router仅保留path／query，不提供HTML锚点滚动或fragment历史。

彩色文字通过 `theme.ReadableForeground` 针对实际承载面达到默认 4.5:1 目标。`BadgeProps.Surface` 可显式提供承载背景；内嵌静态 Outline/Ghost/Link 徽标可设 `InheritForeground: true`，与父动作的文字、图标共享状态前景。原始 Color 仍用于填充、边框及普通焦点反馈。完整策略及原生扩展边界见 [可读彩色文字与 Sidebar](readable-color.md)。
