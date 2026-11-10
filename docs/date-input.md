# 历法与可编辑日期输入

`pkg/display` 的 Calendar 和日期输入都使用真实的 `time.Time`。历法决定民用年月日的划分与显示；解析规则由应用提供。可运行示例见 [date-composition](../examples/date-composition/README.md)，运行 `go run ./examples/date-composition` 或 `make dates-app`（`build/Shadcn Dates.app`）。

`CalendarWith`、`CalendarMultiple` 与 `CalendarRange` 共用 `CalendarProps.AlignSelf`，作用于整个日历的外层容器，零值保留 `ui.Start`。设置 `ui.Center` 可在横排中垂直居中、纵排中水平居中，并保留自然宽度；普通 `Calendar` 的默认布局保持不变，返回元素仍可链式配置对齐。

## CalendarSystem 与 PersianDate

`CalendarProps.System` 默认为 `CalendarGregorian`。`CalendarPersian` 对应固定 shadcn 示例 `calendar-hijri.tsx` 实际导入的 `react-day-picker/persian`：它使用 Persian/Jalali 太阳历。该文件名不表示伊斯兰阴历。

实现依据固定的 `react-day-picker@9.8.1` 和 `date-fns-jalali@4.1.0-0`，转换采用后者 `_lib/jalali.js` 的 **33 年算术算法**。它没有调用天文观测历法或系统 ICU 转换。原文、MIT 许可、包完整性和可再现的 JavaScript oracle 位于 [固定来源档案](../pkg/display/testdata/shadcn-calendar/README.md)。

```go
persian, err := display.PersianDateOf(date)
midnight, err := (display.PersianDate{
    Year: 1404, Month: 3, Day: 22,
}).Time(location)
```

`PersianDateOf` 按 `date` 自身 Location 的民用年月日转换，不改变原始 instant 或 Location。零时间表示未选值，不能转换为 PersianDate；超出范围也返回错误。`PersianDate.Time` 接受完整的正 Persian 年 **1–275138**、月 1–12，以及该月有效的日；此范围位于固定源 JavaScript Date 可表示的完整年份内。无效日期返回错误，不会自动溢出到下个月。

`Time(location)` 返回该民用日的本地午夜，`nil` Location 使用 `time.Local`。时区变更若使整个民用日期不存在，转换返回错误；普通 DST 重叠或午夜跳变沿用 Go `time.Time` 的行为。转换函数本身不携带原来的时分秒，需要保留时钟时应继续持有原始 `time.Time`。

## 方向、周与本地时钟

`CalendarPersian` 默认使用 Farsi 月名、星期名、Eastern Arabic-Indic 数字，星期六为一周起点，方向为 RTL。`CalendarProps.Direction` 可显式覆盖方向，`WeekStartsOn` 可覆盖周起点。Gregorian 在这两个字段为空时分别继承父方向和区域周起点。Persian 周数按固定源的周起点与包含一月一日的第一周计算；既有 Gregorian 周数保持 ISO 行为。

`CalendarProps.Location` 控制 Calendar 的本地民用日计算；未显式设置时沿用既有 Calendar 的选择值/本地区域规则。选择另一天保留原值在该 Location 的时分秒和纳秒，浏览月份或切换历法不会选择新日期。单选、多选、范围、约束与 DST 处理继续共用原来的真实时间路径。`Month`、`MinDate`、`MaxDate`、`StartMonth`、`EndMonth` 都仍是 `time.Time`，无需传入另一种日期对象。

Persian 的 Home/End、PageUp/PageDown、Shift 加翻页键和月份/年份菜单使用 Persian 月年边界；左右方向键与导航按钮遵循实际方向。自定义日期单元仍接收真实 `CalendarDay.Date`。原生字体回退与辅助功能状态属于平台表现，不能据此宣称与网页字体或 DOM 逐字相同。

## DatePickerInput 的状态与解析

`DatePickerInput` 组合真实原生输入框、图标触发器与 `CalendarWith`。应用持有选中日期和地址稳定的 `DatePickerInputState`：

```go
type DatePickerInputState struct {
    Text  string
    Month time.Time
    Open  bool
    Error string
}
```

首次挂载前由应用显式初始化 `Text`；例如选中值是 June 12, 2025 时，可将它初始化成 `"June 12, 2025"`。日期不会在每次构建时覆盖正在编辑的原始文本。应用也可控制 `Month` 与 `Open`，而组件会在解析、浏览、选择和关闭时回写这些状态。调用方应将该 state 存在窗口或页面状态中，避免在每次构建时创建临时值。

`DatePickerInputProps.Parse` 是必填的 `func(string) (time.Time, error)`。文本提交到原生编辑器后，组件以原始文本调用当前构建的 parser；应用决定语言、空值规则、参考日期、时区及错误文案。返回的有效时间保留 parser 产生的 instant、Location 与时钟。Calendar 约束拒绝的日期会被视为错误；成功返回零时间可以清空选值。

解析失败时，原始草稿和 `state.Error` 保留，输入框暴露 Invalid，最近一次有效的日期和显示月份保持可用。解析成功会清除 Error，并把有效日期对应的月份写入 Month。程序直接修改 Text 不等于产生一次原生编辑提交；受控应用同时更新 Text 和日期时应自行执行其解析/校验规则。

`Format` 决定从日历选择后发布的文本。未提供时，Gregorian 使用 `January 02, 2006` 格式，Persian 使用固定 Farsi 完整日期标签。应用应使用与自身 parser 一致的 formatter；组件不会假设任意自然语言 parser 都能识别默认标签。

`OnChange` 仅在接受的日期发生变化时调用，不用于逐字文本通知。`parts.ValueChanged` 区分日期发布；`parts.Root.Changed()` 还包含 InputGroup 传递的原始文本变化。应用回调只从当前构建读取，挂载运行状态不保存 Context、Element 或应用回调。

## Field、真实 Editor 与键盘

返回值 `DatePickerInputParts` 暴露 `Root`、`Editor`、`Trigger`、`Popup` 和 `ValueChanged`。`Root` 可直接返回给 `control.Field`，`Editor` 是保持挂载的真实原生编辑器，浏览日历时不会用文本节点替换它。

```go
control.Field(c, control.FieldProps{
    Label: "Subscription date",
    Error: state.Error,
}, func() ui.Element {
    parts := display.DatePickerInput(c, &date, &state,
        display.DatePickerInputProps{
            Input: control.InputProps{
                Label: "Subscription query",
                OnSubmit: func(text string) { submit(text) },
            },
            Calendar: display.CalendarProps{Location: location},
            Parse: parseDate,
            Format: formatDate,
        })
    return parts.Root
})
```

输入框原有选区、撤销、Field 标签和 `Input.OnSubmit` 路径保持可用。IME 预编辑期间不解析草稿，也不把组合输入的确认键当作应用提交；原生提交标记文本后再解析。正式 native 测试覆盖公开合成输入的组合/提交流程，尚未把真实 Cocoa 取消组合或屏幕阅读器公告计为通过。

编辑器中无修饰键的 ArrowDown 打开日历。图标触发器可切换开关；Disabled 或 ReadOnly 输入不会打开它。弹出日历启用浏览模式：月导航不会修改文本或日期，选择某天才更新日期、格式化 Text、清除错误并关闭，焦点返回触发器。Escape 关闭并返回触发器；外点关闭保留外部目标焦点。日历自身的方向键、月份与年份键盘操作使用现有 Calendar 行为。

默认定位为 bottom-end，间距 10 DIP、对齐偏移 −8 DIP；显式 `SideOffset`、`AlignOffset` 可适配自然语言示例的 8/0。RTL 使用逻辑偏移。该弹层复用现有公开 Popover/motion 策略，没有新增跨窗口能力或宣称网页 DOM 的完整动画行为。

## 自然语言 recipe 的边界

固定网页示例由应用导入 `chrono-node`，不是 Calendar 内置解析。Go 示例同样通过显式应用 `Parse` 接口组合：精确日期接受长英文日期与 ISO 日期；相对日期示范 `tomorrow`、`next week`、`in N days` 和精确日期，参考时间固定为 Asia/Tehran 的 June 10, 2025。

这个小语法用于演示真实编辑、解析、错误保留、月导航和选择反馈，不覆盖 `chrono-node` 的完整语法。应用可接入自己的本地化解析器、业务参考时钟或服务，无需替换原生 Editor 或 Calendar。Linux/Windows 已验证编译，真实操作系统窗口、IME 与 AT 仍需分别验证。
