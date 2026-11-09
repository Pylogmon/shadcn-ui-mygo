# 原生浮层适配

按照用户要求，菜单、下拉菜单和 Popover 优先采用 MyGo 的原生能力，可以越出父窗口；原生平台的交互约定可替代网页 DOM 浮层实现。shadcn 的颜色、组件组合和功能仍作为设计参考。

## 系统菜单

- `overlay.NativeDropdownMenu`、`NativeContextMenu`、`NativeMenubar` 构建 MyGo 的系统菜单。
- `control.NativeSelect` 是系统菜单选择控件；`Select` 保留完全可定制的窗口内选项内容。
- DataTable 的列显隐和分面过滤同样使用 MyGo 的菜单。
- 默认 `ui.SystemMenus` 由操作系统布局、绘制和处理越窗边界。应用显式设置 `ui.PopoverMenus` 时，MyGo 会将同一份菜单绘制在窗口内，便于嵌入预览与特殊风格。
- 完全自绘的 `overlay.DropdownMenu` / `ContextMenu` / `Menubar` 仍可单独使用，不强制所有应用放弃自定义菜单。

原生菜单支持子菜单、复选/单选、禁用项、分组与分隔项。`MenuItem.NativeIcon` 使用一次创建、可复用的 MyGo 平台图标，`Icon` 继续用于自绘菜单中的 SVG。`Shortcut` 当前是显示文本，应用的真实快捷键绑定应单独注册。

```go
choice := overlay.NativeDropdownMenu(c, overlay.MenuProps{
    Items: []overlay.MenuItem{
        {Key: "duplicate", Label: "Duplicate"},
        {Separator: true},
        {Key: "remove", Label: "Remove", Destructive: true},
    },
}, func(trigger ui.Element) {
    control.StyleButton(c, trigger, "Actions", control.ButtonProps{Variant: control.Outline})
})
```

系统菜单在操作系统自己的弹出层中显示，不出现在父窗口的 `ui.Element` 子树。单元测试使用 MyGo 的系统菜单记录/选择接口验证路由、禁用项、勾选状态和回调。它们不能替代真实窗口边缘、屏幕边界、缩放与多显示器的运行验证。

## 任意内容 Popover

MyGo 的 `NewFlyout` 支持拥有独立原生 UI 内容的子窗口，由父窗口及锚点控制位置、工作区碰撞和关闭生命周期。独立窗口必须收到自己的 `*ui.Context`，不能继续使用父窗口某一帧的 Context 构建内容。适配层与 Gallery 接入正在阶段 5.D 实施，完成后补充 API 与验证记录。
