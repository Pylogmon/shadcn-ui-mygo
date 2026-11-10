# 字体角色

`theme.FontRoles` 分别指定正文和标题的原生字体族，支持 MyGo 的逗号分隔回退列表。应用持有配置，`Controller` 每次构建读取其当前值：

```go
fonts := theme.FontRoles{Body: "monospace", Heading: "Georgia"}
controller := theme.Controller{Fonts: &fonts}
// 在原生 View 的每次构建中调用：
controller.Apply(c)
```

Body 为空或 `theme.FontSystem` 时使用系统字体。Heading 为空或 `theme.FontInherit` 时跟随正文；显式 `theme.FontSystem` 使用系统标题字体。改变同一个配置指针的字段也会刷新主题和已有 styled Flyout。`Controller.Fonts == nil` 继承调用方的字体角色。

局部使用 `theme.FontScope(c, fonts).Children(...)`；离开作用域后恢复外层配置。只改变标题时，先复制 `theme.FontsOf(c)` 再修改 Heading。作用域保留原生主题的其它属性和菜单配色。

CardTitle、Dialog/AlertDialog/Drawer/Sheet 标题、EmptyTitle（含自定义标题正文）和问卷标题使用固定源的 `cn-font-heading` 角色。自定义标题可调用 `theme.StyleHeading(c, element)`。继承模式保留普通祖先 `.Font(...)` 的原生继承规则。

styled Flyout 复制正文和标题角色；Native 组件保留平台主题。MyGo 没有公开的祖先计算字体查询，因此需要越窗传递字体时应使用 `Theme.Font`、`Controller.Fonts` 或 `FontScope`。

本模块提供字体角色和传递能力。它接受实际字体族名，不负责下载字体，也不把系统中不存在的族名当作已安装。Create 的 26 种命名字体资产及 Gallery 对应选择仍在后续资产模块实施；固定源及许可见 [来源记录](../pkg/theme/testdata/shadcn-font-roles/README.md)。
