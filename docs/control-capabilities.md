# 控件字号与窗口上下文

Button、Select、NumberInput 以主题的 14 DIP 为基准，按原有 Size 字号比例缩放，并遵守系统 `Preferences.TextScale`。Button / Select 的 ExtraSmall 为基准的 12/14，其余为 14/14；NumberInput 的 ExtraSmall、Small、Large 分别为 12/14、14/14、16/14，Default 保留窄窗 16/14、宽窗 14/14 的规则。控件高度会为大字号留出空间，数字编辑器的光标、选择区和 Field 元数据仍位于真实编辑器。

```go
tokens := theme.Light()
tokens.FontSize = 24
c.SetTheme(tokens.Theme())

control.Button(c, "Save", control.ButtonProps{Size: control.Small})
control.NumberInput(c, &app.amount, control.NumberInputProps{
    FontSize: 20, // 局部逻辑 DIP；系统文字缩放仍生效
})
control.Select(c, &app.plan, control.SelectProps{
    FontSize: 20, // 同时作用于触发器和同窗列表
    Options: app.plans,
})
```

`FontSize=0` 使用主题及 Size 比例；正且有限的值覆盖该控件的逻辑字号。负值、NaN 和无穷值按未覆盖处理。`theme.Controller.Apply` 已包含系统缩放，控件及菜单不会再次乘上它；直接使用 `Tokens.Theme()` 的控件也会遵守系统缩放。字体族继续通过 `Theme.Font` / `theme.FontScope` 设置。`theme.ControlFontSize(c, reference, override)` 可供应用自己的控件采用相同规则。

局部覆盖应传入属性：只在 NumberInput 返回的布局根上调用 `.FontSize()` 无法取得内部编辑器；Select 返回的触发器也不能改变独立弹窗的文本。`SelectTriggerProps.FontSize` 可用于组合槽位，`SelectContentProps.FontSize` 可单独覆盖列表。自定义 ItemContent / TriggerContent 中的显式字号继续由应用掌握。`NativeSelect` 的触发器遵守策略，OS 拥有的系统菜单使用系统字体及文字缩放。

```go
// 每个原生窗口拥有独立 Host；在窗口创建后初始化。
app.host = desktop.NewHost(window)

func (app *application) view(c *ui.Context) {
    app.appearance.Apply(c)
    desktop.BindHost(c, app.host)
    app.host.BeginFrame()
    defer app.host.EndFrame()

    _, err := desktop.Select(c, nil, "plan", &app.plan, control.SelectProps{
        Label: "Plan", Options: app.plans,
        OnChange: func(value string) {
            app.summary = app.planNames[value] // 应用自己的业务映射
        },
    })
    if err != nil { log.Print(err) }
}
```

`desktop.BindHost` / `HostOf` 使用当前窗口 Context 的根状态，跨窗口、跨原生弹窗不共享绑定。每帧 view 开头绑定；nil 清除绑定。所有带 Host 参数的桌面组件保留原签名：非 nil 显式 Host 优先，nil 查找当前窗口；未绑定时保留缺少 Host 的错误行为。BeginFrame / EndFrame 仍负责清理本帧消失的锚点，窗口关闭时 NewHost 会关闭其子窗口。

`SelectProps.OnChange` 在用户选择不同值且存储已更新后执行一次。`desktop.Select` 在独立子窗口 UI 事件内立即调用它，因此回调可直接写应用模型；不要捕获父 Context 或 Element。原有 `Changed()` 在下一次父窗口构建中报告一次，可选任一路径派发业务动作。同值重选、程序同步、取消、打开失败、卸载及旧窗口延迟关闭不产生值变化回调。换绑存储、Host 或 key 会关闭旧弹窗。回调可同步重定向所绑定的业务状态。

同窗 `control.Select` 与系统菜单 `NativeSelect` 也提供相同的 OnChange 通知。组合接口保留 `SelectRootProps.OnValueChange` / `OnValuesChange`。桌面 Select 自行设置 ComboBox / PopupListBox、ListBox 和 Option 语义及选中状态；系统菜单保留 MenuButton 角色。

```go
control.NumberInput(c, &app.quantity, control.NumberInputProps{
    Label: "Quantity", Min: &app.minimum, Max: &app.maximum, Step: .5,
    Empty: &app.empty, ReadOnly: app.readOnly,
    Stepper: &control.NumberInputStepperProps{
        Decrement: control.ButtonProps{
            Variant: control.Outline, Icon: outlined.Remove(), Label: "Decrease quantity",
        },
        Increment: control.ButtonProps{
            Variant: control.Outline, Icon: outlined.Add(), Label: "Increase quantity",
        },
        DecrementTooltip: "Subtract 0.5", IncrementTooltip: "Add 0.5",
    },
})
```

`Stepper=nil` 保持原有单编辑器布局；`&NumberInputStepperProps{}` 使用内嵌在同一输入边框中的加减按钮及默认可访问性标签，已有外层 InputGroup 时不叠加边框。传入 ButtonProps 可选按钮变体、图标、标签、字号和尺寸；未指定尺寸时匹配输入档位。提示自动使用 `desktop.BindHost` 绑定的同窗口 Host，与 `desktop.Tooltip` 共享默认外观及交互；没有 Host 时回退到与 `overlay.InWindowTooltip` 相同的实现。两条路由均设置按钮的可访问性描述。

按钮与 Up/Down 共用 Min、Max、Step、Empty、草稿及回调提交路径。超出边界会钳制到有限端点；无法产生新数值时禁用相应按钮。从空值步进仍以 0 加减 Step，再钳制边界；存在状态回调可拒绝提交或替换存储。无效配置、只读、继承禁用或 IME 组合期间禁用按钮。步进采用最后已提交的值并规范化草稿，保留实际编辑器焦点，不触发 OnSubmit。

Dialog 间距、按钮变体、业务数值/ID 转换继续由应用配置。可运行 [完整示例](../examples/control-capabilities/README.md)，或执行 `go test ./pkg/control ./pkg/desktop ./pkg/theme ./examples/control-capabilities` 验证大字号、即时选择回调、原生弹窗生命周期、空值、边界、只读及 IME。
