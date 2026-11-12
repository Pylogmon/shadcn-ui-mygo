# Sidebar composition

`go run ./examples/sidebar-composition` 打开可操作的原生窗口；`-dark` 使用深色琥珀预设。侧栏条目调用 `SidebarMenuButton`，内容回调由应用定义 28 DIP 图标列及居中文字区域；这是示例布局，不是组件默认值。

Pro 使用 `Color: &tokens.Primary`，Free 使用 `Muted: true`。内嵌静态 Outline Badge 设置 `InheritForeground: true`，文字、图标、Badge 在整个交互状态中共享父操作的前景。普通按钮使用官方 Sidebar tokens，只有选中条目显式传 `Color`。原生按钮负责 Enter/Space、Tab、焦点、禁用及当前页语义，库负责状态动画。

```sh
mkdir -p artifacts/readable-color
go run ./examples/sidebar-composition -dark -snapshot artifacts/readable-color/after-dark.png
go run ./examples/sidebar-composition -dark -state hover -snapshot artifacts/readable-color/after-dark-hover.png
go run ./examples/sidebar-composition -dark -state focus -snapshot artifacts/readable-color/after-dark-focus.png
```

支持 `idle`、`hover`、`pressed`、`focus`、`pro-hover`、`free-hover`、`free-focus`；省略 `-dark` 生成浅色图。每张图通过 MyGo Tester 的真实组件、输入及绘制路径生成。

完整 API、表面作用域、兼容性及对照见 [彩色文字与 Sidebar](../../docs/readable-color.md)。
