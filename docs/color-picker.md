# Color Picker

`control.ColorPicker` 绑定调用方持有的 `ui.Color`，默认提供参考图的八种预设颜色。选中项显示对比色勾选和间隔外圈；重复点击保持选中。

```go
var accent = ui.Hex("#193cb8")

control.ColorPicker(c, &accent, control.ColorPickerProps{
    Label:  "Accent color",
    Shape:  control.ColorSwatchRounded,
    Custom: true,
    OnValueChange: func(color ui.Color) {
        // 用户操作后，accent 已更新；重绘和外部赋值不会触发此回调。
    },
})
```

`Custom: true` 在预设前方显示同尺寸的彩色色块，默认关闭。点击或按 Enter / Space 弹出 MyGo 自带的 `ui.ColorPicker`，支持饱和度／亮度、色相、透明度、Hex 输入与常用颜色。选色实时写回，Escape 或点击外部关闭面板并保留当前值。面板在当前窗口中绘制，无需 `desktop.Host`。`CustomLabel` 可配置其无障碍名称，默认 `Custom color`。

| 配置 | 行为 |
| --- | --- |
| `Shape: ColorSwatchRounded` | 默认圆角方形；48 DIP 时圆角为 14 DIP |
| `Shape: ColorSwatchSquare` | 方形 |
| `Shape: ColorSwatchCircle` | 圆形 |
| `SwatchSize` | 色块边长，默认 48 DIP；自定义色块与预设一致，选中／焦点／错误外圈的间隔和线宽随尺寸同比缩放 |
| `Spacing` | 色块间距指针，默认 24 DIP，可显式设为 0 |
| `AlignSelf` | 整个控件在父容器交叉轴上的对齐，默认 `ui.Start`，可设 `ui.Center`、`ui.End` 或 `ui.Stretch` |
| `Colors` | `nil` 使用默认八色；空切片不显示预设 |
| `Disabled` / `ColorOption.Disabled` | 禁用整个组件／单个预设；同时继承祖先禁用状态 |
| `Invalid` | 错误外圈；也读取所在 Field 的校验状态 |

自定义预设使用 `[]control.ColorOption`；空名称自动使用 Hex。`DefaultColorOptions()` 每次返回独立切片，可安全修改。

```go
colors := []control.ColorOption{
    {Color: ui.Hex("#2563eb"), Label: "Blue"},
    {Color: ui.Hex("#16a34a"), Label: "Green"},
    {Color: ui.Hex("#f9731680"), Label: "Translucent orange"},
}
spacing := float32(16)
control.ColorPicker(c, &accent, control.ColorPickerProps{
    Colors: colors, Custom: true,
    Shape: control.ColorSwatchCircle, SwatchSize: 32, Spacing: &spacing,
})
```

预设使用 MyGo 原生单选语义：Tab 进入当前选中项，左右箭头切换，Home / End 到首尾，跳过禁用项，并支持 RTL。自定义色块是独立的 Tab 焦点。预设按颜色值维持稳定 Key，重排不会丢失焦点；列表内多个组件仍需调用方使用 `c.Key(id)` 区分。窄容器中的预设自动换行；外圈绘制在色块外，建议容器保留至少 12 DIP 内边距。

返回元素的 `Changed()` 与 `OnValueChange` 在实际颜色发生变化的输入帧报告一次。外部状态可以是不在预设中的颜色，此时有自定义色块则显示为选中，未启用时所有预设均不选中。

运行 `go run ./examples/color-picker`，或在 Gallery 搜索 `ColorPicker` / `颜色选择`。演示支持三种形状、可选自定义色块、禁用状态、RGBA 读数和跨页状态保留。
