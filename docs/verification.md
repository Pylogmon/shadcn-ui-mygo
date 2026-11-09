# 第一版阶段验证记录

验证日期：2026-11-09。环境：macOS arm64，Go 1.27.3，MyGo 版本见 `go.mod`。

## 阶段 1：基础系统

- `go test ./pkg/theme/... ./pkg/motion`：10 项测试通过。
- `go vet ./pkg/theme/... ./pkg/motion`：通过。
- 离线执行 `go generate ./pkg/theme` 两次，生成文件一致。
- 26 套颜色各含 11 个色阶；24 套官方主题按固定源码生成。已测试已知颜色坐标、主题覆盖继承、浅/深/系统模式、独立作用域、偏好存储和减少动态效果。
- 数据来源、原始数据 SHA-256 和合并规则见 `pkg/theme/sources/manifest.json`；上游 MIT 声明随数据保存。

## 阶段 2：表单控件第一版

- `go test ./pkg/control`：6 项测试通过。
- `go vet ./pkg/control`：通过。
- 验证实际输入、只读和禁用状态、选择器键盘跳过禁用项、单选和滑块键盘操作、ToggleGroup 选择、OTP 过滤与完成回调。

## 阶段 3：展示、浮层和导航第一版

- `go test ./pkg/display ./pkg/overlay ./pkg/navigation`：23 项测试通过（含图表子测试）。
- `go vet ./pkg/display ./pkg/overlay ./pkg/navigation`：通过。
- 验证日期浏览不提交、选择提交/关闭/焦点恢复、空日期起始月份、单开 Accordion、Carousel 边界、表格排序/过滤/分页/数据不变性、图表 NaN/Inf、代码块复制当前原文。
- 验证菜单、子菜单、复选项、右键菜单、命令过滤、弹窗关闭/焦点、确认弹窗外部点击保护、Toast 动作、Tabs 跳过禁用项、分页及当前路由语义。

## 阶段 4：会话与原生演示第一版

- `go test ./...`：57 项测试通过（包括子测试）；`go vet ./...` 通过。
- Gallery 测试包含六个页面在浅色/深色下渲染，并验证导航、主题切换和对话框关闭联动。
- 会话测试验证问卷必填/多选/自由输入/返回保留/提交，Trace 和附件动作，以及 10,000 行消息虚拟化、历史追加稳定位置与跳至末尾。
- `make snapshots`：生成浅色、深色和主题页 PNG，并人工查看布局与配色。截图位于本地 `artifacts/`，不是所有上游组件的视觉对照证据。
- `make app`：生成 macOS `.app`；通过系统 UI 实际启动并读取原生可访问性树，点击 Overlays、打开 Edit profile，激活窗口后使用 Escape 关闭并返回原页面。
- `CGO_ENABLED=0` 下 macOS arm64、Linux amd64、Windows amd64 的 Gallery 编译通过。Windows 首次编译缺少 `x/net` 校验项，执行 `go mod tidy` 后重新编译通过，依赖已纳入 `go.mod` / `go.sum`。
- `go generate ./...` 后主题生成文件没有差异。

## 尚未通过的最终验收

Windows/Linux 的真实窗口、GPU、输入和系统偏好尚未运行验证。macOS 只完成了上述具体流程；没有验证每个组件、每个变体、所有屏幕阅读器或全部原生行为。

完整组件变体、行为、交互动画和视觉对照仍需按 [parity.md](parity.md) 逐项实施。当前验证证明第一版阶段产物可用，不能证明完整复刻已经完成。CI 已配置，但本地记录不代表远端 CI 已执行。

## 阶段 5：组件行为增强（持续更新）

- 控件模块：`go test ./pkg/control` 15 项通过，`go vet` 通过。新增测试覆盖忙碌按钮、各图标尺寸、分组接缝/RTL/嵌套、复合输入与字段标签、字段自适应容器、分组自定义选项、原生系统选择菜单和多选 Combobox。
- Tabs：`go test ./pkg/navigation` 9 项通过（含动画子测试），`go vet` 通过。新增验证手动激活、面板焦点/控制关系、方向键与 RTL、不循环边界、动态禁用/移除/空集，以及普通/减少动态效果下指示器的调度与收敛。

- 系统菜单适配：`go test ./pkg/overlay` 8 项通过；新增系统下拉/子菜单/勾选/禁用、右键菜单与菜单栏动作测试，并验证菜单内容不被画入父窗口 UI 树。越窗几何仍需真实平台窗口验证。

- 图表与数据表：`go test ./pkg/display -run 'Chart|DataTable|DataRows'` 34 项检查通过；display vet 通过。验证七类图表键盘检查/缺失值、环形几何命中与中空区、极端有限值比例、正负与归一化堆叠、原值 Tooltip，以及真实像素中的多序列颜色。数据表验证稳定键跨页选择、原生列/分面菜单、多个过滤条件、搜索和自定义行操作。已检查六类图表组合渲染，尚不代表完整上游示例矩阵。

- 日期模块：`go test ./pkg/display -run 'Calendar|DatePicker|DateRange'` 24 项检查通过，display vet 通过。实现单/多/范围选择、限定日期/月份、多月与原生时间编辑。独立审计新增回归测试，修复同日键盘导航误清空、禁用导航跨月、隐藏日期预设绕过、时间编辑绕过匹配器及范围时间倒置、翻月后活动日期残留旧月份等问题。其他历法、自由文本日期解析与完整平台矩阵尚未完成。

- 桌面成对组件：`go test ./pkg/desktop` 30 项检查通过；Host、Menu/NativeMenu、Dropdown、Context、Popover 和 Select 分别验证独立子树、系统/自绘路由、锚点跟随、一次性结果、迟到关闭、内容更新先于父刷新、减少动画、键盘/鼠标/子菜单、系统策略恢复。
- macOS 小窗口对照应用 `examples/native-popups` 实际启动：Menu 的独立 shadcn Flyout 超出父窗口下边界；NativeMenu 的系统菜单可访问性与原生外观已观察；Popover 同时跨越父窗口右侧和下边界，子窗口 Done 关闭并恢复到 Popover 触发器。Gallery 的 Inspect theme 也已观察独立窗口与 Escape 关闭。此记录只证明这些具体流程，不证明所有浮层/平台/屏幕边缘均已验收。

- Create 风格 Gallery：27 项检查通过；验证真实主题/色阶/字体/圆角配置、严格预设导入、剪贴板、撤销/重做/随机/重置、独立预览像素变化、1280/800/520 宽布局，以及 Flyout/系统适配切换、选择回传与菜单动作。最终 `go test ./...` 共 171 项检查通过，`go vet ./...`、`go generate ./...` 后主题文件无差异、`git diff --check` 通过。macOS `.app` 构建、Linux/Windows Gallery 交叉编译通过，浅深色 Creator 截图已生成并查看。远端 CI 尚未执行，其他浮层家族的完整成对迁移仍未完成。

## 后续阶段：平台主题与反馈语义

- `theme.NativeScope` 在不重建 Theme 的情况下提供原生语义色别名；测试验证自定义平台颜色、字体、焦点宽度/偏移和 Tooltip 延迟完整保留，作用域退出后父主题不变，后续平台主题变化可更新。
- 反馈组件回归先复现三项失败：负值不确定进度被清零、NaN 可访问性值、视觉标题没有原生 heading rank。修复后新增反馈测试 5 项通过，涵盖值文本、装饰/语义分隔、非有限值与减少动画时静态进度；原有 Spinner 的减少动画行为经运行时测试确认已正确，无需额外替换。

- Flyout 生命周期：Host/Presence 针对性运行 23 项检查和 vet 通过。扩展后端验证实际退场像素、约 100ms 内保留窗口、关闭一次、重开反转而不换窗口、原生关闭通知/焦点仅一次、减少动画立即关闭、锚点/父窗口消失立即清理、NativePopover 不拦截系统动画。新 Tooltip/HoverCard 所需的鼠标穿透、不抢焦点显示和显式键盘进入已由原生后端转发。

- 独立日期组件：`go test ./pkg/desktop -run 'DatePicker|DateRange|DesktopDate'` 10 项检查及 desktop vet 通过，验证两种容器/主题、独立 Context、浏览/一次性提交、日期时间草稿与取消、约束匹配器、范围未完成不发布、DST 两晚跨越47小时、原生小时编辑、预设、外部值/规则更新、无空闲失效、无 Host 不伪造父窗口浮层，以及 Footer 的子 Context 约束。

- 原生日期标题补充回归：月份/年份下拉使用系统菜单，选项不进入日历窗口绘制树，浏览不提交日期。`go test ./pkg/desktop ./pkg/display -run 'Calendar|DatePicker|DateRange|DesktopDate'` 共35项检查及两个模块 vet 通过。

- Tooltip/HoverCard 组件对：`go test ./pkg/desktop -run 'Tooltip|HoverCard|InactiveHover'` 20项检查及 desktop vet 通过。覆盖取消/重入计时、跨窗口指针间隙、键盘进入/返回、Escape 抑制、应用失活、非交互内容、无 Host 错误、主题和独立 Context。真实像素回归先复现 NativeTooltip 缺省反色透明的问题，再验证明暗主题、显式与部分反色覆盖。关闭指针预览不会把已有输入焦点移到触发器并导致重开。
