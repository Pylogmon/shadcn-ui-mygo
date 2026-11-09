# Command

`overlay.Command` 使用原生文本输入和虚拟列表，返回当前事件帧选中的Value；也可通过 `OnSelect` 收到完整条目。查询和异步结果由调用方持有。

```go
choice := overlay.Command(c, &query, overlay.CommandProps{
    Label: "Workspace commands", MaxHeight: 240,
    Items: []overlay.CommandItem{
        {Key: "open", Value: "open", Label: "Open project", Group: "Project",
            Description: "Choose a project from this device",
            Shortcut: "⌘O", Binding: &overlay.CommandBinding{Modifiers: ui.Cmd, Key: ui.KeyO}},
        {Key: "archive", Value: "archive", Label: "Archive project", Disabled: true},
    },
})
if choice == "open" { /* run the action */ }
```

`Shortcut` 仅绘制提示；`Binding` 提供真实修饰键/按键，`ui.Cmd` 在macOS为Command、其他平台为Control。快捷键仅在搜索输入聚焦且没有输入法组字时执行，精确匹配修饰键并抑制按住重复触发；普通、Shift或Alt字符保留给原生文本编辑。

方向键跳过禁用项，Home/End和PageUp/PageDown会把真实活动项滚入视口，页面跳转计入说明、分组标题和分隔符高度。默认循环，`NoLoop` 禁止回绕。输入始终保留键盘焦点，ActiveDescendant仅关联已经构建的可见原生选项。

提供稳定且唯一的 `Key` 可在异步替换/重排结果时保留高亮和行状态，允许不同条目具有相同Value。未提供Key时用Value及其出现次序识别；空Value使用Label。结果的Value或Label应非空，以区别返回的“本帧没有选择”。

默认使用cmdk 1.1.1的模糊评分：支持缩写、词边界优先、轻度相邻交换/重复输入容错，以及大小写和跳过距离的细微排序。原生API以显示Label为主要搜索文本，独立Value作为额外别名，Keywords继续追加；不会因路由ID与显示名称不同而搜不到可见名称。

`CommandScore(value, query, keywords...)` 也可独立使用，与固定上游函数的字符串/别名规则对应。若应用要精确按Value而非显示Label评分，可提供 `Filter: func(item overlay.CommandItem, query string) float64 { return overlay.CommandScore(item.Value, query, item.Keywords...) }`。

`Filter` 返回正数匹配分数，组内按分数稳定排序，分组按最高分排序，未分组结果位于分组之前。空查询保持注册顺序并绕过过滤器，清空后恢复原顺序。`NoFilter` 让应用自行提供过滤后的结果。`Loading` 显示加载状态并暂停选择，`Disabled` 禁止输入/选择，`EmptyContent` / `LoadingContent` 可自定义状态内容。

`ItemContent` 用当前虚拟行的Context构建非交互内容；行高由 `RowHeight` 约束，有Description时至少48 DIP。含交互按钮、编辑器等的复杂行应使用独立组合而非嵌入选项。分组与分隔符和行一起虚拟化，不会创建不可见结果的无障碍节点。

`CommandDialog` 使用窗口内Dialog焦点作用域，默认宽448 DIP，正常窗口从高度1/3处展开，短窗口会上移并约束列表。`DialogTitle` / `DialogDescription` 提供无障碍说明，`ShowClose` 显示关闭按钮；选择后关闭，query仍由调用方保留。完整平台输入法、屏幕阅读器与视觉矩阵仍待验收。

算法源码、MIT声明、SHA-256与2,299条固定JavaScript数值对照保存在 `pkg/overlay/testdata/cmdk`，Go测试无需Node或网络。可选 `node pkg/overlay/testdata/cmdk/generate-fixtures.mjs --check` 离线复验（生成环境Node 26.12.0 / Unicode 17）。对照另记录5条上游因Unicode大小写展开递归溢出的输入；Go实现将越过原始查询长度视为终止，保证安全返回有限评分。
