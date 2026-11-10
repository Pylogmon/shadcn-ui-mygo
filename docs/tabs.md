# Tabs 独立组合

`TabsRoot`、`TabsList`、`TabsTrigger` 和 `TabsContent` 将标签列表与面板布局分开。列表可与输入框、按钮或菜单放在同一 header；面板可以位于独立的容器中。原有 int 状态的 `Tabs` / `TabsProps` 保留。

```go
// selected 和 query 由应用持有，显式初始化 selected = "overview"。
navigation.TabsRoot(c, &selected, navigation.TabsRootProps{Label: "Workspace"},
    func(c *ui.Context) {
        ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
            navigation.TabsList(c, navigation.TabsListProps{AlignSelf: ui.Center}, func(c *ui.Context) {
                navigation.TabsTrigger(c, "overview", navigation.TabsTriggerProps{Label: "Overview"})
                navigation.TabsTrigger(c, "analytics", navigation.TabsTriggerProps{Label: "Analytics"})
            })
            control.Input(c, &query, control.InputProps{Label: "Search"})
        })
        navigation.TabsContent(c, "overview", navigation.TabsContentProps{}, func(c *ui.Context) {
            ui.Text(c, "Overview content")
        })
        navigation.TabsContent(c, "analytics", navigation.TabsContentProps{}, func(c *ui.Context) {
            ui.Text(c, "Analytics content")
        })
    })
```

Root 使用应用的 `*string`，空值不选中任何项；不匹配 trigger 的值也能显示对应 Content。Root 显式设置 `Vertical`、`Manual`、`NoLoop` 和 `OnValueChange`。List 支持 default/line 两种样式。Trigger 的 `HideLabel` 隐藏视觉文字并保留有用的原生 Label，`Content` 可以替换图标和文字。

`TabsListProps.AlignSelf` 对实际参与当前父布局的列表容器生效，零值保留 `ui.Start`。`ui.Center` 可用于横排 header 的垂直居中或纵排的水平居中，Center/End 保留列表自然宽度；返回列表仍可链式覆盖对齐。

所有槽位必须在 Root 当前构建回调中调用，不能替换返回根元素的 Children 后继续调用槽位。槽位返回实际原生 Element，可在当前帧设置样式或绑定 Handle；同步 scope 支持兄弟与嵌套 Root，不跨帧保存 Context、Element 或应用内容回调。

方向键按 orientation 和 RTL 导航，跳过禁用项；Manual 模式需 Enter/Space 激活。原生 tablist 的 Controls / LabelledBy 与当前面板关联，面板是可聚焦的 region。面板可在 List 前声明，当前输入通过原生 invalidation 收敛。Inactive Content 的 builder 不调用，草稿应保存在应用模型中；不提供保留隐藏交互子树的选项。

运行 `go run ./examples/tabs-composition` 或 `make tabs-app` 可操作固定源的13种配置，包含实际 header 菜单、输入/动作与独立面板草稿。示例菜单显式选择同窗适配器；同一槽位也可放入 desktop Flyout 或 NativeMenu。来源与边界见 [source archive](../pkg/navigation/testdata/shadcn-tabs/README.md)。完整平台、字体与辅助技术验收仍待完成。
