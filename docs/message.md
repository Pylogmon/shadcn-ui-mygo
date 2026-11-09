# Message 组合

`Message` 保留Sender、Avatar、Header、Footer与正文的便利入口。`MessageWith`、`MessageGroup`、`MessageAvatar`、`MessageContent`、`MessageHeader`和`MessageFooter`对应固定源的独立槽位，正文可以放Bubble、Attachment、Trace、可选择文字与真实原生操作。

```go
conversation.MessageWith(c, conversation.MessageProps{
    Key: "reply-1", Label: "Assistant reply",
}, conversation.MessageSlots{
    Avatar: func(c *ui.Context) {
        display.Avatar(c, display.AvatarProps{Fallback: "AI", Size: 32})
    },
    Header: func(c *ui.Context) { ui.Text(c, "Assistant · Just now") },
    Content: func(c *ui.Context) {
        conversation.Bubble(c, conversation.BubbleProps{
            Variant: conversation.Secondary,
        }, func() { ui.Text(c, "Ready to review.") })
    },
    Footer: func(c *ui.Context) {
        control.Button(c, "Copy reply", control.ButtonProps{
            Variant: control.Ghost, Size: control.ExtraSmall,
        })
    },
})
```

根行使用14DIP文字、8DIP间距，正文占可用剩余宽，槽位间距10DIP。Avatar至少宽32DIP，位于底部；存在Footer时按源向上移动32DIP。Header/Footer为12DIP、500字重和12DIP水平padding。End反转原生布局，并保留Avatar→Content→Footer的构建和Tab顺序；RTL遵循逻辑方向。

Bubble、Header/Footer会自动采用Message的End规则。其它直接组件可用MessageSlot(c, element)显式应用同一规则，它不增加布局wrapper；普通裸Text不会被当成data-slot强制靠尾。MyGo没有公开后代data-slot查询，因此Ghost正文需声明MessageProps.Ghost以移除metadata水平padding。使用MessageSlots.Children完全自组根内容时，若包含Footer，需声明HasFooter以应用Avatar位移。

Label命名整个原生Article；Sender同时生成默认Header，是兼容入口。只需要无可见Header的语义名称时使用Label。稳定Key保留消息重排/前插后的编辑器身份、焦点与选区；回调和返回的Element只在当前构建中使用。流式内容、动作、分支与错误展示由应用模型组合，不自动启动AI或上传服务。

固定primitive、Vega样式、registry示例及八个独立Message示例均记录在upstream.json。真实原生输入、IME、可选择文字、URL、附件动作和组合已验证；全字体、平台、辅助技术及完整示例视觉矩阵继续跟踪。
