# 实施阶段与验收

目标：将 shadcn/ui 的完整组件、颜色、主题切换与动画移植为 MyGo 原生 Go UI Kit。原始范围始终保留；导出函数存在、编译通过、局部测试通过均不能单独证明完整复刻。

每个阶段按模块实施，完成对应验收后独立 commit。后续发现的问题在所属模块继续修复并提交，不能用阶段验收替代最终验收。

| 阶段 | 模块 | 阶段产物与验收 | 状态 |
| --- | --- | --- | --- |
| 1 | 工程、`theme`、`motion` | 固定依赖；官方颜色/主题数据可离线再生成；OKLCH、主题模式/作用域、偏好存储、减少动画测试 | 已验收，阶段 1 提交 |
| 2 | `control` | 表单与动作组件第一版；真实输入、禁用状态、键盘选择、无障碍状态、OTP 测试 | 已验收，阶段 2 提交 |
| 3 | `display`、`overlay`、`navigation` | 展示、数据与浮层导航第一版；日期提交、排序分页、异常数据、焦点/关闭/键盘测试 | 已验收，阶段 3 提交 |
| 4 | `conversation`、原生 Gallery | 消息与问卷第一版；原生窗口、浅深色截图、Gallery 联动与跨平台编译 | 已验收，阶段 4 提交 |
| 5 | 完整行为与变体对齐 | 对照 66 项上游组件逐一补齐变体、复杂行为、可访问性关联及动画；每一模块独立提交 | 进行中 |
| 6 | 最终验收与发布准备 | 全量组件示例、视觉对照、平台运行验证、文档和完整范围审计 | 待实施 |

各阶段实际执行的验证见 [verification.md](verification.md)。阶段 5 的已知缺口记录在 [parity.md](parity.md)。它们是目标的未完成工作，不是被排除的范围。

## 阶段 5 模块提交

| 子阶段 | 验收范围 | 状态 |
| --- | --- | --- |
| 5.0 | 固定上游源码提交、66 项对照路径和样式校验值 | 已提交 `62155d1` |
| 5.A | 按钮/分组、Field/InputGroup、分组选项、多选 Combobox、原生 NativeSelect | 已提交 `4e7459e`，15 项控件测试和 vet 通过 |
| 5.B | 单/多/范围日期、限制、月份导航、时间与预设 | 已提交 `4be2188`，独立审计修复后 24 项日期检查和 vet 通过 |
| 5.C | 图表家族与交互、数据表选择/过滤/原生菜单 | 已提交 `8c6c6cf`，34 项针对性检查和 vet 通过 |
| 5.D.1 | Menu/NativeMenu 等核心成对组件、跨窗口 Flyout | 已提交 `1859ae3`；30 项 desktop 检查及 macOS 小窗口越窗验证通过 |
| 5.D.2 | Create 风格 Gallery、真实配置/预览和预设 | 已提交 `07c6e2e`，27 项 Gallery 检查及截图通过 |
| 5.D.3a | Flyout 可逆退场、原生关闭拦截、减少动画和生命周期 | 已提交 `a451b0a`，23 项针对性检查及 vet 通过 |
| 5.D.3b | DatePicker/DateRangePicker 与 Native 日期组件对 | 已提交 `6b4cb63`；原生日期标题补充修复 `8f7d8df` |
| 5.D.3c | Tooltip/HoverCard 与 Native 组件对 | 已提交 `0abc803`，键盘交接修复 `524a63d`；macOS Down/Escape 已核验 |
| 5.D.3d | Gallery 日期/悬停组件对接入与共享内容刷新 | 已提交 `ea5111c`，36 项 Gallery 检查、独立审计和构建通过 |
| 5.D.4a | Flyout 原位尺寸更新与父方向作用域传播 | 已提交 `ac345e8`，8 项针对性检查与独立审计通过 |
| 5.D.4b | Menubar/NativeMenubar、跨窗口兄弟切换 | 已提交 `4bf8771`，22 项菜单栏/内容检查及 vet 通过 |
| 5.D.4c | 真实窗口焦点与应用活跃的统一查询 | 已提交 `2c9cc84`，2 项能力检查与组件交接回归通过 |
| 5.D.4d | HoverCard 指针关闭、交接确认及关闭/blur次序 | 已提交 `8979f6b`，34 项悬停相关检查及 vet 通过 |
| 5.E | Tabs 激活模式、关联、方向键、动态指示器 | 已提交 `d734b28`，9 项导航测试和 vet 通过 |
| 5.F | 不确定进度、非有限值、原生标题层级与分隔语义 | 已提交 `25d9ca2`，回归与模块检查通过 |
| 5.G | Slider 单值/范围/多拇指、竖向、步长间距和原生无障碍 | 已提交 `941199d`，23 项 control 检查、独立审计及 vet 通过 |
| 5.H | Carousel 稳定轨道、拖拽、循环、方向与中途切换 | 已提交 `fe0cf45`，93 项 display 检查、独立审计及 vet 通过 |
| 5.I | Accordion/Collapsible 关联、禁用导航、保留状态与反转 | 已提交 `f94cba3`，11 项针对性检查与独立审计通过；运行时显式 Focus 限制继续跟踪 |
| 5.J | 多面板 Resizable、约束、折叠、键盘与布局持久化 | 已提交 `b521448`，16 项针对性检查、118 项 display 检查和 vet 通过 |
| 5.K | 独立导航内容面板、真实几何动画、原生路由菜单 | 已提交 `752f3a8`，42 项 navigation 检查、独立审计及 macOS 实窗路径通过 |
| 5.L | Gallery 菜单栏/导航面板、范围滑块及嵌套布局接入 | 已提交 `c18cf85`，43 项 Gallery 检查、7项新增集成回归及构建通过 |
| 5.M | Avatar 加载/回退、三种尺寸、角标与稳定键头像组 | 已提交 `824ece7`，10 项针对性检查、独立审计及 vet 通过 |
| 5.N | Card 复合槽位、动作、大小与贴边媒体 | 已提交 `2b0f42a`，9 项针对性检查、独立审计及 vet 通过 |
| 5.O | Gallery 头像变体、加载/回退与复合Card接入 | 已提交 `3d3c729`，3 项集成检查、明暗截图检查及 vet 通过 |
| 5.P | Dialog/Alert/Sheet 原生焦点与滚动、Drawer 四边吸附及手势 | 已提交 `91fe521`，38 项针对性检查、独立审计及 vet 通过 |
| 5.Q | Sidebar 响应式布局、移动面板、图标折叠和外部触发器 | 已提交 `7b585fc`，23 项针对性检查、navigation全包64项、独立审计及 vet 通过 |
| 5.R | Command 虚拟列表、稳定选择、快捷键/IME与受限窗口布局 | 已提交 `3645d0d`，21 项针对性检查、overlay全包65项、独立审计及 vet 通过 |
| 5.S | Gallery 响应式侧栏、移动选择及编辑状态接入 | 已提交 `25ad18d`，Gallery全包47项、独立审计及构建通过 |
| 5.T | Command 固定上游模糊评分、Unicode与离线数值对照 | 已提交 `afbf25f`，28项Command检查、overlay全包72项、独立审计和vet通过 |
| 5.U | Item 变体/尺寸/媒体、原生主动作与独立操作区域 | 已提交 `bef64f8`，26项针对性检查、独立审计及vet通过 |
| 5.V | Alert/Empty 完整组合、媒体可访问性与受限动作布局 | 已提交 `304a980`，28项针对性检查、独立审计及vet通过 |
| 5.W | Badge 六种变体、原生动作、视觉槽位与真实下划线 | 已提交 `e402ee3`，21项针对性检查、独立审计及vet通过 |
| 5.X | Gallery 内容/反馈页、统一页面命令源与原生交互 | 已提交 `2015c2f`；Gallery全包51项、独立审计、截图与macOS实窗通过 |
| 5.Y | 可取消的颜色反馈动画基础 API | 已提交 `93bd44f`；3项motion基础检查及隔离验证通过，各模块接入单独提交 |
| 5.Z | Input/Textarea 原生编辑、Vega外观与Field错误传播 | 已提交 `181f277`；13项针对性、87项隔离检查与独立审计通过；拖动resize另见5.AE，平台矩阵继续跟踪 |
| 5.AA | Checkbox/Switch/Radio/Toggle 原生状态、组合与Vega样式 | 已提交 `1cf77eb`；39项针对性、126项隔离检查与独立审计通过 |
| 5.AB | Breadcrumb越窗菜单对与有界原生Pagination | 已提交 `d5b3719`；navigation/Gallery隔离141项及独立审计通过 |
| 5.AC | SDK反馈过渡统一接入动态减少动画 | 已提交 `63aa7f4`；motion共6项检查、SDK隔离582项及迁移审查通过 |
| 5.AD | InputOTP原生编辑、分组、过滤及固定上游行为 | 已提交 `7fdfd6e`；40项针对性与165项隔离检查通过；候选窗/AT及特殊空粘贴撤销限制明确保留 |
| 5.AE | Textarea原生拖拽/键盘调整与组合编辑器目标 | 已提交 `8f61aea`；22项正式检查、187项隔离检查及独立审计通过 |
| 5.AF | Pagination长页码原生文本可读性 | 已提交 `fb340d4`；正式回归及41条独立像素审查通过 |
| 5.AG | Gallery原生表单、混合选择、双菜单祖先及大页数接入 | 已提交 `3d7e3b1`；Gallery57项、6项新增集成、独立审计、明暗截图与macOS实窗通过 |
| 5.AH | Spinner/Skeleton源图形、曲线与减少动画画面稳定性 | 已提交 `6d5e6ae`；8项loading、412项隔离检查及独立审查通过 |
| 5.AI | Typography文档组合与原生Kbd | 已提交 `5642cd5`；21项聚焦与296项隔离检查通过 |
| 5.AJ | CodeBlock原生编辑保留、复合槽位及确认复制 | 已提交 `22de763`；33项正式检查、328项隔离检查与交叉审计通过 |
| 5.AK | ScrollArea真实滚动条与AspectRatio几何 | 已提交 `d7c5880`、`4139801`；51项正式、378项主体隔离与交叉审查通过；原生动画偏好缺口保留 |
| 5.AL | Gallery文本/代码、加载与滚动布局组合 | 已提交 `d73880e`；Gallery70项、独立审查、明暗截图及macOS代码选项/复制/文章详情/暂停通过；另补实窗Start/End、Page Up与滚动条切换路径 |
| 5.AM | Toast/Sonner完整应用队列、堆叠、滑动与自定义正文 | 已提交 `603492d`；95项Toast检查、236项隔离检查、交叉审查与全量929项通过；原生兼容路径及平台/AT差异明确保留 |
| 5.AN | Gallery通知配置、动作与原位完成示例 | 已提交 `f531f69`；Gallery84项（新增14项）、明暗截图及macOS原位更新/快捷键返回/动作保持/取消/跨页返回与混合清除通过 |
| 5.AO | Conversation附件、Bubble与Trace模块拆分 | 已提交 `2a2e839`；仅按模块提取文件，原有会话回归通过 |
| 5.AP | 共享原生文字Shimmer与公开组合入口 | 已提交 `cd5cbc9`；21项聚焦、343项隔离检查、vet与三平台构建通过；字体/颜色传递限制明确记录 |
| 5.AQ | Table静态复合槽位、共享列与虚拟入口 | 已提交 `8900efa`；48项聚焦、455项隔离检查、交叉审查及vet通过；多表离屏补测和缩放反馈环已修复 |
| 5.AR | Bubble七变体、内容/反应槽位及原生触发器 | 已提交 `d8e7cad`；61项聚焦、149项隔离检查、交叉审查与vet通过；源字体继承修正纳入正式回归 |
| 5.AS | Attachment五状态、复合槽位与横向附件组 | 已提交 `e7bb5d7`；68项聚焦、217项隔离检查、交叉审查及vet通过；源scroll-padding和混合轴边界差异保留 |
| 5.AT | Trace独立槽位、原生详情及稳定嵌套状态 | 已提交 `d4757d0`；72项聚焦、288项隔离检查、交叉审查及vet通过；字体继承/缩放测量修正纳入正式回归 |
| 5.AU | Gallery静态Table应用组合与双菜单家族 | 已提交 `1a8bd24`；新增6项集成、Gallery90项、隔离全量1206项、独审明暗宽窄截图、vet与三平台构建通过 |
| 5.AV | Gallery会话/附件/Trace应用流程及窄屏自然高度 | 已提交 `f54630f`；新增27项集成、Gallery117项、全量1233项、交叉审查/明暗宽窄截图、vet、三平台构建及macOS实窗流程通过 |
| 5.AW | Message、Marker与MessageScroller模块拆分 | 已提交 `fc7cfe1`；纯文件拆分，204项会话回归通过 |
| 5.AX | 模块ZIP与空缓存独立消费者验证 | 已提交 `da22eb5`；无workspace/replace，外部编译、跨进程设置恢复像素及Gallery明暗渲染通过；不代表公共发布或GUI平台矩阵完成 |
| 5.AY | 可复用滚动条的手动滚动意图回调 | 已提交 `feca45e`；2项真实原生事件检查、490项隔离display/Gallery检查及vet通过 |
| 5.AZ | Message源槽位、内容宽度与原生顺序 | 已提交 `092d8af`；15项聚焦、336项隔离检查、交叉审查及vet通过；后代选择器采用显式native声明 |
| 5.BA | Marker静态/状态组合、源变体与真实原生链接 | 已提交 `47a604e`；32项聚焦、Root独立conversation/Gallery共368项检查、明暗截图及vet通过；Link视觉分离与扫光显式输入限制保留 |
| 5.BB | Bubble复杂原生链接内容与继承样式修正 | 已提交 `08c4865`；71项Bubble聚焦、331项隔离检查、独审另15项真实原生检查及vet通过 |
| 5.BC | MessageScroller回合控制器、源组合与真实原生滚动 | 已提交 `1dd7536`；66项聚焦、401项隔离检查、独审生命周期/输入复验、明暗截图及vet通过；嵌套overscroll和ARIA字段差异保留 |
| 5.BD | Questionnaire源组合、严格导航与原生成功控件数据 | 已提交 `ea655a0`；63项聚焦、504项隔离检查、交叉审查、明暗截图及vet通过；多自由输入集合、浏览器校验与平台/AT验收保留 |
| 5.BE | Gallery完整消息、聊天控制器与问卷复查流程 | 已提交 `09a5666`；新增21项集成、Gallery138项、全量1439项、交叉审查/明暗宽窄截图、vet、三平台构建及macOS实窗通过 |
| 5.BF | 会话阶段模块ZIP的独立消费者复验 | 已提交 `7b173dc`；`09a5666`制品在空缓存/无workspace/无replace条件下外部构建、跨进程主题恢复像素及Gallery明暗渲染通过 |
| 5.BG | 横向Bar、Cartesian配置与原生坐标/检查修正 | 已提交 `0cdb9a6`；71项聚焦、隔离display/Gallery全包、交叉原生复核、明暗截图及vet通过；源自由组合与几何/路径动画范围继续保留 |
| 5.BH | Chart默认Tooltip/Legend与网格样式对齐 | 已提交 `daaf2f7`；26项聚焦、579项隔离检查、独立原生复核及vet通过；保留窄Legend换行与真实原生字体行盒 |
| 5.BI | 极坐标Chart局部主题的绘制快照 | 已提交 `d397be3`；30项正式回归、609项隔离检查、10项独立原生复核及vet通过；Draw不保留Context或应用回调 |
| 5.BJ | Gallery访问量图表与六家族原生组合 | 已提交 `d72dbc0`；新增16项集成、Gallery154项、全量1553项、交叉原生复核/明暗宽窄截图、vet、三平台构建及macOS实窗通过 |
| 5.BK | 图表阶段模块ZIP的独立消费者复验 | 已提交 `af5274b`；`d72dbc0`制品在空缓存/无workspace/无replace条件下的外部构建、跨进程设置恢复像素及Gallery明暗渲染通过 |
| 5.BL | Radar检查色块与实际系列配色一致 | 已提交 `be7dd98`；5项正式回归、630项隔离检查、6项独立原生复核及vet通过；不混入自由组合候选 |
| 5.BM | Chart独立配置容器与Tooltip/Legend内容组合 | 已提交 `c6517f7`；40项正式、670项隔离检查、10项独立原生复核及vet通过；原生图形与样式由调用方显式组合 |
| 5.BN | 公开API的原生Chart组合应用 | 已提交 `b202c08`；6项真实集成、实际SDK提交隔离复验、明暗/500RTL图、vet、三平台编译及macOS实窗通过 |
| 5.BO | Chart仅检查实际绘制系列与空内容布局 | 已提交 `4f370c1`；13项新回归、683项隔离检查、全量1617项及vet通过；未绘制系列不参与域/命中/提示，空面不占Gap |
| 5.BP | 独立消费者验证器只读缓存清理 | 已提交 `e5fcfc8`；2项真实文件系统回归及vet通过，真实空缓存八步验证完成后缓存自动移除，无须手动清理 |
| 5.BQ | 新Chart组合API的双外部消费者复验 | 已提交 `af6d6e9`；`4f370c1`制品标准八步及独立Chart模块六步通过，空缓存/无workspace/无replace，实际明暗PNG与哈希核验完成 |
| 5.BR | Chart七家族原生数据几何动画 | 已提交 `6ca42ad`；145项动画、828项隔离检查与vet、39项独审及明暗中间帧通过；保留最新原值检查、缺失段并行显现及减少动态效果收敛 |
| 5.BS | Gallery图表动画与独立数据更新演示 | 已提交 `804368c`；新增13项、Gallery167项、全量1777项及vet通过；明暗宽窄与真实时钟中间帧、三平台构建及macOS实窗更新/恢复通过 |
| 5.BT | Line清空/缺失历史后的几何动画续接 | 已提交 `2c544f6`；3项新回归、831项隔离检查、10项独审和实际React五类对照通过；全量1780项、vet、三平台构建及macOS最终重启操作通过 |
| 5.BU | Gallery静态导出完成动画后几何 | 已提交 `edbbf87`；3项明暗真实柱形像素回归先红后绿，CLI明暗PNG逐张复核；全量1783项、vet与三平台构建通过，不改变应用动画偏好 |
| 5.BV | 动画SDK及静态Gallery最终模块ZIP验收 | 已提交 `2beb4bf`；`edbbf87`同制品标准八步及独立动画消费者七命令/11项公开检查通过，全新缓存/无workspace/无replace、明暗最终柱形与真实时钟中间帧逐张复核 |
| 5.BW | Base/Vega Flyout菜单与子面板动画 | 已提交 `51ce6e7`；32项动画/19项几何正式检查、332项隔离与vet、18项独立原生复核通过；100ms进退、当前side/origin、退出期inert、反转/编辑器与Menubar真实点击修正 |
| 5.BX | Gallery菜单/右键菜单与Export子菜单演示 | 已提交 `905fbe9`；新增21项、Gallery188项、全量1855项及vet通过；明暗逐帧图、三平台构建和macOS Flyout/系统菜单实际动作回写通过 |
| 5.BY | Select两类菜单触发器统一矢量箭头 | 已提交 `1d9601e`；新增32项真实像素/交互检查、13项独审、全量1887项及vet通过；固定16DIP图标、明暗/RTL/字体对照和macOS实窗两种路由一致 |
| 5.BZ | 菜单/Select阶段模块ZIP外部消费者复验 | 已提交 `526b066`；`1d9601e`同制品标准八步与第二个外部模块五命令/五项公开检查通过，全新缓存、无workspace/replace，几何能力/退场及三路箭头像素验证 |
| 5.CA | Material Symbols按需函数与独立全量目录 | 已提交 `66d83a8`；4157名称/24942 SVG、离线重复生成、11项正式/13项独审及独立SDK1898项通过；真实链接裁剪/像素、三平台构建及macOS浏览切换通过 |
| 5.CB | Combobox选值/query、清除、共框标签及弹层搜索 | 已提交 `e454b8a`；新增53项、377项隔离、77项独审与vet通过；禁用/IME旧缺陷、RTL/Tab/原生编辑/组合回归已修，源100ms保留退出另阶段完成 |
| 5.CC | Gallery可清除/分组/内部搜索/多选Combobox | 已提交 `344311e`；新增17项、Gallery205项、全量1968项及vet通过；十张明暗/窄RTL/打开态图与三平台构建、实际选择回写验证 |
| 5.CD | Material/Combobox同ZIP的独立消费者验收 | 已提交 `bfe9324`；`344311e`标准八步与第二外部模块15命令、4项Combo/13项图标检查通过；双轮空缓存/无replace、527份原文、实际按需裁剪/像素及缓存清理核验 |
| 5.CE | Gallery全局搜索与示例草稿独立持有 | 已提交 `b857d62`；两条真实输入回归先红后绿，隔离Gallery207项及vet通过；网址、内嵌Command与组件搜索跨页互不修改 |
| 5.CF | Combobox源进退动画与原生搜索组 | 已提交 `dd685e9`；新增24项、420项隔离与22项独审、全量1994项/vet/三平台编译通过；100ms/实际side和origin、退出inert/保留过滤及编辑选区、反转和32DIP搜索组 |
| 5.CG | Carousel多项几何、吸附位置与独立组合控件 | 已提交 `4d39f7a`；Carousel48项/display703项、324源函数对照和195项独审通过；全量2020项/vet/三平台编译及许可打包通过；保留完整controller/物理/Autoplay后续范围 |
| 5.CH | Carousel/Combobox动画制品的双外部消费者复验 | 已提交 `17e01f3`；`4d39f7a`同ZIP标准八步及第二外部模块八步/4叶公开检查通过，双轮空缓存/无replace、574份原文匹配、原生明暗图及实际100ms时钟验证 |
| 5.CI | Combobox Base输入、导航、typeahead与弹层语义 | 已提交 `4073227`；新增46项、control259/Gallery207、161叶独审及全量2066项/vet/三平台编译通过；九类源差距修正，motion产品保持不变，Gallery仅更新必要角色定位与强化语义断言 |
| 5.CJ | Gallery多项Carousel、项目草稿及垂直/RTL演示 | 已提交 `c5f0017`；新增12项、Gallery219项及全量2078项/vet/三平台构建通过；布局/间距/loop/短集合/重排/原生编辑/方向键/窄屏滚动与四张明暗宽窄图验证 |
| 5.CK | Carousel应用控制器、部分选项、事件与原生观察器 | 已提交 `069c355`；新增9项、display712项及18项独立公开检查通过，全量2087项/vet/三平台构建通过；FIFO、源事件中间状态、Handle子集、累计resize及单owner生命周期验证 |
| 5.CL | Combobox对象状态、独立槽位、Content配置与交互示例 | 已提交 `1d50ec2`；新增9项control/3项Gallery，268/222项及4项独立公开检查通过；全量2099项/vet/三平台构建、四张明暗宽窄图通过；With共用运行时、单构建手写集合、Grid/原生虚拟窗与12+20例配置映射完成 |
| 5.CM | 新公开API的独立制品复验与Chart源例范围核对 | 已提交 `3402c85`；`1d50ec2`同ZIP两轮空缓存共16命令/1条公开集成流程通过，605文件匹配；132份固定Chart原文/70例逐项盘点，静态需求清单不计作功能或平台验收 |
| 5.CN | Carousel源物理、数值选项与背景原生拖动 | 已提交 `cd3f14c`；12条官方Engine轨迹/985时点、43纯核场景通过，display732/kernel48结果，全量2167项/vet/三平台构建及四张真实时钟逐帧图通过；子控件接管/click取消和多触点SDK缺口仍保留 |
| 5.CO | Create菜单颜色、强调与原生透明材质 | 已提交 `d2ec705`；六种有效配置、同窗Gaussian/饱和度/tint和独立Flyout系统材质、原有Scope兼容及Creator预设流程完成；全量2174结果/vet/三平台构建和macOS实窗选值通过，系统材质核不声明CSS逐像素等价 |
| 5.CP | Chart Cartesian源能力与30例应用 | 已提交 `9092002`；内边距/轴预算、四曲线、渐变/标记/LabelList与既有动画接通；30个可运行源配置，Polar/Tooltip/Gallery目录继续实施 |
| 5.CQ | Carousel typed插件、Autoplay与源Gallery配置 | 已提交 `9d21a84`（零间隔异步修正`9bf0c55`）；公开生命周期/插件动作、源timer/事件状态与原生hover/focus/hidden接通；Gallery使用controller与两秒自动播放，固定八例配置可操作 |
| 5.CR | Tooltip富内容、实际侧向箭头与八例应用 | 已提交 `22f766e`；独立子Context组合、明确描述、Kbd/段落与revision、六侧/RTL和原生几何箭头，保留系统主题对与旧字符串API |
| 5.CS | Tabs独立组合与13例应用 | 已提交 `3e9c7c3`；Root/List/Trigger/Content、同栏输入/按钮/菜单与外置面板，保留旧Tabs，真实键盘/AX/RTL与显式icon名称 |
| 5.CT | Select独立组合、无编辑器多选与22例映射 | 已提交 `c371168`；复用旧便捷入口，六侧/RTL、item alignment、100项滚动/箭头、富内容及当前表单；修复临时绑定身份兼容与消费角色定位 |
| 5.CU | Polar源几何/内容/命中与31例应用 | 已提交 `15ad058`；源角度/半径/扇区/网格/中心/沿弧标签与既有动画、真实ring/axis检查；Tooltip/Gallery整合继续实施 |
| 5.CV | Gallery独立Tabs、多选Select与富Tooltip | 已提交 `6f6d947`；现有模型同步格式选择与跨页保留、header动作/面板关联及真实全局搜索快捷键，保留旧manual Tabs与popup清理 |
| 5.CW | 类型化数值、日期、时间和文件输入 | 已提交 `3f668e7`；真实原生编辑器/分段/系统对话框、应用模型绑定、Field/InputGroup与当前回调，提供完整可运行输入示例 |
| 5.CX | Chart Tooltip/原始数据及完整目录整合 | 已提交 `157c3ca`；9种Tooltip、按检查定位/边缘翻转、类别反转与原始payload，70目录/文档8例/Create6例接入Gallery和独立应用 |
| 5.CY | Persian历法与可编辑日期组合 | 已提交 `1b993fa`；固定33年Jalali模型、Farsi/RTL/历月导航与真实time.Time，应用Parse/Format、错误草稿、真实编辑器/Field/Submit及可操作示例 |
| 5.CZ | 新增API的正常Go模块消费 | 已提交 `c4dedf8`；实际`1b993fa`的canonical ZIP/pseudo-version通过独立项目正常require，无replace；单条跨组件原生表单/图表/菜单流程首次通过，970份源码与Git/ZIP/modulecache一致 |
| 5.DA | Gallery Progress/Radio/Avatar/Skeleton源组合 | 已提交 `1d77483`；同值进度slider、整卡及网格原生单选、真实灰度肖像、非交互form/table占位与共享暂停；保留原独立模型与旧流程 |
| 5.DB | Gallery类型化预约与InputGroup动作组合 | 已提交 `6822d90`；同日期的原生/可编辑输入、文件选择与保存快照，真实clipboard/收藏/搜索/textarea动作、Flyout/OS菜单及窄栏跨页/搜索接线 |
| 5.DC | Field多条错误与自定义原生内容 | 已提交 `97e6e16`；去重列表/独立槽位、真实InputGroup编辑器关联、当前构建正文及应用验证示例；隔离全量2313结果、vet与三平台构建通过 |
| 5.DD | 原生正文/标题字体角色与作用域 | 已提交 `5435ac8`；固定七类标题、继承/system分离、Controller同指针更新和styled Flyout字体快照；隔离全量2318结果、vet与三平台构建通过；26种命名字体资产另行实施 |
| 5.DE | Gallery Payments完整应用组合 | 已提交 `e6ab7be`；固定五条源数据、邮件过滤/排序、列显隐、混合页选择/分页与确认复制；实际窄RTL页面搜索/编辑/返回，Gallery239结果、vet与macOS构建通过 |
| 5.DF | Gallery条件/恢复/跨答案校验问卷 | 已提交 `34ecb6d`；三份源配置复用既有问卷API，独立草稿/保存检查点/Reset/真实校验回焦；窄RTL整页搜索与返回保留当前选择，Gallery241结果、vet与macOS构建通过 |
| 5.DG | Create 26族离线字体与按需/全量入口 | 已提交 `8d6ad7c`；49原始TTF与一份可再生600实例、各族原许可、显式原生注册和真实链接裁剪；隔离全量2324结果、vet与三平台构建通过；Gallery目录接线另行实施 |
| 5.DH | Gallery命名字体选择与完整角色传递 | 已提交 `2f22d8b`；26族正文/标题选择、Inter/继承默认、旧预设/JSON/撤销重做和真实编辑选区；有界Flyout列表与既有浮层字体刷新；Gallery242结果、vet与macOS构建通过 |
| 5.DI | Questionnaire同题独立原生输入集合 | 已提交 `e32e933`；自动/显式Key、独立草稿与成功值、同值控件提交、校验失败准确回焦及默认选择Reset；提供可运行原生组合示例；隔离全量2326结果、vet与三平台构建通过 |
| 5.DJ | Persian示例原始Vazirmatn字体 | 已提交 `1ea4ca1`；显式按需资产/原许可、示例局部字体与原生日期选择，Create目录保持26族；隔离全量2327结果、vet与三平台构建通过 |
| 5.DK | Combobox固定Unicode筛选与JS trim | 已提交 `cb13d72`；固定UTF-16窗口、区域search/标点/零权重、原始query与选值策略；64官方Intl向量及同原生流程，隔离全量2329结果、vet与三平台构建通过 |
| 5.DL | Gallery完整ButtonGroup组合与Avatar菜单 | 已提交 `151577a`；13配置的输入/选择/动作/独立task与原生双路菜单，窄RTL整页搜索/编辑/跨页保留；Gallery244结果、vet与macOS构建通过 |
| 5.DM | 独立可编辑Combobox组件对 | 已提交 `f4d1601`；当前窗口provider、typed单/多选、真实搜索编辑器与即时Filter，取消外点保留同一编辑器；Host统一退出/清理、平台主题版本及可运行小窗口示例；隔离全量2332结果、vet与两应用三平台构建通过 |
| 5.DN | Gallery完整Card源组合 | 已提交 `ff4f839`；Create12配置与文档不同内容共17配置，原始图片离线、间距/边框/贴边滚动、真实登录编辑与本地动作、会议笔记和RTL；复用现有Card API，Gallery246结果、vet与macOS构建通过 |
| 5.DO | Field独立组合与完整源表单 | 已提交 `19b4471`；顺序自由的真实目标关联、前置说明/丰富标签/整卡选择、Legend/Separator及Slider独立Invalid；Gallery映射13docs与Create10分区，旧便利布局保留；隔离全量2337结果、vet与三平台构建通过 |
| 5.DP | Gallery完整Alert与Empty源组合 | 已提交 `74a4176`；20通知/13空状态、真实链接/搜索/离线头像、明暗颜色与RTL，独立本地动作与跨页草稿；复用已有SDK，Gallery250结果、vet及macOS构建通过 |
| 5.DQ | Disclosure独立部件与完整源组合 | 已提交 `e0f2659`；任意布局Trigger/Content、当前帧主题载体、真实inner样式及Accordion StyleItem，修复新旧首次展开测量；17源配置与原生编辑/文件树/RTL；隔离全量2342结果、vet与三平台构建通过 |
| 5.DR | Gallery完整Item源组合 | 已提交 `c562a6b`；35配置的尺寸/变体/原始图像、真实主链接和独立动作、Header/Footer/分组及成员菜单，补充自绘菜单可配置行高；隔离全量2345结果、vet与三平台构建通过 |
| 5.DS | Gallery完整Popover与HoverCard源组合 | 已提交 `de7ba53`；12配置的真实独立字段、Title/Description、align/6side、Dialog嵌套、10/100ms预览与RTL；复用已有桌面API，Gallery256结果、vet与macOS构建通过 |
| 5.DT | Gallery完整Dropdown Menu与源菜单选择契约 | 已提交 `067fef3`；27配置的原始组／快捷键／禁用与状态、多层动作、双头像、六侧／Dialog／RTL，补充逻辑行／标题缩进、独立前置图标与尾端SVG、逐项CloseOnSelect及Select兼容、SVG方向箭头；隔离全量2350结果、vet与三平台构建通过 |
| 5.DU | Gallery完整Context Menu源组合 | 已提交 `0a7485f`；22配置的真实右键／菜单键、分组状态与原生SVG模板、六侧／Dialog／Inset及翻译RTL；隔离Gallery／cmd共260结果、vet及macOS构建通过 |
| 5.DV | Gallery完整Menubar源组合 | 已提交 `a021b3e`；15配置映射18源分区，保留完整命令／Format／Insert／图标差异、同child单选互斥、Native模板图与六侧／Inset／Dialog／RTL；隔离Gallery／cmd共263结果、vet及macOS构建通过 |
| 5.DW | Menubar逐顶层宽度与源尺寸接通 | 已提交 `5385ef6`；逐菜单宽度覆盖共享默认，同Surface兄弟切换／child提前输入／动态更新与有限值守卫，保留所有显式源尺寸；隔离全量2356结果、vet与三平台构建通过 |
| 5.DX | Gallery完整Badge源组合 | 已提交 `324572b`；14配置的六变体／双侧图标／加载／真实链接、完整长文本横滚、显式明暗颜色与RTL；隔离Gallery／cmd共265结果、vet及macOS构建通过 |
| 5.DY | Gallery完整Button源组合与样式修正 | 本阶段提交；20配置与五个4×6矩阵、真实Link／图标标记／RTL，修正单层禁用透明度、小尺寸gap／圆角与逻辑起点内距；隔离全量2360结果、vet与三平台构建通过 |

用户已明确允许原生平台适配：Popover、menu、下拉菜单可以使用 MyGo 的系统菜单与独立 Flyout，超出父窗口范围；无需逐像素照搬网页浮层。Gallery 以 [shadcn/create](https://ui.shadcn.com/create) 的配置与实时预览流程为参考。完整组件功能、主题、动画范围保持不变。

## 下一轮的明确缺口

- Combobox九类Base交互差距及原生组合API已在5.CI/5.CL交付：Content定位/anchor宽度、Empty/List、可取消状态事件/Always/Grid/虚拟窗口及12+20例配置完成。菜单可选translucent preset已在5.CO交付；独立越窗普通搜索集合由desktop组件对提供，任意跨窗槽位/Grid/虚拟窗口、完整locale/Unicode及平台IME/AT继续验收。
- macOS实窗无active Enter复验：本轮自动化出现旧bundle/AX目标失效及首次Return仍留列表的观测。补充12项公开settled输入检查全绿；Tester无公开batch接口，无帧输入时序仍未证实。先辨明实际进程、IME与事件时序，再依据实证修正。
- Carousel多项布局、controller/options/events及源物理/数值选项/背景拖动已在5.CG/5.CK/5.CN完成；Autoplay/插件与源Gallery配置已在5.CQ交付。从任意子控件起拖并抑制其click、完整多触点信息仍缺MyGo公开能力；保持原生子控件正常输入，未计完整源gesture闭环。按parity表补齐Chart高级配置及所有组件例子；Windows/Linux实际窗口、真实AT/IME、多显示器/缩放和完整视觉矩阵仍是最终验收要求。
- Chart按[固定源需求清单](chart-source-coverage.json)实施三个完整模块：Cartesian30例的margin/曲线/渐变/标记/LabelList与既有动画已在5.CP交付；Polar31例的角度/半径/网格/扇区/中心内容已在5.CU交付；Tooltip9例及交互、docs/RTL/Create与Gallery目录整合在5.CX交付。Brush、ReferenceLine和额外显式动画参数未在132份固定原文中使用，列为Recharts扩展能力，不以其替代真正缺失的源例功能。该清单记录`1d50ec2`时的静态差距，不能当通过计数。
- Select组合、独立Tabs、富Tooltip与类型化输入/历法/可编辑日期模块已在5.CR至5.CY交付。Gallery中的Progress标签/值、Radio cards、Avatar灰度及Skeleton表单/表格在5.DA交付；类型化预约表单和InputGroup动作/textarea/menu组合在5.DB交付，复用原有SDK语义。后续继续核对其余固定源组合与原生适配边界，实际平台/AT/IME验收保持独立记录。


## 实施约定

- 组件返回 `ui.Element`，允许调用方继续设置样式；状态由调用方持有。
- 使用 MyGo 原生 Base 控件处理输入、焦点、键盘及可访问性；稳定数据用 `Context.Key` 或 `ListState.Key`。
- 所有颜色使用语义 token；完整主题按 Context 作用域传递。
- 渲染函数在 UI 线程使用；异步数据通过窗口的 `Update` 调度，按需刷新。
- 动画读取桌面减少动态效果偏好；主题偏好采用临时文件加原子替换保存。
- 参考 Music 的应用架构经验，独立实现；不复制其 AGPL 源码。

## 参考基线

阶段 5 的组件与样式源码基线固定在 [upstream.md](upstream.md)，详细组件路径和内容校验值见 [upstream.json](upstream.json)。

- Lobe MyGo：`40d9ea87a3c037531a83cdb1578b5e6102b74d03`。
- MyGo Music：`1af0d8efd77fdf630c27cfc635f12e9cb089ce10`。
- MyGo 运行时：`v0.0.0-20261108125120-b5db0b40dbbc`。
- shadcn 组件目录：[官方 66 项组件目录](https://ui.shadcn.com/docs/components)，2026-11-09 核对。AI Elements 独立目录不混入这 66 项。
