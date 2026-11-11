# Message Scroller

`MessageScrollerWith`及其Provider、Root、Viewport、Content、Item、Button部件移植固定shadcn primitive的回合锚定与跳转行为。状态由应用持有，每个同时挂载的Provider使用独立的`MessageScrollerState`。

```go
conversation.MessageScrollerWith(c, &scroll, conversation.MessageScrollerProviderProps{
    AutoScroll: true,
}, func(c *ui.Context) {
    conversation.MessageScrollerViewport(c, conversation.MessageScrollerViewportProps{}, func(c *ui.Context) {
        conversation.MessageScrollerContent(c, conversation.MessageScrollerContentProps{}, func(c *ui.Context) {
            for _, message := range messages {
                conversation.MessageScrollerItem(c, conversation.MessageScrollerItemProps{
                    ID: message.ID, ScrollAnchor: message.IsUser,
                }, func(c *ui.Context) {
                    ui.Text(c, message.Text).Selectable()
                })
            }
        })
    })
    conversation.MessageScrollerButton(c, conversation.MessageScrollerButtonProps{})
})
```

应用把`scroll`保存在窗口或页面状态中，放在有明确高度约束的容器内。ID应稳定且唯一；默认ID也作为原生部件Key，可显式提供可比较的Key。正文可以放Bubble、Message、附件及实际原生编辑器。

默认`AutoScroll`为false，初始位置为End；还可选择Start或LastAnchor。默认边缘阈值8DIP、上一回合预览64DIP、命令偏移0DIP。可选数值使用指针，允许显式传0。初次按End或LastAnchor打开、且没有显式ID恢复时，Viewport在等待定位期间暂时隐藏。已排队的消息ID命令优先于默认位置。

开启自动滚动后，停在末端时跟随流式正文；读者滚轮、键盘或操作滚动条会中断跟随。新单个锚点回合放到阅读线附近，末尾暂留空白让后续回答增长；空白被消耗后交回末端跟随。历史前插默认按可见消息ID保留相对位置；开启跟随时，批量新增多个锚点仍保持末端。返回末端可恢复跟随。卸载Root、Viewport或Content会清理过期布局、即时命令及原生平滑滚动；等待尚未出现ID的命令单独保留。

`ScrollToMessage(id, options...)`支持Start、Center、End、Nearest四种对齐及Smooth。空列表可预先排队一个ID；非空列表的未知ID返回false。`ScrollToStart`和`ScrollToEnd`要求Viewport已挂载。`Scroll`是真实的`ui.ScrollState`，可读取原生坐标；命令和应用状态变更在UI线程执行。

`Scrollable()`区分两端是否还有真实内容，排除回合尾部占位；Following期间也会抑制流式追加尚未追上的短暂End缺口。`Visibility()`按阅读线返回当前锚点及可见消息ID。`Mode()`报告Following、Free、Anchored或Jumping，`Pending()`、`AutoScrolling()`及`TailSpace()`提供对应状态。跳转按钮默认平滑滚动，可设置Instant；OnClick返回false取消默认跳转。按钮离场动画期间已退出输入、Tab及无障碍导航。

Content默认32DIP行间距。请通过`MessageScrollerContentProps`声明Gap、PaddingStart、PaddingEnd，命令会纳入这些值；原生公开API没有计算后样式读取入口，fluent覆盖这些几何值不会同步到控制器。内容继承祖先字体和颜色。Viewport直接使用`ui.Scroll`及其标准滚动条，颜色、宽度、RTL位置、悬停显示和拖拽均遵循MyGo，不保留额外轨道留白，也不在自动滚动时单独改变滚动条样式。动态减少动画会停止本组件已知目标的平滑滚动。

Viewport默认保留嵌套原生滚动与常规链式滚动。仅在没有嵌套Scroll的平面正文中，可显式设置`ContainScrollAtEdges`阻止边缘事件向外传播。MyGo先向祖先分发自定义输入，再执行子Scroll默认行为，公开API没有CSS overscroll-behavior或已处理wheel查询，不能对任意嵌套正文自动复刻该CSS规则。

Content使用原生RoleLog，Viewport为有名称的Region；ARIA的独立relevant/live配置没有公开对应字段。真实读屏、触摸、跨平台及全部源示例视觉矩阵仍须验收。

旧`MessageScroller(c, *ui.ListState, count, props, row)`保留MyGo可变高度虚拟列表接口，适合大列表；FollowEnd、HoldPlace及按索引跳转遵循原生ListState。新的回合锚定组合构建完整原生Scroll内容，没有声称提供虚拟化锚定或VirtualViewport。

完整primitive、几何/控制器/命令源码及测试固定归档于`pkg/conversation/testdata/message-scroller`；文件校验与许可随归档保存，来源索引见[固定上游](upstream.md)。
