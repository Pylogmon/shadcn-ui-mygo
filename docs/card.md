# Compound cards

`CardWith` 按默认24 DIP或 `CardSmall` 的16 DIP间距布局头部、正文和页脚；`Spacing` 可以覆盖间距。每个内容回调收到当前卡片作用域的 Context。

```go
display.CardWith(c, display.CardProps{
    Label: "Meeting notes", Size: display.CardSmall,
    HeaderBorder: true, FooterBorder: true,
}, display.CardSlots{
    Title: "Design review", Description: "Updated today",
    Action: func(child *ui.Context) {
        control.Button(child, "Edit", control.ButtonProps{Variant: control.Ghost})
    },
    Content: func(child *ui.Context) { ui.Text(child, notes) },
    Footer: func(child *ui.Context) { control.Button(child, "Save", control.ButtonProps{}) },
})
```

`Action` 位于逻辑末端，RTL下会换边；标题和说明可以在剩余宽度换行。`Header` 可替换默认标题/说明栈，`Action` 仍独立布局。

`Media` 默认放在顶部并贴边；`MediaBottom` 将它放到页脚之后并贴住底边。媒体需要自己指定高度或宽高比，卡片负责外圆角裁剪。`ContentBleed` 让正文的表格、图像等贴左右边缘。分隔线和对应内侧留白由 `HeaderBorder` / `FooterBorder` 提供。

原有 `Card(c, content)` 继续提供统一24 DIP内边距，兼容已有 `CardHeader` / `CardContent` / `CardFooter` 组合。新槽位 API 已管理各区内边距，回调中直接构建内容即可。`CardTitle`、`CardDescription`、`CardAction` 也可单独组合。
