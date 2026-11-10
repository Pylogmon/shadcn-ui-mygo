# Item

`display.Item` 保留原有Title/Description/Leading/Trailing组合，新增默认/描边/弱底色三种变体，以及默认/Small/ExtraSmall三种尺寸。

```go
display.Item(c, display.ItemProps{
    Key: "report", Title: "Quarterly report", Description: "Updated today",
    Variant: display.ItemOutline, Size: display.ItemSmall,
    Media: &display.ItemMediaProps{Variant: display.ItemMediaIcon, Icon: fileIcon},
    Primary: &display.ItemPrimary{
        Label: "Open quarterly report",
        OnActivate: func() { openReport() },
    },
    Actions: func(child *ui.Context) {
        if control.Button(child, "Pin", control.ButtonProps{}).Clicked() { pinReport() }
    },
})
```

`Primary` 只负责媒体/正文区域的主动作；Header、Actions、Footer始终是独立兄弟，自己的按钮和编辑器不会触发主动作。主区域在设置Primary后作为非交互视觉内容，不应包含编辑器或其他按钮。自定义内容没有Title时，应明确提供Primary.Label。

默认Primary使用原生ButtonBase；`Kind: ItemPrimaryLink, URL: "…"` 使用实际MyGo Link，保留URL、Router和Enter激活语义。`Primary.Handle` 绑定真实交互目标，外层Item仍是容器。通过 `Primary.OnActivate` 或返回Element.Changed处理动作，选择其中一条通知路径。

`Media` 的图像尺寸为40/32/24 DIP，使用已经解码的MyGo ImageSource并Cover裁剪；Icon默认16 DIP。自定义媒体保留原生控件的可访问性，有Label时是命名分组；简单命名图像/图标为image角色。`Decorative` 明确隐藏并禁止媒体输入。无Primary时可在自定义媒体中组合原生选择框等控件。

`ItemGroup(c, ItemGroupProps{Items: …, Label: …})` 提供原生list/list-item语义。重排时给每个成员唯一稳定Key，可保留动作编辑器的草稿、caret与焦点。`Separators` 增加装饰分隔，`Gap` 指针可覆盖自动间距。

`Content` / `Header` / `Actions` / `Footer` 接收当前子Context，独立compound helpers也可组合。标题单行、说明最多两行。过窄时主区域可收缩、动作换行；调用方明确指定的超宽固定控件使用有界原生横向滚动，保留所有操作入口。

Gallery 的 Content & feedback 页提供[35个固定源配置](../examples/item-composition/README.md)，对应11个独立文档示例和 Create 的24个分区。通过 `Item configuration` 切换尺寸/变体、原生主链接与独立按钮、分组/分隔、Header/Footer、音乐时长栏、模型图片和翻译RTL。九张新原图离线嵌入，三个头像复用已有原始素材；灰度在应用图像预处理阶段完成。

成员菜单的 Flyout 使用原生菜单项与丰富 Item/头像内容，Native 路由使用系统文字菜单；两者共享本地选择模型。示例动作只更新本地读数，不连接业务后台。搜索 Item 可定位本页，切页保留配置和最后动作。

多行菜单内容可通过 `desktop.MenuProps.ItemHeight` / `MenuContentProps.ItemHeight` 声明行高；默认32 DIP。该值同时用于真实行、窗口预算、子菜单定位和键盘滚动，变更会刷新已打开的同一窗口。Item 成员示例使用60 DIP以容纳原始头像、两行文字及padding；系统菜单保持操作系统自己的行高。
