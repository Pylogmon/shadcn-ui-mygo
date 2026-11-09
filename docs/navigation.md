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
