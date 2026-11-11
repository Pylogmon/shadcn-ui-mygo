# Badge

`Badge(c,text,variant)` 是内容自适应的静态徽章；`BadgeWith` 增加图标、原生动作和自定义视觉槽位。六种变体为 Default、Secondary、Outline、Soft、Ghost、Link。

```go
display.Badge(c, "New", display.BadgeSecondary)
// Soft 默认使用当前主题的 Primary 色。
display.Badge(c, "Ready", display.BadgeSoft)
// 自定义颜色同时用于文字、图标、浅色背景、悬停及焦点反馈。
green := ui.RGB(22, 163, 74)
display.BadgeWith(c, "Success", display.BadgeProps{
    Variant: display.BadgeSoft, Color: &green,
})
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

`BadgeProps.Color` 类型为 `*ui.Color`，仅作用于 `BadgeSoft`；nil 跟随当前作用域的主题主色，也可显式指定黑色、透明色或半透明色。与按钮的 Soft 一致，背景透明度在亮色主题下为 10%，交互徽标悬停／按下为 20%；暗色主题下分别为 20%／30%，与颜色自身的透明度相乘。焦点边框和光环使用相同颜色，`Invalid` 优先使用主题危险色。

原 `display.BadgeDestructive` 已替换为 `display.BadgeSoft`，枚举位置不变。旧用法迁移为 `BadgeWith(c, text, BadgeProps{Variant: display.BadgeSoft, Color: &tokens.Destructive})`，可保留危险状态外观。Gallery 的 Badges 展示默认 Soft，Themes 页演示成功、警告、危险色；固定源组合中的 Destructive 示例也通过此方式保留。

`BadgeLink` 单独只选择视觉样式；提供URL才使用实际MyGo Link，保留系统URL/Router和Enter-only激活。Button或OnActivate请求原生按钮，URL优先。返回Element是整个徽章容器，可继续设置样式；`Handle` 绑定真实按钮/Link。`OnActivate` 与Element.Changed是两条可选通知路径。

`BadgeProps.AlignSelf` 作用于整个徽章，包括组合链接的外层容器；零值保留 `ui.Start`。设置 `ui.Center` 可在横排中垂直居中、纵排中水平居中，并保留内容自适应宽度。原 `Badge` 入口与返回元素的链式对齐继续可用。

高度固定20 DIP、文字12 DIP，图标槽12 DIP；逻辑起止图标使对应内边距缩为6 DIP，RTL自动换边。`Leading` / `Trailing` 覆盖图标，`Content` 覆盖文字，只应构建非交互视觉内容。没有文字或含义不同的自定义内容应提供Label。Disabled/Busy禁止动作，Invalid与键盘可见焦点使用语义边框/光环。

Link变体悬停为默认和自定义文字提供继承下划线。原生字体采用1 DIP基线偏移；照搬CSS的4 DIP偏移会让下划线被20 DIP容器裁掉，此处按原生绘制适配。

Gallery 的 Content & feedback 页提供[完整源组合](../examples/badge-composition/README.md)，包括7份文档和7个Create分区。14个配置保留双侧图标、加载、明暗颜色、原始长文本横滚与RTL。示例原链接#／#link通过局部原生Router激活本地状态；当前Router仅保留path／query，不提供HTML锚点滚动或fragment历史。
