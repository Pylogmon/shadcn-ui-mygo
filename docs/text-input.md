# 原生文本输入

`control.Input` 与 `Textarea` 返回真正的 MyGo 原生编辑器。选择、剪贴板、撤销/重做、组字、密码保护和光标滚动由同一个编辑器处理；调用方持有字符串状态。

```go
control.Field(c, control.FieldProps{
    Label: "Display name", Error: nameError,
}, func() ui.Element {
    return control.Input(c, &name, control.InputProps{
        Placeholder: "Your name", MaxLength: 40,
        OnChange: func(value string) { nameError = validateName(value) },
        OnSubmit: func(value string) { saveName(value) },
    })
})
```

默认 Input 高度为 36 DIP；XS/Small/Large 分别为 24/32/40 DIP。输入占满可用宽度，并允许收缩。默认字体在宽度 768 DIP 以下为 16、以上为 14；系统文字缩放会相应增高输入框，避免裁切原生光标。Vega 的浅色背景透明，深色使用 Input token 的 30% 透明度；错误和键盘焦点使用 3 DIP 外环。

`MaxLength` 使用原生编辑器的 UTF-16 单位限制，覆盖输入、粘贴和 IME 提交，零表示无限。应用直接替换字符串不会被截断，也不会触发 `OnChange`。`OnPaste` 返回 true 消费本次粘贴；返回 false 保持原生选择、插入和撤销。`OnTextKey` 接收原生非文本按键，可通过当次 `TextKey.SetText` 修改文字并消费该键；组字期间交由原生输入法处理。

`OnSubmit` 只报告普通回车。用于提交 IME 候选文字的同一次回车不会同时提交表单；随后再按回车才提交。`Password` 使用原生受保护文本，不向可访问性值或复制操作暴露内容；`Search` 使用原生搜索框语义。`ReadOnly` 保留选择与复制。

```go
control.Textarea(c, &notes, control.TextareaProps{
    Label: "Notes", MinRows: 2, MaxRows: 6,
})
```

Textarea 默认最小高度 64 DIP，按换行和折行内容增长；`MaxRows: 0` 不额外限制增长，`Rows > 0` 固定原生行数。达到显式上限后，原生编辑器滚动以保持光标可见。用户拖拽/键盘调整高度使用 [TextareaWith](textarea.md)，保留原 Textarea 的原生编辑器契约。

`Field.Error` 在构建编辑器时作用于错误外观，并关联原生错误描述。`InputGroup` 把环与边框放在组合根节点，保留内部编辑器自己的 Disabled/Invalid、原生回调和选择状态；独立输入不继承相邻字段的错误状态。

本阶段验证了实际像素、键鼠、剪贴板、原生 IME 客户端和光标几何。真实中文候选窗、辅助技术播报和完整跨平台视觉矩阵仍需验收；对照基线见 [upstream.md](upstream.md)。
