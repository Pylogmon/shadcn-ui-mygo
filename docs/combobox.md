# Combobox

`ComboboxWith`和`MultiComboboxWith`使用真实MyGo文字编辑器，区分已选值、显示标签和搜索文本。
应用持有状态；组件使用稳定的`Option.Value`标识建议和已选标签。
需要对象值、独立槽位或自定义列表布局时，可使用共享同一实现的`ComboboxRoot`。

```go
selected := control.ComboboxValue{Value: "mygo", Valid: true}
query := ""

control.ComboboxWith(c, &selected, &query, control.ComboboxWithProps{
    Label: "Framework",
    Placeholder: "Choose a framework",
    ShowClear: true,
    Options: []control.Option{
        {Value: "mygo", Label: "MyGo"},
        {Value: "gio", Label: "Gio"},
        {Value: "", Label: "None"},
    },
})
```

实际应用应把`selected`和`query`放在自己的持久状态中。
`Valid`区分合法的空字符串选项与未选择。选择写入Value，外部输入显示Label；
输入过滤不会把未提交文字当成已选值；单选外部编辑器被删除或剪切为空时，也会清除已选值。
退出动画完成后恢复所选标签，弹层内部搜索则清空搜索文本。
程序更新选值/选项不触发用户值回调；动态标签在未编辑时同步显示。

默认不自动高亮首个匹配项，`AutoHighlight`显式开启；方向键循环并跳过禁用项和分隔线。
手动模式经过“输入/无高亮”位置后再循环，自动高亮直接循环，`NoLoop`停在边界。
重新打开时定位已选项并滚入视口；鼠标离开由指针高亮的项目会取消该高亮。
Enter选择当前高亮项，Escape关闭；已关闭时Escape清除选择。
输入法候选确认不会同时选择建议，组合输入中的退格不会删除多选标签。
原生query可显示预编辑文字，但过滤、`OnInputChange`和空输入清值等到提交后执行。

## 多选与弹层搜索

```go
control.MultiComboboxWith(c, &technologies, &technologyQuery, control.ComboboxWithProps{
    Label: "Technologies",
    Placeholder: "Add technology",
    Options: options,
    AutoHighlight: true,
    MaxSelected: 4,
})
```

已选项继续显示在列表中，再次选择会取消；未过滤时选择保持列表打开，
外部输入的过滤结果被选中后关闭并清空query。达到`MaxSelected`只禁止新增，
现有选择仍可取消。值回调接收独立slice。

已选标签与输入在同一个可换行容器中。输入光标位于起点时，LTR按Left、RTL按Right进入最后一个标签；
标签上的左右键按LTR/RTL移动，Backspace/Delete删除当前项，Enter/Space返回输入，
Up/Down打开选项。可打印字符（含数字、标点和Shift字符）把焦点交回真实编辑器，保留原生文字/IME/撤销路径。
`HideRemove`隐藏移除按钮，保留标签键盘删除；只读和禁用阻止修改。

`PopupInput: true`把搜索框移入弹层，外部显示原生按钮；单选显示所选标签，多选显示摘要。
支持`ShowClear`、自定义`InputPlaceholder`，Escape和Tab保持原生回焦及后续控件顺序。
弹层按钮可直接组合到`ButtonGroup`，不会注册内部重复按钮、重复标签或额外偏移。
触发器提供combobox/haspopup=dialog语义，内部搜索为list-autocomplete combobox，弹层为非模态dialog。
无障碍消费者需用HasPopup区分同名触发器与编辑器。
关闭状态的单选触发器支持750ms前缀typeahead；按固定Base源，它可以匹配标为disabled的标签，
而展开列表中的禁用项仍不可用方向键选中。关闭态多选、只读、禁用和IME候选不会执行该选择。

## 内容与表单组合

`Option.Group`和`Separator`提供命名原生分组/分隔语义；列表标记选中状态，
多选列表提供原生Multiselectable，活动建议通过ActiveDescendant关联到编辑器。
`ItemContent`和`TriggerContent`可定制视觉，仍需明确的Option.Label。
默认trim查询后按标签做大小写/重音/字宽宽松匹配，`Locale`指定区域；
自定义`Filter(option,query)`仍接收原始query，可搜索描述等元数据。
过滤后空组及首尾/连续分隔线被归一化。独立Group/Collection结构见下文组合API。
应用更新Options即可替换建议；请求调度、加载与数据竞争处理由应用负责。

`Field`、`LabelFor`、`InputGroup`通过实际editor-target关联原生输入，保留标签、描述、错误、
只读复制、原生选择/IME以及焦点外观。`Size`与普通输入使用相同控件尺寸；减少动画服从主题设置。

旧`Combobox(c,*string,ComboboxProps)`保留自由文本和首项高亮行为，旧`MultiCombobox`
保留独立标签行与选择策略；两者继续可用，并修复禁用项、IME及重排焦点缺陷。
需要上述选值/标签/query契约的新代码可使用With或Root入口。

## 组合API与对象值

`ComboboxState[T]`由应用持有，包含Value、Values、Query和Open。地址在挂载期间保持稳定；
`ComboboxSelection[T].Valid`区分未选中与合法的零值。Root不增加包围槽位的布局容器，
内部状态由零尺寸、无输入/AX的原生挂载节点管理，卸载时通过公开Local清理。

```go
// state保存在应用状态中。
control.ComboboxRoot(c, &state, control.ComboboxRootProps[string]{
    Items: []string{"Go", "Rust", "Swift"},
    Adapter: control.ComboboxStringAdapter(),
    AutoHighlight: control.ComboboxHighlightOnInput,
}, func(b *control.ComboboxScope[string]) {
    parts := b.Input(c, control.ComboboxInputProps{
        Label: "Language", Placeholder: "Search languages", ShowClear: true,
    })
    parts.Editor.OnSubmit(submitForm)
    b.Content(c, control.ComboboxContentProps{}, func() {
        b.Empty(c, "No languages found.")
        b.List(c, control.ComboboxListProps{}, func() {
            b.Collection(c, func(c *ui.Context, value string, _ int) {
                b.Item(c, value, control.ComboboxItemProps{}, nil)
            })
        })
    })
})
```

对象值通过`ComboboxAdapter[T]`提供Key、Label、FormValue、Equal及Disabled。
Key必须唯一，Equal默认比较Key；显式Equal可独立决定选择相等性，查找按集合声明顺序进行。
`FormValues()`返回当前选择的提交字符串，原生Editor可继续接OnSubmit、Field目标或其他编辑能力。
适配器、渲染器和事件回调每次构建重新提供，不保存在内部长期状态中。

Root支持Items或Groups数据，也可省略二者，通过手写Group/Item注册集合。
Group、Label、Collection、Separator、Item、Empty和List可独立组合；Empty是List的兄弟状态节点，
有结果时保留空的status节点，空结果时显示文字。自定义Item正文仍需适配器提供可访问名称。
Scope及其返回Element只在Root同步回调内有效，不能放入应用状态或延迟列表回调。
所有公开Scope调用在回调结束后都会拒绝继续使用。

Trigger搭配Value显示所选值，Content内放Input即可形成弹层内部搜索；
Chips、Chip和ChipsInput可组成对象多选，Chip.HideRemove隐藏移除按钮并保留原生标签键盘操作。
Clear可独立放置，Input返回Root和实际Editor，支持InputGroup的内联/上下addon。

`QueryControlled`保留显式的空Query，不由已选标签覆盖；否则外部单选编辑器在初次挂载及
非编辑期间选值变化时显示所选标签。AutoHighlight支持Off、OnInput和Always。
OnValueChange、OnValuesChange、OnQueryChange、OnOpenChange和OnHighlighted接收当前构建的
ComboboxChange，调用Cancel可否决本次变化；OnOpenChangeComplete在进退动画完成后通知。
程序直接修改应用状态不会伪造用户选值事件。

## Content定位与自定义列表

Content支持Top/Bottom/Left/Right/InlineStart/InlineEnd、Start/Center/End、SideOffset和AlignOffset。
默认在下方、start对齐、间距6DIP，SideOffset指向0可明确取消间距。逻辑侧向随RTL解析。
未提供Anchor时，外部输入跟随实际Editor，弹层搜索跟随Trigger，默认最小宽度为anchor加28DIP；
显式Anchor使用公开Handle，采用源data-chips分支，最小宽度等于anchor。
Chips容器作为Anchor时，应在本帧先Bind再传入Content。
Width可扩大该宽度，MaxHeight默认288DIP；原生AttachTo翻转与可用窗口尺寸共同限制弹层。
Style回调可继续调整当前Element的外观。

Grid开启四方向导航和原生grid/row/cell语义，Row可声明不等长行；GridColumns默认2。
自定义渲染决定布局，禁用项跳过。普通列表继续使用原有的上下键与未高亮输入位置。
Virtualized不自动创建延迟渲染器：应用使用FilteredItems、自己的ScrollState和当前帧Item窗口，
Reveal收到完整过滤集合的索引并滚动到对应项；Item保留完整PosInSet/SetSize和活动项关联。
Gallery演示100项集合只构建约12项的原生滚动窗口。

## 来源与尚待验收

行为对照固定shadcn `bases/base`、Vega及Base UI1.6.0，原文、许可、Git blob和发布包完整性见
[`testdata/shadcn-combobox`](../pkg/control/testdata/shadcn-combobox/README.md)。

With入口使用100ms CSS `ease`淡入淡出和95%缩放，进入时按实际布局侧向移动8DIP；
变换原点根据公开布局测量投影到触发器，不能证明侧向时使用中心淡缩。
关闭立即排除弹层输入和可访问性，原生编辑器与过滤结果保留至退出完成；
快速重开从当前姿态反转并保留编辑选区。外部多选输入立即清空，但退出列表保留原过滤结果。
减少动态效果或窗口隐藏时立即收敛。内部搜索组默认32DIP，文本缩放时可增高。

Content定位、独立槽位、anchor宽度分支、空态/List结构、可取消状态回调、Always、Grid和
调用方持有的虚拟窗口已提供。可选translucent预设、完整locale/Unicode路径仍需补齐与验收。
独立越窗编辑式Combobox、平台IME候选窗、真实读屏及完整视觉矩阵仍需验收；
这些范围不因原生Tester和三平台编译通过而视为完成。

Gallery的Forms & controls页提供可清除框架、国家/城市描述过滤、Office内部搜索、技术多选及
只读/禁用/必填错误示例。搜索Combobox、MultiCombobox或Chips可匹配该页面；选择与草稿跨页保留。
另有Content placement、Destination form、Grid及Large collection操作区，覆盖六侧/对齐/RTL、
对象多选、分组分隔线、隐藏移除、只读/错误、真实表单提交及Dialog内搜索。
固定源的12个独立例子和20个集合例子去重映射见[示例对应表](combobox-examples.json)；
这些配置共用真实操作区，不重复生成32张相同功能的卡片。
