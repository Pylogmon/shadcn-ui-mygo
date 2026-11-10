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

- Gallery通知集成：全包84项通过，新增14项实际联动检查。类型/六位置到达真实窗口坐标，动作PreventDefault保留通知且取消后移除，loading跨页/10秒仍在、完成更新同一ID并恢复到期，堆叠/清除/配置保留，以及Alt+T/Tab/Escape回到实际可用触发器均通过。Complete upload执行后禁用自身，因此返回目标明确选为可用Start upload，避免把焦点送入禁用控件；嵌套preview作用域下右/下边距仍精确24DIP。
- Gallery通知明暗渲染：生成并查看 `artifacts/gallery-notifications.png`、`gallery-notifications-dark.png`，控制项/动作/中性色success图标与关闭按钮可读，无重叠。临时截图导出测试已删除，图片不作为全矩阵验收。
- Gallery通知macOS实窗：UI退出旧进程并启动新bundle，已观察Notifications页、持续loading通知；Complete upload后原row身份原位变为Upload complete，按钮禁用且状态提示更新。Alt+T进入通知原生list，Escape实际返回Start upload。勾选Keep open后Review触发Notification action selected并保留通知，Later触发Notification postponed并移除通知。实窗未验证本阶段的指针swipe、触控和所有位置；这些继续保留为平台验收工作。
- Gallery最终交叉审查：先以完整State.View复现跨页后Escape请求已卸载旧触发器、Clear仅清新队列两条红测；改为当前activeTab绑定独立Handle作为跨页fallback，并同时清理MyGo兼容队列后转绿。测试继续验证切回Overlays没有旧Start抢焦点、下一Tab到Edit profile，以及实际Show toast与Start upload均被清除。独立只读14项复核通过。
- Gallery修正版macOS补验：重新启动最终bundle，Start upload后切Text & layout，Alt+T/Tab/Escape实际返回当前Text & layout标签，通知仍在。返回Overlays后叠加Show toast与loading，Clear实际移除两组文字并禁用Complete。切回页后下一Tab的AX焦点没有明确显示，未把这一步记为实窗通过；完整原生键盘/AT矩阵仍保留。最终全量931项、13个包、vet/diff及三平台构建通过。

- Shimmer基础模块：固定CSS/MDX的Git blob逐字节核对；20项内部检查及1项外部package公开API检查通过。实际原生字形像素验证空间扫光、2秒线性、RTL/reverse、一次播放恢复、明暗高光与透明度、OKLCH中间色/OKLab插值和sRGB色域映射；原生中文/字体继承/ellipsis/AX及动态减少动画、Instant和停止后的idle均通过。默认3ch估值需要已知字体，Base需显式传递非Theme祖先颜色；公开运行时没有计算样式getter，不读取私有状态。
- Shimmer独立阶段：从已提交HEAD仅加入内部实现和公开display入口，internal/shimmer与display共343项检查通过，vet及macOS/Linux/Windows构建通过；没有把尚未完成的Table、Attachment或Trace候选算入隔离验收。完整实窗、字体、广色域和辅助技术矩阵继续跟踪。

- Table静态组合：48项聚焦检查通过；header/body/footer共享测量列、row header/原生索引与跨列、caption16DIP间距、自然行高、Checkbox物理右padding、RTL、selected/expanded/hover像素、明暗自定义token、原生Select/Input/动作、稳定列/记录Key、实际编辑选区与IME、指针取消和100,000行虚拟入口均验证。Style回调配置当前真实Row/Cell，内容只按原生build执行一次，不重复调用来测量。
- Table离屏与调度：先实证离屏第十列漏测、同Key编辑器宽度变化后陈旧、多表独立成对帧永久互相触发以及Scale坐标系来回覆盖四类失败，再改为公开LayoutBounds加窗口共享postpaint补测。无绘制/AX/输入的1×1 overlay只携带数值和持久model，所有Table共享一次纠正波次；首尾插入、重排、删除/重挂、离屏多表宽度增长/缩小及normal/reduced静止后无待帧均通过。不保留Context/Element，也不读取运行时私有树。
- Table独立阶段：从已提交HEAD仅加入data.go与Table实现/正式检查，display/Gallery共455项通过，vet/diff通过；新Gallery表格演示未混入模块提交。共享原生轨道不是完整HTML intrinsic/rowspan引擎，虚拟Table保留运行时样式，所有平台/字体/AT/视觉矩阵仍需完成。

- Bubble：61项聚焦检查、独立阶段conversation/Gallery共149项与vet通过。实际像素覆盖七变体明暗/hover、Destructive浅色背景、OKLCH相对tint与混色、比例圆角、祖先颜色/字重继承及动态减少动画；普通、Button和Link的字重覆盖缺陷先红后修。原生Button/URL/Router、disabled在press中切换/取消、反应层四侧RTL物理布局/溢出可点击、原生pressed状态、长字换行与80%/Ghost100%上限、同Key流式追加/重排的编辑器焦点/选区和Message祖先End作用域均通过。源示例没有单独tails/size API，旧泛化待办已改正。
- Bubble当时保留的Link差异已在5.BB继续修正：默认组合Link分离真实交互目标与惰性视觉内容，支持原生Box/Icon并继承quiet颜色；直接文字Link仍可显式TextOnly选择。所有token已是sRGB，不能从中恢复原始广色域坐标。完整源示例/字体/平台/AT/视觉矩阵仍未完成；5.AR独立验证基于`cd5cbc9`加Bubble文件，没有提前纳入Attachment/Trace或新Gallery示例。

- Attachment：68项聚焦检查通过；固定源五状态×三尺寸×两方向、纯内容/图标/图片与Cover/圆角/状态透明度、源卡片/媒体/错误文字token、整卡Trigger与独立动作/编辑器、真实IME/文件拖入路径副本/禁用/取消与重试/进度有限值、原生字体继承、共享文字扫光及动态减少动画均验证。独立审查另验证processing→done保留输入焦点和全选、最新动作callback及真实wheel→snap；临时探针已清理。
- AttachmentGroup：真实原生横向viewport保留TrackScroll、焦点滚入、首尾键、snap和按接受位置计算的CSS边缘渐隐；稳定ID重排保留原生编辑器/选区，普通竖直滚动仍到父级，纯水平边界包含。源4DIP scroll-padding缺公开API（实测第二卡native192，CSS inset应188）及边界混合轴事件无法只转发一个轴，两项明确保留；未实现上传引擎。
- Attachment独立阶段：基于已提交`d8e7cad`仅加入附件实现/正式检查和11枚固定Lucide SVG，conversation/Gallery共217项通过，vet/diff通过；资产SHA-256/XML均核对。已生成并查看浅/深附件矩阵PNG，实际展示状态色、三尺寸、两个方向与图像状态；完整源示例/平台/触控/拖入/AT/视觉矩阵仍继续验收。Gallery新会话流程未提前纳入本模块提交。

- Trace：72项聚焦检查通过。固定源静态Line不会误变按钮、Complete与默认均为Muted、Running标题共享扫光且完成后idle、Pending整行透明度、Error与Link hover错误色、Description alpha/tnum、8DIP间距/16DIP图标/标题截断、正文物理左缩进/border/RTL、源150ms颜色和200ms高度均验证。真实原生Button/URL、可选CodeBlock复制、Dialog/Sheet、Controlled回调替换/反转、惰性KeepMounted输入选区/AutoFocus与关闭回焦、嵌套同名/重复/空ID隔离均通过。
- Trace独立审查先复现Scale后的Bounds/Draw坐标系反馈环、强制Theme字体与400字重覆盖祖先、Description/Content多余字号重置；修为公开LayoutBounds和按源继承后全部转绿。32项family/weight/ancestor/fluent实际像素/AX/idle与额外22DIP继承参考验证通过；离屏动态正文、首帧显示和多实例静止无待帧通过，不引入独立成对帧循环。
- Trace独立阶段：基于已提交`e7bb5d7`仅加入Trace实现/正式检查及固定Chevron SVG，conversation/Gallery共288项通过，vet/diff通过；16个固定示例Git blob/SHA-256以及SVG XML/校验值均核对。Native Link整行装饰点击/hover下划线、缺计算祖先字体getter导致3ch估值、compound显式Status传播和完整源示例/平台/触控/AT/视觉矩阵继续保留；Gallery会话流程单独提交。

- Gallery Table：新增6项真实集成、Gallery90项通过。选中行AX状态、备注编辑、重排前原生全选保留并在重排后实际替换、跨页状态、空表六列跨度/总额恢复、窄窗口横向滚动到Actions均验证。Status/Actions分别实际打开独立shadcn Flyout和OS Menu并回写正确记录；新组件名进入统一Search命令源。独审另验窄窗口Tab自动显露Note/Actions、空表关闭已开菜单与两类菜单键盘动作；临时探针已清理。
- Gallery Table明暗/宽窄：导出并查看1100与620DIP两种宽度的四张PNG；页脚/行/控件清晰、窄窗口按源水平溢出，额外内容通过真实滚动或键盘焦点查看。该证据不代表所有平台视觉矩阵。以`d4757d0`仅加入Table Gallery的隔离阶段，全量1206项、14个包（10个有测试、4个无测试）、vet/diff及macOS/Linux/Windows构建通过；未提前包含尚在完善的会话Gallery。

- Gallery会话：新增27项正式集成、Gallery全包117项通过。实际原生反应pressed/回复草稿/共享备注Sheet、附件五状态/三尺寸/两方向/取消重试完成/移除恢复/真实代码预览Dialog、Trace状态/原生编辑器/URL/展开保持，以及五个组件名的完整Search路由均验证。普通动画下四种opener关闭回焦、退出动画中三种离页不重开/抢焦，以及整页中文IME/跨页/reduced idle通过。离页清理通过公开ui.Local Closer持有应用状态和Handle，关闭临时弹层并取消旧目标的待回焦，不保留帧对象。
- Gallery窄屏实证修正：500宽整页中columns切纵向后仍给子列Basis(0)/Grow(1)，导致Card重叠且Vertical files被其它卡片挡住；改为纵向FillWidth/Shrink(0)自然高度，横向才等分增长。500/1100/1450的真实滚动、竖排卡120DIP边界、取消/重试与Dialog回焦由红转绿。竖排附件的长文字Preview动作移到group下方，绝对动作区只保留图标，避免120DIP卡片溢出。
- Gallery会话明暗/宽窄：已导出并查看artifacts/gallery-conversation-wide-{light,dark}.png与narrow-{light,dark}.png；四Card在宽屏两列、窄屏按自然高度堆叠，文字/控件可读，无重叠。截图使用减少动画审静态样式，不代替正常动画/所有字体与平台验收。
- macOS会话实窗：原生Helpful从3/off变4/on；气泡打开Review notes，原生Textarea编辑并Save后实际返回Open review notes焦点；Trace展开显示同一备注，原生Input编辑成功，Finish review实际禁用。附件Cancel后变Failed/Retry，Retry变Uploading25%，Complete变Done并恢复Preview report；实际Dialog显示可选择代码。以上1280宽会话流程在窄屏helper修正前的同分支构建完成，随后已重新打包最终版本。
- 最终macOS Table补验：重启最终bundle，通过Table搜索进入Data；Status INV001实际打开独立Flyout，选择Pending后父表回写且焦点返回触发器；切System native后实际打开系统Menu并选Processing，父表值正确。未将辅助技术接口点击推导成全部指针/触控验收。最终全量1233项、14个包、vet/diff、macOS .app和Linux/Windows构建通过；bundle中的Lucide及shadcn许可与归档逐字节一致。完整复刻仍未完成，平台/AT/字体/广色域/全示例视觉矩阵继续跟踪。

- 独立消费者：新增仅用公开API的examples/consumer和可重复的cmd/verify-consumer。实际以`fc7cfe1`生成模块ZIP，外部require其伪版本；GOWORK关闭、整个依赖图无replace、GOMODCACHE/GOCACHE均从新临时目录开始，固定依赖由官方代理下载并正常校验，SDK从本地标准模块代理解析而非工作树。完成外部编译、消费应用渲染、Dark/Blue/Radius12保存与第二进程仅Load后的560×400实际像素一致，以及从模块ZIP运行Gallery的明暗渲染。记录/log/PNG/ZIP保存在build/consumer-verification/run-4177292664，临时缓存已清理；已查看消费者与Gallery渲染。工具和示例vet/build通过。
- 消费者验证范围：该证据覆盖本地模块制品、公开API、无本地替换/预热缓存依赖及跨进程设置恢复。SDK没有因此发布到公共代理；PNG运行不是原生窗口启动，系统主题监听、真实窗口重启和Windows/Linux运行矩阵继续跟踪。此验证锁定已提交SDK，未把并行Message/Marker/Questionnaire/Scroller候选提前计入。

- ScrollBar意图回调：新增OnScrollIntent在有效手动操作改变原生offset之前通知，供会话自动跟随释放使用。2项真实原生检查验证wheel/key之前的位置、最新callback、程序滚动静默、禁用/非滚动键/零轴输入排除及callback更新不取消现有捕获；回调没有加入comparable布局配置。仅HEAD加这两个文件的隔离display/Gallery共490项通过，vet通过；没有提前包含Scroller候选。

- Message：15项聚焦、独立conversation/Gallery共336项与vet通过。旧90%正文二次限宽导致Bubble288而源应320、存在Footer时Avatar少32DIP位移先红后修。完整slot/group/custom Children、metadata12/500/padding12与Ghost0、8/10DIP间距、原生Reverse/RTL视觉与原DOM Tab序、空Avatar、非slot裸内容、Bubble/显式MessageSlot的End作用域恢复、稳定Key前插/空Header切换后实际IME/全选替换，以及附件/URL/原生可选文字均验证。独审另10项字体/继承/长无断点换行/复杂嵌套/RTL编辑检查与Root旧便利入口空Header全选探针通过；临时探针已清理。
- Message原生适配：无公开data-slot后代查询，Ghost与完全自组Children中的HasFooter需调用方声明，其它slot可用零wrapper MessageSlot显式对齐；不读取私有树或保留帧对象。固定registry、base文档与八个独立示例共11份Git blob/SHA-256核对。完整示例/字体/平台/AT/视觉矩阵继续跟踪，未将函数名或组合入口本身视为全矩阵完成。

- Marker：32项聚焦检查、Root独立conversation/Gallery共368项与vet通过。旧默认RoleStatus误报、12DIP字号、Border全边圆角框先红后修为源静态语义、14DIP/20DIP文字、16DIP最小高度与底边框/padding8。实际像素验证原生字体/字重与token、Separator两侧物理间距/RTL、持久下划线3DIP和明暗字形扫光；真实URL/Router/Enter-only、整行padding命中、Handle/fluent Label、禁用、原生选择/IME、稳定Key前插与空图标槽，以及折叠/Drawer组合通过。明暗PNG已导出并查看；截图与局部检查不代表完整平台/字体/AT视觉矩阵。
- Marker原生适配：保留Marker便利签名与旧变体常量，但默认不再强制live状态；仅Status/Busy显式使用RoleStatus，Busy另外携带原生busy。根Anchor用实际native Link覆盖惰性视觉行，避免原生Link的text-only子节点限制；MarkerText提供自然名，任意自定义native内容需Label，否则使用URL，fluent Label仍命名实际Link。Shimmer自定义颜色与3ch字体估值显式传递，不读取未知计算祖先样式。primitive/registry/base文档/八独立示例共11份固定Git blob与SHA-256核对，完整示例/平台/辅助技术与视觉矩阵继续跟踪。

- Bubble复杂Link修正：先复现MyGo文字Link包含Column/Icon时的真实构建panic，再用实际native Link覆盖惰性视觉面修复。71项Bubble聚焦检查、基于`feca45e`加Bubble三文件的conversation/Gallery共331项隔离检查及vet通过。原生Handle、URL/Router、Enter-only、整面padding点击、fluent Label、无重复AX和无Label时URL回退均通过；Muted/Outline/Ghost在LTR/RTL与祖先字体/颜色下像素匹配普通内容面。直接文字Link自然名称原回归保留并显式TextOnly。
- Bubble独审另15项真实原生检查覆盖明暗、长文与复杂嵌套、空callback、同Key前插及内容清空、实际焦点/Space/Enter/URL；全部通过，临时探针已删除。默认复杂Link与显式TextOnly的内容/名称/样式边界已写入指南；未读取私有树或修改MyGo模块缓存，完整平台/AT/视觉矩阵继续跟踪。

- MessageScroller：66项聚焦（含旧virtual入口）、基于`092d8af`加本模块13文件与源归档的conversation/Gallery共401项隔离检查及vet通过。跟随/自由/锚定/跳转、初始位置和待出现ID、四种对齐/负margin/显式padding、单/批量锚点、流式增长/尾部空白消耗、前插与同帧回答折叠、可见ID/阅读线、真实细滚动条拖动/取消/透明时操作、原生wheel/键盘打断、动态减少动画、隐藏/卸载/重挂载、原生编辑器选择/IME与稳定Key均验证。
- MessageScroller交叉审查：先复现并修复Viewport独立卸载残留可见性、空→非空未恢复follow、排队ID被边缘命令覆盖、边缘处理器吞嵌套原生Scroll、手动输入未撤销尚未应用命令，以及Content卸载原生glide残留。独审原始8项探针最终全绿并清理；后加入第二实例、离屏、0.75缩放和三实例idle验证，避免补测量反馈环。明暗1320×790双实例PNG已导出并由Root查看，左侧锚定留白、右侧历史/跳转/fade符合该状态。
- MessageScroller适配范围：组合入口使用完整原生Scroll；旧10k可变高度virtual List入口继续保留，不声称提供VirtualViewport锚定。默认保留任意嵌套原生滚动，ContainScrollAtEdges仅为平面正文显式开启，缺少公开overscroll/wheel处理后查询；RoleLog缺独立ARIA relevant/live字段。31份参考文件blob/SHA核对，归档15原文与MIT许可、ArrowDown来源校验均通过。完整示例、触摸、平台、字体/AT/视觉矩阵仍未完成。

- Questionnaire：63项聚焦、基于`1dd7536`加本模块6个Go文件与源归档的conversation/Gallery共504项隔离检查及vet通过。严格answered/skipped导航、optional显式Skip、inactive挂载/输入与AX排除、single草稿互斥及multiple成功值顺序、受控步骤/Checked/Value、最新默认值Reset及取消事件、原生radio/checkbox/Input/customTextarea、scoped快捷键/IME/方向/光标、校验失败回焦点及Error关系、Progress变更通知、动态Questions删除和nil/非nil choice metadata均验证；旧required-only向导与Result fixed+freeform行为保留。
- Questionnaire交叉审查：16项独立原生检查先红后修并最终通过，覆盖runtime-only choices与Tab入口、最新DefaultItem、相同/全disabled Reset回调、Choice卸载、uncontrolled PreventDefault保留实际文字与先前选择状态、换项清IME、custom Editor ReadOnly、空字符串choice、disabled快捷AX及single→multiple不复活草稿。最后补完整Questions移除重插和runtime快捷顺序，正式回归已吸收。临时探针已清理。560×640明暗PNG已导出，Owner与Root均实际查看，multi/描述/原生自由输入及导航无重叠截断。
- Questionnaire来源与适配：32份固定参考文件blob/SHA核对，14原始primitive/collection/hooks/测试/说明及MIT许可归档。每题一个自由输入沿用Answer.Text，任意多个独立自由输入集合仍待实现；HTML email/date/number类型、ValidityState与浏览器约束气泡没有公开原生对应，由Go验证/实际Editor适配。整体With离页卸载保caller-owned草稿是Go扩展，显式Questions移除则清运行时与答案。原生Group替代fieldset、RoleAlert交runtime自动live、Progress单独变更通知均未冒称真实AT验收；完整示例/字体/平台/AT/视觉矩阵继续跟踪。

- Gallery会话组合：新增12项聊天/消息/Marker和9项问卷集成，Gallery共138项；最终`go test ./...`共1439项、16包（10个有测试、6个无测试）及全量vet通过。完整Message槽位、复杂Bubble原生URL与实际剪贴板、静态三种Marker、真实340DIP聊天、600ms分段回复、Send/Stop回焦、Follow与Start/End、手动阅读/历史前插、原生IME提交与离页暂停/回来续写均覆盖。问卷改用严格源组合，覆盖受控步骤、多选、Unicode字数Go校验、实际Textarea、显式Skip、独立保存快照与Review/Edit/Save changes/Reset及跨页焦点。
- Gallery交叉审查补充6项聊天原生检查和2组问卷完整路径：离页24小时/3小时不追赶、回页600ms单tick，流中下一草稿/IME与Stop再Send，500/800/1450完整页面手动阅读/前插/Jump与不压缩边界；480深色/RTL、280个emoji边界、空白多行、281汉字校验、编辑固定题保留原notes/Skip、重置及idle。截图检查发现custom Editor只同步Invalid语义而未传Textarea样式，现显式传`item.Invalid`，实际Destructive边框像素回归先红后绿；SDK未为此改动。
- Gallery视觉/构建：组合明暗620/1100四图及问卷form/notes/review明暗六图均由Owner查看，Root抽查宽浅/窄深组合、错误输入浅色与复查深色。最终macOS app、Linux/Windows amd64构建通过。macOS真实窗口验证必答错误回焦、optional直接Next提示/Skip、ASCII多行输入→结果→Edit保留→Reset；聊天发送→停止回焦、完整分段回复、Start、前插后原消息屏幕位置保持、Jump到最新及新构建输入错误红边框均观察到。主页面End键滚动已核验；合成wheel未作为该次实窗证据。实际系统输入法、读屏与Windows/Linux运行矩阵仍待最终验收。

- 会话阶段消费者复验：最终代码提交`09a56667eac94250dc89a8fbc42534c5458faf77`的本地模块ZIP，以`v0.0.0-20261109111738-09a56667eac9`在全新模块/构建缓存中验证。外部require依赖、实际模块图无replace、GOWORK关闭、编译、消费者保存→第二进程恢复的像素对比、模块Gallery明暗渲染均通过。证据`build/consumer-verification/run-1723552001`；ZIP SHA-256为`1f1e7069991f4a38803b4bec0d4630f42a3a28cdc378b9821e4c8593934a0ba4`。Root核验receipt/模块图/哈希，并实际查看恢复后消费者与模块Gallery深色PNG。该步骤仍是本地制品验证，未公开发布，也未替代原生系统主题监听、实际设置重启或三平台窗口验收。

- 横向Bar SDK：71项Chart/Donut/Polar聚焦（28顶层、43子例）、基于`7b173dc`加本模块三文件的隔离display/Gallery全包及vet通过；旧chart_test.go未改。新配置覆盖物理X数值/Y分类、分组与正负/绝对比例堆叠、单系列分类配色及对应Tooltip/Legend、独立轴格式/隐藏与轨道回收、圆角0/无效值、原始标签/数值与missing跳过。原生实际像素、Scoped Up/Down或Left/Right/Home/End/Escape、外部editor隔离、动态方向/resize/数据缺失后mark身份、颜色过渡中切ReduceMotion与最终idle均验证。
- Cartesian真实修正：独审先复现RTL把native Grid轴翻向而marks仍物理放置、PointerPosition为local却除以变换后Bounds、长类别把绘图区挤到0；现锁Plot物理LTR、轴文本恢复应用方向，按LayoutBounds归一化命中、横向默认60DIP分类预算。还修正小于1/1e-99数值域被旧hi=1压扁，以及normalized极值轴错误乘回原量级。独审4项非等比Scale/Translate/RTL/长标签交叉、实际zero baseline、静止pointer与键盘、数据清空重载/双图/原生editor/keyed sibling复核通过，临时探针已清理。
- 横向Bar来源/视觉：三示例与base文档四份原文blob/SHA和共享MIT许可核对，base包装组件blob补录；源码明确mixed是单series逐category颜色，negative并非横向且含未实现的LabelList。1280×760明暗四面板PNG（grouped/diverging/normalized/category fills）均由Owner与Root实际查看。Native归一化按每分类sum(abs)保持正负方向，未冒称混合符号的Recharts任意stackOffset完全等价；任意ChartContainer/Recharts组合、Brush/Reference/LabelList/free margin、bar几何和path插值、全字体/AT/平台/视觉矩阵继续跟踪。此次隔离不纳入并行Gallery新面板或Tooltip/Legend样式候选。

- Chart默认面：基于`0cdb9a6`加两候选Go文件的26项聚焦、隔离display441/Gallery138共579项及vet通过。Vega的min128、padding10×6、RadiusLG、Background、Border/50、dot10与双shadow-xl、标题/series/value字体和颜色、空payload/空标题/宽格式化值；Legend swatch8、gap6/16、top12与Chart组合不重复留白均覆盖。Cartesian grid横向/纵向×明暗分别用独立RGBA像素检查，修正两方向均为Border/50。
- Chart面交叉复核：6项独立原生检查验证A128×.5后A64的手工src-over、真实阴影足迹区别于任一单层、serif/monospace祖先28px/700下字形crop及颜色、260↔400动态宽度×LTR/RTL整项图例无重叠和idle。曾用同公式reference掩盖的重复Alpha已按锁定MyGo公共API原文纠正，reference改独立RGBA构造后先红再绿；未修改已经正确的Area/Radar Alpha乘法。去掉原Legend.Wrap的候选在300宽会把Other裁成Ot，现保留native换行并加入300/400双方向正式回归。所有诊断探针已清理。
- Chart面来源/适配：shadow-xl用shadcn lockfile锁定的官方Tailwind4.3.0包完整SHA-512验证，原theme.css/MIT及证据归档；参数与早前main缓存一致，未混淆来源。打包脚本现随原生app携带对应Tailwind许可。保留系统字体LineHeight1时14.13DIP左右的度量下限，没有硬设CSS26DIP tooltip高度来裁字；locale、其余indicator/icon/formatter组合和完整平台/字体/AT/视觉矩阵继续跟踪。

- 极坐标Chart局部主题：Gallery截图复核先发现Blue作用域中Pie/Donut等实际图形使用Neutral配色而Legend正确；根因是延迟Draw在作用域恢复后读取Context。现在构建时解析全部类别/系列颜色及数值几何，Draw只持有独立快照，不保留Context、ChartProps或应用回调。Pie/Donut单双环、Radial单/堆叠、Radar完整/缺失数据共8分支的明暗16子例先红后绿；显式CategoryColors/Series.Color及局部回退、三作用域动态换色与原生焦点/选择、serif/monospace原生字形精确对照共30项正式检查通过。
- 极坐标隔离与交叉复核：以`daaf2f7`仅加入chart.go和chart_polar_theme_test.go，display471/Gallery138共609项及双包vet通过；不纳入Gallery在途示例。独审同一组10项检查从基线全红到候选全绿，覆盖实际sector/Legend像素、所有绘制分支、Blue→Orange动态主题、并列Blue/Rose独立点击/Tooltip/AX焦点、单边换色不串及最终idle；诊断探针已清理。1200×850明暗PNG均由Owner与Root实际查看，Neutral窗口中Blue Donut、Green Pie、Violet堆叠Radial及Rose Radar使用各自配色；此修复未替代完整主题/字体/平台/AT/视觉矩阵。

- Gallery图表组合：基于已提交SDK`d397be3`，新增16项正式集成，Gallery共154项；最终`go test ./... -count=1`共1553项、16包（10个有测试、6个无测试）通过。Visitor explorer覆盖方向及对应真实按键、分组/堆叠/绝对比例与原计数、正负变化、五种实际Bar分类色/Legend、数据原位更新/恢复、离页保配置及500 RTL/1450整页滚动可达。第二卡六类图表覆盖原生检查、240方形与Cartesian几何、Pie实心/Donut中空/Radial环轨和Radar实际轮廓像素，两个选择状态互不影响。
- Gallery图表交叉复核：既有Bar独审路径外，新增整页500 RTL/1450的3项原生检查；真实滚到第二卡后逐一家族操作，Radio的Home/End、图表Home/Right/Escape及上卡选择保持均通过，局部Blue配色和最终idle正常。620/1100组合明暗四图、500 RTL深色及1450浅色整页两图在最终SDK修复后重导并由独审逐张查看；Root抽查窄深组合及宽浅整页，控件自然换行、图表不压缩、Donut颜色与Legend一致。导图用4个Gallery文件哈希与workspace保持一致，临时诊断导出器未入正式回归。
- Gallery图表构建与实窗：最终macOS app、Linux/Windows amd64编译及全量vet通过，app继续携带Tailwind许可。macOS实际切Blue局部图表主题并观察Donut五段/Legend一致，Home/Right检查Safari原值200；Update/Restore确实改变并恢复Bar原生marks，Shift+Tab自动滚入上卡，Space打开正负变化后Home/Down检查March -207。Share of total配合坐标/网格/Legend使用-100%到100%轴而AX/Tooltip保留-207/-90；改Vertical后Home/Right检查February原值205/160。主页面End和Tab焦点揭示已观察；合成wheel只作部分滚动尝试，不代替完整滚轮/触摸或Windows/Linux真实运行验收。

- 图表阶段消费者复验：代码提交`d72dbc0238b7a6455be8838917b34d665469b224`以`v0.0.0-20261109122650-d72dbc0238b7`从标准模块ZIP读取，空模块/构建缓存、GOWORK关闭、实际模块图无replace。外部tidy/build、消费者首次渲染与第二进程恢复后的实际像素、模块Gallery明暗渲染和环境检查八步通过。证据`build/consumer-verification/run-3091013942`；ZIP SHA-256为`1fa04441a4665e63c2f909afe2ce822dd1ca90d03ef56a86474fc8a6cf17ba69`，消费者源码SHA-256仍为`b678ffb61c408baf655f9a26dfc89effa00a3c54041f25949d347a4c85efa91f`。独审与Root实际查看恢复消费者和Gallery深色PNG，Root核验receipt、实际模块图与文件哈希；本地制品验证不代表公共发布或原生三平台运行完成，不混入后续在途Go修改。

- Radar提示配色：先复现CategoryColors把两个系列的提示色块都替换为同一类别色，而实际Radar轮廓/Legend仍为不同系列色。修复仅让Pie、Donut与未堆叠Radial采用分类色，Radar保持系列颜色。5项正式检查覆盖明暗×默认局部Blue/显式系列色、原生提示色块实际像素与Legend一致及Home/End检查原值；显式色错误场景先红后绿。以`af5274b`仅加入本模块chart.go与chart_radar_tooltip_test.go，display476/Gallery154共630项及双包vet通过，文件哈希与workspace候选一致。
- Radar交叉复核：6项独立原生检查直接操作默认Tooltip，色块与实际选中mark和Legend一致，East/West键盘切换、实时系列换色不受分类色影响；Pie/Donut/未堆叠Radial的分类色及堆叠Radial的系列色未退化。复核另发现未堆叠极坐标图表会提示未绘制系列，作为独立后续修正保留，不包含在本次通过结论中。

- Chart自由组合SDK：40项正式检查（17顶层、23子例）及基于`be7dd98`仅新增chart_composition.go/test.go的隔离display516/Gallery154共670项、双包vet通过。独立ChartContainer与Config的统一/明暗颜色、严格live作用域与深复制、nested/sibling恢复、真实raw Plot/Rule/Anchor/编辑器、动态主题/字体/方向/禁用/原IME及选区、Local生命周期与idle均验证，不依赖高层Chart。源base chart.tsx和完整Vega样式原文已补归档，与另外四份参考及共享MIT许可的blob/SHA逐一核验。
- Chart内容组合覆盖：Config原生标签/图标，payload顶层→datum→原键查找，Name/DataKey的nil/数字/显式空值，空NameKey/LabelKey的显式presence，dot/line/dashed真实像素、单raw payload嵌标题、过滤none后formatter索引、null/undefined与原生可编辑formatter、label formatter和en-US/de-DE数值格式，Legend上下留白/隐藏图标回退/原生标签身份均通过。默认JS式逗号数组仅声明[]any/[]string/[]float64/[]int，其它Go值按fmt.Sprint或应用formatter；不把Go默认格式冒称完整Intl等价。
- Chart组合独审10项通过：scope/raw的两实例重排、嵌套配置、主题/字体/RTL/实际选区/Local释放原本正常；另两条条件标题/指示器移除导致editor失焦先红后以typed slot keys修绿。显式空selector此前不可表达，新增兼容HasNameKey/HasLabelKey后独立操作Tooltip名称/标题及Legend原生按钮，动作一次且配置正确。Inactive Tooltip/empty Legend原0尺寸Box仍占父Gap，两个实际布局子例从40≠30转为Hidden(true)后的30。所有独审探针已清理；自由绘图没有自动注入CSS后代样式，Draw需捕获已解析值，独立窗口需重新提供Config，完整平台/字体/AT/视觉及几何动画矩阵继续保留。

- 原生Chart组合示例：examples/chart-composition只用公开ChartContainer/config/content、MyGo Plot/Axes/Rect/RuleY/Anchor和原生控件，不调用display.Chart。6项正式集成验证真实mark比例/配色/柱上标签/阈值像素，指针和Home/End/Right/Escape检查，双thumb范围排除实际marks，Input校验/Enter/按钮重绘，三指示器/隐藏/主题，原IME候选跨范围/主题保留与选区剪贴板，500 RTL布局和idle。Root复核发现移出图表后Tooltip保留选择但AX值为空，纯hover→移出实际先红后由app-owned selected同步Value修绿；Escape也同步清空。
- 组合示例隔离/视觉：以实际SDK提交`c6517f701450e524f2e019ad5ec683b0934668f8`只加入示例两个Go文件，6项与vet、CLI明暗导图通过，SDK及示例哈希与workspace三方一致。1000×960明暗、500×1000 RTL/Dashed三图由独审逐张实际查看，Root抽查窄RTL和宽浅图；控件换行、图表物理时间轴、Tooltip/编辑器、阈值和柱顶label无压缩遮挡。临时导图器已删除，未计入正式6项。范围筛选/参考线/标签是应用组合，未冒称完整Recharts Brush/ReferenceLine/LabelList API。
- 组合示例原生构建：新增make charts-app，生成build/Shadcn Charts.app并携带既有第三方许可；macOS app与Linux/Windows amd64编译通过。macOS真实窗口Home/Right检查March237，实际Input改380并Enter后mark、Tooltip、AX均到380；切Dashed和深色，原生range thumb缩至April–June时真实移除March、清空检查并禁用数据编辑。原生图形与图例同时换色，笔记保留。该流程不替代真正系统IME/AT或Windows/Linux窗口验收。

- Chart实际绘制数据：独审先复现未堆叠Pie/Donut/Radial只绘制首系列，却把第二系列加入默认Tooltip/AX；Radial的domain/scaling也会受该未绘制巨大值影响。现在在Chart入口派生first-only数据，统一交给绘制、域、命中和检查，不改调用方Series或Stacked/Radar行为。10项正式检查从基线5红/5绿到全绿，真实默认提示、原生AX、全图像素、额外MaxFloat64对首系列弧形/命中的影响，以及四种多系列极坐标图形确实保留后续系列均验证；原141项Chart聚焦和vet也通过。
- Chart旧空内容布局：Tooltip/Legend空内容原0尺寸Box仍把父Gap算两次，两个实际布局子例先红；改Hidden(true)后，3项正式检查通过，包含空→非空→空布局和相邻原生编辑器原选区保留。自由组合入口已在5.BM独立修复，此处补齐旧的无Config兼容入口。以`c6517f7`仅加入本阶段三个Go文件，display529/Gallery154共683项隔离检查与vet通过，三个文件与workspace哈希一致；不纳入并行示例候选。
- 本轮收尾：最终`go test ./... -count=1`共1617项、17包（11个有测试、6个无测试），全量`go vet ./...`与diff检查通过，最终Gallery macOS app及Linux/Windows amd64重新编译通过。完整平台、读屏/实际系统IME、字体/主题/视觉矩阵、其余组件复杂变体及几何/路径动画仍须按原目标继续实施，未把本次绿色检查视为最终交付完成。

- 消费者验证器缓存清理：实际分发复验发现原defer os.RemoveAll忽略错误，Go只读模块目录残留。新增真实只读嵌套目录回归先红，再恢复本次独占缓存的owner权限并删除；符号链接外目标内容/权限不变、重复清理也成功，共2项与cmd vet通过。清理错误通过named return与原错误合并，CLI会失败而不静默报完成。改后以代码制品`4f370c1`执行完整八步空缓存验证，`build/consumer-verification/run-3088365774`退出0；Root核对go-env中的模块/构建缓存共同父目录已自动消失，cleanup-verification.json记录准确工具源码SHA。该修正未包含并行图表动画或Gallery候选。

- Chart组合外部消费者：SDK提交`4f370c1e01d68ca72b0eb016b65700e571469cb1`、版本`v0.0.0-20261109125141-4f370c1e01d6`的ZIP SHA-256为`614923e93cd5f662c04d95f119ad63437a43e476ca6e01bd208a6f1d3b33a40f`。标准消费者八步通过，再从ZIP复制公开Chart示例至独立example.invalid/chart-composition-consumer，用另一套全新缓存、GOWORK关闭、无replace的正常require构建，六条实际命令及明暗运行均通过。模块图确认SDK来自新modulecache；源码SHA-256为`d1f7ffd55c70bede08fa60203caa03341edc300dcd25269f22bd715591b55946`，明暗PNG与先前隔离示例完全相同。独审实际查看恢复消费者/Gallery明暗/Chart明暗五图，Root核验14命令、模块图、ZIP/source/PNG并查看恢复消费者及外部Chart深色图。证据`build/consumer-verification/run-772084584/distribution-phase-receipt.json`；旧标准cache需手工清理的历史已如实记录，5.BP随后修正并实跑验证自动清理。本地制品验证不冒称公共发布、真实系统偏好恢复或跨平台GUI完成。

- Chart几何动画SDK：Bar、Line、Area、Pie、Donut、Radar、Radial新增145项正式检查（30顶层、115子例）。默认时序和CSS ease取固定Recharts 3.8.0：Bar400ms，其余1500ms，Pie/Donut延迟400ms。实际矩形、曲线坐标/长度、Area裁剪、扇区/环轨角度及Radar多边形随原生时钟变化；不是整体透明度或根节点缩放。覆盖正负反转中间零尺寸、分组/正负堆叠、比例域、缺失值、连续快速更新、参数更换、局部主题与真实检查、隐藏祖先/窗口/离屏、Instant/Disabled/ReduceMotion立即收敛及最终idle。时长/延迟可显式设置，Ease仅在当前build调用，不保留Context或应用回调。
- Chart动画隔离与独审：基于`af6d6e9`仅加入6个Chart源/测试文件，`GOWORK=off go test ./pkg/display ./internal/gallery -count=1 -json`得到display674/Gallery154共828项通过，双包vet通过，未混入并行Gallery新演示。独审另以同一提交加2产品文件运行39项：17项独立长度数学/public Trim工具、3项public Dash对子路径重置、19项真实绘制/输入。全missing或singleton后第一条连通Line、Pie/Donut/Radial部分正值归零连续缩小、缺失段并行显现共7个失败场景先红后绿。当前可见Pie/Donut扇区决定指针分类，原值始终来自最新数据；正在缩小的零目标扇区仍可指针检查，键盘选择目标正值且不会被静止指针抢回。收据`/private/tmp/chart-motion-final-audit.zl6w1O/audit-receipt.json`，产品SHA与workspace相符，临时Go探针已清理。
- Line来源校正：SVG的dash模式在每个子路径重置，正式与独立已知长度像素检查验证各缺失段并行收到相同绝对显现长度，不能用全路径顺序Trim代替。固定Recharts组件的实际React生命周期数值oracle另证实旧短线变成长线时可能保留旧dash周期而重复出缺口；原生使用当前总长度和最长已显现绝对长度，明确作为修正而非逐帧照搬该故障。33份参考文件及metadata逐一SHA-256验证，11份TS/TSX匹配官方Git tree blob，22份npm成员匹配与lockfile SHA-512一致的官方tarball，MIT许可随macOS包携带。
- Chart动画视觉：`artifacts/chart-data-animation-light.png`与dark图均由Owner及Root实际查看，七家族每行展示0/100/250ms三帧，明确使用受控200ms linear/Begin0，未冒称默认时序截图。现有7个静态像素/命中测试只显式settle或关闭入口动画以继续验证其原契约，断言未删除，默认动画由新回归独立覆盖。完整三平台实窗、系统IME/读屏、字体/视觉矩阵及其余Recharts高级配置仍待后续验收。

- Gallery图表动画：基于实际SDK提交`6ca42ad2bbe3b87693becc7a8c77f9d1bc9768af`导出，仅加入chart_preview.go与chart_motion_preview_test.go；新增13项聚焦、全Gallery167项与vet通过。默认开启的Animate chart changes作用于两卡，story卡有独立更新/恢复数据；覆盖七家族原生中间/终态像素、最新原值、关闭动画/减少动态效果、跨页与500RTL。两个UI文件SHA-256分别为`588e3e4c61e09defd9678af995f955f25ec85fb604df1896cf7eb6f1389e0878`与`c6f82e7881bd073cc28ea07229b9ff69109cd16af1085c814e43f614998ed008`。收据`/private/tmp/gallery-chart-motion-final-f2_x1nnc/gallery-chart-motion-receipt.json`记录隔离源码、命令和PNG校验值。
- Gallery动画视觉：620RTL/1100LTR×明暗四张最终图，以及六家族Before/Middle/Finished的真实时钟图均由Owner逐张查看，Root复核窄深/宽浅及完整中间帧图。使用真正Update action及NewTester.Advance，Line/Area/Radar/Radial中间450ms，Pie/Donut900ms含默认400ms延迟，无时序覆盖或像素缩放。自然换行、完整图表高度、独立数据、Tooltip原值及Legend可读；两个诊断导出函数已从隔离目录删除，未计入正式167项。
- Gallery动画构建与实窗：当前SDK与两个UI候选的全量`go test ./... -count=1`共1777项、17包（12个有测试、5个无测试），全量vet通过；macOS app及Linux/Windows amd64构建通过，Recharts随包许可与归档逐字一致。重新启动macOS app，真实Update story counts后Line Home/Right得到February Desktop145/Mobile110；Donut检查Safari310，关闭Animate后Restore回Safari200。切浅色、重新开Animate与Reduce motion并更新后Safari又到310，最后关闭Reduce motion，实际Donut及原生提示保持一致。该操作验证真实控件/数据，不代替慢动作实窗捕获或Windows/Linux运行。复核另发现已有Line经清空后重新加载的长度衔接问题，作为独立后续修正，不混入本阶段隔离结论。

- Line缺失/清空历史：先用真实原生像素复现已有短线→全部missing→新长线直接露出终点，以及已有路径→分类清空→新点跳过源的进入位置。最小修复只改变Chart动画状态：已有drawable历史的空点集采用(2倍绘图区宽度,垂直中点)种子，Line目标改变时即使坐标已相同也继续绝对长度时间线；缺失Y保留数值X、Y直接取新目标的既有正确分支不改。新增3项正式回归，前两条先红后绿，第三条验证2/3/5点的真实坐标继承；动画共148项（33顶层、115子例）。
- Line历史交叉证据：固定官方CurveWithAnimation及帮助函数在真实React19.2.3 state/effect中运行，数值线性SVG度量确认已有路径后的dense-null和empty-array两种历史，以及初次empty/singleton/dense-null三种对照；收据`/private/tmp/recharts-3.8.0-reference.Jzhaz6/oracle/line-history-receipt.json`。独审以实际SDK基线加冻结动画产品运行10项原生端点/进入/首次路径检查通过，同一探针在旧版5绿/5红；收据为`/private/tmp/chart-line-history-audit.ncsUoX/audit-receipt.json`，临时Go探针已清理。原有旧测量周期故障仍采用已记录的原生当前长度修正，不将数值oracle冒称浏览器视觉验证。
- Line历史最终验证：`6ca42ad`加本阶段两个Go文件的隔离display677/Gallery154共831项及双包vet通过。Root对最终workspace重跑全量1780项、17包（12个有测试、5个无测试），全量vet、diff与macOS/Linux/Windows构建均通过；前后源码SHA不变，动画实现SHA-256为`2db14b2b2ed324c90e2786a3b349b5767d9e516742e4878cd9494014d92981b0`，测试为`1c7c22bc52f4dd7bbcd2be6de569c1a07e1b0dc79c26e7d17a985889bbf20d09`。关闭旧窗口后启动重新构建的macOS app，Data页真实Update story counts，Line Home/Right再次得到February Desktop145/Mobile110；新历史路径由正式Tester与独审覆盖，Gallery没有添加只用于测试的清空按钮。正常七家族截图未被伪装为新历史路径证据，完整平台/AT/字体/视觉与高级配置继续保留。

- Gallery静态导出：提交`2c544f6`的标准模块八步和独立动画消费者11项/七命令均通过，但Owner及Root实际查看标准Gallery明暗PNG时发现Activity只有轴和网格，柱停在入口第0帧；该组标准图片明确作为诊断，不计视觉验收。原因是CLI直接调用ui.Render，未推进时间，也未设置Instant；SDK本身已正确尊重c.Instant。现在仅在`-snapshot`构建前调用公开c.DisableAnimations，再执行原View，不改应用Reduce motion偏好或窗口动画。
- 静态导出回归与验证：新增1顶层/2明暗子例共3项，先在原ui.Render路径复现真实柱中心像素缺失，再验证六个已完成原生mark的像素与独立正常时钟Advance后参考一致，且偏好未变。实际CLI明暗PNG均由Root逐张查看，六柱完整、布局正常；`go test ./... -count=1`共1783项、17包（13个有测试、4个无测试），全量vet、diff与macOS/Linux/Windows构建通过。动画SDK文件及既有Gallery演示文件未改，后续标准模块ZIP验收仍需使用此次提交，不能沿用旧空柱图片。

- 最终动画分发：实际提交`edbbf871116923d2a31348d50813aaa655af1e3b`、版本`v0.0.0-20261109135517-edbbf8711169`的ZIP SHA-256为`1626954f38f42faedeee69f0f856b204b97253145cc905321ae63237a0e2b5c4`。标准八步与独立example.invalid/chart-animation-consumer七命令/11项公开检查均退出0。外部测试真正使用Duration400ms、Begin100ms及Linear，半程原生像素对照独立静态数据；Disabled/ReduceMotion立即完成、最新原值AX、dense missing后目标Y与未完整绝对prefix、真正Labels=[]后的半程离屏/后续进入/最终几何均验证。该工程main/test SHA-256为`e785533d60970428a4755702757c86fc4af740c6b597522fb2da6b9ef9e8f6a0`/`88587f9df0f89fe0e175257985c37b736a4e25c5edfb6262fa94f4b2027db80e`，不导入internal包或工作区替换。
- 分发最终证据：`build/consumer-verification/run-3352262293/animation-distribution-receipt.json`记录同一SDK ref的15条命令、完整模块图/环境、ZIP/源码/PNG哈希；模块与构建缓存每轮均为空，GOWORK关闭、无replace且SDK确来自该ZIP，两套临时cache均实际确认删除。Owner逐张实际查看恢复consumer、Gallery明暗及外部动画明暗五图，Root核验完整收据和哈希、ZIP内SDK/CLI及两个Gallery源文件与workspace一致，并查看最终Gallery深色及外部动画深色图。Gallery六柱齐全且Reduce motion仍off；外部图使用真正时钟Advance，没有缩放/伪动画，其哈希与旧SDK探针相同但已在新ref完整重跑。旧空柱诊断证据未覆盖，也未把全部命令通过当作全部视觉通过；公共发布、Windows/Linux实窗和完整平台/AT/字体/视觉矩阵仍未完成。

- Flyout菜单动画SDK：新增32项正式动画检查（13顶层/19子例）及19项几何/释放/policy检查。按既有Base入口核Vega100ms CSS ease、95%缩放、8DIP实际侧向进入、退出无位移目标，根与submenu按各自面板变换。完整路径保留关闭层和尺寸，同层替换重叠但不另占一列；退出立即移出AX/输入，快速重开保持当前pose。真实TextInput选区/IME、同path反向/同depth A↔B Local释放已验证；已有Input显式Focus、Handle.Focus与关闭期新增AutoFocus三路均未偷回焦点，因此没有为假设增加Interactive API。Menubar实际点击兄弟trigger的Focus-before-Clicked误关闭已先复现，再用按压入口快照修正；子窗口连续按键先于父重建也同步新root身份。
- 菜单来源与几何：固定Base MenubarContent组合DropdownMenuContent而继承closed classes，不用Radix直接入口的无exit规则覆盖既有Base目标。Base UI1.6.0的虚拟零尺寸箭头以anchor中心投影/clamp决定原点，普通默认padding5、Context root0，主轴原点可在gap外。可选GeometrySurface仅报告公共Window.ContentBounds与实际整型anchor的屏幕DIP数值，窗口失效、Windows尺度不一致、可能Wayland及不能证明side的重叠/对角关系保持unknown。19项覆盖负屏幕/小数外包anchor、真实child rect、parent移动、nil/无面积、listener重入/重复释放及8种Wayland信号策略；它们不是实际OS窗口或后端检测证明。42份来源/metadata/许可的SHA-256、9份shadcn Git tree blob与5份官方npm tarball的lockfile SHA-512及成员原文均由Root核验，应用打包加入Base UI/Radix/tw-animate-css许可。
- 菜单隔离与独审：以实际`2beb4bf`导出，仅覆盖7产品与3测试Go文件，desktop165/旧Gallery167共332项及双包vet通过；不包含并行Gallery新UI。`/private/tmp/desktop-menu-motion-final-46lpn8yr/verification-receipt.json`及candidate-hashes.json记录全部结果，Root逐一确认10文件SHA与workspace一致。独审fresh同基线仅加7冻结产品，18项真正native像素/AX/按键/指针与vet全绿，收据`/private/tmp/desktop-menu-final-audit.XF2VKm/audit-receipt.json`。50ms独立红marker验证CSS ease，running hidden/Reduced立即完成且idle，Base center/offset原点及actual side反向均验证。未知首布局→known或首次side翻转会更新进入变量而不重启时钟，两条早期真实像素红已修正；反转的一次假红源于Advance落在非刷新边界，按Frame同步采样纠正后已撤销，不冒称产品修复。
- 菜单阶段边界：现有两个静态相邻Bounds测试仅在原即时状态断言之后各Advance110ms，再保留原布局/方向断言，没有删弱键盘/焦点/选项断言。NativeMenu系统行为与通用Popover原动画保留；continuously retarget是原生交互适配，不宣称任意中断CSS帧逐像素等价。临时原生与几何探针已清理。真实window事件、碰撞、缩放/混DPI、Windows/Linux/Wayland、读屏及完整视觉矩阵仍需后续验收。

- Gallery菜单演示：以实际SDK `51ce6e7`仅加入四个Gallery文件，新增21项（7顶层/14子例）、Gallery188项及vet通过；三类菜单共享Export子菜单，选择只回写一次并显示可读动作名。检查覆盖真实默认100ms时钟、退出期AX/输入排除、submenu反转、reduced、500宽RTL右键菜单、键盘回焦与离页释放。收据`/private/tmp/gallery-menu-motion-final-vgeiwmsw/gallery-menu-motion-receipt.json`，Root核对全部14个SDK/测试/Gallery文件哈希。
- Gallery菜单视觉：实际导出并查看roots/submenus明暗四图，Root复看根菜单浅色与子菜单深色。入口/退出0、40、140ms，LTR/RTL关闭40ms与立即重开姿态一致、最终层消失。注入Backend使用600×400 Tester画布，只证明面板像素、AX、输入及给定GeometrySurface事实；不证明原生FitContent、系统碰撞或合成器阴影。全量1855项、17包（13有测试/4无测试）、全量vet、macOS app与Linux/Windows amd64编译通过，打包三份新增许可与源归档逐字节一致。
- macOS菜单实窗：最终构建的Project actions Flyout键盘进入Export→Preset JSON，右键Project workspace的Flyout进入Export→Go source，均观察到Last action/通知及触发器回焦；切System native后，真实系统右键菜单选择Preset JSON并正确回写。Menubar只实证File打开和Right后Edit取得焦点，未把该次操作记为完整Edit选择链通过。真实窗口移动/缩放、混DPI及所有平台/读屏/视觉矩阵继续保留。

- Select箭头修复：Flyout原用文本`⌄`，Gallery12DIP字号下实际着墨6×5、中心Y46.5，而44DIP触发器中心为42；大字号还会变成13×11、中心Y50.5。现在三个入口共享从既有control.Select原样迁移的SVG，固定16DIP布局/字形、禁止压缩、垂直居中及IgnoreInput/AX hidden；所有字体和LTR/RTL的实际着墨均为10×6、中心Y42。未新增Material Symbols资源或公共图标集，迁移来源和literal/helper SHA记录于indicator manifest。
- 箭头检查：新增32项（3顶层/29子例），包括字体/尺寸/长自定义标签/RTL的独立SVG参考像素、点击箭头打开菜单及选择/AX；基于`51ce6e7`的隔离desktop/control/旧Gallery共500项与vet通过。独审仅覆盖三个产品文件，原8绿5红、修复13项全绿，收据`/private/tmp/select-chevron-final-audit.avojMI/audit-receipt.json`。Root核验四个候选哈希，最终全量1887项、17包（13有测试/4无测试）、全量vet、macOS app与Linux/Windows amd64构建通过。
- 箭头视觉：Owner与Root实际查看浅/深三入口对照PNG；Root重启最终macOS bundle，查看Gallery左侧Flyout各选择框，再切System native查看相同位置，箭头大小和居中一致。系统菜单展开内容仍由OS绘制；该证据覆盖关闭触发器，不替代其它平台/字体/AT完整矩阵。

- 菜单/Select分发：实际`1d9601e`标准ZIP的八步验证和第二个外部模块五命令/五项公开检查通过，覆盖可选GeometrySurface/DismissInterceptor、实际left侧向进入像素、100ms完成、退出期AX排除/保留窗口与最终一次回调，以及Flyout/Native/窗口内三路Select的真实箭头像素。两套缓存初始为空且完成后删除，GOWORK关闭、无replace；Root核对十个ZIP产品文件与该提交、模块图/源码/图片哈希，并实际查看恢复消费者与Gallery明暗。证据`build/consumer-verification/run-1641041268/menu-distribution-receipt.json`，未纳入并行Combobox候选，也不代替OS放置或三平台GUI验收。

- Material Symbols：固定Google官方`10ecc14790612ca8fd164e3f4ffe9c16fc50d5b7`，4157名称×三风格×FILL0/1共24942份原始SVG，wght400/GRAD0/opsz24。全部Git blob/SHA、XML/尺寸验证，3个独立函数包无catalog/init；全量all包单独内置固定ZIP。8份资产重复离线生成逐字节一致，收据`/private/tmp/material-symbols-generation-receipt.json`。Root核对17候选SHA与独立工程一致，原Apache许可随示例bundle复制。
- 图标正式/独审：全部24942份SVG通过真实MyGo解析；11项正式覆盖原生颜色/16、24、32DIP与字体隔离、填充/圆角、真实Button点击、decorative/named AX、非有限尺寸、名称/variant错误、并发及按需/全量相同像素。fresh实际`526b066`仅覆盖17份图标/示例/打包文件，完整SDK1898项、25包（15有测试/10无测试）、全量vet及icons-app通过；未带入并行Combobox。另13项独立公开API检查通过，见`/private/tmp/material-icons-linker-audit-mskqbzvt/audit-artifacts/verification-receipt.json`。
- 按需链接实证：同Go1.27.3/macOS-arm64/trimpath/stripped参数的四个真实应用均通过SVG.Image渲染，像素/PNG哈希完全相同。baseline、空导入outlined、outlined.Home均4271650字节，仅后者保留Home的两份SVG，无其它官方完整字面量或全ZIP；all.Must应用15637314字节且包含完整ZIP。Root读取完整`linker-receipt.json`、命令/候选哈希及原始产物，未把source函数计数当作裁剪证明。
- 图标视觉/平台：Root查看示例明暗PNG，并在macOS新bundle实际点击Save观察Saved、切Filled/Dark、通过OS菜单选择Rounded，再搜索arrow_drop_down得到2项，查看正确填充/颜色/尺寸。初次键盘菜单操作未改变值，随后真实Rounded菜单项点击才作为通过证据。当前workspace全量1959项及vet、Gallery/图标示例的macOS与Linux/Windows amd64构建通过；此数含并行Combobox早期候选，不作其最终验收。图标静态配置完整名称已提供，其它variable-font轴及完整平台/AT矩阵不计已完成。

- Combobox行为：新增53项（27顶层/26子例），typed选值存在性/合法空值、Label/query、默认无高亮及显式auto、禁用跳过/循环/NoLoop、clear/check、group/separator/custom Filter、toggle/共框chips、RTL键删/重排、内部搜索及回焦/9路Tab、动态集合/readonly/Field/InputGroup/尺寸与TextScale2均验证。旧API保留兼容行为，原11项disabled/IME/Changed/稳定移除焦点真红已纳入正式回归。source100ms retained motion尚未实现，不把原motion.Popup preset计为对齐。
- Combobox来源：46归档文件包含16份固定shadcn原文、27份Base UI发布成员及metadata、manifest/README；Root核44份原文/metadata SHA、16个Git blob、两个复用CSS blob、官方完整tarball SHA-512/lockfile与27成员原文。共享check/x从已验证Lucide资产原样复制，16DIP固定字形/AX hidden/IgnoreInput，旧Select箭头与其它choice图形未改。
- Combobox隔离/独审：fresh实际`1d9601e`仅覆盖5产品/资产与1正式测试，control189/旧Gallery188共377项、三包vet通过，收据`/private/tmp/combobox-sdk-final-bf3905mb/verification-receipt.json`。另一fresh同基线仅5产品/资产、无owner测试/Gallery，77项独审及vet通过，见`/private/tmp/combobox-native-final-audit-4qj6utg9/audit-receipt.json`。Root核对6冻结SHA。早期Tab/AutoFocus/重复Bind、尺寸强制与动态标签红均修正；PopupInput的真实ButtonGroup first/last×RTL×明暗1DIP重叠、四角8×8像素与普通OutlineButton逐一匹配，重复label也已消除。运行时仅保留值/Handle，无跨帧Context/Element/应用callback。
- Combobox回车源契约：固定ComboboxInput.js390–405明确无active Enter关闭列表并保留表单提交；先用原型复现popup留开，再修正常Enter关闭/恢复已选Label或清query、保持值回调不变。正式包内用既有fieldTarget定位真实native编辑器并验证Submitted一次，独立公开检查不冒称可见内部Submitted；inline/PopupInput × 匹配/无结果/已有选择/多选/IME候选后正常Enter均通过。源动画、真正平台候选窗/读屏、独立越窗编辑式弹层和完整compound/视觉矩阵继续验收。

- Gallery Combobox：基于实际SDK `e454b8a`仅覆盖四个UI文件，17项（9顶层/8子例）、全Gallery205项与vet/diff通过。覆盖清除/明确键盘选择、国家描述过滤/disabled、内部编辑器/选择/Escape回焦、多选toggle/RTL删除、IME及集合重排保选区、只读/禁用/required错误消除、离页新输入焦点、500/800/1450整页滚动与三词Search匹配。最后两条无active Enter旧预期按固定源改正，原typing/disabled/选值前置断言保留，未为这些测试改产品。
- Gallery图像/最终检查：Owner逐张查看1100宽/500RTL、Groups/custom结果、内部搜索、Multi打开态十张明暗PNG并重复导出得到相同字节。Root核四UI与57个源/SDK快照SHA，复看wide-light、narrow-rtl-dark和popup-dark；五卡自然高度、选中check/clear、描述与表单状态可读。收据`/private/tmp/gallery-combobox-final-ie3gonpk/artifacts/gallery-combobox-receipt.json`。导图显式DisableAnimations，只证明静态样式；临时导出Go已删除。最终全量1968项、25包（15有测试/10无测试）及全量vet通过；macOS app和Linux/Windows amd64构建通过。
- macOS Gallery范围：实际观察Combobox搜索词匹配Forms & controls、通过tab进入页面、建议点击回写Gio，再用真实输入/Down/Return回写Fyne，Value、Selected文本与Framework编辑器焦点一致。本次自动化存在AX目标失效、焦点与旧bundle变更提示；搜索结果键盘激活、滚动、无active正常Enter和Office popup实窗链没有记为通过。首次批量输入/Return后仍见Fy列表，尚需分辨输入法、事件时序或旧进程，不将Tester检查泛化为该实窗链已验收。源100ms退出、popup搜索32DIP外观、平台/AT矩阵继续跟踪。

- 实窗Enter补充调查：fresh`344311e`的12项纯公开native检查全绿，覆盖真实Gio建议选择回焦、CmdA/菜单selectAll、Field有无、Fy普通Enter恢复Gio/关闭及DownEnter选Fyne；未改产品或运行时。只读macOS事件路径确认普通Return走KeyPressed，insertNewline只补发IME withheld原键。公开Tester所有输入自动settle，没有no-frame batch接口；Changed/Submitted同帧与旧open状态的时序只是推演，未判为SDK故障。收据`/private/tmp/combobox-enter-order-investigation-1wyiiq9a/investigation-receipt.json`，临时Go已移文本模板。
- 最终图标/Combo分发：实际`344311e`同ZIP标准八步、第二外部模块15命令与4项Combo/13项图标公开检查全0；两轮全新缓存、GOWORK关闭、无replace，527份ZIP文件匹配新模块缓存且缓存已删除。Root核对ref/伪版本/ZIP SHA、全部日志/外部源码/四份二进制及PNG哈希；按需Home与全量图标的真实像素相同，完整ZIP只在全量应用中，体积与私有DCE探针相同。新标准图片与本轮已查看对应图逐字节同；外部raw Home透明PNG已查看。完整收据`build/consumer-verification/run-1850267044/material-combobox-distribution-receipt.json`，不把这些无窗口证据计为公共发布或平台GUI完成。

- Gallery搜索状态：网址输入、内嵌Command与全局组件搜索原先共同绑定State.Query。两项真实native输入回归先在`bfe9324`精确复现全局搜索被改成native-project/settings，再将两种示例分别绑定调用方独立草稿。组件搜索仍保留原Query；增加清楚的原生编辑器名称。验证搜索进入Combobox后网址保留、网址继续编辑不影响全局词、内嵌命令在导航离开/返回后保留且反向不影响搜索。fresh基线仅覆盖两个产品文件和一份回归，聚焦2项、Gallery207项及vet全绿，三个SHA与workspace一致；收据`/private/tmp/gallery-query-final-bnnfgfp2/verification-receipt.json`。未提前混入Combobox动画候选。


- Combobox源动画：新增24项正式检查，固定Base/Vega的100ms CSS ease、95%淡缩与仅进入时的实际side 8DIP位移。默认5DIP箭头投影和实际gap决定原点；未知/对角/重叠/无效布局保持中心。首帧opacity0跳过Draw时也能从公开LayoutBounds取得已完成布局，首个可见刷新帧的真实质心已验证。单选/多选×外部/内部编辑器关闭时立即移除输入/AX，保留原过滤结果；外部多选立即清query，其它query在退出完成后清理/恢复。快速重开保留原生编辑器和选区；减少动态效果/隐藏立即收敛并停止请求帧。
- Combobox源/隔离复验：补入Base UI1.6.0 Positioner、useAnchorPositioning与mount-transition三份原文，完整tarball SHA1/SHA256/SHA512及固定lockfile完整性保持；现49归档文件含47份原文/metadata，30份npm成员逐字节核对，另复用6份已验证CSS/动画包文件。fresh实际`b857d62`仅叠4个产品/测试文件，control213、已提交Gallery207共420项及vet通过，原53项行为守卫保留，只把完成关闭后的query断言移至100ms完成点。收据`/private/tmp/combobox-motion-stage-rcj5wd4r/combobox-motion-receipt.json`；Root核四个SHA与源码。
- Combobox独审：fresh`bfe9324`只叠两产品文件，不含owner测试/Gallery；同一公开native探针的22叶项从baseline12红/10绿变为22绿。20/50/110ms曲线、top/bottom/RTL原点、退出inert/query、真实32DIP搜索组、Tab/Escape/外点/IME/原生选区、反转、局部主题和另一overlay Escape均验证。收据`/private/tmp/combobox-motion-independent-audit/candidate3-result.json`及manifest；Root核raw JSON/probe/产品SHA并阅读断言。两条早期假红分别为Advance未绘制准确Now及把30DIP editor当32DIP外组，已按刷新帧/真实绿色像素测量修正并重跑baseline，未迁就错误探针改产品。
- Combobox阶段收尾：Root fresh提交基线仅覆盖同四个产品/测试和五个来源文件，全量1994项、25包（15有测试/10无测试）、全量vet及macOS/Linux/Windows amd64 Gallery编译通过。收据`/private/tmp/combobox-motion-root-final-cn4_cb22/artifacts/verification-receipt.json`。Root用纯公开API真实时钟导出并查看inline浅色、popup明暗及top/RTL深色四张六帧图（entry0/20/50/100ms、exit50/100ms）；未禁用动画，也未把图片当作OS窗口测试。Content定位/宽度、Empty结构/compound槽位、独立越窗编辑式弹层、Cocoa候选窗/VoiceOver及完整平台视觉矩阵仍需完成。


- Carousel多项组合：新增Container/Content/Item/Previous/Next和值快照，fraction/fixed basis、默认16DIP补偿间距、三种align/contain、snap与slide索引分离、loop空间不足回退、独立外侧Material chevron控件均实现。可见及部分可见项保留真实输入/AX，完全离屏排除；稳定keys保留原生编辑器/选区，非法/重复keys安全回退。旧carousel.go及原22项无gap分页守卫未改，兼容动画不宣称Embla物理。
- Carousel固定来源：15份shadcn原文逐一核固定Git tree/blob/SHA，四份Embla8.5.2完整tarball核lockfile SHA512/SHA256与12份原文成员；四tar均无LICENSE，Root重新取得固定gitHead官方MIT许可并核1066字节原文，macOS bundle的embla.txt与归档一致。归档保留完整原始lock及再生脚本；Root实际执行官方CJS函数得到1296组，与独立oracle数据相同，归档324条LTR数据逐项匹配。数据oracle是合成实测矩形下的官方几何函数，不冒充浏览器CSS/像素测试。
- Carousel隔离与独审：fresh实际`dd685e9`仅覆盖40个本模块文件，display703项、Carousel48项、vet及1296→324再生通过，见`/private/tmp/carousel-phase1-final2-o3eboq2d/verification-result.json`。独立fresh同基线仅两个产品、无owner测试/fixture/Gallery，最终195叶195绿，收据`/private/tmp/carousel-independent-audit/candidate-final2-result.json`。前175项覆盖真实像素/AX/可见编辑器/控制键/重配/loop；额外12项前置Snapshot及8项极小basis在旧candidate各先红再绿。负Duration被AnimateWith夹到1ms的问题改用公开Jump；前置快照采样统一到Container，极小outer按原生padding至少Spacing。没有把合法前置调用改成panic，也未删弱旧矩阵。
- Carousel阶段收尾：Root fresh提交基线仅覆盖冻结产品/正式测试、源归档及本阶段文档/许可打包，全量2020项、25包（15有测试/10无测试）、全量vet、macOS bundle与Linux/Windows amd64编译通过；manifest最后仅增加许可获取URL后，源完整性专项再次通过。四张实际native三帧图（half浅色、third深色、mixed-loop RTL浅色、vertical深色，rest/100ms/650ms）已逐张查看；使用公开Tester、真实600ms兼容时钟，未DisableAnimations。收据`/private/tmp/carousel-root-final-r4ltc8yr/artifacts/verification-receipt.json`。新组合的Gallery完整源例、长寿命controller/options/events、Embla物理/手势/速度投影、Autoplay/plugins及真实平台/AT/视觉矩阵仍需后续实现与验收。


- Carousel/Combo动画分发：实际`4d39f7a`同ZIP标准八步、第二个真正外部模块八步及4叶公开native检查（含父项共5条通过结果）均通过。后者实际验证三方向Carousel前置组合/五slide四snap/选值边界/快照独立，以及真实时钟Combo CSS-ease进入和100ms保留退出的像素与AX。双轮空缓存、GOWORK关闭、无replace，574份ZIP文件匹配实际新缓存且两套缓存均已删除。Root核对完整16命令/日志/模块图/源码/ZIP/二进制/PNG哈希，标准四图与此前实际查看对应图逐字节一致，另查看外部Carousel明暗两图。收据`build/consumer-verification/run-1157201957/carousel-combobox-distribution-receipt.json`；未混入并行Combo行为候选，未替代source物理/controller/Autoplay或平台GUI/AT验收。


- Combobox Base行为：新增46项正式结果（36叶），同一测试fresh`dd685e9`基线31红/15绿，候选全部通过；control原213项保留，新总259项。修复外部单选空输入清值、默认trim/自定义raw过滤、native IME committed-filter/回调、manual无active循环点/auto/NoLoop、已选起点及实际reveal、pointer离开/移入padding重置、chip数字标点Shift/BMP回编辑器、closed单选750ms typeahead与combobox/dialog语义、过滤空组/flat分隔线归一化。固定Base closed typeahead可以匹配disabled label，展开键盘仍跳过disabled；该源差异没有擅自改成普通选择规则。motion产品SHA保持原值，24项时间/像素守卫只适配新Role定位，未降低断言。
- Combobox来源/独审：新增四份Base1.6.0导航/typeahead/filter原文，与完整官方tarball SHA256/SHA512及成员字节一致；现53归档文件含51份原文/metadata、34份npm成员，复用6份CSS/动画来源。独立fresh`dd685e9`只叠一份combobox.go、无owner测试/来源/Gallery，161叶从94红/67绿变为161绿；包含95个可打印原生输入（94 ASCII+1 BMP）、readonly/disabled、九类行为及原22项动画/native守卫。Root核源码、完整probe/raw JSON/来源SHA与冻结一致，收据`/private/tmp/combobox-behavior-independent-audit/candidate1-result.json`。native query显示preedit但过滤/用户事件仅提交，custom Filter继续raw，flat separator归一化不等于已实现source compound Collection。
- Combobox Gallery契约：先在冻结SDK+旧Gallery真实跑出201绿/6红（5叶），均为Office旧RoleButton定位；仅将同名trigger/editor按PopupDialog/PopupListBox区分，保留实际键盘选择/Value/query/Tab/Escape/离页焦点与三宽度滚动断言，另补关闭后无残留Dialog。没有为了更新AX而回退源语义或移除交互检查。正式隔离回执`/private/tmp/combobox-behavior-candidate-final-lem0iyvw/manifest.json`保留原红日志。
- Combobox行为收尾：Root fresh`17e01f3`只叠四个SDK/正式测试、Gallery必要测试适配和六份来源文件，全量2066项、25包（15有测试/10无测试）、全量vet及macOS bundle/Linux/Windows amd64编译通过。收据`/private/tmp/combobox-behavior-root-final-hwh8llqy/artifacts/verification-receipt.json`。Content/宽度/Empty结构、完整compound/controlled/always/grid能力、源全例和完整locale/Unicode/平台/AT仍保留；真实Cocoa候选窗/取消不因公开Tester Compose检查通过而计完成。


- Gallery Carousel：基于实际SDK`4073227`仅增加四份Gallery产品/测试文件，12项新检查、Gallery219项通过。单项/双项/三项/固定220DIP/640与1024断点、16→4间距、loop边界/短集合回退、真实备注编辑与按key重排/跨页保留均验证。正常时钟100ms实际可见0/1/2三项，250ms后收敛；垂直按钮+Down、RTL按钮+Left分别导航独立track，ReduceMotion即时完成。500×600原生Scroll到末尾操作RTL，再返回首项编辑可达；500/800/1280控件横向边界与三个搜索词真实进入Data页通过。
- Gallery图像/收尾：Root逐张查看1100宽明暗、500宽浅色与全局RTL深色四图，项目内editor、外侧控件、270DIP垂直viewport及窄屏自然高度均可读；导图使用真实时钟350ms收敛，未DisableAnimations。最后修正loop提示，避免内容仍需滚动但不足以loop时误称全部已显示；专项与最终全量重跑，四图再导出字节一致。最终隔离全量2078项、25包（15有测试/10无测试）、全量vet及macOS bundle/Linux/Windows amd64构建通过，收据`/private/tmp/gallery-carousel-stage-ojvuu8qq/artifacts/verification-receipt.json`。未混入并行C1/compound候选，没有为新Gallery重建既有SDK几何审计；controller/物理/Autoplay源例及真实平台/AT继续保留。
- Carousel控制器：应用持有的Controller只存值、Handle及公开弱Services，单owner/代际Local清理不保留Context/Element；导航与部分配置按FIFO处理，事件中间快照、显式监听器身份、跳转scroll→select及已有RAF尾settle依固定8.5.2源语义实现。分组/Auto/纯值aligner、有序原生窗口断点、Handle非连续/逆序子集原始矩形、inactive及watch开关/当前帧否决钩子均有正式与独立检查。resize投递与自上次reInit的0.5DIP累计基线分开，错误恢复及无效第二owner键盘输入也验证。
- Carousel控制器独审与收尾：固定官方CJS的Engine/EmblaCarousel/EventHandler生成器及完整JSON原文再生一致；正式源几何/计数器对照103组。独立fresh实际`4073227`仅六产品、无owner测试/fixture/Gallery，冻结同18项18绿，原始SHA`28c6811ce24bef185a9f6862116ed6443826ea3dcc36dcd52665f656e5e5d01e`；candidate1的原始13绿/5红及三处探针语义纠正、四处真实缺陷分类分别保留，未伪称修正后的旧版全量计数。Root fresh实际`c5f0017`仅本阶段文件，全量2087项（display712/Gallery219）、25包、vet及macOS bundle/Linux/Windows amd64构建通过，收据`/private/tmp/carousel-c1-root-final-6zoznns0/artifacts/verification-receipt.json`。未混入Combobox候选；源物理曲线/拖动、物理同名选项、Autoplay/plugins及真实平台/AT仍未计完成。
- Combobox组合模块：Root/Scope独立Input、Trigger、Content、List/Empty、Group/Collection、Item、Value、Chips及Clear与With共用运行时；应用typed值、Key/Equal/Label/FormValue、显式空query、可取消值/query/open/highlight、Always、六侧/RTL定位、offset和anchor宽度分支、原生Editor.OnSubmit/Field、Grid行列导航及调用方ScrollState虚拟窗均实现。当前帧槽位和回调不跨挂载保存；公开Scope过期拒绝、卸载清理及新回调验证。正式新增9项，control268项全绿，旧259项和24项动画断言保留，motion产品SHA保持`4b4a9a0b23ca90bd5470979ae8afa277eccd572da824f9b4995e587b1deb220f`。
- Combobox单构建收敛：手写集合在Root结束前同步完整registry、Empty/List和各Item选中/活动/PosInSet/SetSize，不重放应用builder或依靠下一帧暖机。独立原四usage中，普通构造后稳定AX/预声明Group.Items版本旧版四绿；真正手写Item、Tester就绪后首次挂载Open=true、关闭动画、公开Draw中AccessAll采样的同第3项旧版Empty/partial SetSize红，最终同probe四项全绿。owner原“首AX”补充断言叠回旧版仍绿，两种条件分别保留，未合并成虚假的全量先红后绿；这证明公开构建metadata一致性，不冒充真实AT首发布或平台IME。收据`/private/tmp/combobox-compound-independent-audit/final2-result.json`，原始SHA`300485ee3f94cd21d20c09d355ed11a208fe9dc3402c3cfb65c9547095d14cdf`。
- Combobox示例/来源/收尾：Gallery新增三个实际流程检查（共222项），Content配置、对象单/多选表单、权限/错误/隐藏移除、Dialog、Grid和100项原生滚动窗可操作；12独立+20集合例子按源SHA/定义行/测试位置逐一映射。补3份Base npm成员、通用InputGroup及3份create转换原文，58个归档条目、37个完整tar成员和6份复用引用核验；create默认menuColor移除透明marker，未把可选translucent的无动画/模糊能力冒充完成。Root fresh实际`069c355`仅本阶段25文件，全量2099项、25包、vet及macOS bundle/Linux/Windows amd64构建通过；逐张查看四张实际native明暗/500宽/RTL图，含展开搜索和错误/chips，真实150ms时钟、未DisableAnimations。收据`/private/tmp/combobox-compound-root-final-jkdbd514/artifacts/verification-receipt.json`；未混入并行Carousel物理候选，独立窗口编辑、完整locale/Unicode、真实平台IME/AT及可选菜单预设继续保留。
- 控制器/组合API制品：实际`1d50ec2`同ZIP标准八步、第二个真正外部模块八步及一个公开原生集成流程通过。对象型Root/Scope选Rust、Editor.OnSubmit/FormValues、分组controller事件、部分ReInit接ScrollNext的FIFO均从正常require使用，未导入internal。两轮初始空缓存、无workspace/replace；605份SDK原文/ZIP/实际新缓存逐字节一致，临时缓存退出后确已删除。标准四图与已查看的`run-1157201957`相同，新的外部明暗图逐张查看，Project03/04对应实际分组跳转。收据`build/consumer-verification/run-1326536727/controller-compound-distribution-receipt.json`；减少动态效果下的静态/输入验证不重复背书源运动、公共发布或平台GUI/AT。
- Chart需求核对：固定官方tree完整枚举的132份Chart命名原文共561821字节，Root复核Git blob/SHA、70个catalog ID和逐条证据行界。真实使用margin为27个catalog文件，LabelList为7个；natural/linear/step、渐变/标记和polar角度/半径/网格/center等是实际功能差距，包括Create本身所用。Brush/ReferenceLine/额外显式动画属性在132份原文均未出现，不能把底层Recharts所有泛化接口另算固定源必选。需求、源SHA与示例映射保存在`docs/chart-source-coverage.json`，明确是`1d50ec2`时的静态需求快照，没有运行或宣称Chart新像素/平台/AT通过；后续模块仍须实现和验收。
- Carousel源物理：控制器/Options路径替换250ms独立ease/Draggable，接通原始ScrollBody/Bounds/Target/Looper/DragTracker与60Hz双暖机累积器，Duration/DragThreshold/DragFree/SkipSnaps完成复制/部分更新/断点合并。Root原43纯核场景/48结果的三个产品原文哈希保持一致；实际官方Engine/DragHandler共同执行的12轨迹、985时点和每个事件中间状态与原生对照，代表时点色块像素和AX也匹配。数值比较显式拒绝NaN/Inf；零Duration、retarget、inflight jump、loop、snap/free/skip、单指touch及隐藏重置均覆盖。前置组合Snapshot原C1 Visible空列表问题改为采样上次声明和当前solver姿态，12条轨迹同步检查前后可见性。
- Carousel物理原生边界：公开Dragged只声明viewport捕获，非消费HandleInput取其完整越界move/up，保留子按钮tap且不叠加SDK平移动画。取消/失去输入返回当前选择，不提交甩动；触摸跨轴取消、RTL/纵向、单snap/空集合、减少动态效果均验证。交互子按钮赢得指针后，祖先无完整move/up流和取消其click能力；多触点数量/cancelable等元数据也缺失，明确未计完成。Root只读核对官方HEAD `5a2f9c65`的handler与pinned逐字节相同，未无效升级、改模块缓存或使用私有flags。
- Carousel物理收尾：Root fresh实际`3402c85`仅本模块23文件，全量2167结果（display732/kernel48）、26包、vet及macOS bundle/Linux/Windows amd64构建通过。额外公开API检查验证首次60Hz插值offset=-8.16及Duration0即时目标；四张实际原生逐帧图已查看，水平浅色、RTL深色、垂直浅色、DragFree深色各含rest/100ms/650ms/3s，正常时钟未关闭动画。收据`/private/tmp/carousel-c2-root-final-6yn7d_i6/artifacts/verification-receipt.json`。未混入menu/Chart候选；当前Gallery仍用兼容组合路径，Autoplay/插件/完整源例及平台gesture/AT继续完成。

- Create菜单预设：Root fresh实际`cd3f14c`仅叠28份冻结文件，全量2174结果、26包、vet及macOS bundle/Linux/Windows amd64构建通过；所有冻结SHA保持一致，未混入Chart/Carousel候选。收据`/private/tmp/menu-preset-root-final-cewclnop/artifacts/verification-receipt.json`。覆盖六种有效配置、Creator导入/导出/撤销/恢复、标准/自定义scope配对、所有styled菜单及Combo/Select、solid100ms与透明即时切换。Gaussian CDF、饱和度/tint和前景像素验证同窗真实shader，独立Flyout材质更新不重挂。
- 菜单原生观察：实际macOS短生命周期probe报告Painter.Accelerated=true、shader已draw、Flyout.Vibrancy=true及SystemBackdrop，收据`/private/tmp/menu-preset-native-probe/accelerated-result.json`；Root另外运行上述隔离bundle，Creator切到Inverted translucent/subtle，观察独立菜单和同窗Combobox透明面，再选择Fyne并确认Selected同步且菜单关闭。此次AX首次click后文本输入未改变，不作为搜索/键盘或IME通过证据；OS材质核、Win/Linux窗口、真实AT/IME仍未计完成。

- Chart Cartesian源能力：30个固定Area/Bar/Line配置共用公开样式接口，signed margin/轴预算/刻度间距、natural/linear/step/monotone、渐变透明度、普通/活动点与柱形、四角和LabelList均接通既有几何动画。官方D3四类factory的16输入与native Path命令对照；当前297项Chart/Donut叶检查和两条源例应用检查通过。构建30配置的单条inventory不冒充30条独立视觉验收。Root查看渐变Area、自定义Line dots及堆叠Bar三图，原生字体/图标/刻度/渐变差异保留。
- Chart透明度修正与收尾：Root独立公开API探针将半透明bar同色native Box对比，candidate1确实失败，final2同probe通过；修正Color.Alpha已有乘法语义上的二次透明度，既有渐变检查补实际stop×FillOpacity。fresh已提交Menu `d2ec705`仅叠71个冻结候选/包装说明文件，2182结果、27包、全量vet、Gallery和Cartesian macOS bundle、Linux/Windows amd64构建通过；两个bundle所带D3 ISC许可与归档字节相同。收据`/private/tmp/chart-cartesian-menu-final2-root-final-oeqe_ak2/artifacts/verification-receipt.json`，修正前red证据`/private/tmp/chart-cartesian-root-final-3mlpafkj/artifacts/root-alpha-red-observation.json`；无并行C3/Select候选，Polar31/Tooltip9/Gallery和真实平台AT未计完成。

- Carousel插件/Autoplay：typed应用插件与当前回调Actions、ordered lifecycle/重复名称及身份、FIFO外部命令、完整source timer和事件中间状态、pure delay/root resolver、媒体重建及原生hover/Tab/focus/hidden接通。纯core13场景/209时点，14结果；实际原始PluginsHandler/EventHandler/Autoplay执行器覆盖4生命周期及4嵌套订阅顺序。Controller/Options源物理不回退；旧12项Gallery结果保留，物理settle等待替代旧250ms兼容等待，新增实际API/配置/播放/离页重挂流程。
- Carousel C3收尾：Root fresh实际`9092002`只叠19份冻结模块文件与3份说明，全量2211结果、28包（display752/Gallery224/core14）、vet及macOS bundle/Linux/Windows amd64构建通过。公开原生Tester正常时钟0→2000ms→3500ms逐帧运行，实际Autoplay由01/Discover切到02/Create，播放保持有效；两张原生图逐张查看，未关闭动画。收据`/private/tmp/carousel-c3-root-final-igpqmxig/artifacts/verification-receipt.json`，附独立native visual原文与日志；未混入并行Polar/Select。八份固定源例有显式可操作配置映射；同slide任意editor间pointer focusout、交互子控件拖动/click取消、多触点及完整平台/AT仍保留SDK/验收边界。

- Autoplay零间隔修正：最小公开原生探针在9d21a84初次挂载未Advance时已选下一项，同probe定向修正后保持首项、Advance1ms才触发一次并被listener停止。源码setTimeout是异步，MyGo windowHost.afterTimer最短1ms；调度deadline据此归一化，TimeUntilNext继续保留source当前delay查询语义。fresh9d21a84仅两份候选文件，Carousel92结果和display vet通过，收据`/private/tmp/carousel-zero-delay-root-j2jokr5b/receipt.json`；未重跑无关全库/三平台。

- Tooltip富内容：TooltipWith/NativeTooltipWith保留独立child Context、明确语义描述与revision，native Kbd/分段/长文本、六side/RTL、真实几何翻转箭头和未知几何省略均有正式检查。原Tooltip/HoverCard30结果保留；新增6个SDK结果与1条真实八例应用流程（含disabled wrapper、实际link、状态按钮和描述更新）通过，未把例子列表当视觉验收。MyGo原生窗口不抢焦点/鼠标穿透；Native仍用平台主题，source10DIP圆角旋转箭头使用7DIP透明窗内空间，Gap量到外窗口。
- Tooltip收尾：Root fresh实际9bf0c55仅19冻结文件，全量2219结果、29包、desktop204、全量vet、Gallery与Tooltips macOS bundle及Linux/Windows Gallery编译通过；Material Symbols许可在两个bundle字节一致。两张原生明暗图逐张查看。Root实际启动此隔离Tooltips app，Tab聚焦Basic显示正确箭头，再Tab到Save changes，观察真实Kbd及方向箭头，焦点仍在父按钮；未计真实AT/所有侧向/跨display通过。收据`/private/tmp/tooltip-rich-root-final-xkhblxt6/artifacts/verification-receipt.json`。工具首次启动返回延迟很长，后续UI观察成功；未因此重复重启或虚报未见结果。无并行Select/Tabs/Polar候选。

- Tabs组合：两个产品提供当前帧Root/List/Trigger/Content、独立header/panel、native Controls/LabelledBy、icon-only可用名称与兄弟/嵌套scope，legacy tabs.go逐字不变。navigation96结果、example15结果（13实际配置+真实输入/菜单/草稿流程）及vet通过。Root额外公开paint中AccessAll采样，确认方向键选中Beta时实际绘制已关联Beta面板；原生invalidation在绘制前收敛，未基于代码猜测改动产品。
- Tabs收尾：Root fresh实际22f766e仅22冻结模块/包装说明文件，全量2240结果、30包、vet、Gallery/Tabs macOS bundle及Linux/Windows Gallery构建通过，Root逐张查看input浅色、vertical深色及双variant对照三图。收据`/private/tmp/tabs-composition-root-final-o8urh0hb/artifacts/verification-receipt.json`；公开paint补验`/private/tmp/tabs-first-paint-root-rh3zqcm8/root-probe-corrected.json`。无并行Select/Polar/typed-input候选；13配置构建与三图不冒充完整平台/字体/AT验收。

- Select组合：无editor单/多选、独立Root/Trigger/Value/Content/List/Group/Item/滚动箭头、100项真实reveal、selected文字对齐或六侧/RTL定位、可取消状态与富内容均共用当前帧核心。22源例映射到实际workbench/manual/dialog与OS流程，35原文归档；Control285结果、example3结果与vet通过。FormValues同build应用赋值曾实际漏读，保留同一提交断言并改读应用State后通过。
- Select消费兼容：两个旧Calendar/Table locator实际先失败；仅换ComboBox角色后Table通过、Calendar仍无法选择月份。根因是便捷Select用每帧临时绑定地址作为mount身份；改用原生位置/调用方key，保留显式Root稳定地址合同，原Month End/Enter、原日期不变、Table越行/焦点全部完整断言通过。已有MultiSelect流程也改为每帧重建derived绑定，连续两次选择与open保持验证。原始红、仅角色与最终绿日志在`/private/tmp/select-module-delivery/consumer-roles-{red,new-role,green}.json`；最终49收据`/private/tmp/select-module-delivery/final49-result.json`。
- Select收尾：Root fresh实际3e9c7c3仅55冻结候选/包装说明文件，全量2259结果、31包、vet、Gallery/Selects macOS bundle及Linux/Windows Gallery构建通过。最终隔离app重新导出的明暗打开态PNG与Root已逐张查看的两图字节一致；收据`/private/tmp/select-composition-root-final-57iujs5t/artifacts/verification-receipt.json`。不混入Polar/typed-input；原生completed-click、DOM drag-release差异及真实平台/触摸/AT/IME边界保留。

- Polar31：可选source profile提供degree角度、fraction+DIP半径、实际Sector圆角、活动层/空隙、circle/polygon grids、axes/native center/curved glyph labels，既有动画与原值检查共用。Pie按当前ring/item，Radar/Radial按axis区域（含空track）命中；Source与legacy动画key分离。304 Chart/Donut叶与两应用各2检查通过；31配置构建不计31独立视觉。官方完整CJS直接执行4Pie/5Radial/5Sector SVG/1Radar与2轴命中，Root再生输出字节一致。
- Polar收尾：Root只在原7正式检查的数值比较器补NaN/Inf拒绝，不改7产品、不增矩阵。fresh实际c371168仅92冻结候选/包装说明文件，全量2268结果、32包、display760、全量vet、Gallery/Polar macOS bundle及Linux/Windows Gallery构建通过。四张最终隔离app图已逐张查看，含活动外环/center、沿弧标签、填充圆网格及两行自定义Radar tick；新Select的trigger呈现改变后已从最终app重新导出，未沿用旧图冒充。收据`/private/tmp/chart-polar-root-final-rtc2spwx/artifacts/verification-receipt.json`及官方oracle再生日志。
- Polar并行消费失败归因：原共享display两处Calendar/Table角色定位失败经Select阶段进一步检查，发现Calendar还有真实临时绑定mount问题；均在c371168修复并完整保留断言，Polar本次合并全量已绿。字体/图标与1设备像素边缘栅格差异、完整平台/AT及Tooltip9/Gallery目录继续保留，未将先前修正的探针currentColor/AA/1LSB期望计作产品红绿。

- Gallery组合整合：现有格式模型同时连接ToggleGroup和真正无编辑器MultiSelect，独立Tabs header动作/面板关联与跨页状态保留，Tooltip包含可读说明、平台快捷键及styled Kbd。新增两条实际原生流程覆盖双向选值、键盘、窄布局和离页返回；旧popup检查改为真正hover后核富内容，再使用实际全局搜索快捷键。Tooltip完成click会关闭，先前测试入口修正不计产品缺陷。
- Gallery组合收尾：Root fresh实际15ad058仅叠8冻结文件，全量2270结果、32包、Gallery226结果、全量vet、macOS Gallery bundle及Linux/Windows编译通过；收据`/private/tmp/gallery-primitive-root-final-65egcs70/artifacts/verification-receipt.json`。Root查看实际多选打开态及360宽深色Tabs原生图，确认三项选值/计数和窄栏换行。私有静态表单画布高度600，底部裁切不作整页视觉证据；未混入typed-input、Calendar或Chart目录候选，也未扩展为真实平台/AT验收。

- 类型化输入：NumberInput使用真实TextInput编辑器，有限范围/步进、独立draft与提交、IME/Field/InputGroup均保留；Date/Time复用公开本地化日期/分段，File由用户激活后异步打开系统选择器，取消保留值、卸载/换绑丢弃旧结果、当前构建回调。Time元数据位于SDK公开group，未虚构内部段ReadOnly/AutoFocus。外部NaN模型曾因NaN!=NaN导致每帧Changed，原定向检查修正后通过。
- 类型化输入收尾：fresh实际6f6d947仅叠30冻结模块/包装说明文件，全量2288结果、33包（control295、example8）、全量vet、macOS Gallery/Inputs bundle及Linux/Windows Gallery编译通过。Root核对冻结文件和收据，逐张查看由exact最终Inputs可执行文件导出的三张原生PNG，含types明暗与Field/InputGroup/file组合；两bundle各14份许可字节相同。收据`/private/tmp/typed-input-stage-final-e1baevri/receipt.json`；未重复运行独立消费者或旧检查矩阵，未混入Calendar/Chart3。实际OS文件选择器、平台候选窗和AT仍待验收。

- Chart目录：九种Tooltip复用现有content/legend并接入最新原始行payload，支持series数据覆盖、DataKey/TooltipName、复制明确的无环JSON形状；新增按检查位置的10DIP间距/边缘翻转及显式类别反转，默认旧定位和物理顺序保留。目录70项、文档8项、Create6项共用模块内构建器，原两个应用的数据JSON逐字迁移，Gallery搜索/草稿与旧探索器保留。84配置的一条inventory不计84独立视觉验收。
- Chart目录位置修正：官方defaultIndex非类别坐标包含完整图表高度与axis offset，旧plot-only midpoint实际tip.Y=138.9336，源期望150；同一公开原生检查修正后通过。Root仅补原位置检查的有限数守卫，未改产品或扩展检查矩阵。Root核对19个固定Git tree blob/原文SHA及8个官方npm成员、完整tgz SHA-512，离线核验不依赖缺失promisor blob下载。
- Chart目录收尾：fresh实际3f668e7仅叠63冻结文件并删除两个已迁移JSON，全量2296结果、35包（display763、Gallery227、catalogue4）、全量vet、macOS Gallery/Catalogue bundle及Linux/Windows Gallery构建通过。收据`/private/tmp/chart3-root-final-vp1hh22_/artifacts/verification-receipt.json`。Root逐张查看advanced total、Arabic RTL、anatomy和完整六卡Create原生图；最终打包可执行文件重新导出四图与已查看图片逐字一致。未混入Calendar候选；字体/栅格差异、其它Go对象借用边界及真实平台/AT验收保留。

- Calendar/date：固定date-fns-jalali 4.1.0-0的33年算术Persian模型、Farsi/RTL/周六默认与历月/年导航共用真实time.Time/Gregorian/DST路径。DatePickerInput组合真实编辑器、应用Parse/Format与text/month/open/error，保留错误草稿、Field/Submit、IME提交和日历浏览/选值；示例相对日期小语法不冒充完整chrono-node。432转换/月长向量及3个faIR标签/周/月案例由固定官方npm执行并逐字再生；Root核对5个Git tree原文和两个完整官方tgz及其成员，共28原文。
- Calendar/date收尾：Root fresh实际157c3ca仅叠51冻结模块/包装说明文件，全量2304结果、36包（display769、desktop204、example2）、全量vet、macOS Gallery/Dates bundle与Linux/Windows Gallery构建通过；收据`/private/tmp/calendar-root-final-2_bacds7/artifacts/verification-receipt.json`。Root查看实际明暗Persian+编辑器popup图，最终Dates可执行文件导出与所查看字节相同；新增date-fns-jalali/react-day-picker两份MIT许可在两个bundle均逐字核对。真实Cocoa取消、平台窗口/AT与完整NLP不是该检查结论。

- 合并API外部消费：精确提交`1b993fa1e7fabda4cb6f02a7b4e948ab47733afd`、版本`v0.0.0-20261109222000-1b993fa1e7fa`，通过本地file proxy提供canonical模块ZIP/.mod/.info；独立项目只用正常require，GOWORK关闭、无replace，新版本实际进入正常GOMODCACHE。保留现有依赖/构建缓存，只下载本次kit版本；未重跑旧空缓存八步。唯一公开原生Report流程首次运行通过，无失败/重试：真实标题/数值步进/日期/时间段、多选ListBox、异步文件结果、独立Tabs、Cartesian/Polar原始行检查及菜单保存同一应用模型。
- 消费制品复核：Root核对全部970份Git tree/ZIP/实际modulecache文件、外部源码与图片SHA，ZIP SHA-256为`9829ca8b1bbe27c2485ab9d587b9a7f810a6760bdbe2b89fd23bd51f6395bf70`。实际保存Q4 review、金额126、November12/16:30、desktop/mobile与一个附件；键盘检查时分类索引1和原始row ID先被断言，离开图表点击菜单后的最终索引-1未冒充仍选中。Root逐张查看三张真实native表单/图形/反色菜单图，按需Material无all包。收据`/private/tmp/kit-external-final-wupgmk3e/receipt.json`与`root-review.json`；本地模块协议通过不计公共发布，注入picker/同窗菜单不计OS对话框、越窗tracking、IME或AT平台验收。

- Gallery小组合：七个固定源原文的blob/长度/SHA已核对并记录于`examples/gallery-recipes/source-map.json`。Progress标签/值与独立0–100 Slider、整卡Plus/Pro/Enterprise和两列尺寸radio、真实灰度/彩色Avatar、非交互busy Skeleton form/table及共享暂停接入实际页面；保留既有Plan/Volume/Avatar/loading模型。三条新原生流程验证整卡边缘选择/键盘与跨页、实际Progress/AX同值、图片切换与占位语义，不把构建列表或纯样式计成流程。
- Gallery小组合收尾：Root fresh实际c4dedf8仅叠9冻结模块/映射/说明文件，Gallery230与cmd/gallery3共233结果、两包vet及macOS Gallery bundle通过；收据`/private/tmp/gallery-recipes-root-final-agwcq1r9/artifacts/verification-receipt.json`。Root逐张查看四张冻结模块原生图，含真实灰/彩色肖像、两种占位、标签数值及整卡/grid。SDK未改，未重复全库SDK/三平台/外部消费者矩阵，也未混入后续Booking/InputGroup候选。灰度通过应用图像预处理提供，不宣称Box滤镜会影响后代或额外Avatar API。

- Gallery Forms组合：预约模型连接Number/原生Date/Time/File和可编辑DatePickerInput、Gregorian/Persian切换、错误草稿恢复及独立保存快照；文件窗口使用当前Gallery原生Window作为parent，只在激活后打开。InputGroup把实际clipboard拒绝/成功、收藏、local search、独立Flyout/OS菜单、真实textarea/UTF16选区与caret、copy/reset/文本preview接通。模型由Gallery跨页持有，三个新增单模块流程和一个实际500宽RTL全页面编辑/全局搜索/返回流程构成固定检查范围；三个InputGroup源原文映射加入既有source-map。
- Forms修正证据：Root发现无Host时Info在未激活前就报错；原missing-host case增加初始quiet断言后真红，改为只在激活时报错后同case绿。合并全量另发现长关键词串联让`cntfdbck`误选Forms，Gallery改为分别评分页面名/单个组件名，不改SDK cmdk算法；原Content路由断言及新DatePickerInput路由流程均保留并通过。
- Forms收尾：fresh实际1d77483叠13冻结文件，初次全量只出现上述Gallery搜索回归；修正后完整复验受影响Gallery234/cmd3共237结果，其余不变包保留本轮全量通过证据，去重覆盖2311结果、36包。全库vet、macOS Gallery bundle与Linux/Windows Gallery编译通过；收据`/private/tmp/gallery-forms-root-final-3wwithgc/artifacts/verification-receipt-final.json`明确记录初始失败与最终复验来源，未将两轮结果简单相加。Root逐张查看Booking/InputGroup四张冻结模型明暗图，另导出并查看实际State.View的两个500宽RTL滚动页面；可操作区域无横向溢出。未重复外部消费者，未把注入picker、合成输入或headless场景计为真实OS/IME/AT验收。

- macOS实际文件选择：使用6822d90已验证Gallery可执行文件（SHA-256 `0469f09f11d38d6b937b102044ff4a08ad5ceac1874016cfb4c06ced3d639a61`）的独立标识bundle，真实键盘将预约表单滚动到附件按钮，Space打开标题为“Attach booking documents”的系统open-panel sheet。点击Cancel后对话框消失、附件按钮恢复焦点，仍为No file selected与0附件。该检查证明初始空值取消及回焦，不代表非空值取消、所有系统或读屏验收；未选择或读取文件。记录见`/private/tmp/gallery-macos-acceptance-8i6mb051/provenance.json`。自动化首次连接长时间阻塞未计作应用启动性能。
- Field错误：固定Field原文与文档的blob/长度/SHA和既有Vega样式已核对。Error与Errors精确去重、空值省略，单条文字/多条原生alert-list及独立FieldError共用实现；ErrorContent使用当前构建回调与可访问摘要。两条新增原生流程覆盖真实InputGroup编辑、错误更新/清除后选择与草稿保留、旧Error兼容、自定义动作更新及独立空槽位；验证规则/时机仍由示例应用持有。
- Field收尾：fresh实际6822d90仅叠10冻结模块/示例/来源文件，全量2313通过结果、37包、全量vet、macOS Gallery bundle/Field示例及Linux/Windows Gallery构建通过；收据`/private/tmp/field-errors-root-final-a4orn395/artifacts/verification-receipt.json`。独立只读审查未发现必须修复问题，确认没有新增跨帧Context/Element或回调存储。Root逐张查看exact最终示例可执行文件生成的浅色列表与深色自定义正文PNG；没有源码产品失败，不计红绿修复。未混入字体或Gallery Payments/Questionnaire候选，未重复消费者矩阵；真实AT播报与平台IME仍待验收。

- 字体角色：固定12份Create/组件原文中七类cn-font-heading标记接入原生Body/Heading角色、FontScope与Controller，空/inherit标题保留祖先字体，显式system与正文独立。相同公开native probe在旧基线的monospace父宽296.71875/child274.93405、切字体invalidates0；候选父/child同宽，切Georgia后同为289.27734且invalidates1。该probe使用Backend创建独立SDK Context，不计真实OS窗口。独立system-body像素参考宽66、mono174，确认system不是继承的同一伪oracle。
- 字体标题审查：只读独审发现EmptyWith自定义Title遗漏wrapper角色；原生同case标题宽175、独立Georgia参考101先失败，给既有wrapper应用StyleHeading后通过。33项早期聚焦结果不冒充最后修改后的全量；最终Empty聚焦10结果/9叶、vet及原公开Flyout probe均通过。Root逐张查看`/private/tmp/font-role-phase1/font-roles-native.png`，无跨帧Context/Element或应用回调新增持有。
- 字体角色收尾：fresh实际97e6e16仅叠25冻结模块/测试/来源和4份使用说明文件，全量2318通过结果、37包、全量vet、macOS Gallery bundle及Linux/Windows Gallery构建通过；收据`/private/tmp/font-role-root-final-m0j386l0/artifacts/verification-receipt.json`。源和候选最终哈希一致，未混入Payments/Questionnaire。此阶段不包含26种字体资产、字体许可证或Gallery目录，原生平台字体/读屏/全视觉验收继续保留，未重复外部消费者矩阵。

- Payments组合：固定五行和默认十行页保留，2/5页大小是明确的演示选择；既有DataRows/TableWith与app模型提供邮件过滤/排序、保护选择/动作列的显隐、当前页三态选择/跨页key及filtered计数。复制使用真实异步clipboard completion，客户/明细仅显示本地数据。三份原文的Git tree blob/长度/SHA核对并写入examples/payments/source-map.json；未重写SDK表格。
- Payments收尾：独立只读审查未发现产品必修问题，提交前仅格式化原测试。Root接入真实State字段/Data页及Payment/Payments搜索词，保留Recent invoices与旧Table composition；新增单条500宽RTL实际State.View搜索、滚动到真实filter、编辑及离页返回流程首次通过。fresh实际5435ac8只叠6冻结产品/测试/说明文件，Gallery236与cmd3共239通过结果、两包vet及macOS bundle通过；收据`/private/tmp/gallery-payments-root-final-gyg_6ka_/artifacts/verification-receipt.json`。Root逐张查看冻结模型的浅色与深色RTL原生图；完整State流程是交互/边界证据，不冒充该两张图来自完整页面。未重复全库/外部消费者矩阵；真实系统tracking/读屏/跨平台视觉继续保留。

- 问卷recipes：固定conditional/resume/validation三份原文的Git tree blob/长度/SHA已重新核对，复用现有QuestionnaireWith/Disabled/DefaultAnswer/DefaultItem/ValidateAll。条件环境随local/cloud加入/排除导航及成功提交，保存值/第二步checkpoint可编辑与Reset，Public+Concise拒绝后回到早先Detail并聚焦；修正后提交当前值。三份应用草稿独立，原Workspace复查/编辑流程保留；原集中native流程贯穿三个配置。独立只读审查未发现必修问题，Root逐张查看三张原生状态图。
- 问卷整合：真实State持有recipe模型，追加到原问卷区域，新增Migration/Conditional搜索别名。单条500宽RTL实际State.View流程通过原生配置控件恢复第二步、修改checkbox、离页并搜索返回，当前值/检查点保留且原Workspace答案不变。首轮唯一失败来自Root测试假设原Workspace Answers应是空map；诊断证实SDK预先建立三条空答案，页面/配置/检查点均正确。断言改为和进入时实际内容比较，未改产品或削弱独立性要求，不计产品红绿修复。
- 问卷收尾：fresh实际e6ab7be仅叠7冻结模块/测试/映射/说明文件，最终Gallery238与cmd3共241结果、两包vet及macOS bundle通过；收据`/private/tmp/gallery-questionnaire-root-final-hzfptp1n/artifacts/verification-receipt.json`，初轮记录保留在`/private/tmp/gallery-questionnaire-root-final-1qc7a4q9/artifacts/test.log`。未混入命名字体资产，未重复SDK全库或外部消费者矩阵；恢复示例由应用默认值hydration，不冒充磁盘持久化服务，真实AT/平台视觉继续保留。
