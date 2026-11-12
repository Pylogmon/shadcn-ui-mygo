# Responsive Sidebar

`SidebarWith` 组合侧栏和主内容区域，状态由调用方持有；旧 `Sidebar` 继续提供静态导航列表。

```go
var sidebar navigation.SidebarState
var route = "home"

_, err := navigation.SidebarWith(c, host, "workspace", &route, &sidebar,
    navigation.SidebarLayoutProps{
        Label: "Workspace", Variant: navigation.SidebarInset,
        Collapse: navigation.SidebarIcon,
        Items: []navigation.SidebarItem{
            {Key: "home", Label: "Home", Group: "Workspace"},
            {Key: "projects", Label: "Projects", DefaultOpen: true,
                Children: []navigation.SidebarItem{{Key: "current", Label: "Current project"}}},
        },
    }, func(child *ui.Context) { ui.Text(child, "Selected: " + route) })
_ = err // 按应用约定显示错误
```

默认桌面宽256 DIP、图标模式48 DIP、移动面板288 DIP；根据布局自身测量宽度在768 DIP切换为Sheet。嵌入其他容器时不会一直使用外层窗口宽度。`AvailableWidth` 可以显式指定可用宽度，`Mobile` 可以覆盖响应式判断。桌面 `Collapsed` 与 `MobileOpen` 相互独立。

`SidebarStandard` / `SidebarFloating` / `SidebarInset` 提供标准、悬浮、内嵌主区变体。`SidebarOffcanvas` / `SidebarIcon` / `SidebarFixed` 控制完全移出、图标折叠或固定布局。`SidebarStart` / `SidebarEnd` 跟随RTL，`SidebarLeft` / `SidebarRight` 始终是物理边。

可选Header/Footer回调支持搜索、工作区和账号操作；条目支持分组、子项、角标、独立Action、尺寸、描边和Tooltip。条目Key必须全树唯一且非空，分组展开状态保存在 `SidebarState.Groups`。`SidebarState` 可直接JSON保存/载入导出状态。

默认主区包含 `SidebarTrigger`，也可隐藏后自己放置。轨道可点击折叠，Cmd/Ctrl+B仅作用于焦点所在侧栏布局。移动面板提供模态焦点约束，选择路由/Escape关闭后返回触发器；触发器移除时返回主区。`OnToggle`、`OnGroupChange`、`OnSelect` 通知实际用户变化，返回Element的Changed供另一种处理路径。

图标模式的Tooltip使用显式 `desktop.Host` 的独立shadcn Flyout，可越出父窗口；必须在父View中调用Host.BeginFrame/EndFrame。没有Host而实际请求打开Tooltip会返回错误，不伪造窗口内替代品。主面板和移动Sheet本身在当前窗口中布局。

首次未显示的offcanvas内容延迟构建以避免AutoFocus抢主区焦点。已经保留的隐藏内容中动态插入显式Focus/AutoFocus仍受MyGo的祖先禁用检查限制，应先检查 `element.IsDisabled()`，详见 [parity.md](parity.md)。

独立 `SidebarMenuButton` / `StyleSidebarMenuButton` 支持自定义非交互内容、字号、active 与禁用状态。nil Color 遵循官方 SidebarForeground / SidebarAccent / SidebarAccentForeground；显式 Color 提供原生彩色选中态，Muted 用于免费版状态文字。固定图标列、居中文字区域由调用方配置。见 [公开 API、MangGo 消费示例与视觉验证](readable-color.md) 和 [可运行示例](../examples/sidebar-composition/README.md)。
