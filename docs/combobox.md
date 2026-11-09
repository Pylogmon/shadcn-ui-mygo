# Combobox

`ComboboxWith`和`MultiComboboxWith`使用真实MyGo文字编辑器，区分已选值、显示标签和搜索文本。
应用持有状态；组件使用稳定的`Option.Value`标识建议和已选标签。

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
过滤后空组及首尾/连续分隔线被归一化；这不代替独立的compound Group/Collection结构。
应用更新Options即可替换建议；请求调度、加载与数据竞争处理由应用负责。

`Field`、`LabelFor`、`InputGroup`通过实际editor-target关联原生输入，保留标签、描述、错误、
只读复制、原生选择/IME以及焦点外观。`Size`与普通输入使用相同控件尺寸；减少动画服从主题设置。

旧`Combobox(c,*string,ComboboxProps)`保留自由文本和首项高亮行为，旧`MultiCombobox`
保留独立标签行与选择策略；两者继续可用，并修复禁用项、IME及重排焦点缺陷。
需要上述选值/标签/query契约的新代码应使用With入口。

## 来源与尚待验收

行为对照固定shadcn `bases/base`、Vega及Base UI1.6.0，原文、许可、Git blob和发布包完整性见
[`testdata/shadcn-combobox`](../pkg/control/testdata/shadcn-combobox/README.md)。

With入口使用100ms CSS `ease`淡入淡出和95%缩放，进入时按实际布局侧向移动8DIP；
变换原点根据公开布局测量投影到触发器，不能证明侧向时使用中心淡缩。
关闭立即排除弹层输入和可访问性，原生编辑器与过滤结果保留至退出完成；
快速重开从当前姿态反转并保留编辑选区。外部多选输入立即清空，但退出列表保留原过滤结果。
减少动态效果或窗口隐藏时立即收敛。内部搜索组默认32DIP，文本缩放时可增高。

源Content定位参数、独立compound槽位、anchor宽度偏移与空态结构仍待补齐。
完整controlled事件、autoHighlight的always模式、grid/virtualized选项及全部locale/Unicode路径也未计为完成。
独立越窗编辑式Combobox、平台IME候选窗、真实读屏及完整视觉矩阵仍需验收；
这些范围不因原生Tester和三平台编译通过而视为完成。

Gallery的Forms & controls页提供可清除框架、国家/城市描述过滤、Office内部搜索、技术多选及
只读/禁用/必填错误示例。搜索Combobox、MultiCombobox或Chips可匹配该页面；选择与草稿跨页保留。
