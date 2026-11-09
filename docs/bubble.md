# Bubble 组合

`Bubble` 保留原来的单内容面便利入口。`BubbleWith`、`BubbleContent`、`BubbleGroup` 和 `BubbleReactions` 对应固定源中的独立组合槽位，可在一个外壳中放多个内容面与独立反应层。

```go
conversation.BubbleWith(c, conversation.BubbleProps{
    Key: "reply-1", Variant: conversation.Secondary,
}, conversation.BubbleSlots{
    Content: func(c *ui.Context, p conversation.BubbleContentProps) {
        conversation.BubbleContent(c, p, func() { ui.Text(c, "Hello") })
    },
    Reactions: func(c *ui.Context) {
        conversation.BubbleReactions(c, conversation.BubbleReactionsProps{
            Interactive: true,
        }, func() {
            conversation.BubbleReaction(c, conversation.Reaction{
                Key: "like", Label: "Like", Count: 3, Selected: liked,
                OnSelectedChange: func(v bool) { liked = v },
            })
        })
    },
})
```

七种变体为Primary（BubbleDefault）、Secondary、Muted、Tinted、Outline、Ghost、Destructive。默认最大宽度80%，Ghost为100%；长文字自然换行，MaxWidth可附加DIP上限。Align沿文本方向，Message的End祖先规则也传入子Bubble。源中没有单独的尾巴或size API；大小示例通过内容和组合实现。

内容面保留祖先字体和字重继承；固定14DIP字号、1.625行高与源一致。普通内容、Button和Link都不额外重置为400字重。

Reaction为调用方控制的真实原生pressed按钮。Key保留重排后的焦点与按下状态。反应层不占正文高度，支持Top/Bottom和Start/End；这里Start/End对应源中的物理left/right，RTL也保持物理侧。被动pill保留内边距，包含按钮时设置Interactive。

BubbleTrigger选择真实Button或Link；可提供Label、Handle及OnActivate。Link保留原生URL/Router与Enter语义，Button保留Space/Enter。`BubbleContent`返回内容面，激活后其`Changed()`为true；Handle绑定实际的原生交互目标，OnActivate也由该目标驱动。交互控件宜放在普通内容面中，避免把按钮嵌入另一个按钮。Bubble的Expanded/Summary便利接口仍由调用方控制。

默认Link内容面使用覆盖整个面的真实原生Link接收点击、键盘及URL路由，视觉内容是不可交互的原生兄弟节点。可组合Column、Row、Icon和富文本，不受MyGo原生Link只接受文字子节点的限制。请用Label给复杂内容命名；省略时使用URL作为无障碍名称，框架没有公开的后代名称聚合接口。视觉内容不重复进入无障碍树。

```go
conversation.BubbleContent(c, conversation.BubbleContentProps{
    Variant: conversation.Outline,
    Trigger: &conversation.BubbleTrigger{
        Kind: conversation.BubbleLink,
        Label: "打开构建报告", URL: "https://example.com/report",
    },
}, func() {
    ui.Row(c).Gap(8).Children(func() {
        ui.Icon(c, reportIcon).Size(16, 16)
        ui.Text(c, "构建报告")
    })
})
```

仅包含原生文字时，可显式设置`BubbleTrigger.TextOnly: true`，直接使用MyGo的文字Link并从文字获得自然名称。该入口不能放Box、Row、Column或Icon；它保留MyGo的hover下划线，quiet变体使用语义Foreground，可在返回Element上显式TextColor覆盖。

Muted、Outline、Ghost的普通内容、Button及默认组合Link继承祖先文字颜色和字重。默认组合Link不额外添加文字下划线。Tinted与hover按源的OKLCH相对色和混色计算；输入token已转为sRGB，不能恢复其原始广色域坐标。
