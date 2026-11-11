# 通用控件能力

运行 `go run ./examples/control-capabilities`。通过 Change theme size 切换主题字号；Local 18 DIP font 演示局部覆盖。系统文字缩放会同时作用于控件、实际数字编辑器和独立 Select 列表。

每个窗口创建自己的 `desktop.NewHost(window)`，在 view 开头调用 `desktop.BindHost(c, host)` 并保留 `BeginFrame` / `EndFrame`。Select、Menu、Tooltip 传 nil 即继承当前窗口的 Host；旧的显式 Host 参数仍优先。Select 的 `OnChange` 在子窗口选择发生时直接更新应用状态，不需要等父窗口读取 Changed。

Quantity 使用可选空值、0–10 边界、0.5 步长、Outline 图标按钮、可访问性标签和 Tooltip。步进按钮与 `overlay.InWindowTooltip` 共用提示实现，无需 Host；Help 演示独立原生 Tooltip。清空、只读、无效草稿和 IME 使用同一个实际编辑器。

应用在示例中选择按钮变体、布局间距和业务 ID 映射。Dialog 的间距继续由应用配置。

无桌面截图：`go run ./examples/control-capabilities -snapshot /tmp/control-capabilities.png`，可加 `-dark`。
