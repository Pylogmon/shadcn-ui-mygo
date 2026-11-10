# Alert 与 Empty

`Alert` 保留旧的Title/Description/Content API，`AlertWith` 增加前导图标、正文和独立操作槽位。返回的Element提供MyGo原生Alert角色（assertive/atomic），组件本身不会请求键盘焦点。

```go
display.AlertWith(c, display.AlertProps{
    Title: "Changes saved", Description: "Your workspace is ready.", Icon: checkIcon,
}, display.AlertSlots{
    Action: func(child *ui.Context) {
        if control.Button(child, "Undo", control.ButtonProps{Size: control.ExtraSmall}).Clicked() { undo() }
    },
})
```

`AlertSlots` 的Icon/Title/Description覆盖对应属性，Content补充正文；旧Content回调仍可使用。前导图标是非交互装饰，操作控件应放在Action或正文。Destructive使用语义错误色，独立 `AlertDescription(c,text,true)` 可明确选择错误样式。

动作按实际原生宽度留位。过窄时它移动到正文下方，超宽固定控件使用有界横向滚动；同一操作编辑器在布局变化时保持草稿、caret和焦点。实际屏幕阅读器播报、全平台及完整视觉矩阵尚待验收。

`Empty` / `EmptyWith` 提供媒体、标题、说明和操作区。默认48 DIP内边距，Header与Content最大宽384 DIP；`Padding` 指针允许明确0，`ContentWidth` 可调整限宽。`Bordered` 使用虚线边，`Muted` 增加弱底色。

```go
display.EmptyWith(c, display.EmptyProps{
    Title: "No projects yet", Description: "Create a project to get started.",
    Icon: inboxIcon, MediaVariant: display.EmptyMediaIcon,
}, display.EmptySlots{
    Content: func(child *ui.Context) {
        if control.Button(child, "Create project", control.ButtonProps{}).Clicked() { createProject() }
    },
})
```

`EmptySlots.Header` 可替换整个媒体/标题/说明区域，其他槽位替换各自内容。Content支持原生输入、链接、按钮等组合；旧Action在Content之后执行。Empty默认不会自动宣布状态，可由应用在返回Element上明确选择Status角色。

`EmptyMedia` 的Icon变体是40 DIP容器、24 DIP图标。简单图标默认装饰性；命名图标提供image角色；自定义媒体保留其原生可访问性，有Label时使用命名分组。明确 `Decorative` 后隐藏且禁止交互，避免出现无障碍树中不存在的键盘焦点。独立的Header/Title/Description/Media/Content helpers也可以组合。

Gallery 的 Content & feedback 页提供[完整源组合](../examples/status-composition/README.md)：Alert 的20种通知配置涵盖纯标题/说明、多个真实内联链接、明暗 amber、destructive正文及独立动作；Empty 的13种配置涵盖头像/灰度组、404 InputGroup/Kbd搜索、通知刷新、云存储、帖子、Card内嵌和ar/he翻译。两个配置区使用独立模型，搜索草稿跨页保留；菜单跟随全局Flyout／Native设置。原始头像复用已离线提供的固定源素材。

这些示例的创建、刷新、上传、联系与登录动作只记录本地意图，不连接后端。`Kbd` 是可视键帽；这里没有注册会抢走编辑器输入的全局斜杠快捷键。业务传输和路由由使用者接入。
