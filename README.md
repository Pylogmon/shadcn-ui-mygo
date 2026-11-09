# shadcn/ui for Native Go

基于 [MyGo](https://github.com/egoist/mygo) 的原生 Go UI Kit，移植 shadcn/ui 的组件和设计系统。采用 Go UI 树和 GPU 渲染，可组合、可主题化。

项目按模块、分阶段实施，每阶段验证后独立提交。**目前处于组件第一版和行为对齐阶段，完整复刻尚未完成。** 进度与完整目标见 [实施阶段](docs/implementation.md)，具体差异见 [对齐记录](docs/parity.md)。

需要 Go 1.27.3 或更新版本。

```sh
make test       # 原生交互和模型测试
make vet
make run        # 原生 Gallery
make snapshots  # 浅色、深色和主题页 PNG，输出 artifacts/
make build     # CGO_ENABLED=0 构建
make app       # macOS .app 输出 build/Shadcn Gallery.app
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
| `pkg/control` | 按钮、输入、选择、表单、OTP、开关、滑块 |
| `pkg/display` | 卡片、排版、日历、数据表、图表、代码块、布局与反馈 |
| `pkg/overlay` | 对话框、抽屉、菜单、命令、提示、通知 |
| `pkg/navigation` | 标签页、面包屑、分页、导航、侧栏 |
| `pkg/conversation` | 附件、消息、虚拟滚动、标记、Trace、多步问卷 |
| `internal/gallery` | 原生演示与集成验收 |

组件使用调用方持有的状态；在动态列表中使用稳定的 `c.Key(id)`。颜色通过 `theme.TokensOf(c)` 获取，局部完整主题使用 `theme.Scope(c, tokens).Children(...)`，不会影响其他窗口。

参考 [Lobe MyGo](https://github.com/lobehub/lobe-mygo) 的分层组件库架构、[MyGo Music](https://github.com/egoist/mygo-music) 的原生应用实践，以及 [shadcn/ui](https://github.com/shadcn-ui/ui) 的组件与主题规范。来源及固定版本见实施文档和主题数据目录。

已执行的验证和平台范围见 [验证记录](docs/verification.md)。
