# 桌面浮层：成对 API

按照用户确认的命名，`Menu` 使用独立原生 Flyout 加统一 shadcn 绘制；`NativeMenu` 使用系统菜单，保持系统样式。双组件的区别是外观与交互实现，而不是能否越出父窗口：两者都可以越窗。

## 已实现的组件对

桌面组件入口统一放在 `pkg/desktop`，显式绑定父窗口，不依赖全局当前窗口。

| Shadcn Flyout | 系统样式版本 | 行为 |
| --- | --- | --- |
| `Menu` | `NativeMenu` | 自绘可测量菜单与系统菜单 |
| `DropdownMenu` | `NativeDropdownMenu` | 对应菜单对的下拉命名 |
| `ContextMenu` | `NativeContextMenu` | 指针、Control-click、菜单键、Shift+F10 |
| `Popover` | `NativePopover` | 自绘主题与平台原生 Popover 外观 |
| `Select` | `NativeSelect` | 独立 Flyout 选项与系统菜单选择 |

`NativeMenu` 家族和 `NativeSelect` 强制走 `ui.SystemMenus`，并在同一帧恢复调用方的菜单策略；不会因为预览设置而偷偷改成自绘菜单。`NativePopover` 在 macOS 使用 NSPopover 材质/箭头与平台主题；其他平台使用 MyGo 的原生 Flyout 和该窗口的平台默认主题，不声称存在对应的操作系统通用 Popover 控件。

自绘 Menu 的内容实际位于 Flyout 子窗口，使用 shadcn 颜色、字体、圆角与状态反馈。子菜单以可测量的相邻面板呈现，可使用键盘、typeahead、禁用项、复选/单选和长菜单滚动；原生窗口负责显示器工作区约束。它没有用父窗口内的绝对定位覆盖层冒充越窗实现。

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

## 窗口内与迁移

`pkg/overlay` 保留 `InWindowDropdownMenu`、`InWindowContextMenu`、`InWindowMenubar`、`InWindowPopover`、`InWindowTooltip`、`InWindowHoverCard`，专门用于明确需要父窗口内表面的场景。第一版同名简写已标记 Deprecated。没有 Host 时，桌面组件可以绘制关闭状态的触发器，但尝试打开会报错，不会静默退回窗口内浮层。

Tooltip、HoverCard、DatePicker 等其余弹出家族的完整成对迁移仍待后续阶段。`Native*` 不会用一个自绘窗口冒充运行时不存在的操作系统控件；适配范围必须在实现与文档中明确。

## 验证范围

- 运行 `go run ./examples/native-popups`；macOS 可用 `make popups-app` 生成小窗口对照应用。
- 已在 macOS 实际观察 Menu 的 shadcn Flyout 越过父窗口下边界、NativeMenu 的系统菜单样式，以及自绘 Popover 同时越过父窗口右侧/下边界。子窗口 Done 关闭后恢复父窗口焦点。
- 自动测试验证独立 Context、系统与主题路由、锚点、无空闲帧循环、关闭保护、选择回传、键盘、子菜单与自定义选项。
- Plain Flyout 已增加可逆的短退场、原生外部/Escape 关闭拦截、减少动画立即关闭、迟到回调保护与所有者清理。NativePopover 保留操作系统动画。真实平台的全部关闭路径、多显示器/缩放边界和 Windows/Linux 运行矩阵仍需验收。
- `DismissInterceptor` 是扩展 Surface 能力：标准 MyGo 后端支持保留窗口完成退场；显式注入且没有该能力的后端保持立即关闭，避免留下无法驱动的子窗口。无窗口测试不能替代真实平台验收。
