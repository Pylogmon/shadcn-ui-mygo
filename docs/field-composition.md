# Field 独立组合

`control.FieldWith` 允许按当前界面所需的顺序组合 Label、Content、Description、真实 Control 与 Error。旧 `Field`、`FieldSet`、`FieldGroup` 便利入口保持现有布局；新的组合入口对应固定 Base/Vega 的 12 DIP Field 间距、4 DIP Content 间距和两种 Legend 字号。

```go
control.FieldWith(c, control.FieldProps{
    Orientation: control.FieldResponsive,
    Required: true,
}, func(f *control.FieldScope) {
    f.Content(c, func() {
        f.Label(c, control.FieldLabelProps{Text: "Name"}, nil)
        f.Description(c, "Provide your full name for identification")
    })
    f.Control(c, func() ui.Element {
        return control.Input(c, &model.Name, control.InputProps{}).Grow(1)
    })
})
```

`Content` 把 label 与 description 放入同一个弹性列。水平与响应式横排时，这一列位于实际编辑器或 Switch 的旁边；窄容器回到纵排。响应式阈值复用 `FieldGroup` 的实际测量宽度，或显式 `FieldProps.ResponsiveWidth`。原生布局的 `Basis(0)` 会作用于当前主轴；需要同时横排与纵排的内容应保留自然高度，不给编辑器设置零纵轴 basis。

Scope 只在 `FieldWith` 构建回调内有效，离开后使用会 panic。可以先声明 Control，再声明 Label；所有关系在该次构建末尾建立。Scope 不是应用状态，不应保存 Context、Element、scope 或帧处理器。`Control` 显式注册唯一真实目标；InputGroup、TextareaWith 等按既有公开 target 合同关联内部编辑器。多个输入应使用多个 Field。只有动作按钮的布局 Field 可以不注册 Control。

`Label` 的 `Text` 是默认可见文字，`Label` 是自定义原生子内容的可访问名称。回调可组合 `FieldTitle` 与 Badge。整个 Choice 卡片可使用 `Choice: true`，并在回调内声明 `Content` 和 `Control`；`Selected` 由应用传入，决定外框和淡色填充。组件通过 MyGo 公开 `LabelledBy` 把实际控件命名并转发点击，不创建额外的 Radio、Checkbox 或 Tab stop。键盘与 checked 状态仍属于真实选择控件。`FieldProps.Disabled` 禁用整个组。

`Description` 可以在 Control 前，也可以在后；`DescriptionWith` 接受完整可访问摘要和当前构建的原生内容，可包含真实 Link。未显式声明 Description 时，`FieldProps.Description` 在正文末尾作为便利回退。未声明 Label 时，`FieldProps.Label` 只作目标的可访问名称；需要可见标签时应明确调用 Label。Error 默认追加在最后，也可显式调用 `f.Error(c)` 放在任意位置；复用既有单条、多条去重列表和自定义 ErrorContent，并准确关联实际控制。

`FieldProps.Invalid` 允许没有错误正文的验证状态；不会凭空创建 alert。Input、Radio、Checkbox、Switch、Select 等读取既有 Field 外观作用域。Slider 的应用 Invalid 也传给真实 thumb；它与不可能的数值约束分开，验证失败时有效范围仍可用指针和键盘调整。

独立部件包括 `FieldContent`、`FieldLabel`、`FieldTitle`、`FieldDescription` / `FieldDescriptionWith`、`FieldLegend` 与 `FieldSeparator`。独立 Label 的 `Target` 只接受当前构建的真实 Element；Scope 中改用 Control 注册，以支持任意声明顺序。Legend 的 `FieldLegendDefault` 为 16 DIP，`FieldLegendLabel` 为 14 DIP；带文字 Separator 保留真实文本，线条是装饰内容。

`FieldSetWith` 接受 `FieldSetSlots`：Legend、Description 返回实际 Element，Content 接受当前 Context。原生 group 与这两部分建立名称/描述关系；单独 Legend 的后间距为 12 DIP，Legend 后 Description 采用源 -6 DIP 调整，之后才进入 24 DIP 内容间距。原生组合不会注入网页的后代 CSS；源 FieldGroup 的 28 DIP、紧凑选择组 12 DIP、嵌套组 16 DIP 可在返回的原生组上显式设置 Gap。

Gallery 的[源组合说明](../examples/field-composition/README.md)映射 13 个文档示例和 Create 的 10 个分区。固定 shadcn 版本为 `8ab27715521cbd4f87ae6886cd989cbb678f6de7`，已有组件/文档原文及 Vega 样式保存在 [`testdata/shadcn-field`](../pkg/control/testdata/shadcn-field/README.md)，新增应用映射在对应 source-map 中记录原文 Git blob 与 SHA。主题、菜单、校验和保存均是当前应用组合，不发送业务数据。

正式原生流程覆盖响应式重排时的真实编辑/选区保留、密码前置描述、丰富标签、当前错误列表/清除、整卡空白区域激活 Radio 与箭头导航、禁用选择、水平 Switch、FieldSet 关系和单 Legend 间距，以及验证失败但仍可编辑的 Slider。公开 Tester 的角色/关系与像素证据不能替代 Windows/Linux 实窗、真实读屏和所有输入法验收。
