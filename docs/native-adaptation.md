# 桌面浮层：成对 API

按照用户确认的命名，`Menu` 使用独立原生 Flyout 加统一 shadcn 绘制；`NativeMenu` 使用系统菜单，保持系统样式。双组件的区别是外观与交互实现，而不是能否越出父窗口：两者都可以越窗。

## 已实现的组件对

桌面组件入口统一放在 `pkg/desktop`，显式绑定父窗口，不依赖全局当前窗口。

| Shadcn Flyout | 系统样式版本 | 行为 |
| --- | --- | --- |
| `Menu` | `NativeMenu` | 自绘可测量菜单与系统菜单 |
| `DropdownMenu` | `NativeDropdownMenu` | 对应菜单对的下拉命名 |
| `ContextMenu` | `NativeContextMenu` | 指针、Control-click、菜单键、Shift+F10 |
| `Menubar` | `NativeMenubar` | 同一 Flyout 的兄弟菜单切换与系统弹出菜单触发器组 |
| `Popover` | `NativePopover` | 自绘主题与平台原生 Popover 外观 |
| `Select` | `NativeSelect` | 独立 Flyout 选项与系统菜单选择 |
| `DatePicker` | `NativeDatePicker` | 独立日期/时间窗口与平台样式容器 |
| `DateRangePicker` | `NativeDateRangePicker` | 隔离未完成范围、完整范围提交 |
| `Tooltip` | `NativeTooltip` | 鼠标穿透、不抢焦点的独立提示窗口 |
| `HoverCard` | `NativeHoverCard` | 延迟预览、跨窗口指针保持、显式键盘进入 |
| `navigation.NavigationMenuWith` | `navigation.NativeNavigationMenu` | 独立富内容导航面板与系统路由菜单组 |

`NativeMenu` 家族和 `NativeSelect` 强制走 `ui.SystemMenus`，并在同一帧恢复调用方的菜单策略；不会因为预览设置而偷偷改成自绘菜单。`NativePopover` 在 macOS 使用 NSPopover 材质/箭头与平台主题；其他平台使用 MyGo 的原生 Flyout 和该窗口的平台默认主题，不声称存在对应的操作系统通用 Popover 控件。

`Select` 与 `NativeSelect` 的关闭触发器都采用主题绘制，共享固定16DIP的SVG下箭头；系统负责后者展开后的菜单。箭头不继承文字字号/行高，禁止压缩并垂直居中，隐藏于无障碍树且不拦截触发器点击。

自绘 Menu 的内容实际位于 Flyout 子窗口，使用 shadcn 颜色、字体、圆角与状态反馈。子菜单以可测量的相邻面板呈现，可使用键盘、typeahead、禁用项、复选/单选和长菜单滚动；原生窗口负责显示器工作区约束。它没有用父窗口内的绝对定位覆盖层冒充越窗实现。

菜单面板使用固定Base/Vega的100ms进退动画，退出时保留测量空间并立即移出输入及无障碍树，同路径重开接续当前姿态。方向与缩放原点采用可观测的原生几何；自定义Backend的可选GeometrySurface、未知side及平台边界见[菜单动画](menu-motion.md)。

`Menubar` 使用调用方持有的 `MenubarState{Open,Active}`，同一个 Flyout 在顶层标题间重锚。左右键与RTL、指针切换、子菜单优先、动态移除及结果只回传一次均已测试；子窗口连续按键可在父窗口尚未刷新前操作新菜单。Tab 关闭并恢复当前触发器，可通过 `MenubarProps.OnTab` 用真实控件 Handle 指定相邻目标。`NativeMenubar` 是系统弹出菜单的触发器组，不是应用级操作系统菜单栏；MyGo 目前没有暴露系统 tracking 期间跨顶层触发器切换的 hook。

## 父窗口与生命周期

```go
// 在 App.WhenReady 中、创建父窗口后：
state.Popovers = desktop.NewHost(parentWindow)

// 每个父窗口的 view 中：
state.Popovers.BeginFrame()
defer state.Popovers.EndFrame()

anchor := control.Button(c, "Menu", control.ButtonProps{Variant: control.Outline})
choice, err := desktop.Menu(c, state.Popovers, "project-actions", anchor,
    &state.MenuOpen, &state.MenuState, desktop.MenuProps{Items: items})
```

`Host` 关闭已消失的锚点、跟踪布局后的锚点位置、同步主题/减少动态效果、处理关闭与焦点恢复，并拒绝旧窗口的迟到关闭回调。Host 绑定明确的父窗口，子窗口接收自己的 `*ui.Context`；回调不得捕获并复用父窗口某一帧的 Context。子窗口可以共享应用状态，但必须在 UI 线程修改。

Host 同步父调用作用域的 LTR/RTL，打开期间方向变化会刷新现有子内容。标准 MyGo 后端支持 `SizeSurface.SetSize(width,height)`：仅宽高变化时原位调整窗口，保留子树和编辑状态；其他不可变窗口选项仍重建。显式注入且不支持该能力的后端保留原来的重建行为。

```go
err := desktop.Popover(c, host, "settings", anchor, &open,
    desktop.PopoverProps{Width: 320, Height: 220, Label: "Settings"},
    func(child *ui.Context) {
        control.Input(child, &name, control.InputProps{Label: "Name"})
    })
```

`theme.NativeScope(child)` 保留完整平台 Theme（字体、焦点尺寸、Tooltip 延迟等）并提供原生颜色的语义别名，可用于操作系统没有对应 API 的受约束 Go 内容；它不把平台主题重建成 shadcn 主题。

`Popover` 自动选择 plain Flyout 与完整 shadcn token；`NativePopover` 选择平台容器与子窗口自己的默认主题。原生样式内容使用 MyGo 的 `ui.*` 控件；不要在该回调中又强制安装 shadcn 主题。

菜单动作既提供 `OnChoose`，也可在下一次父窗口构建中返回结果；应用选其中一条路径处理业务动作，避免自己重复执行。需要外部数据变化刷新子内容时使用 `Revision` 或 `Host.Refresh(key)`。普通空闲帧不会不断重绘子窗口。

Gallery 的日期、日期范围、Tooltip 和作者 HoverCard 已接入组件对，随 Menu style 配置切换。共享的层名称与订阅状态变化会更新子内容版本，避免父窗口已更新而子窗口仍显示旧值。小窗口对照示例同时展示两套 API。

Navigation 页的 Menubar 和导航面板也随配置切换；普通模式支持同窗兄弟菜单及富内容，系统模式使用菜单栏触发器组和系统路由菜单。Forms 页新增范围/多拇指/竖向滑块，Data 页新增带禁用项的边框Accordion、保留笔记内容的Collapsible和可折叠/嵌套Resizable工作区。

Navigation 页的 Project actions、File 和 Project workspace 右键菜单共享 Export 子菜单，可选择 Go source 或 Preset JSON；workspace 支持 Shift+F10。Menu style 同时切换这三处的 Flyout/系统菜单路由，选择结果在 Last action 与通知中显示可读名称。

## 窗口内与迁移

`pkg/overlay` 保留 `InWindowDropdownMenu`、`InWindowContextMenu`、`InWindowMenubar`、`InWindowPopover`、`InWindowTooltip`、`InWindowHoverCard`，专门用于明确需要父窗口内表面的场景。第一版同名简写已标记 Deprecated。没有 Host 时，桌面组件可以绘制关闭状态的触发器，但尝试打开会报错，不会静默退回窗口内浮层。

日期组件对已实现；`NativeDatePicker` / `NativeDateRangePicker` 使用原生 Popover/Flyout 容器与平台主题的受约束 Go 日历，并不声称 MyGo 暴露了系统日期网格。原生版月份/年份下拉默认走系统菜单；`CalendarProps.CaptionSelect` 允许显式注入接收当前日历 Context 的选择器。普通选择提交一次；范围选择在完成有效端点对后才发布；`ShowTime` / `Confirm` 使用 Apply/Cancel 草稿。Footer 接受明确的子 Context，父窗口零参数 Calendar.Footer 不可跨窗口复用。

`Tooltip` 默认立即显示，`HoverCard` 默认延迟700ms打开、300ms关闭；关闭时取消计时，Escape 后不会因触发器仍有焦点而立即重开。指针打开的 HoverCard 保持父窗口键盘焦点；触发器上的 Down/Tab 可以进入内容，Escape/反向 Tab 返回。`NativeTooltip` 和 `NativeHoverCard` 使用平台主题的普通原生 Flyout：MyGo 没有暴露系统 Tooltip/HoverCard 控件，macOS NSPopover 会抢焦点，也不适合指针预览。NativeTooltip 保留显式反色配置，缺省时沿用 MyGo 的平台文字/背景回退。

键盘进入的窗口交接期间，父窗口和子窗口的绘制活跃状态可能短暂都未更新。`Host.FocusState(key)` 通过可选 `FocusSurface.IsFocused()` 查询实际键盘所有权；`Host.ApplicationState()` 通过可选后端 `ApplicationActive() bool` 区分应用交接与失活。标准 MyGo 后端提供两种查询，未知后端保留明确的 unknown 状态，不把继承的活跃外观猜成真实焦点。

HoverCard 仅在尚未确认的键盘交接期间使用应用活跃状态保持窗口；确认后按实际焦点或指针保持。两窗指针均离开、实际子焦点已丢失时按关闭延迟退出；原生关闭早于 blur 刷新的路径不会抢回父窗口新输入框的焦点。

`Native*` 的平台适配范围在 API 和文档中明确；拥有系统控件 API 的菜单使用系统实现，其余采用上述平台主题和原生窗口能力。

导航内容面板放在 `pkg/navigation`，通过明确的 Host 参数使用上述窗口能力。`NavigationMenuWith` 支持描述链接网格与任意 `Content(child,choose)`；同一窗口在标题间移动并调整实际宽高，350ms尺寸/位置与内容过渡服从减少动画偏好。`NativeNavigationMenu` 仅提供系统路由菜单组，复杂富内容不伪装成系统菜单。原 `NavigationMenu(c,items)` 是已废弃的窗口内兼容入口，显式名为 `InWindowNavigationMenu`。新 API 示例及边界见 [navigation.md](navigation.md)。

## 验证范围

- 运行 `go run ./examples/native-popups`；macOS 可用 `make popups-app` 生成小窗口对照应用。
- 已在 macOS 实际观察 Menu 的 shadcn Flyout 越过父窗口下边界、NativeMenu 的系统菜单样式，以及自绘 Popover 同时越过父窗口右侧/下边界。子窗口 Done 关闭后恢复父窗口焦点。
- 自动测试验证独立 Context、系统与主题路由、锚点、无空闲帧循环、关闭保护、选择回传、键盘、子菜单与自定义选项。
- Plain Flyout 已增加可逆的短退场、原生外部/Escape 关闭拦截、减少动画立即关闭、迟到回调保护与所有者清理。NativePopover 保留操作系统动画。真实平台的全部关闭路径、多显示器/缩放边界和 Windows/Linux 运行矩阵仍需验收。
- `DismissInterceptor` 是扩展 Surface 能力：标准 MyGo 后端支持保留窗口完成退场；显式注入且没有该能力的后端保持立即关闭，避免留下无法驱动的子窗口。无窗口测试不能替代真实平台验收。
