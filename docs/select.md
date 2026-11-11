# Select 组合

`SelectRoot` / `SelectScope` 提供独立 Trigger、Value、Content、List、Group、Label、Item、Separator 和滚动箭头。`SelectWith` 使用相同核心提供标准布局；原有 `Select(*string, SelectProps)` 保留，新增 `MultiSelect(*[]string, SelectProps)`。这些接口使用无编辑器的原生 listbox，选多项时保持展开。

```go
// state 保存在应用中，挂载期间地址稳定。
state := control.SelectState{
    Value: control.SelectSelection{Value: "pro", Valid: true},
}
control.SelectWith(c, &state, control.SelectWithProps{
    Root: control.SelectRootProps{Options: []control.Option{
        {Value: "free", Label: "Free"},
        {Value: "pro", Label: "Pro"},
    }},
    Trigger: control.SelectTriggerProps{Label: "Plan"},
})
```

`SelectSelection.Valid` 区分未选择与有效的空字符串；`UnselectedValue` 可显式指定一个清空选值的项目。对象保存在应用模型中，用稳定字符串 identity、ItemContent / ValueContent 组合固定源 Plan 的富内容。多选使用 `SelectRootProps.Multiple` 与 `SelectState.Values`，不会创建搜索输入框。

Root 不添加视觉布局，Scope 只在当前构建回调中有效。手写 Content 需显式组合 List、Group / Item；传入 Options 可在关闭状态提供标签与 typeahead，未传入时按实际 Item 注册集合。所有渲染和状态事件使用当前构建回调；持久状态仅保存值、几何和 Handle。`FormValues` 返回当前应用状态的副本，适合应用自己提交表单。

默认 Content 把选中项的实际文字对齐到触发器值，并提供原生有界滚动、选中项 reveal 和 hover 滚动箭头；此配置依源包装器关闭动画。`AlignItemWithTrigger` 显式为 false 时使用普通六侧、RTL、alignment/offset 定位和既有100ms保留退出动画。透明菜单使用已提交的真实场景材质和即时切换配置。MaxHeight、Width、Anchor 与 Style 提供当前内容配置。

逐项对齐使用实际窗口内的拟合位置，长列表展开保持窗口边距。键盘打开／切换活动项时，定位会避开覆盖在列表上的上下滚动箭头；End／Home目标可用真实指针点击。该定位请求只执行到目标可见，随后用户可自由滚动，不持续锁定活动项。

弹层打开期间缩小／恢复窗口或修改MaxHeight，会重新限制已测量的宽高和位置，并重新露出当前活动项一次。显式禁用的Trigger仅应用单层50%透明度；继承禁用仍由原生祖先负责绘制和输入阻止。

新 Root 保留 source 的无循环键盘导航及可聚焦的 aria-disabled 项（不可提交）；旧 Select 便捷入口保留跳过 disabled 的行为。触发器使用原生 ComboBox / PopupListBox 语义，消费测试或自动化应按新角色定位。便捷 Select / MultiSelect 允许每次构建重建临时绑定，挂载身份沿用原生位置/调用方 key；显式 Root 的状态地址仍需稳定。

`NativeSelect` 继续使用 OS 单选菜单及系统绘制和定位，不接入任意 ItemContent 或多选状态。新组合接口是同窗控件，独立窗口适配器继续使用 desktop API。指针操作采用原生按钮完整 click；未重造 DOM mousedown 打开后的拖曳释放选择机制。真实触摸、IME候选窗口、跨显示器与辅助技术仍需平台验收。

运行 `go run ./examples/select-composition` 或 `make selects-app`，三个操作区覆盖22份固定源例的实际配置及提交、Dialog、NativeSelect 流程。来源映射与边界见 [示例](../examples/select-composition/README.md)。

Gallery另有逐项映射的22个源配置。源列表使用当前窗口中的完整组合API；配置／语言菜单随Gallery Flyout／Native切换。`desktop.Select`目前仍是单字符串菜单适配器，不提供富多选或逐项对齐的跨窗槽位，这一缺口未计为完成。

便捷入口支持 `SelectProps.FontSize`，同时覆盖触发器和列表；`Content.FontSize` 可单独覆盖列表。`SelectProps.OnChange` 在存储更新后通知不同的用户选值；`desktop.Select` 从独立弹窗中立即通知，原 `Changed()` 继续在下一帧报告。窗口可用 `desktop.BindHost(c, host)` 绑定 Host，并对 Select 传 nil 继承；完整代码及生命周期约定见 [通用控件能力](control-capabilities.md)。
