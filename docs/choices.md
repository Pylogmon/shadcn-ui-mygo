# 选择与切换控件

Checkbox、Switch、Radio 和 Toggle 使用 MyGo 原生 Base 控件处理键鼠、焦点和可访问性；视觉层采用主题语义 token 与固定 Vega 基线。错误状态同时接受显式 `Invalid` 与外围 `Field.Error`。

```go
state := ui.CheckMixed // 保存到应用状态中
control.CheckboxState(c, &state, "Select all", control.CheckboxProps{
    Description: "Select every project",
    OnStateChange: func(next ui.CheckState) { updateChildren(next) },
})
```

`CheckboxState` 提供完整受控 Off/On/Mixed 状态，Mixed 激活变 On。原 `Checkbox` 的 `Indeterminate` 是调用方控制的视觉状态；若使用该兼容入口，调用方可在 `OnStateChange` 中清除此标志。

Switch 默认轨道 32×18.4 DIP、拇指 16 DIP；`Size: control.Small` 为 24×14、拇指 12。RTL 使用镜像轨迹；在动画进行中改值、切方向或开启减少动态效果不会留下旧位置或持续重绘。Label/Description 保留在原生 Switch 上。

```go
control.RadioGroup(c, &plan, control.RadioGroupProps{
    Label: "Plan", Orientation: ui.Vertical, NoLoop: true,
    Options: []control.Option{
        {Value: "free", Label: "Free", Description: "Personal projects"},
        {Value: "pro", Label: "Pro", Description: "Your growing team"},
    },
})
```

显式 Orientation 约束方向键并设置原生方向语义；零值保留原 API 的四方向键行为。Horizontal 控制旧版水平布局。NoLoop 在有效首尾停住，跳过禁用项；RTL 使用逻辑左右。Radio/Toggle 的值采用稳定 key，重排不会把焦点错误转给另一个值。

```go
gap := float32(0)
control.ToggleGroup(c, &formatting, control.ToggleGroupProps{
    Label: "Text formatting", Multiple: true, Outline: true, Spacing: &gap,
    Items: []control.Option{{Value: "bold", Label: "Bold"}, {Value: "italic", Label: "Italic"}},
    OnValueChange: func(values []string) { saveFormatting(values) },
})
```

Spacing 为 nil 使用 8 DIP，指向零则合并边框和外角；Vertical 支持竖向组合。单选组允许取消当前值。`Changed()` 与值回调在同一次用户事件中报告，空闲帧不重复；回调得到独立的 slice。

Toggle 可用 Icon/EndIcon 或 Content 定制视觉。所有 Content/ItemContent 槽位为非交互内容；真实输入和可访问性由外围原生控件承担。图标按钮应提供明确 Label，Rich Content 仍应同时提供可访问名称与描述。

Select的窗口内、独立Flyout和系统菜单入口使用同一个16×16DIP SVG下箭头，保持主题MutedForeground。图标尺寸独立于字体与行高，长标签不压缩图标，RTL位于逻辑末端；自定义Trigger内容仍由外围真实控件处理点击与名称。独立窗口与系统菜单的组件对见[桌面浮层](native-adaptation.md)。

可搜索单选/多选、清除、分组、自定义建议、共框标签与弹层内部搜索见[Combobox](combobox.md)。With入口区分选值和搜索文本；旧自由文本入口继续兼容。

39 项针对性检查覆盖原生状态、键鼠、RTL、尺寸/变体明暗像素、焦点、无效/禁用、稳定身份、同事件回调和动态减少动画。完整平台、屏幕阅读器和示例视觉矩阵仍需验收；固定上游基线见 [upstream.md](upstream.md)。

Gallery Forms 页的[Switch源组合](../examples/switch-composition/README.md)提供7份文档与5个Create分区的12个配置，保持各自默认、label／description、尺寸、禁用／错误与整卡布局。FieldLabel／整卡关联实际原生Switch，按键和指针更新独立的应用草稿；配置与页面切换后仍保留。

`control.FieldGroup` 默认间距为固定Vega的28DIP；源checkbox-group的12DIP和嵌套group的16DIP可以在返回Element上用Gap显式组合。

[Checkbox源组合](../examples/checkbox-composition/README.md)映射8份文档和7个Create分区，共15个配置；原label、不同默认、组／整卡、错误／禁用、译文和四行表格独立保存。源Table的header只在全部行选中时checked，部分选择保持普通未勾选；原Workspace access示例另演示受控Mixed，不把该扩展替代源表格合同。
