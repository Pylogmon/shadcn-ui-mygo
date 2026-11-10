# Badge

`Badge(c,text,variant)` 是内容自适应的静态徽章；`BadgeWith` 增加图标、原生动作和自定义视觉槽位。六种变体为Default、Secondary、Outline、Destructive、Ghost、Link；保留原四种常量的数值。

```go
display.Badge(c, "New", display.BadgeSecondary)
display.BadgeWith(c, "Remove filter", display.BadgeProps{
    Variant: display.BadgeOutline, EndIcon: closeIcon, Button: true,
    OnActivate: func() { removeFilter() },
})
display.BadgeWith(c, "Documentation", display.BadgeProps{
    Variant: display.BadgeLink, URL: "https://ui.shadcn.com/docs/components/badge",
})
```

`BadgeLink` 单独只选择视觉样式；提供URL才使用实际MyGo Link，保留系统URL/Router和Enter-only激活。Button或OnActivate请求原生按钮，URL优先。返回Element是整个徽章容器，可继续设置样式；`Handle` 绑定真实按钮/Link。`OnActivate` 与Element.Changed是两条可选通知路径。

`BadgeProps.AlignSelf` 作用于整个徽章，包括组合链接的外层容器；零值保留 `ui.Start`。设置 `ui.Center` 可在横排中垂直居中、纵排中水平居中，并保留内容自适应宽度。原 `Badge` 入口与返回元素的链式对齐继续可用。

高度固定20 DIP、文字12 DIP，图标槽12 DIP；逻辑起止图标使对应内边距缩为6 DIP，RTL自动换边。`Leading` / `Trailing` 覆盖图标，`Content` 覆盖文字，只应构建非交互视觉内容。没有文字或含义不同的自定义内容应提供Label。Disabled/Busy禁止动作，Invalid与键盘可见焦点使用语义边框/光环。

Link变体悬停为默认和自定义文字提供继承下划线。原生字体采用1 DIP基线偏移；照搬CSS的4 DIP偏移会让下划线被20 DIP容器裁掉，此处按原生绘制适配。

Gallery 的 Content & feedback 页提供[完整源组合](../examples/badge-composition/README.md)，包括7份文档和7个Create分区。14个配置保留双侧图标、加载、明暗颜色、原始长文本横滚与RTL。示例原链接#／#link通过局部原生Router激活本地状态；当前Router仅保留path／query，不提供HTML锚点滚动或fragment历史。
