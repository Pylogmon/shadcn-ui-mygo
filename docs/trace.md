# Trace 组合

`TraceGroup`、`TraceLine`、`TraceIcon`、`TraceTitle`、`TraceDescription` 与可折叠槽位对应固定源中的独立组合。静态Line不自动变成按钮；Action、Link、Collapsible分别使用真实原生动作、URL链接与受控展开。

```go
conversation.TraceGroup(c, func(c *ui.Context) {
    conversation.TraceLine(c, conversation.TraceLineProps{
        Status: conversation.Complete,
    }, func(c *ui.Context) {
        conversation.TraceTitle(c, "Read project files")
        conversation.TraceDescription(c, "12 files")
    })
    conversation.TraceCollapsible(c, &open, conversation.TraceCollapsibleProps{
        Label: "Build details", Status: conversation.Running,
        KeepMounted: true,
        Trigger: func(c *ui.Context) {
            conversation.TraceTitle(c, "Building native views")
        },
        Content: func(c *ui.Context) {
            control.Input(c, &note, control.InputProps{Label: "Build note"})
        },
    })
})
```

Running使直接Title使用共享Shimmer；Pending降低整行透明度，Error使用Destructive，Complete保持源的MutedForeground。Description为MutedForeground的70%并使用等宽数字。Title默认继承所在Line的状态；传TraceTitleProps可显式覆盖。静态Title不自动增加live announcement。

Line固定源的14DIP字号与行高，其余槽位按源继承祖先字体/字重；Description还继承字号和行高，Content继承字号并使用1.625行高。MyGo没有计算祖先字体getter，Running标题默认3ch带宽只能用原生Theme字体估值；自定义祖先字体不影响实际字形。需要精确带宽时可组合公开ShimmerText并显式传MeasureFont或Spread。

TraceAction返回真实Button，可以在当前构建中读取Clicked；TraceLink保留原生URL/Router与Enter语义。装饰Icon不会抢输入或单独朗读。原生Link目前没有公开LinkBase或关闭hover下划线入口，因此组合图标放在外层Line，只有文字是原生链接目标；整行图标点击范围和完整链接样式差异继续保留。

TraceCollapsible由调用方持有open；OnOpenChange报告原生触发。KeepMounted在首次展开前保持惰性，关闭后立即从输入/AX排除正文，并在需要时把焦点返回Trigger。200ms高度与箭头动画可反转，减少动画时直接切换。正文可以组合真实输入、可选择文字、CodeBlock、Dialog、Sheet或嵌套Trace，不保存帧Context/Element。

自定义Trigger中的Error状态应同时传入TraceCollapsibleProps.Status，使外层箭头应用错误色；原生公开API没有后代属性查询。旧TraceStep便利入口已自动传播Status。

旧的Trace(c, state, steps)便利入口继续可用。TraceStep新增Content、Open、KeepMounted、Disabled、URL与OnOpenChange；唯一顶层ID仍使用公开Expanded map，嵌套路径/重复ID/空ID另行隔离，避免同名子步骤互相开关。稳定唯一ID可在重排时保留各记录的原生状态；重复ID使用出现次序作为回退，不能替代应用的记录键。

固定16个Trace示例和组件/Vega样式的来源校验见upstream.json。AI请求、工具执行与持久会话属于应用模型，本库只提供UI组合；完整示例、平台、触控、辅助技术及视觉矩阵继续跟踪。
