# 独立导航面板

`navigation.NavigationMenuWith` 将描述链接或任意 Go 内容放入独立 shadcn Flyout，接收明确的父窗口 Host。`NativeNavigationMenu` 使用系统菜单，适合简单路由树。原始 `NavigationMenu(c,items)` 保留为已废弃的窗口内接口，迁移时选择上述两套 API。

```go
// host 在 App.WhenReady 中、创建父窗口后由 desktop.NewHost(window) 初始化。
var route string
var afterNavigation ui.Handle

func view(c *ui.Context) {
    theme.Apply(c)
    host.BeginFrame()
    defer host.EndFrame()

    _, err := navigation.NavigationMenuWith(c, host, "main-navigation", &route,
        navigation.NavigationMenuProps{
            Label: "Main navigation",
            Items: []navigation.NavigationMenuItem{
                {Key: "docs", Label: "Documentation", Width: 420,
                    Links: []navigation.NavItem{
                        {Key: "start", Label: "Getting started",
                            Description: "Create your first native application."},
                        {Key: "theme", Label: "Themes",
                            Description: "Colors and appearance."},
                    }},
                {Key: "components", Label: "Components", Width: 520,
                    Content: func(child *ui.Context, choose func(string)) {
                        ui.Text(child, "Composable by design").FontWeight(600)
                        if control.Button(child, "Browse components", control.ButtonProps{}).Clicked() {
                            choose("buttons")
                        }
                    }},
            },
            OnTab: func(backward bool) { afterNavigation.Focus() },
        })
    // 按应用约定显示 err；没有 Host 时不会退回窗口内浮层。
    _ = err
    control.Button(c, "Continue", control.ButtonProps{}).Bind(&afterNavigation)
}
```

Content 必须使用传入的子 Context，不能引用父窗口某帧的 Element。`choose` 仅在面板仍处于活动状态时提交非空路由；同一选择回调的重复通知不会重复执行。选中值写入调用方 `*string`，父窗口下一帧的返回 Element 报告 Changed。`OnSelect` 是另一条业务动作通知路径，调用方选择其中一条处理。

零 Height 自动测量内容，零 Width 使用组件默认宽度；各项可以有不同尺寸。标准 MyGo 后端在同一窗口内调整几何并保留已访问面板的局部状态；只有其他不可变原生选项变化才重建。注入后端若不实现 `desktop.SizeSurface`，保留原来的尺寸变化重建行为。`Duration` 为负或桌面减少动画时直接采用目标几何。

方向键在标题与描述链接间导航，文本编辑器保留自己的 caret 键；任意内容中的原生 Tab 顺序保留。内容首末边界通过真实可访问的面板区域检测；`OnTab` 可使用稳定 Handle 把焦点交给父窗口真实相邻控件。没有回调时恢复当前标题。

外部闭包数据变化时增加 `Revision`。首次未访问面板延迟构建，避免隐藏内容的初始 AutoFocus 争抢。MyGo 当前显式 `Focus` / `AutoFocus` 可以绕过禁用祖先，已访问的隐藏内容动态创建新编辑器时，应先检查 `element.IsDisabled()` 再请求焦点。运行时统一 inert-focus 约束仍是完整目标的未完成项。

## 面包屑与分页

`BreadcrumbWith` 的折叠祖先使用独立 shadcn Flyout；`NativeBreadcrumbWith` 使用系统菜单。两者保持相同的面包屑外观和选择结果，原简单 `Breadcrumb` 入口继续可用。

```go
items := []navigation.BreadcrumbItem{
    {Key: "home", Label: "Home", URL: "/"},
    {Key: "projects", Label: "Projects", URL: "/projects"},
    {Key: "team", Label: "Your team"},
    {Key: "current", Label: "Current project"},
}
selected, err := navigation.BreadcrumbWith(c, host, "project-trail", items,
    navigation.BreadcrumbProps{MaxItems: 3})
_ = selected
_ = err
// 系统菜单入口：navigation.NativeBreadcrumbWith(c, "project-trail", items, props)
```

URL 祖先是真实原生 Link，按 Enter 导航；相对 URL 优先使用当前 MyGo Router，绝对 URL 使用原生 OpenURL。没有 URL 的祖先是原生 Button。当前页始终为不可激活的 CurrentPage Link，即使提供 URL 也不会导航。MaxItems 折叠中间祖先并保留首项和当前页；小于三时按三项处理。Separator 和 Item.Content 是非交互视觉槽位，语义目标由原生控件持有。

选择以返回 Key 或 OnSelect 通知，避免同时把两条通知当成两个业务事件。Flyout 结果在当前父帧重查祖先：已移除、禁用、变为当前页或 URL 改变的旧结果不会触发路由。菜单内容变化自动增加 Revision；Host 的 BeginFrame/EndFrame 负责离页清理。系统菜单的 tracking/cancel 能力取决于 MyGo 平台接口。

完全自定义组合可用 BreadcrumbNav/List/Link/Page/Separator/Ellipsis。BreadcrumbLink 的可选 Handle 绑定真实 Link，而返回值是可继续设置样式的视觉容器。

```go
var page = 1
zero, one := 0, 1
navigation.PaginationWith(c, &page, navigation.PaginationProps{
    Pages: 1_000_000, SiblingCount: &zero, BoundaryCount: &one,
    URL: func(n int) string { return fmt.Sprintf("?page=%d", n) },
})
```

不提供 URL 时使用原生 Button；提供时使用真实 Link。PaginationWith 返回组合根 Element；兼容入口 Pagination 返回本次用户变化的 bool。受控页码自动夹到一至 Pages，夹取本身不报告用户事件；零页显示空导航，负页数、nil 页码或不合法生成窗口标记 Invalid。

SiblingCount/BoundaryCount 的指针允许明确配置零；旧 Siblings 零值仍默认一个相邻页。算法只遍历边界和相邻窗口，支持 MaxInt 总页数；默认最多生成 4096 个数字控件，超出请求标记 Invalid/inert，不静默截断。上一页/下一页的身份不随目标页变化，连续 Enter 保持焦点。

窄窗口自动省略前后按钮文字，Compact 可显式启用；AlwaysShowText 保留文字。箭头适应 RTL；PageContent 为非交互视觉回调，PageLabel 设置完整可访问名称。一至三位页码保持 36 DIP，长页码按原生文字内容宽度展开，避免多个边界变成无法区分的省略文本。极大数字保持有限控件区域和完整 AX 名称；完整超窄容器/平台/视觉矩阵仍需验收。
