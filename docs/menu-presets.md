# 菜单预设

Gallery 的 Menu color / Menu accent 对应固定 shadcn/create 配置。应用可直接配置主题控制器：

```go
appearance := theme.Controller{
    Preset: "rose",
    MenuColor: theme.MenuInvertedTranslucent,
    MenuAccent: theme.MenuSubtle,
}
// 在每次 view 构建开始时调用。
appearance.Apply(c)
```

四种颜色为 `MenuDefault`、`MenuInverted`、`MenuDefaultTranslucent`、`MenuInvertedTranslucent`；强调为 `MenuSubtle`、`MenuBold`。空值使用 default/subtle。透明配置总是规范化为 subtle，共六种有效组合。`MenuOptions.Normalized` 返回规范值，`Validate` 检查未知值。控制器保存与 Gallery 导入、导出、撤销、重做均使用规范配置；Gallery 切回不透明菜单时恢复此前的强调选择。

Inverted 始终使用对应深色调色板。Bold 将当前主题的 Accent / AccentForeground 映射到 Primary / PrimaryForeground，因此也影响使用这对语义色的其他控件；SidebarAccent 保持原值。

局部配置使用 `theme.MenuScope(c, tokens, options)`，返回的原生容器支持 `.Contain(...)`；独立主题使用 `tokens.MenuTheme(options)`。原有 `theme.Scope(c, tokens)` 签名保留。标准预设会自动匹配自己的深色菜单；完全自定义的浅色主题若显式请求 inverted，需通过 `MenuOptions.InvertedTokens` 提供对应深色 `Tokens`。普通自定义 Scope 保留自身颜色，不借用祖先调色板；可用 `theme.MenuOf(c).HasInvertedPalette()` 检查配对情况。

`theme.MenuOf(c)` 返回可复制的纯值快照，包含实际菜单 Tokens、颜色、强调及动画/透明标志。预设覆盖 Combobox、Select、styled Menu / DropdownMenu / ContextMenu / Menubar 及子菜单，也覆盖保留的同窗菜单。通用 Popover、Tooltip、HoverCard、Dialog、Command 使用各自的动画。`NativeMenu` 等系统组件继续使用系统外观。

不透明菜单使用既有 100ms 源动画；透明菜单立即切换，不保留退出内容。同窗透明面通过 MyGo 公开 shader 采样背景、40 DIP Gaussian blur、1.5 饱和度与 70% Popover tint 绘制，文字和子控件保持清晰。独立 Flyout 使用系统 Vibrancy 材质，无法直接采样父窗口 framebuffer；系统不提供材质或启用减少透明度/高对比度时使用不透明面。系统材质的模糊核不声明与 CSS 逐像素相同。

固定原文、许可与 SHA 清单见 [source archive](../pkg/theme/testdata/shadcn-menu-preset/README.md)。macOS 真实加速绘制及系统材质已运行验证；Windows/Linux 当前完成编译，实际窗口和辅助技术验收仍待完成。
