# Questionnaire

`QuestionnaireWith`实现固定源中回答或显式跳过的逐步问卷。状态由应用持有；默认构建提供Progress、Item、Title、Description、Choice/Input、Error与导航按钮，也可通过Parts组合Card、Dialog或自定义进度内容。

```go
conversation.QuestionnaireWith(c, &form, conversation.QuestionnaireProps{
    Shortcuts: conversation.QuestionnaireLetters,
    Questions: []conversation.Question{
        {
            ID: "project", Title: "你准备构建什么？", Required: true,
            Choices: []control.Option{
                {Value: "desktop", Label: "桌面应用"},
                {Value: "tool", Label: "开发工具"},
            },
        },
        {
            ID: "notes", Title: "还有什么需要考虑？",
            Freeform: true, InputLabel: "补充说明",
        },
    },
    OnSubmit: func(answers map[string]conversation.Answer) {
        // 保存应用数据；回调接收独立副本。
    },
}, nil)
```

把`form`保存在窗口或页面状态中。稳定且唯一的Question.ID保持重排后的部件身份；省略时使用位置ID。每个挂载问卷使用独立状态。Parts及ItemParts仅在本次构建有效，不保存Context、Element或Parts。

Required题不能跳过。可选题也必须回答或显式Skip，直接Next/Submit会显示错误并聚焦问题中的实际原生控件。Disabled题不进入进度、导航与结果。Inactive题仍挂载，隐藏并退出输入和无障碍导航，原生编辑器的草稿、选区与撤销状态因而能够保留。

单选题的固定选项与自由输入互斥；选项切换保留未激活的文字草稿，提交时排除该草稿。多选题可同时提交多个固定选项和输入。Answer.Mode区分自动兼容值、Choice、Input、Combined和Draft，应用直接构造复杂默认答案时可显式选择。Skipped排除整题，修改答案会撤销跳过状态。`Result`返回清理后的Answer副本，`FormValues`按实际控件顺序输出成功值，并排除未选择、禁用或跳过的控件。

`DefaultItem`用于首次打开和Reset；`Item *string`与OnItemChange用于受控步骤，应用可以接受或拒绝导航请求。`DefaultAnswer`、Choice.DefaultChecked和Input.DefaultValue设置默认值，Reset读取最新默认值；Choice.Checked及Input.Value提供受控值。OnReset及按钮OnClick可以PreventDefault。`State.JumpTo`供非受控恢复或应用的复查入口使用；受控步骤应修改Item值。从Questions显式删除题目会清理其回答、错误与运行时注册，再加入同ID时使用默认值。整个问卷离页卸载保留调用方持有的Answers，支持原生应用跨页恢复。

自定义构建通过`Parts.Item(id, ...)`提供每题内容，用Item.Header、Title、Description、Choices、Choice、Input、Error组合；Parts.Progress、Actions、Previous、Skip、Next、Submit、Reset提供独立部件。所有步骤应保持构建，仅让组件处理inactive状态。复查、结果编辑和条件问题列表由应用用受控步骤、状态及Card/Dialog组合，固定源没有单独Review/Edit部件。

Choice使用实际原生Radio或Checkbox，可以提供描述、装饰性Content、Handle及可取消OnChange。空字符串也是有效选项值。Input默认使用原生文本编辑器，支持ReadOnly、Password、Search及Handle；Editor可返回应用自定义的真实可聚焦原生编辑器，例如Textarea。内部字符串模型持续存在，保留原生IME、选择与编辑行为。

同题可以声明多个Input。可选的`QuestionnaireInputProps.Key`在重排或条件内容中保持编辑器身份，同题显式Key必须唯一；省略时按未命名输入的声明顺序分配独立槽位，第一项继续使用兼容的`Answer.Text`。独立草稿及选中状态保存在`Answer.Inputs []QuestionnaireInputValue`，固定选项仍保存在Values。单选题选择一个输入或固定选项，多选题可以同时选择多个输入与固定选项。`Result`只返回成功输入的独立副本，`FormValues`按实际声明顺序输出各控件；不同输入之间、输入与固定选项之间的相同文本都保留。

禁用输入保留草稿并退出成功值，移除槽位清理对应注册值；显式DefaultAnswer.Inputs的Selected标记（包括非空且未选中的草稿）在Reset时保留，非空Input.DefaultValue可以提供自己的默认值。Input.Validate失败后，原生焦点回到实际失败的编辑器。运行`go run ./examples/questionnaire-inputs`试用两输入与checkbox、同值提交、保存快照和Reset；说明见[独立输入示例](../examples/questionnaire-inputs/README.md)。所有值仍属于同一个Question ID，并未引入额外表单schema。

Input.OnInput的PreventDefault取消回答的激活状态更新。对于非受控输入，原生已编辑文字仍保留；之前未激活的输入仍是草稿，之前已激活的输入仍参与结果。受控Value始终由应用决定。OnChange提供接受编辑时的简便值通知。

可配置Letters的A–Z或Numbers的1–9选择快捷键，禁用选项不占用名额。Question.Choices为nil时按实际注册的原生Choice顺序分配；非nil时以定义集合为准，显式空slice不生成选项快捷键。Enter确认已选或已填答案，Cmd/Ctrl+Enter在当前问卷内确认。Up/Down在选项和空输入间移动；实际输入的光标、单选方向键及IME组合优先保留原生语义。快捷键仅作用于焦点所在问卷，不能抢占其它表单或外部编辑器。

Question.Validate、Input.Validate和ValidateAll分别提供单题、输入与跨题Go校验，返回错误文字。外部Invalid/Error可表达服务端错误。验证失败会返回首个无效题，并把Error关联到问题与实际控件；原生RoleAlert处理错误播报，Progress变化另作一次polite通知。实际读屏验证仍须进行。

浏览器的email/date/number等HTML类型、ValidityState和约束验证气泡没有MyGo公开等价API；可用Go校验与自定义Editor表达应用需要的输入。隐藏内容中的无条件显式Focus/AutoFocus同样受原生运行时限制，自定义编辑器应只在Item.Active时请求焦点。完整源示例、字体/主题/缩放、跨平台、读屏与视觉矩阵继续跟踪。

旧`Questionnaire`及State.Next/Submit/Skip保留原便利向导的required-only验证契约。新的源组合入口使用上述严格导航语义；现有调用方可按所需行为迁移。

固定primitive的完整14份源码、测试、说明和MIT许可归档于`pkg/conversation/testdata/questionnaire`，校验与来源见[固定上游](upstream.md)。
