# 数值、日期、时间与文件输入

`NumberInput`、`DateInput`、`TimeInput`、`FileInput` 直接绑定应用拥有的数值、时间或文件路径；原有字符串 `Input` 的密码、搜索、只读、选区、撤销与 IME 行为保留。`StyleInput` 可把相同的主题、尺寸、禁用和错误外观应用到公开的原生编辑器根节点。

```go
// 状态保存在应用模型中，不在每次 View 中重新初始化。
amount := 123.0
minimum, maximum := 0.0, 10000.0

control.Field(c, control.FieldProps{
    Label: "Amount", Description: "Enter a unit price",
}, func() ui.Element {
    return control.InputGroup(c, control.InputGroupProps{Prefix: "USD"},
        func() ui.Element {
            return control.NumberInput(c, &amount, control.NumberInputProps{
                Min: &minimum, Max: &maximum, Step: .5,
                OnSubmit: func(value float64) { saveAmount(value) },
            })
        })
})
```

数值输入使用与 `Input` 相同的公开原生文本编辑器。可编辑 draft 与已提交的 `float64` 分开：符号、未完成指数或超出范围的文字保留在编辑器中，不覆盖最后有效值；有效、有限且在范围内的编辑触发 `OnChange`。Enter 和失焦规范化文字，无法提交的 draft 恢复为应用值。应用修改数值会更新 draft。默认 Step 为 1，Up/Down 递增并按可选上下界限制；Step 不决定输入精度，也不提供浏览器的 step-mismatch 验证。无效范围或非正、非有限 Step 会使编辑器只读并标记错误；非有限应用数值标记错误。

`NumberInputProps.Handle` 指向实际文本编辑器，返回的 Element 是布局根。Field、LabelFor、InputGroup 把标签、描述、错误和 addon 焦点关联到同一实际编辑器；同时提供上下界时，Range/Step 元数据也位于该编辑器。它保留原生文本选择、IME 和撤销，角色仍为文本字段。可用 `Stepper: &control.NumberInputStepperProps{}` 添加加减按钮；不提供浏览器约束验证气泡。

`NumberInputProps.FontSize` 直接覆盖真实编辑器字号，并保留系统文字缩放；0 使用主题和 Size 比例。Stepper 可分别配置 Decrement / Increment 的 ButtonProps（变体、图标、可访问性 Label）和 Tooltip；按钮复用键盘的范围、空值及草稿提交，边界、只读、禁用和 IME 组合期间统一禁用。配置示例见 [通用控件能力](control-capabilities.md)。

需要可选数值时，传入应用持有的 `Empty *bool`。`true` 显示真实空编辑器与 Placeholder；清空只改变 Empty 并保留最后数值，输入0则得到存在的0。未提供 Empty 时保留原来的必有值行为。`OnEmptyChange` 只通知用户引起的存在状态转换，应用外部同步不会触发；`OnChange` 只通知有效数值提交。从空变为数值会分别通知 `OnEmptyChange(false)` 与 `OnChange(value)`，应用若统计编辑次数，应分别处理存在与数值通道。空值Enter不提交，Up／Down从0开始再加减Step并遵守边界。空值不暴露有值的Range元数据。

两个回调均可同步重定向模型。`OnEmptyChange(false)` 若恢复Empty=true，将保留旧数值；回调主动写入的新数值优先保留。失焦／无效提交恢复当前空值或有效模型，不将空值伪造为0。原生选择、IME候选提交和撤销仍由同一编辑器处理。

```go
control.DateInput(c, &app.date, control.DateInputProps{Label: "Date"})
control.TimeInput(c, &app.time, control.TimeInputProps{Label: "Time"})
control.FileInput(c, &app.paths, control.FileInputProps{
    Label: "Picture", Parent: app.window,
    Filters: []mygo.FileFilter{{Name: "Images", Extensions: []string{"png", "jpg"}}},
    OnChange: func(paths []string) { app.selectionChanged(paths) },
})
```

日期使用 MyGo 的本地化日期按钮和日历，支持系统键盘选择和 Escape 取消；更改日期保留 `time.Time` 的时钟字段与 Location。时间使用原生小时、分钟及 AM/PM 分段，按用户的 12/24 小时制显示；数字、Up/Down 和段间左右键保留原生行为。它们绑定明确的 `time.Time`，不把浏览器日期/时间的空字符串当作独立状态。

Time 的 Handle、Invalid、描述和错误关联位于原生 group。Field 和 LabelFor 同时设置标签值与 LabelledBy，内部段会继承可读的标签前缀。SDK 未公开段部件、每段的 Field 元数据、指定段焦点或只读能力，因此 Time 不暴露 ReadOnly/AutoFocus；InputGroup 可以组合外观，但点击前缀不能指定内部时间段。内部段的焦点高亮由 SDK 负责。Date 支持实际按钮的 Handle 和 AutoFocus。

文件输入是主题化原生按钮，用户通过点击或键盘激活后才调用 `mygo.Dialog.Open`。对话框在 View 外运行；结果通过值和弱 Services 请求新构建，再使用当前属性中的 `OnChange` / `OnError`。取消保留原选择，等待时禁用并标记 Busy；卸载或换绑状态会丢弃旧结果。Filters 与回调路径均复制，默认单选只保留首个路径，Multiple 支持多选。应用也可提供长期拥有的 `FilePicker`，它只能携带值与原生 Window，不能捕获 Context 或 Element。绑定值是桌面本地路径，不是浏览器 File/Blob。

固定源实际使用的 email、tel、URL 继续使用字符串 Input；示例由应用校验 email/URL 并显示 Field 错误，不承诺网页自动填充、移动键盘提示或浏览器约束气泡。`StyleInput` 只改变实际传入根节点的外观，不会取得复合 SDK 控件的私有编辑器部件。

运行 `go run ./examples/typed-input` 或 `make inputs-app`。Input Types、Basic and Field states、With controls、Form 四个区域覆盖固定源输入类型、独立文件例、Field/InputGroup、选择器、搜索按钮与表单校验/提交/取消。可用 `-section types|states|composition|form`、`-dark`、`-snapshot /path/input.png` 生成实际原生画面。来源、18 份原文摘要与原生边界见 [源归档](../pkg/control/testdata/shadcn-input/README.md)；真实系统文件选择器、输入法候选窗和辅助技术仍属于平台验收。
