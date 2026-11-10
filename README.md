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
make charts-app # 自由组合图表：build/Shadcn Charts.app
make cartesian-app # 30个Cartesian源例：build/Shadcn Cartesian.app
make catalog-app # Chart70例、文档及Create预览：build/Shadcn Catalogue.app
make icons-app # Material Symbols 图标浏览：build/Shadcn Symbols.app
make verify-consumer # 空缓存、无replace的外部消费者和模块ZIP验收
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
| `pkg/fonts` / `pkg/fonts/all` | Create 的 26 族离线字体、按需函数、独立全量目录与显式原生注册 |
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
数值、日期、时间和系统文件选择可直接绑定应用模型，见[类型化输入](docs/typed-input.md)；运行 `go run ./examples/typed-input` 或 `make inputs-app`。
Persian/Jalali 日历和应用控制解析的可编辑日期见[历法与日期输入](docs/date-input.md)，运行 `go run ./examples/date-composition` 或 `make dates-app`。
Gallery 的整卡单选、进度标签与数值、灰度头像及 Skeleton 表单/表格见[组合示例](examples/gallery-recipes/README.md)。
独立展开布局、文件树、Radius输入和FAQ组合见 [Disclosure API](docs/disclosure-composition.md)；Data页的17个[源配置](examples/disclosure-composition/README.md)保留真实输入、同组键盘导航和跨页草稿。
Forms 页还提供可保存的预约表单和带复制、收藏、菜单、搜索与草稿预览的 InputGroup，编辑状态跨页保留。

Data 页也展示头像尺寸、状态角标、加载/回退和头像组。对应 API 见 [Avatar](docs/avatar.md) 与 [Card 槽位](docs/card.md)。
Table composition 展示真实单元格选择、状态菜单、可编辑备注、稳定键重排、页脚总额与空表；窄窗口可横向滚动，状态和动作菜单跟随全局 Flyout／系统菜单选择。
Cards 页提供 [Card 源组合](examples/card-composition/README.md)：尺寸、间距、边框、原图、独立登录草稿、会议笔记、条款滚动及 RTL；保留原有账号和偏好卡片。
同页的 [Payments 示例](examples/payments/README.md) 提供邮件过滤/排序、列显隐、当前页混合选择、分页及确认后的 ID 复制，状态跨页保留。

Content & feedback 页展示 [Badge](docs/badge.md)、[Item](docs/item.md)、[Alert/Empty](docs/status-surfaces.md) 的组合示例，包含真实原生按钮/链接、独立主动作、通知操作和空状态表单。
其中 [Alert／Empty 源组合](examples/status-composition/README.md) 提供20种通知与13种空状态配置，包含内联链接、明暗 amber 色、真实头像、404搜索和翻译RTL。
同页的 [Item 源组合](examples/item-composition/README.md) 提供35个配置，包含三尺寸／三变体、独立动作、音乐／模型原图、分组分隔、Header／Footer和成员菜单。
该页的 [Badge 源组合](examples/badge-composition/README.md) 提供14个配置，保留六变体、双侧图标、Spinner、真实原生链接、完整长文本横滚、自定义明暗颜色及翻译RTL。

Text & layout 页集中展示 [Typography/Kbd](docs/typography.md)、[CodeBlock](docs/code-block.md)、[Spinner/Skeleton](docs/loading.md) 与 [ScrollArea/AspectRatio](docs/scroll-area.md)。文章详情、确认后的复制反馈、加载暂停和滚动条样式切换均可交互；代码 Buffer、输入草稿与滚动位置跨页保留。通过 Search 输入组件名，或运行 `go run ./cmd/gallery -page text-layout` 可直接进入。

响应式导航布局见 [Sidebar](docs/sidebar.md)；对话框焦点、固定页脚与抽屉吸附手势见 [Modal API](docs/modals.md)。
命令面板的虚拟列表、过滤、快捷键与自定义内容见 [Command](docs/command.md)。
选值/搜索文本分离、可清除单选、多选标签、对象值与独立组合槽位见[Combobox](docs/combobox.md)。
越出父窗口的真实搜索编辑器使用 [`desktop.Combobox` / `NativeCombobox`](docs/combobox-hosted.md)，可运行[独立窗口示例](examples/desktop-combobox/README.md)或 `make comboboxes-app`；Native 使用平台主题的 Flyout。
Field 的多条错误、自定义原生正文与实际编辑器关联见 [Field errors 示例](examples/field-errors/README.md)。
独立 Label、Content、前置说明、整卡选择和 Legend/Separator 见 [Field 组合 API](docs/field-composition.md)；Forms 页提供13个文档示例和 Create 10 分区的[可操作组合](examples/field-composition/README.md)。
多项可见布局、独立轮播控件、吸附位置、应用控制器与原生编辑内容见[Carousel](docs/carousel.md)。
Gallery的Data & display页可切换Carousel布局、间距、循环和项目顺序，编辑卡片备注，并试用垂直/RTL导航。
Forms & controls页已接入框架搜索、国家分组/城市描述过滤、Office弹层搜索、多选技术标签与表单状态示例。
同页可切换 [ButtonGroup 的13种组合](examples/control-compounds/README.md)，Data页提供同模块的Avatar下拉菜单；组内编辑与菜单动作随全局Flyout／Native模式切换。
Forms 页的 [Button 源组合](examples/button-composition/README.md) 提供20个配置，包含四尺寸／六变体矩阵、左右图标、独立原生链接、禁用／加载／Invalid、圆角与翻译RTL。
同页的 [Switch 源组合](examples/switch-composition/README.md) 提供12个配置，保留两尺寸、前置说明、整卡选择、禁用／错误状态和译文，操作真实原生Switch。
该页的 [Checkbox 源组合](examples/checkbox-composition/README.md) 提供15个配置，包含独立默认值、描述／整卡、分组、禁用／错误、真实表格选择和RTL；表格全选保持源布尔规则。
同页的 [RadioGroup 源组合](examples/radio-composition/README.md) 提供13个配置，包含原生键盘互斥、整卡计划、价格／Fieldset、两列网格、单项／整组禁用、错误与翻译RTL。
同页的 [Textarea 源组合](examples/textarea-composition/README.md) 提供11个配置，保留原生多行编辑、说明前后顺序、固定行数、禁用／错误、真实发送按钮与独立本地草稿。
该页还可配置Combobox定位/RTL/高亮，试用对象多选表单、Dialog内搜索、Grid及100项滚动窗口。
通知的应用队列、六位置堆叠、loading更新、可阻止关闭的动作、原生滑动及完整自定义正文见 [Toast/Sonner](docs/toast.md)。
Overlays 页的 [Popover／HoverCard 源组合](examples/popup-composition/README.md) 提供12个配置：独立窗口表单、对齐和逻辑侧、Dialog内弹层、延迟预览及翻译RTL。
同页的 [Dropdown Menu 源组合](examples/dropdown-composition/README.md) 提供27个配置：嵌套动作、带图标的勾选／单选、独立组与行缩进、头像、六侧、Dialog及翻译RTL，随全局 Flyout／Native 模式切换。
Navigation 页的 [Context Menu 源组合](examples/context-menu-composition/README.md) 提供22个配置，可通过右键／菜单键操作，包含分组状态、子菜单、六侧、Dialog、独立缩进与翻译RTL。
同页的 [Menubar 源组合](examples/menubar-composition/README.md) 用15个配置保留6份文档与12个Create分区：完整File／Edit／View／Profiles、格式／插入动作、独立状态与双路菜单。
可组合的原生文字扫光、颜色与字体传递见 [Shimmer](docs/shimmer.md)。
独立正文/标题字体、局部作用域和 styled Flyout 传递见 [字体角色](docs/font-roles.md)。
可分发的字体文件、许可、按需/全量注册与固定来源见 [Create 字体包](pkg/fonts/README.md)。
Persian 日历示例另提供显式可选的 `fonts.Vazirmatn()`，保持原示例字体及应用字体作用域。
Gallery 的 Font/Heading 可分别选择 26 族字体，默认 Inter/继承；旧预设仍可导入，选择和作用域见 [Gallery 字体](examples/font-gallery/README.md)。
静态共享列组合与原生虚拟表格见 [Table](docs/table.md)。
会话内容面、原生按钮/链接与反应层组合见 [Bubble](docs/bubble.md)。
文件/图片卡片、应用持有的上传状态与横向附件组见 [Attachment](docs/attachment.md)。
静态/动作步骤、运行状态扫光、原生展开详情与嵌套状态见 [Trace](docs/trace.md)。
消息的完整槽位、布局/Tab顺序和稳定原生内容见 [Message](docs/message.md)。
静态标记、分隔线、状态扫光与原生链接组合见 [Marker](docs/marker.md)。
回合锚定、流式跟随、历史位置保持与跳转见 [Message Scroller](docs/message-scroller.md)。
严格逐步导航、原生选项/输入、草稿与受控校验组合见 [Questionnaire](docs/questionnaire.md)。
Gallery 同页的 [Questionnaire workflows](examples/questionnaire-recipes/README.md) 提供条件步骤、保存进度恢复和跨答案校验，三个草稿各自保留。
同题多个独立原生输入、稳定 Key、同值提交及校验回焦可运行 [Questionnaire inputs](examples/questionnaire-inputs/README.md)。
横向柱形、分组/堆叠、原值检查与坐标轴配置见 [Chart](docs/chart.md)。
30个Area/Bar/Line源例及曲线、渐变、标签和自定义标记见[Cartesian配置](docs/chart-cartesian.md)。
31个Pie/Radar/Radial源例及角度、半径、扇区、网格与中心内容见[Polar配置](docs/chart-polar.md)。
70例目录、9种Tooltip、原始行payload及文档/RTL/Create流程见[Chart目录](docs/chart-catalog.md)，运行`go run ./examples/chart-catalog`；Gallery的Data & display页也可筛选和操作。
七类图表的数据几何动画、时序配置与减少动态效果见 [Chart动画](docs/chart-animation.md)。
独立容器、配置与原生Tooltip/Legend内容见 [Chart自由组合](docs/chart-composition.md)，可直接搭配MyGo Plot、Rule、Anchor与原生控件。
可运行的公开API示例见 [chart-composition](examples/chart-composition/README.md)，包含原生阈值线、柱顶标签、月份范围筛选及可编辑数据。
Gallery 的格式多选与切换按钮共用状态；Navigation页演示独立Tabs header动作，Tooltip展示可实际打开搜索的快捷键。
无编辑器多选、独立槽位和选中项对齐见[Select组合](docs/select.md)，运行`go run ./examples/select-composition`。
独立Tabs列表、同栏输入/菜单和外置面板见[Tabs组合](docs/tabs.md)，运行`go run ./examples/tabs-composition`。
独立窗口的富内容提示、Kbd与方向箭头见[Tooltip](docs/tooltips.md)，运行`go run ./examples/tooltip-composition`。
菜单颜色、强调及透明材质可在 Gallery 实时切换，应用接入见[菜单预设](docs/menu-presets.md)。
Material Symbols 提供按需函数与独立全量名称目录，包含4157名称、三种风格和填充状态；用法、尺寸/颜色及离线来源见[图标模块](docs/material-symbols.md)，可运行`go run ./examples/material-icons`。
Gallery 的 Data & display 页提供可切换方向、数据模式、坐标和图例的访问量图表，以及 Line、Area、Pie、Donut、Radar、Radial 六类预览。
Messages & questions 页展示反应按钮、共享备注Sheet、五种附件状态与取消/重试/完成、文件预览和可编辑Trace；草稿、反应与展开状态跨页保留，临时弹层在离页时关闭。还可试用完整消息/标记组合、发送与停止流式回复、前插历史与跳回最新，以及问卷的校验、跳过、结果复查、编辑和重置。支持用Message、Marker、MessageScroller、Questionnaire等组件名搜索进入。
Overlays页包含通知类型/位置、状态色与关闭按钮、堆叠、动作保持打开和上传完成更新的交互示例；通知与配置跨预览页保留。

固定上游源码、离线评分对照和对应许可见 [第三方声明](THIRD_PARTY_NOTICES.md)。

预设可从 Gallery 复制，并用 `go run ./cmd/gallery -preset path/to/preset.json` 加载；无效预设会在修改配置之前被拒绝。

参考 [Lobe MyGo](https://github.com/lobehub/lobe-mygo) 的分层组件库架构、[MyGo Music](https://github.com/egoist/mygo-music) 的原生应用实践，以及 [shadcn/ui](https://github.com/shadcn-ui/ui) 的组件与主题规范。来源及固定版本见实施文档和主题数据目录。

已执行的验证和平台范围见 [验证记录](docs/verification.md)。
公开API应用示例、模块ZIP和独立消费者构建流程见 [消费者验证](docs/consumer-verification.md)。
