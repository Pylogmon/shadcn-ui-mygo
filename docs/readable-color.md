# 可读彩色文字与可组合 Sidebar

本库增加 `theme.ReadableForeground`、承载表面作用域和独立 `navigation.SidebarMenuButton`。这是原生扩展；官方预设、`Tokens.Primary` 与 `PrimaryForeground` 的值没有修改。可运行的完整消费示例见 [sidebar-composition](../examples/sidebar-composition/main.go)。

## 官方语义与扩展边界

2026-11-12 核对了官方 [Theming](https://ui.shadcn.com/docs/theming)、[Sidebar](https://ui.shadcn.com/docs/components/sidebar)、[Base Sidebar 源码](https://github.com/shadcn-ui/ui/blob/main/apps/v4/registry/bases/base/ui/sidebar.tsx) 及具有显式颜色类名的 [Radix Sidebar 源码](https://github.com/shadcn-ui/ui/blob/main/apps/v4/registry/new-york-v4/ui/sidebar.tsx)。

官方的 Primary / PrimaryForeground 是实色表面及其内容的语义配对。Sidebar 普通内容使用 SidebarForeground，选中及悬停内容使用 SidebarAccent / SidebarAccentForeground，焦点使用 SidebarRing；SidebarPrimary 是另一组实色语义。官方没有自动派生的 PrimaryText token。官方 [Button](https://github.com/shadcn-ui/ui/blob/main/apps/v4/registry/new-york-v4/ui/button.tsx) / [Badge](https://github.com/shadcn-ui/ui/blob/main/apps/v4/registry/new-york-v4/ui/badge.tsx) 的 Link 直接使用 Primary；本库对彩色小字的自动校正、Soft、自定义 Color 和下面的组合扩展不属于官方 token 机制。

原版只读 Manggo 的 `Theme.cpp` 从强调色向黑或白调整，在相关背景上满足 4.5:1 后得到 PrimaryText；`StyleSheetBuilder.cpp` 将它用于选中导航及 Pro 内容，背景仍用 PrimarySoft。这里把该能力实现为按承载表面求值的函数，避免一个全局派生 token 无法覆盖不同颜色的 Sidebar/Card/Popover。

## 统一前景解析

```go
// 任意小字号彩色文字，实际背景由作用域或参数指定。
fg := theme.ReadableForeground(c, tokens.Primary, tokens.Sidebar)
ui.Text(c, "Professional Version").TextColor(fg)

// 同一文字跨多个交互背景时求共同前景，避免背景动画中途失去对比度。
fg = theme.ReadableForeground(c, tokens.Primary,
    theme.Composite(tokens.Primary.Alpha(.2), tokens.Sidebar),
    theme.Composite(tokens.Primary.Alpha(.3), tokens.Sidebar))

// 无 UI Context 的纯颜色计算。background 应为不透明显示 sRGB。
fg = theme.EnsureContrast(accent, background, 4.5)
```

`EnsureContrast` 使用 WCAG sRGB 相对亮度。已经达标的颜色保持原值；否则分别寻找向黑、白的最短混合距离，选择调整较少的一条路径。半透明文字根据 source-over 后的显示颜色判断，必要时增加不透明度。它不改写强调色、填充、边框或 ring。

`ReadableForeground` 默认目标是 4.5:1；系统 HighContrast 开启后以 7:1 为目标。它同时考虑 8 位半透明合成的一步舍入误差。若给定表面无法达到 7:1，选择对比度最高的端点。若多个状态不存在共同前景，使用第一个实际状态背景；本库按钮会在这种少见的自定义表面上停止颜色插值，避免跨越不可读区间。普通反馈继续使用 `motion.ApplyTransition`；没有新增消费侧动画系统。

## 实际承载表面

`Card` / `CardWith`、`Sidebar` / `SidebarWith`、同窗 Popover / HoverCard 和独立窗口 `desktop.Popover` 的内容自动进入对应表面作用域。`theme.SurfaceOf(c)` 返回该作用域的合成色；无作用域时回退到 Tokens.Background。独立 SidebarMenuButton 无作用域时回退到 Tokens.Sidebar。

自定义容器需要声明与绘制相同的背景。作用域不插入额外布局节点，嵌套半透明表面按顺序合成，并在返回时恢复兄弟节点的作用域。

```go
surface := ui.RGBA(40, 60, 85, .8)
ui.Column(c).Background(surface).Children(func() {
    theme.WithSurface(c, surface, func() {
        control.Button(c, "详情", control.ButtonProps{Variant: control.Soft})
        display.BadgeWith(c, "Pro", display.BadgeProps{
            Variant: display.BadgeOutline, Color: &tokens.Primary,
        })
    })
})

// 也可以只为一个控件显式提供承载面；半透明值叠在当前作用域上。
control.Button(c, "详情", control.ButtonProps{
    Variant: control.Soft, Surface: &surface,
})
```

`ButtonProps.Surface`、`BadgeProps.Surface`、`SidebarMenuButtonProps.Surface` 都描述控件下面的表面；控件自己的浅填充及状态背景仍由库合成。普通 ui.Box 的背景不会通过私有 MyGo 节点反射猜测，应使用作用域或 Surface。图像、渐变及系统材质无法由单个颜色完整描述；需要调用方提供实际内容区域适用的背景色。禁用态仍按现有语义降低透明度，不承诺禁用文字达到普通可用文字的对比度。

## MangGo 导航消费示例

```go
t := theme.TokensOf(c)
for _, route := range routes {
    selected := currentPath == route.Path
    p := navigation.SidebarMenuButtonProps{
        Label: route.Title, Active: selected, FontSize: 14,
    }
    if selected {
        p.Color = &t.Primary // 原生扩展；nil 保留官方 Sidebar 配色。
    }
    navigation.SidebarMenuButton(c.Key(route.Path), p, func(c *ui.Context) {
        // 这些尺寸、字重和对齐方式是 MangGo 的选择。
        ui.Box(c).Width(28).Shrink(0).Center().Children(func() {
            ui.Icon(c, route.Icon).Size(18, 18).AccessHidden(true)
        })
        ui.Text(c, route.Title).Basis(0).Grow(1).MinWidth(0).
            TextAlign(ui.Center).SingleLine()
    }).Role(ui.RoleLink).FillWidth().Height(34).PaddingX(10).Gap(0).
        OnClick(func() { navigate(route.Path) })
}
```

`SidebarMenuButton(c, props, content)` 不依赖 SidebarWith 或路由器。`content == nil` 时提供普通起始对齐的图标和标签；非 nil 回调构建非交互内容，并继承统一前景及当前表面。`Size` 使用 SidebarItemDefault/Small/Large，`FontSize` 遵循系统文字缩放并扩大所需高度。`Active` 设置 CurrentPage。`Disabled` 及祖先禁用由原生控件处理；Tab、Enter、Space、焦点 ring 和减少动态效果仍由库与 MyGo 管理。

`StyleSidebarMenuButton` 可将同一外观应用到已有原生 trigger，例如 Collapsible；保留原 trigger 的行为及可访问性。应用可通过返回 Element 的 OnClick 或 Clicked 处理动作。自定义内容需要提供可访问的 Label，装饰图标应 AccessHidden。

## MangGo 底部状态消费示例

```go
p := navigation.SidebarMenuButtonProps{
    Label: "License Management", Active: currentPath == "/license",
}
version, badge := "Free Version", "Free"
if activated {
    version, badge = "Professional Version", "Pro"
    p.Color = &t.Primary
} else {
    p.Muted = true // 静止 MutedForeground；hover / press / focus 为 Foreground。
}
navigation.SidebarMenuButton(c.Key("/license"), p, func(c *ui.Context) {
    display.BadgeWith(c, badge, display.BadgeProps{
        Variant: display.BadgeOutline, InheritForeground: true,
    })
    ui.Text(c, version).MinWidth(0).Shrink(1).SingleLine()
}).FillWidth().OnClick(func() { navigate("/license") })
```

`BadgeProps.InheritForeground` 适用于 Outline/Ghost/Link，供非交互内嵌徽标继承父动作的状态文字色；独立可操作的 Badge 仍使用自身 Color 与 Surface。不要给继承的徽标另设文字色。普通内嵌图标、文字和自定义视觉槽位自然继承同一状态前景。

## 验证与视觉对照

测试覆盖全部 24 个官方预设的浅色/深色、Background/Sidebar/Card/Popover、自定义明暗表面、黑白/彩色/透明/半透明输入、idle/hover/pressed/focus、系统高对比度、文字缩放、禁用及运行中切换 ReduceMotion。真实组件渲染检查图标实心像素、14 DIP 文字与 12 DIP Badge 的一致颜色；每 10ms 检查彩色反馈进入和离开的动画中间帧，并覆盖无法共用前景的自定义表面。独立窗口 Popover 通过库的测试 backend 运行同一内容 view 和实际绘制路径。

本次验证命令及结果记录在 [验证产物](../artifacts/readable-color/verification.txt)。当前工作区另有忽略的 `build/priority-milestone-2026-11-10` 测试，需要指定它自己的产物目录，因此完整命令为：

```sh
PRIORITY_CONSUMER_ARTIFACTS="$PWD/artifacts/readable-color/priority-consumer" \
GOCACHE="$PWD/.cache/go-build" go test ./...
GOCACHE="$PWD/.cache/go-build" go vet ./...
```

基线截图在修改实现前由 6d2dd97 的真实 control.Button / BadgeWith 生成；修改后由公开 SidebarMenuButton 示例生成。对照图左侧为基线、右侧为新实现：

| 深色琥珀 | 浅色琥珀 |
| --- | --- |
| ![深色前后对照](../artifacts/readable-color/comparison-dark.png) | ![浅色前后对照](../artifacts/readable-color/comparison-light.png) |

| 实际像素颜色对比度 | 深色：前 → 后 | 浅色：前 → 后 |
| --- | --- | --- |
| 导航选中项文字/图标 | 2.24 → 5.02 | 4.21 → 5.34 |
| Pro 徽标文字 | 2.53 → 5.18 | 4.82 → 5.23 |
| Pro 版本说明文字 | 8.40 → 5.18 | 3.07 → 5.23 |

基线消费代码的 Pro 徽标使用 Primary，版本说明使用 SidebarPrimary，深色琥珀下并非同色；2.53:1 对应徽标。新示例统一从 Primary 派生。静止值高于 4.5 是因为同一前景还覆盖更强的交互填充。数据见 [contrast.json](../artifacts/readable-color/contrast.json)。

还检查了 [hover](../artifacts/readable-color/after-dark-hover.png)、[pressed](../artifacts/readable-color/after-dark-pressed.png)、[键盘焦点](../artifacts/readable-color/after-dark-focus.png)、[Pro hover](../artifacts/readable-color/after-dark-pro-hover.png)、[Free hover](../artifacts/readable-color/after-dark-free-hover.png) 和 [Free focus](../artifacts/readable-color/after-dark-free-focus.png)，以及对应的浅色截图。产物位于被 git 忽略的 artifacts 目录，可用示例命令重新生成修改后截图；基线原始 view 也保存在该产物目录。

## 行为兼容性

- 新 API 为增量扩展。旧 Sidebar/SidebarWith、普通中性变体、实色 Primary / PrimaryForeground 的配对语义保留；自定义实色填充继续按实际合成色选择黑/白文字。
- Soft、Link（含默认主题色）以及显式 Color 的 Ghost/Outline、对应 Badge 文字现在会自动校正。原始强调色继续用于背景、边框、焦点；小字不再直接等于原始 Color。透明/半透明文字必要时也会变得更不透明。
- 彩色 Outline Button 静止背景改为透明，正确显示 Sidebar/Card/Popover；未指定 Color 的普通 Outline 保留原有背景和边框语义。
- 状态前景同时满足相关静止/交互背景，通常不随 hover 突变。无法跨状态共用前景时，库直接切换该颜色反馈；ReduceMotion 始终立即生效。HighContrast 使用增强文字目标及主题焦点色。
- 单独的彩色 Badge 仍管理自身状态。内嵌状态徽标需显式 InheritForeground，避免它单独解析成不同颜色。
- 修改限于本组件库及其文档、示例和验证产物；原版 Manggo 与 MangGo 消费仓库未修改。
