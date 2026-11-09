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

- Slider：保留单值 API，新增 `RangeSlider`、`SliderValues`。`go test ./pkg/control` 23项检查及 vet 通过，独立审计通过。新增8项检查覆盖原生指针捕获/取消、最近拇指选择、不可交叉和步长间距、每拇指 Tab/方向键、竖向/RTL、受控 slice 替换的稳定焦点、格式化值和受约束 AX 数字范围，以及小数步长/极大有限值/不合法约束。

- Carousel：`go test ./pkg/display` 93项检查、其中 `-run Carousel` 22项检查及 vet 通过。保留原 API，新增 `CarouselWith`。验证横纵向/RTL、原生拖拽捕获与取消、循环及两页回绕、键盘/圆点/状态播报、隐藏页输入与 AX 排除、受控切换焦点回收、减少动画、稳定页内状态。独立审计复现快速/受控中途切页搬走仍可见源页造成空白；改为依据当前动画/拖拽位置循环布局后，多方向像素回归、连续切换与空闲收敛均通过。

- HoverCard 原生焦点交接：补充回归后 `-run 'Tooltip|HoverCard|InactiveHover'` 26项检查及 vet 通过。先复现父失活/子活跃状态尚未更新时被误关，再验证平台主题/自绘的交接、真实应用失活关闭、未知后端保守关闭和无空闲重绘。macOS 更新后的应用已实际观察 NativeHoverCard 的 Down 进入 Native author 子窗口及 Escape 返回触发器；自绘 HoverCard 同路径通过。DatePicker 也已实际观察越过父窗口下边界、选择日期回传并关闭；NativeDatePicker 显示平台 Popover 材质和箭头。仍不代表全平台/全部关闭路径验收。

- Gallery 补充集成：日期/范围、Tooltip/HoverCard 接入桌面组件对；36项 Gallery 检查通过。新增9项检查覆盖独立日期选择、指针穿透提示、共享名称/订阅状态变化、双向样式切换和离页清理；审计先复现共享状态更新缺少子窗口失效，再补内容 Revision。小窗口示例同步覆盖两套 API 和共享值刷新。
- 本次阶段收尾：`go test ./...` 264项检查通过，`go vet ./...`、`git diff --check` 通过；离线 `go generate ./...` 后 colors_gen.go 无差异。macOS Gallery/Popups `.app` 构建、Linux/Windows amd64 Gallery 交叉编译通过；浅深色与主题页截图重新生成，浅深色布局已查看。最终完整复刻与全平台运行矩阵仍未完成。

- Accordion/Collapsible：`go test ./pkg/display -run 'Accordion|Disclosure|Collapsible'` 11项检查通过。回归先复现方向键落在禁用标题、缺少受控面板关联和单开回调含两个打开项；修复后覆盖标题级别、边框/循环、稳定键重排、一次性变化、关闭内容立即失去输入/AX、减少动画时焦点回收、外部焦点不被夺取、真实高度反转及空闲收敛。独立审计补充初始闭合保留子树的 AutoFocus 争抢，改为首次打开时创建后保留状态，并在关闭期间禁用但不染淡内容。框架 direct Focus 无视祖先的剩余限制明确记录在对齐清单；未修改或私自访问运行时内部。

- Resizable：`go test ./pkg/display -run Resizable` 16项检查通过，display全包118项和 vet 通过。覆盖多面板约束推挤、折叠/恢复、真实拖动/取消、横纵向/RTL键盘、原生AX合成方向键、稳定ID/焦点、JSON回载、外部布局替换取消旧拖动、嵌套及禁用拖动时内容仍可编辑。1500次确定性约束分配先复现两侧折叠转移不匹配导致总和96，再修复并验证总和100、有限值和每个范围；非凸折叠模式无解不覆盖调用方布局。旧DIP包装保留原生Split，只修已复现的非有限受控尺寸。

- Flyout 尺寸/方向：`go test ./pkg/desktop -run 'SizeCapability|ParentDirection'` 8项检查通过，独立审计通过。验证普通/平台容器的仅宽高变化保留原子窗口与局部计数、不可变选项仍替换、无尺寸能力后端保留重建，以及父RTL作用域跨独立Context真实改变子布局、打开时切回LTR只刷新一次、空闲不重绘。

- Menubar：`go test ./pkg/desktop -run 'Menubar|MenuContent'` 22项检查及 vet 通过，其中17项针对Menubar。验证独立Context、上下打开首/末项、左右/RTL、同窗口指针切换/重锚、子菜单优先、快速child Right/Down/Enter在父刷新前选中新菜单、空菜单、禁用/移除/重复key、结构/外部Revision刷新、结果/回调各一次、显式Tab邻接目标、关闭/迟到回调/所有者消失。NativeMenubar选项走系统菜单、不画入父树并恢复调用方菜单策略。没有把缺少的OS tracking hook计为已验收。

- 实际窗口焦点能力：Host的 `FocusState` / `ApplicationState` 与对应可选后端能力统一查询原生窗口焦点和应用活跃。2项专门检查验证app活跃但ShowInactive子窗口未聚焦、实际焦点/失活变化、关闭后不报告所有权、缺失key/后端能力为unknown及缺失原生父窗口为inactive。

- HoverCard 停留/关闭补充：`-run 'Tooltip|HoverCard|InactiveHover'` 34项检查通过。先复现默认drawn-active子窗口在指针离开后永久保留，以及完成键盘交接后app仍活跃导致失焦窗口永久保留；改为真实focus与仅pending交接桥接。补充原生关闭通知先于blur时父输入框焦点不被抢回；最初的交接、应用失活及未知后端保守关闭仍通过。desktop全包114项及 vet 通过。

- Navigation Menu：`go test ./pkg/navigation` 42项检查及 vet 通过，含新增33项和4项独立临时审计探针。覆盖独立Context、描述/当前路由、网格/自定义内容、一次选择、原生系统策略、hover intent/间隙/取消、RTL/兄弟/动态项、缺Host错误不重试、Revision、真实尺寸/位置350ms与减少动画、连续快速重定向像素/局部状态/caret、Tab首末/空内容/编辑器吞Tab、失焦后取消交接保活。审计先复现禁用分组后代仍可选、自定义Tab陷阱、选择后同hover重开和静止指针反抢键盘切换；修复后严格回归通过。未访问内容延迟构建，已访问隐藏内容动态显式AutoFocus的运行时限制继续保留。
- macOS 最终构建实际启动：Menubar File→Right→Edit→Enter 后菜单关闭、焦点恢复Edit；Navigation Documentation独立窗口显示描述链接，Right切到更宽的Components富内容窗口，Tab/Enter激活Browse components后父界面显示`Selected: buttons`并恢复Components触发器。布局页读取一度超时，进程采样显示正常AppKit事件循环；改用应用完整路径连接后恢复读取。未将未完成的实窗面板拖动或所有窗口边缘行为记作通过。

- Gallery 高级示例：43项检查及 vet 通过，新增7项验证Menubar两种路由/同窗切换、富导航选择/静止指针下的键盘切换/离页清理、RangeSlider及嵌套Resizable实际改变State并跨页保留。测试先复现选择后重开和旧指针反抢兄弟切换，组件修复后原断言通过。浅深色/主题页截图重新生成，另渲染并查看Forms、Data和Navigation页面；截图不能替代未完成的原生输入矩阵。
- 本次阶段收尾：`go test ./...` 共364项检查通过；`go vet ./...`、`git diff --check`通过。`go generate ./...` 后颜色生成文件无差异。macOS `.app` 构建及最终导航实窗路径通过，Linux/Windows amd64 Gallery交叉编译通过。完整组件/变体视觉对照、运行时inert焦点约束及Windows/Linux真实运行仍未完成，完整目标保持进行中。

- Avatar：10项针对性检查及 display vet 通过。先复现默认直径偏离Vega、没有原生image角色和无效解码图像隐藏回退，再验证三种尺寸、延迟背景/回退、应用控制的加载/错误/成功状态及一次通知、Cover真实像素、圆形裁剪外的角标、头像组重叠/计数/RTL/重排和空闲收敛。独立审计复现成员Key="remaining"撞上内部计数标识导致panic；内部使用私有类型的标识后，延迟/删除计数回归通过。组件接受MyGo已解码ImageSource，不隐式联网；应用异步加载后经Window.Update发布结果。完整图像滤镜/示例和跨平台视觉验证仍待完成。

- Card：`CardWith` 新增默认/小号与可覆盖间距、动作/自定义头部、分隔页脚、贴边媒体和正文；保留旧Card的统一内边距契约。9项针对性检查和 vet 通过，display/Gallery合计183项通过。独立审计先复现空Header/Media/Content/Footer回调留下未消费的Context.Key、造成后续编辑器身份改变与失焦；移除回调边界多余Key，保留私有容器身份后四种回归与正文局部草稿/焦点保持通过。已检查窄宽度换行、RTL动作位置、媒体真实边缘像素；完整示例视觉矩阵仍待完成。

- Gallery Avatar：3项集成检查通过，验证明暗主题下示例图像→首字母→加载→错误→成功恢复的真实控件联动及剩余成员计数。Data页使用小号复合Card，三个尺寸和角标实际绘制；重新生成并查看 `artifacts/gallery-display.png` / `gallery-display-dark.png`，布局无重叠。内置SVG为原生示例插画，不依赖网络或外部图像缓存。

- Modal/Drawer：38项针对性检查、overlay vet及diff检查通过。先复现关闭内容仍接收动作、长内容末按钮溢出，改为原生模态scope、立即inert及真实Scroll；覆盖嵌套trap、明确初始/返回Handle、退出反转、Alert响应式Media/Small/Pending、Sheet四边/RTL与固定页脚。Drawer覆盖原生capture/速度投影/取消/失焦、DIP与比例吸附、四边手柄AX方向、键盘、NonModal、滚动上下边界与嵌套。独立审计复现捕获中禁用重启/改变snaps仍由旧手势误关闭；暂停旧Draggable直到释放与收敛后恢复，五种重配置回归通过，正文编辑器只初始化一次。精确嵌套返回使用ReturnFocus；运行时子控件触摸接力与真实触屏/全平台矩阵未完成。
- 浮层阶段独立提交验证：将已提交HEAD导出到临时目录，仅覆盖本阶段4个Modal/Drawer文件，overlay/Gallery共91项通过，确认不依赖同时开发的Sidebar/Command改动。

- Sidebar：23项针对性检查、navigation全包64项及 vet 通过。覆盖容器自身宽度测量、移动Sheet trap/选择/Escape/回桌面、独立折叠与移动状态、nested/disabled/JSON、RTL与物理边、三种变体200ms反转、sidebar语义色、独立Host Tooltip、rail和局部快捷键。独立审计先复现移动resize关闭漏通知、宽Header/Footer越界绘制/点击、移除触发器失焦和布局前外部trigger丢Controls；修复后真实像素/命中、前后布局trigger关联/实际回焦、无有效opener回主区与空闲收敛通过。未显示offcanvas内容延迟构建，已保留隐藏内容动态显式AutoFocus限制继续记录。

- Command：21项针对性检查、overlay全包65项和 vet 通过。实际10000行虚拟化、稳定Key/重复Value/异步重排、禁用/loop/Home/End/Page、精确Binding/重复抑制、原生caret/全选/组字、custom/group/status及Dialog返回焦点均有回归。独立审计先复现Page忽略description/heading高度，按实际行高度累计后1000项混合行999→998；后续修复评分打散分组、首行孤立分隔符、overscan隐藏项继续AX聚焦和零宽容器重复失效。CommandDialog默认448宽、窗口1/3定位和180×64极小窗口限制经检查；默认子串筛选与完整上游模糊评分的差异继续记录。

- Gallery Sidebar：全包47项通过，新增集成验证真实路由选择、owner输入草稿、图标折叠、容器变窄切移动Sheet、移动选择关闭和实际返回Toggle焦点；桌面折叠状态与草稿均保持。独立审计通过。macOS实际应用命令搜索Navigation后选择，页面切换成功；通过AX激活Sidebar触发器后Header从Acme Studio变为AC并折叠分组。实窗自动滚动未可靠移动到侧栏，未将侧栏视觉/拖动记为完成。
- macOS最终构建重启后实际使用Tab聚焦Search→Return打开命令面板→Escape关闭，AX确认焦点返回Search components。macOS app构建与Linux/Windows amd64交叉编译均通过；其他平台真实运行、所有组件视觉/触摸/屏幕阅读器矩阵继续待验收。
- 本轮模块收尾：最终 `go test ./... -count=1` 共465项检查、12个包通过；`go vet ./...`、`git diff --check`通过，`go generate ./...`后生成源码无差异。所有阶段分别提交，完整复刻目标仍进行中。

- Command模糊评分：28项针对性检查、overlay全包72项和vet通过。默认子串实现先在缩写排序、清空顺序与默认分组评分3项回归失败，再接入cmdk 1.1.1的原生Go移植。固定2,299条JavaScript数值（ASCII/空白/分隔符/转置/重复键/别名/UTF-16/Unicode大小写）逐条误差≤1e-14；独立新seed生成15,986条对照再次零差异。另有5条固定及14条临时上游RangeError输入，Go终止越过原始长度的递归。清空搜索绕过自定义Filter，默认也按分数排序并保持分组连续、未分组在前；真实键盘选择仍返回业务Value。10,000条候选本机3次benchmark约7.6ms/轮、5.6MB分配，仅为M4 Pro本机证据。源文件/许可证/哈希归档，可离线再验，macOS包脚本带上cmdk许可。
- 评分阶段将已提交HEAD导出到临时目录，只覆盖本阶段Command/go.mod/许可脚本文件，overlay/Gallery共119项通过，确认不依赖并行展示组件变更。

- Item：26项针对性检查及display vet通过，先复现默认Vega留白偏差；覆盖9种size×variant真实媒体像素、两行说明/RTL、Button鼠标/Space/Enter、实际Link/Router/Enter-only、独立Header/Actions/Footer编辑器、稳定重排/list语义、空callback身份和减少动画收敛。根审计补充窄60DIP主目标溢出与自定义媒体有键盘焦点却无AX节点；改为可收缩basis与保留custom native children后回归通过。超宽固定动作在原生横向Scroll中，检查实际viewport/滚动显露/像素/外部命中及可见动作键鼠；没有伪造运行时完整scrollchild的AX几何。
- Item独立阶段验证：HEAD加本阶段Item源文件/测试及仅Item旧块提取，display/Gallery共210项通过；没有提前包含Badge或Alert/Empty实现。

- Alert/Empty：28项针对性检查及display vet通过。先复现旧Alert留白/文本间距与Empty默认Padding/空描述问题；覆盖16DIP前导图标、独立操作鼠标/键盘、RTL/错误语义色/原生Alert角色、Empty384限宽/40DIP媒体/75%输入、虚线与有限参数、10种空callback身份。独立审计复现命名自定义媒体隐藏可聚焦Button及固定190DIP动作越界；修复为命名Group保留原生子节点、Decorative真正inert、响应式堆叠和有界横向Scroll。操作编辑器跨容器宽度保持状态/焦点；实际viewport/像素/命中与Enter验证通过，滚动子节点AX保留运行时原始几何。图标槽位无隐藏键盘停靠点的额外探针通过；真实AT播报与完整平台矩阵未完成。
- Alert/Empty独立阶段验证：HEAD只加本阶段4个文件和对应旧块提取，display/Gallery共238项通过；未包含并行Badge或Gallery新页。

- Badge：21项针对性检查与vet通过。先复现旧徽章高度20.13DIP/父级stretch和Destructive实底色偏差，完成6种变体明暗真实像素、原生Button/Link/Router/Enter-only、逻辑12DIP图标/6DIP边距、AX/RTL/禁用祖先/Busy/焦点/Invalid、空callback身份和有限宽长文本。独立审计用真实像素先复现静态/Button/URL的默认与custom Link文字hover均0变化；CSS偏移4在native20DIP裁剪外，改为继承下划线及1DIP偏移后6种回归通过，独立默认/custom探针各出现30个可见下划线像素。
- Badge独立阶段验证：HEAD只加Badge源码/测试和旧块提取，display/Gallery共259项通过，Gallery新页仍单独提交。

- Gallery内容/反馈页：全包51项与独立审计通过。新页包含6种Badge、实际Button/Link、3种Item/媒体、独立Pin/主动作、Alert undo/dismiss与Empty输入/创建；2项集成检查验证动作隔离、URL、跨页草稿/状态和模糊命令注册。命令项从统一pages注册表生成，新增页不会遗漏搜索入口。明暗 `artifacts/gallery-content*.png` 已重新生成并查看，修正示例icon-only按钮的显式Label及输入起始对齐。
- macOS最终app重启实测：键盘打开Command，输入 `cntfdbck` 仅留下Content & feedback并Enter进入，焦点返回Search；AX激活Pin只变Unpin且没有触发主动作，再激活Open显示Opened quarterly report。真实New project name输入Runtime project后Create显示Created project: Runtime project。未打开任何外部示例URL。静态截图与这些流程不替代完整平台/AT/视觉矩阵。
- 本轮收尾：`go test ./... -count=1` 共551项、12包通过，`go vet ./...`与`git diff --check`通过；macOS app及Linux/Windows amd64交叉编译通过。cmdk归档源码/许可证与不可变提交逐字节一致，app内携带的许可也匹配。完整组件复刻继续进行，未改变原始范围。

- Motion基础：新增 `ApplyTransition(c, element)`；先复现进行中的背景/边框颜色过渡在切换ReduceMotion后仍保留混色，再验证立即到目标像素、取消待绘帧、后续即时更新及恢复普通动画。底层 `Transition` 值 API 保持兼容；SDK调用点迁移作为后续独立子阶段。

- Input/Textarea：13项针对性检查、control vet与独立审计通过。先复现父line-height改变Input尺寸、错误实底色和Textarea八行增长上限，再验证全宽/24·32·36·40DIP、明暗真实像素、文字缩放光标、UTF-16 typed/paste/undo、readonly/password/search、可消费OnPaste/OnTextKey、候选Enter仅提交文字及第二Enter提交表单、Field作用域与InputGroup保留显式Disabled/Invalid及回调。用户拖动resize仍待实施；原生客户端检查不替代真实输入法候选窗与AT验收。

- Choice：39项针对性检查、control vet和独立审计通过。补齐受控Mixed、Vega Switch默认/小号尺寸及RTL拇指、Radio方向约束/循环/描述、Toggle自定义视觉和多/单选组即时Changed。明暗像素、禁用祖先、原生Space/方向键、稳定重排、边框合并与运行中切ReduceMotion均验证。文字与视觉槽位保持单一原生输入/AX目标；真实AT与完整跨平台视觉矩阵继续跟踪。
- Choice独立阶段验证：HEAD仅覆盖本阶段button/choice源码与choice测试，control/Gallery共126项检查通过；未提前包含OTP或Routes增强。

- Routes：原生Breadcrumb链接/当前页、独立Flyout与系统祖先菜单、结构Revision及迟到选择重查；Pagination的MaxInt有界计算、显式零/预算、不合法配置、真实URL/Router、compact/RTL像素与长文本包含均通过。独立审计复现前后按钮目标页作为Key造成连续Enter失焦，改为稳定身份后Button/Link两路径连续激活与回调通过。完整navigation 90项及vet通过；真实平台/AT及超窄视觉矩阵继续跟踪。

- SDK反馈减少动画：所有标准颜色反馈调用改用ApplyTransition，Tabs匹配指示器也在Reduced时跳过注册。motion核心加组件共6项检查通过：运行中改变Button/Badge/Input/OTP/Tabs/Pagination后切偏好，浅/深整幅图像立即等于独立静态Reduced参考，取消全部待绘帧，恢复普通动画并正常收敛。迁移独立diff审查通过；Popup/Modal等presence presets保持原API，完整presence/平台矩阵继续跟踪。
- SDK反馈迁移独立阶段验证：HEAD仅覆盖本阶段调用点与components_test，`go test ./pkg/... -count=1` 共582项、9包通过，未包含待提交OTP与Gallery集成。

- InputOTP：40项检查与独立审计通过。覆盖36DIP与自定义槽/分组、默认不限字符、显式过滤/正则、原生选择/覆盖/拖选/RTL、剪贴板/非空转换粘贴undo/redo、IME选区取消及clear→commit、明暗像素/caret居中、reentrant回调和idle。审计先复现取消preedit静默丢字、转换粘贴raw未Changed导致漏回调、极大间距/组数造成非有限几何，修复后均有正式回归。完成回调与转换粘贴后的末尾选择核对精确input-otp1.4.2源码。真实逐格候选窗定位及特殊空clipboard转非空的undo事务缺少公开API，明确记录限制，没有伪造运行时或撤销历史。
- OTP源归档：官方不可变commit的input.tsx与LICENSE逐字节核对，SHA-256分别为 `9045a8bbed30895c1749ae71b98460181bb8f68bfca54e212591643c4545dcc2` 与 `079031192fea535a245aa64a637d8433c406b82a14f91c3606b1d63e41c00152`。仅对该归档源码保留其原有行尾空格；MIT通知与macOS许可打包同步补齐。
- OTP独立阶段验证：HEAD仅加入本阶段OTP源码/测试、旧块提取及数字过滤兼容测试调整，control/Gallery共165项检查与control vet通过；未依赖同时开发的Textarea resize或Gallery新示例。

- Textarea resize：22项正式检查与独立审计通过。新增TextareaWith保留原生Input并用独立sibling splitter提供drag/capture/cancel、上下/Home/End/Shift及AX桥接键、RTL、caller JSON高度、auto恢复、外部重配置/blur/卸载取消和无idle。独审先复现同批原生Down事件读取旧Draw高度导致丢增量，改为从已接受caller高度继续后累计/反向回归通过。原生undo、caret、编辑器局部状态、只读/仅禁resize与多实例均验证。
- 组合输入回归：隔离HEAD加新Textarea模块、保留旧InputGroup时，Field真实编辑器名称缺失和内层重复边框两项失败；InputGroup分别处理布局container与fieldTarget原生editor后两项通过，addon焦点/undo/resize仍正常。macOS实窗已编辑Bio并Tab到手柄，连续两次Down后原生AX高度64→80；真实指针/触屏与AT/跨平台矩阵未宣称完成。
- Textarea独立阶段验证：HEAD仅覆盖新resize/source测试及InputGroup目标解析，control/Gallery共187项与control vet通过；未提前包含Gallery新示例。

- 长页码补充：明暗截图发现百万页边界被截为“10…”。增加独立native Text测量回归后，旧36DIP控件先失败；长页码改为内容宽度，一至三位仍36DIP。navigation/Gallery共147项通过，文本不再依赖只有AX可读的完整标签；超窄容器/完整文字缩放平台矩阵继续跟踪。
- 长页码独立审查：41条临时公共API探针覆盖Button/URL、LTR/RTL、Compact和1/2/100/百万/MaxInt页码；数字实际像素逐项等于完整native文字参考，普通页码宽36与长数字留白均正确。临时探针未写入仓库，正式MaxInt文字/边界回归保留。

- Gallery表单/路由集成：全包57项、6项新增集成及独立审计通过。验证码3+3分组显式数字过滤/完成反馈，2+4邀请代码不限字符且转换粘贴；原生Mixed权限联动、小号Switch与多选格式ToggleGroup实际改变State。Bio使用TextareaWith保存草稿/高度，跨页保留。面包屑按Menu style选Flyout或OS菜单并回传祖先选择；百万页分页支持连续Enter/远端边界，长数字可读。隔离HEAD加本阶段Gallery文件后57项通过。
- Gallery明暗像素：重新生成并查看 `artifacts/gallery-controls{,-dark}.png`、`gallery-navigation{,-dark}.png`、`gallery-form-choices{,-dark}.png`；新模块布局与颜色正常。修正Header的Native Go徽章显式居中；详细页截图验证完整百万页标签及底部状态行。截图不是完整组件视觉矩阵。
- macOS实窗：重启本次构建，Bio输入Runtime bio draft后Tab/Down/Down，原生splitter的AX高度64→80；粘贴123-456后验证码为123456并出现完成状态；AX激活Mixed All permissions后其与Edit projects均为On。Breadcrumb独立Flyout选择Forms，回传Selected parent: Forms并回焦Show parent routes；随后键盘切System native，OS菜单选择Patterns回传Selected parent: Patterns并关闭。未打开外部示例URL；未实测真实中文候选窗、指针resize/触屏或所有AT。
- 本轮最终收尾：`go test ./... -count=1` 共700项、12包通过；`go vet ./...`、`git diff --check`通过。macOS最终app重建、Linux/Windows amd64交叉编译通过。OTP归档源码、Git内原始字节、manifest SHA与app内MIT许可匹配。各模块分别提交；完整复刻与平台/AT/视觉验收仍进行中，未缩减原始目标。

- Loading：4项基线回归先复现Spinner不继承当前颜色、减少动画重绘仍改变转轮、Skeleton的pulse easing偏差以及instant frame继续脉冲；修复后含扩展共8项检查通过。Spinner用固定Loader2 SVG并匀速一秒一周，实际图形在0/90度逐像素等于独立SVG参考；Skeleton在250ms像素与CSS曲线数值参考一致。动态reduce/static/恢复、原生未知进度、decorative、尺寸/百分比/圆形及卸载idle均覆盖，Busy Button使用同一视觉且原生busy/disabled检查通过。
- Loading来源与独立阶段：核对0.474.0 tag及不可变Lucide提交的SVG/ISC许可字节一致。HEAD仅覆盖本阶段loading、共享indicator、Button接入与旧块提取后，display/control/Gallery共412项通过，未依赖并行Typography/ScrollArea/CodeBlock改动。radius×.8由本次固定globals.css确认，本阶段没有把正确圆角当作实现bug；旧版减2px对照已撤回。
- Loading独立审查：2条临时公共API探针覆盖Busy XS按钮单一原生busy/disabled目标、运行中全图静态参考对比、解除loading后实际激活，以及instant/disabled祖先下无idle/动作；全部通过，临时文件已删除。

- Typography/Kbd：21项聚焦检查与display vet通过。先复现Kbd在Column内stretch、H2缺pb8/分隔线，以及局部mono父字体污染键帽；修复后完整body family/fluent font像素参考、原生RichText选择/复制/链接、代码背景padding、文档间距/Scope、RTL引用与列表标记、带稳定Key的交互列表/表格、empty callback身份、Tooltip明暗逆色及真实外层Button键鼠均通过。TypographyTable为等宽静态语义表格，不把此项当成完整HTML排版引擎。固定MDX blob为 `21953c1ed037c79181e7ca03e1898f31eefe4901`，与已缓存checkout tree一致。
- Typography独立阶段：HEAD仅加入Typography/Kbd源码与正式测试、保留AspectRatio旧块，display/Gallery共296项通过；17个新增对照文件（MDX、15个base排版示例与globals.css）逐个git blob核对一致，再记录SHA-256。未依赖并行CodeBlock/ScrollArea或Gallery新页。

- CodeBlock：33项正式检查通过，独立审计另5条探针通过并清理。先复现Language切换丢选区、没有确认后的两秒复制反馈、非有限Height进入AX；再覆盖readonly原生编辑、UTF-8局部Edit/流追加、折行/上下限、十万行与真实水平reveal、source variant/size、startLine与不进入复制的diff/highlight背景。独审实际像素复现浮动Copy缺背景/环及Destructive Footer错误继承红色，修复后明暗回归通过；title按当前Theme.Font恢复body字体。原生显示切换快捷键不覆盖受控Wrap/gutter，其他app bindings不增长。Copy/Check SVG另与不可变Lucide提交逐字节归档。
- CodeBlock独立阶段：HEAD只加入CodeBlock源/测试与SVG assets后display/Gallery共328项通过；未依赖正在实施的ScrollArea/AspectRatio或Gallery新页。

- ScrollArea/AspectRatio：50项正式聚焦与独立审查通过。实际native内容位移/accepted callback、三轴模式×RTL、Border token 7DIP拇指、父font/currentColor/custom theme、键盘与内部编辑器、真实drag/track/capture/cancel、wheel/ScrollTo中途接管、滚动链、隐藏/重配/移除、可选原生slider range及有限参数均有回归。独审先复现corner缩短轨道后的比例误差，再按native实际inner行程修复，四方向半行程对应精确半范围。Auto条自身focus及custom→MyGo样式切换回viewport也有先红后绿回归。
- Aspect图片/限制：16:9、21:9、1:1、9:16四种比例实际Cover像素、Max约束、invalid有限正方形、With默认visible overflow与legacy clip及动态内容/空callback身份均通过。没有修改运行时私有缓存。
- 明确运行时缺口：真实probe复现1s ScrollState.ScrollTo在100ms到40后开启ReduceMotion，下一100ms仍到80且有待帧；ui/scroll.go只在新target时读取Reduced，公开ScrollState没有target getter。未通过丢弃intent伪造通过，该项保留在最终motion gate。
- Scroll独立阶段：HEAD仅加入Scroll/Aspect源码、正式测试及旧块提取后，display/Gallery共378项通过；新Gallery页未提前纳入。
- Scroll最后复核追加Invalid轴配置向真实native viewport传播，保留内容输入而不只标装饰Root；51项聚焦检查通过，作为模块修正独立提交。

- Gallery Text & layout：全包70项检查通过。新页注册在统一Tabs/Command数据源中，可用组件名和模糊页名进入。原生只读代码的语言、换行与紧凑配置实际到达editor，切换选项和离页保留Buffer，确认复制只含源文；文章链接/列表/详情表格和搜索按钮保留真实语义及动作。加载检查在正常动画下先确认像素变化/待帧，再暂停并确认静止/无待帧，避免通过测试环境禁用动画造成假阳性。输入草稿、实际内容位移及滚动条切换/离页位置保留均通过。
- Gallery字体作用域：预览与Inspector的native Theme.Font同步选择的正文字体，同时保留语义token metadata；键帽和恢复body字体的代码标题因此遵循选择，不修改窗口全局主题。
- macOS实窗：已观察新页、Plain text/Wrap/Compact/Highlight切换、Example copied成功反馈、文章详情表格和暂停加载后Sync按钮恢复可用。此前可访问性读取超时后重新连接恢复；后续下方滚动调用连续被自动化接口以“窗口状态已变化”拒绝，未观察到下方滚动结果，不计作实窗验收。头部明暗截图与下方模块的离屏明暗截图仅证明对应原生渲染路径；全平台、AT及完整视觉矩阵继续保留。
- 本次阶段收尾：全量823项检查、13个包通过；`go vet ./...`、`git diff --check`、macOS `.app`及Linux/Windows交叉构建通过。新增Gallery测试加强后另运行全包与vet复核。完整复刻仍在阶段5，未将测试数量视为最终验收。

- Gallery下方实窗补验：在同一已构建应用中通过Workspace lookup后的Tab焦点自动滚入下方区域；End按钮键盘激活后实际显示Note 17–20，聚焦Preview notes并Page Up后显示Note 14–17，切换Use MyGo scrollbars后保留当前位置，Start返回Note 01–04。原生焦点环、加载组合、裁剪的16:9插画和两个滚动条样式均已实际观察。这补齐前条记录中未确认的Start/End、键盘滚动和样式切换路径；没有将其扩展为触控、指针拖拽或全平台验收。

- Toast/Sonner：固定shadcn锁定的Sonner2.0.7官方tag/不可变提交、原始源码与MIT许可逐字节一致；四枚Lucide状态图标与同一固定提交一致，Sonner CloseIcon保留原几何。归档11个source/icon校验值与5个SVG XML核对，macOS bundle内Sonner/Lucide许可字节一致。
- Toast公开运行时审计：13项临时原生API探针验证4秒/8秒便利API、准确剩余时间/整组暂停、三项上限、同ID更新/焦点、关闭回调、隐藏后台计时、无swipe、卸载fallback及原生减少动画差异；临时文件已清理。这些证据决定新队列使用显式应用状态，旧路径保留兼容语义。
- Toast正式95项检查与交叉审查通过。覆盖默认无close、普通error中性色/富色明暗像素、Vega比例圆角、六物理位置/RTL/mobile、自定义visible数量/真实测量高度、窗口隐藏暂停与失焦区别、loading更新/原生控件身份、动作阻止关闭/取消/回调快照、ID碰撞和关闭重开、更新朗读、闭包清理、viewport卸载重挂及挂起关闭。自定义Hotkey使用实际原生KeyShortcuts并真正进入通知；Unstyled正文的编辑器、动作和更新焦点/选区检查通过。
- Toast滑动包含51项正式检查：真实native捕获越界、按钮/选中文字/右键排除、六位置×LTR/RTL、45DIP及0.11DIP/ms边界、轴锁/反向阻尼、loading/不可手动关闭、取消/失焦、更新/卸载、减少动画、400ms归位与200ms退出。普通退出先复现直接透明后改为按源400ms曲线绘制前200ms，100ms仍可见且移除后无ghost；手势100ms无错误纵移叠加。原生最小1ms定时器唤醒窗口明确保留。
- Toast独立阶段：已提交HEAD仅覆盖Toast源/正式测试、图标/参考归档和许可包脚本后，overlay/Gallery共236项检查通过；没有提前包含Gallery通知示例。Root另2项临时混合队列顺序/隐藏编辑器焦点探针通过并清理。当前全工作区929项检查、13个包（9个有测试、4个无测试）通过，vet/diff与macOS、Linux/Windows构建通过；完整原生平台/触控/辅助技术与视觉矩阵继续验收。
