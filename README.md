# shadcn/ui for Native Go

基于 [MyGo](https://github.com/egoist/mygo) 的原生 Go UI Kit，移植 shadcn/ui 的组件和设计系统。采用 Go UI 树和 GPU 渲染，可组合、可主题化。

项目按模块、分阶段实施，每阶段验证后独立提交。**目前处于组件第一版和行为对齐阶段，完整复刻尚未完成。** 进度与完整目标见 [实施阶段](docs/implementation.md)，具体差异见 [对齐记录](docs/parity.md)。

Gallery 参考 [shadcn/create](https://ui.shadcn.com/create)：紧凑配置面板搭配独立实时预览，支持基础色、强调色、图表色、字体、圆角、外观、严格预设导入导出与撤销重做。默认菜单使用 shadcn Flyout，可切换系统原生样式。

需要 Go 1.27.3 或更新版本。

```sh
make test       # 原生交互和模型测试
make vet
make run        # 原生 Gallery
make snapshots  # 浅色、深色和主题页 PNG，输出 artifacts/
make build     # CGO_ENABLED=0 构建
make app       # macOS .app 输出 build/Shadcn Gallery.app
make popups-app # 小窗口对照：菜单、Popover、日期与悬停组件对
```

```go
import (
    "github.com/egoist/mygo/ui"
    "github.com/pylogmon/shadcn-ui-mygo/pkg/control"
    "github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

var appearance theme.Controller // 默认跟随系统
var enabled bool

func view(c *ui.Context) {
    appearance.Apply(c)
    ui.Column(c).Padding(24).Gap(12).Children(func() {
        control.Button(c, "Save changes", control.ButtonProps{})
        control.Switch(c, &enabled, control.SwitchProps{Label: "Notifications"})
    })
}
```

| 模块 | 内容 |
| --- | --- |
| `pkg/theme` | 语义颜色、OKLCH、官方主题与色阶、作用域主题、浅/深/系统模式、偏好存储 |
| `pkg/motion` | 交互过渡、浮层动画、减少动态效果支持 |
| `pkg/control` | 按钮、输入、选择、表单、OTP、开关、单值/范围/多拇指滑块 |
| `pkg/display` | 卡片、排版、日历、数据表、图表、代码块、可折叠/多面板布局与反馈 |
| `pkg/overlay` | 对话框、抽屉、菜单、命令、提示、通知 |
| `pkg/navigation` | 标签页、面包屑、分页、独立富内容导航面板/系统路由菜单、侧栏 |
| `pkg/desktop` | 成对桌面组件：Menu / NativeMenu、DropdownMenu、ContextMenu、Menubar、Popover、Select、日期、Tooltip、HoverCard；独立 Flyout 与系统样式 |
| `pkg/conversation` | 附件、消息、虚拟滚动、标记、Trace、多步问卷 |
| `internal/gallery` | 原生演示与集成验收 |

组件使用调用方持有的状态；在动态列表中使用稳定的 `c.Key(id)`。颜色通过 `theme.TokensOf(c)` 获取，局部完整主题使用 `theme.Scope(c, tokens).Children(...)`，不会影响其他窗口。

桌面弹出组件显式使用 `desktop.NewHost(parentWindow)`。`Menu` 使用统一 shadcn 样式的独立 Flyout；`NativeMenu` 使用系统菜单。没有 Host 时不会默默退回父窗口内绘制。窗口内旧组件以 `InWindow*` 明确命名；API、平台范围与迁移说明见 [原生浮层适配](docs/native-adaptation.md)，完整示例见 [native-popups](examples/native-popups/main.go)。

Gallery 的 Navigation 页展示菜单栏、可切换尺寸的导航内容面板和响应式侧栏，Forms 页展示范围/多拇指滑块，Data 页包含保留状态的折叠内容和嵌套可调整面板。导航面板的子 Context、路由回传和 Tab 边界用法见 [导航 API](docs/navigation.md)。

Forms 页也包含 [原生文本输入](docs/text-input.md)、[可调整高度的 Textarea](docs/textarea.md)、[分组 OTP](docs/input-otp.md) 与 [三态/单多选控件](docs/choices.md)。Bio 草稿和高度、验证码与选择状态跨页保留；Navigation 页的面包屑可切换独立 Flyout／系统菜单，并展示有界的大页数分页。

Data 页也展示头像尺寸、状态角标、加载/回退和头像组。对应 API 见 [Avatar](docs/avatar.md) 与 [Card 槽位](docs/card.md)。

Content & feedback 页展示 [Badge](docs/badge.md)、[Item](docs/item.md)、[Alert/Empty](docs/status-surfaces.md) 的组合示例，包含真实原生按钮/链接、独立主动作、通知操作和空状态表单。

Text & layout 页集中展示 [Typography/Kbd](docs/typography.md)、[CodeBlock](docs/code-block.md)、[Spinner/Skeleton](docs/loading.md) 与 [ScrollArea/AspectRatio](docs/scroll-area.md)。文章详情、确认后的复制反馈、加载暂停和滚动条样式切换均可交互；代码 Buffer、输入草稿与滚动位置跨页保留。通过 Search 输入组件名，或运行 `go run ./cmd/gallery -page text-layout` 可直接进入。

响应式导航布局见 [Sidebar](docs/sidebar.md)；对话框焦点、固定页脚与抽屉吸附手势见 [Modal API](docs/modals.md)。
命令面板的虚拟列表、过滤、快捷键与自定义内容见 [Command](docs/command.md)。
通知的应用队列、六位置堆叠、loading更新、可阻止关闭的动作、原生滑动及完整自定义正文见 [Toast/Sonner](docs/toast.md)。

固定上游源码、离线评分对照和对应许可见 [第三方声明](THIRD_PARTY_NOTICES.md)。

预设可从 Gallery 复制，并用 `go run ./cmd/gallery -preset path/to/preset.json` 加载；无效预设会在修改配置之前被拒绝。

参考 [Lobe MyGo](https://github.com/lobehub/lobe-mygo) 的分层组件库架构、[MyGo Music](https://github.com/egoist/mygo-music) 的原生应用实践，以及 [shadcn/ui](https://github.com/shadcn-ui/ui) 的组件与主题规范。来源及固定版本见实施文档和主题数据目录。

已执行的验证和平台范围见 [验证记录](docs/verification.md)。
