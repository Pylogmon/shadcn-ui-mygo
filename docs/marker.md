# Marker 组合

`Marker(c, text, props)` 保留文字/图标便利入口；`MarkerWith`、`MarkerIcon`、`MarkerContent`、`MarkerText` 和 `MarkerLink` 提供源中的独立组合。默认Marker是静态内容，不创建按钮或live region。需要状态播报时显式设置Status或Busy；它们使用原生RoleStatus，Busy另外提供原生busy状态，不自动添加Spinner。

```go
conversation.Marker(c, "Conversation compacted", conversation.MarkerProps{
    Variant: conversation.MarkerSeparator,
})

conversation.MarkerWith(c, conversation.MarkerProps{
    Key: "checking-files", Status: true, Busy: checking,
}, func(c *ui.Context) {
    conversation.MarkerIcon(c, nil, conversation.MarkerIconProps{
        Content: func(c *ui.Context) { display.Spinner(c) },
    })
    conversation.MarkerContent(c, func(c *ui.Context) {
        conversation.MarkerText(c, "Checking files", conversation.MarkerTextProps{
            Shimmer: checking,
        })
    })
})
```

三种变体为MarkerDefault、MarkerBorder和MarkerSeparator；旧名Inline、Bordered、Separator保持可用。根行占满可用宽度，字号14DIP、行高20DIP、最小高度16DIP、间距8DIP，使用MutedForeground并继承原生字体与字重。图标槽为16×16DIP，只作装饰，不进入辅助技术或指针目标。普通长文本自然换行，可在MarkerText返回值上启用原生Selectable。

Border按源只绘制底边框，并增加8DIP底部padding；旧版的全边框、圆角和12DIP字号已修正。Separator用两侧可收缩的1DIP线条包围内容，保留源中的物理右/左4DIP附加间距，内容不收缩并居中。其它布局可在返回Element上用Column、AlignItems等公开方法组合，例如竖排图标与文字。

```go
conversation.Marker(c, "View the pull request", conversation.MarkerProps{
    Trigger: &conversation.MarkerTrigger{
        Kind: conversation.MarkerAnchor,
        URL: "https://example.com/pull/42",
    },
})

conversation.MarkerWith(c, conversation.MarkerProps{}, func(c *ui.Context) {
    conversation.MarkerContent(c, func(c *ui.Context) {
        ui.RichText(c).Children(func() {
            ui.Text(c, "Preview at ")
            conversation.MarkerLink(c, "example.com", "https://example.com")
            ui.Text(c, " is ready.")
        })
    })
})
```

Trigger选择真实原生Button或Link。Button保留Space/Enter；Link保留URL、Router与Enter激活，不用Space模拟按钮。根链接和MarkerLink在hover前就有下划线，偏移为3DIP，hover使用Foreground。OnActivate接收动作通知，返回根Element的Changed也可检测激活；Handle绑定实际原生焦点目标。禁用状态同时作用于指针和键盘。

MyGo公开Link只接受文本子节点，因此根Anchor由实际native Link覆盖视觉行，图标、Box及其它视觉内容位于独立的惰性行中；视觉行不产生重复辅助技术目标。MarkerText和MarkerLink提供的文字可构成自然名称，装饰图标中的文字不参与。自定义原生内容应提供MarkerProps.Label或Trigger.Label，没有文字或Label时使用URL；返回根Element上的fluent Label也会命名实际Link。需要独立按钮、编辑器、折叠或Drawer触发器时，将它们放在静态Marker中。

Shimmer使用共享的原生字形扫光，不替换文字排版与语义。公开API无法读取计算后的祖先颜色、字体或后续fluent覆盖；自定义扫光颜色通过MarkerTextProps.Color显式传入，MeasureFont只校准3ch扫光宽度估值，不改变实际原生字体。便利入口默认使用当前主题MutedForeground；根Anchor hover时采用对应Foreground。稳定Key及子内容的Context.Key保留重排后的原生编辑器身份；回调只在当前构建中使用，不保存帧Context或Element。

本模块32项聚焦检查、Root独立conversation/Gallery共368项检查及vet通过，明暗截图已导出并查看。覆盖实际像素、RTL、整行链接命中、原生URL/Router、焦点、禁用、选择与IME、空槽身份，以及静态Marker中的折叠/Drawer组合。固定primitive、Vega样式、registry示例、base文档和八个独立示例的来源记录见[upstream.json](upstream.json)；完整字体、平台、辅助技术及全示例视觉矩阵继续跟踪。
