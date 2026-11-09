# 独立消费者与模块 ZIP 验证

`examples/consumer/main.go` 是只使用公开包的原生应用示例：主题控制器、局部Rose主题、原生输入、Flyout/系统菜单选择，以及调用方指定路径的主题保存与恢复。复制到自己的模块后可通过正常Go依赖构建；不需要导入本项目internal包。

```sh
go run ./examples/consumer
go run ./examples/consumer -preferences /tmp/native-theme.json
make verify-consumer
```

`verify-consumer` 将已提交HEAD打包为标准Go模块ZIP，并创建临时文件代理。新建的外部消费者模块使用require引用该版本；GOWORK关闭，不使用replace。模块缓存和构建缓存在每次验证开始时均为空，SDK从生成的ZIP读取，其余固定依赖从官方Go代理下载并按正常校验数据库验证。SDK本地制品尚未公开发布，因此仅它的校验数据库查询被禁用。

流程实际编译外部消费者、运行它的PNG渲染和主题保存，再启动第二个消费者进程只加载设置，比较主题与局部控件的实际像素。随后从模块代理运行Gallery的浅色和深色渲染。工作区未提交的SDK改动不会混入模块ZIP；消费者示例独立复制，其源码校验值另行记录。

输出保留在 `build/consumer-verification/run-*`：模块ZIP/版本信息、外部go.mod/go.sum、依赖图、命令日志、设置文件、四张PNG与receipt.json。记录包含SDK提交、伪版本、ZIP和示例SHA-256、Go版本及目标平台。临时模块/构建缓存完成后删除。

清理会先恢复本次独占缓存中只读目录/文件的owner权限，避免Go下载模块的只读目录残留；不沿符号链接进入其它目录。清理失败会与原验证错误合并并返回非零退出状态，不再静默忽略。此清理只作用于验证器自行创建的临时缓存，不处理用户现有缓存。

```sh
go run ./cmd/verify-consumer -ref <commit> -out build/consumer-verification
```

该流程证明本地模块制品可被独立消费者使用并运行无窗口原生渲染，没有发布到公共Go代理，也没有替代真正窗口的启动、系统主题变化、三平台运行或完整组件验收。新模块实现提交后可再次使用同一流程验证该提交。

会话SDK与Gallery阶段完成后，于2026-11-09再次验证提交`09a56667eac94250dc89a8fbc42534c5458faf77`，伪版本为`v0.0.0-20261109111738-09a56667eac9`。空模块/构建缓存、关闭workspace、无replace的外部构建、跨进程主题恢复像素及Gallery明暗渲染均通过。证据保存在`build/consumer-verification/run-1723552001`，模块ZIP的SHA-256为`1f1e7069991f4a38803b4bec0d4630f42a3a28cdc378b9821e4c8593934a0ba4`；消费者源码SHA-256为`b678ffb61c408baf655f9a26dfc89effa00a3c54041f25949d347a4c85efa91f`。Root另外核对receipt、实际依赖图、ZIP哈希，并查看恢复后的消费者与模块Gallery深色PNG。

图表SDK与Gallery阶段完成后，同日验证提交`d72dbc0238b7a6455be8838917b34d665469b224`，伪版本为`v0.0.0-20261109122650-d72dbc0238b7`。全新缓存、关闭workspace、无replace条件下的tidy、模块图、消费者编译、首次渲染、第二进程恢复、Gallery明暗渲染与环境检查八步通过。证据为`build/consumer-verification/run-3091013942`；ZIP SHA-256为`1fa04441a4665e63c2f909afe2ce822dd1ca90d03ef56a86474fc8a6cf17ba69`，消费者源码校验值不变。独审与Root均实际查看恢复消费者和Gallery深色PNG，Root另核验收据、实际模块图及两个文件哈希。该记录只覆盖该提交，不包含后续在途Chart自由组合实现。

自由组合API与示例完成后，提交`4f370c1e01d68ca72b0eb016b65700e571469cb1`以`v0.0.0-20261109125141-4f370c1e01d6`复验，ZIP SHA-256为`614923e93cd5f662c04d95f119ad63437a43e476ca6e01bd208a6f1d3b33a40f`。标准八步通过后，另建外部模块`example.invalid/chart-composition-consumer`，直接从该ZIP复制examples/chart-composition/main.go，使用另一套全新模块/构建缓存引用同一SDK版本。tidy、实际模块图、build、明暗渲染及环境检查六步均通过；GOWORK关闭、无replace，SDK确从新模块缓存解析。额外源码SHA-256为`d1f7ffd55c70bede08fa60203caa03341edc300dcd25269f22bd715591b55946`，两张PNG与先前隔离示例逐字节一致，确实使用新ChartContainer/config/Tooltip/Legend内容API。

完整标准及Chart证据在`build/consumer-verification/run-772084584/distribution-phase-receipt.json`，Root核验14条命令状态、模块图、ZIP/源码/PNG哈希，并实际查看恢复消费者及外部Chart深色图。首个标准运行的只读缓存曾需人工清理，已记录于该收据；随后缓存清理修复在`run-3088365774`用同一SDK制品完成标准八步，退出后缓存自动消失。额外Chart缓存也已删除。此次仍为本地制品/无窗口渲染验收，不包含后续在途动画，也未替代公共发布或平台GUI验收。

动画SDK、Line历史补修及静态导出修复完成后，最终提交`edbbf871116923d2a31348d50813aaa655af1e3b`以`v0.0.0-20261109135517-edbbf8711169`验证，模块ZIP SHA-256为`1626954f38f42faedeee69f0f856b204b97253145cc905321ae63237a0e2b5c4`。同一制品标准八步，以及另一个真正外部模块`example.invalid/chart-animation-consumer`的七条命令和11项公开API检查均通过。两轮模块/构建缓存均从空开始、GOWORK关闭且无replace，SDK从新缓存解析；两套临时缓存均已自动删除。外部工程实际使用Animation的Duration、Begin、Ease和Disabled，验证Bar/Line的100ms延迟、400ms linear半程像素、最新AX原值、减少动态效果及Line缺失/清空后恢复，另导出真实时钟明暗对照图。

最终证据位于`build/consumer-verification/run-3352262293/animation-distribution-receipt.json`，包含两份完整收据、15条命令、模块图、环境、SDK/消费者源码及PNG哈希。五张图由Owner逐张实际查看，Root核验ZIP内SDK/CLI源码与workspace逐字一致，并查看最终Gallery深色和外部动画深色图。旧提交`2c544f6`虽通过标准命令，但Gallery静态图停在动画入口；其诊断证据保留于`run-1965657269`，不冒充视觉通过。最终同ref已重新执行所有步骤，Gallery明暗六柱完整，不复用旧图或只靠文件大小判完成。此步骤仍是本地制品验收，不是公共发布或三平台实窗验收。

菜单动画、Gallery子菜单与Select箭头阶段提交`1d9601ef61be21cefe9585689b62d852fca84749`以`v0.0.0-20261109144504-1d9601ef61be`复验，ZIP SHA-256为`af88a0645342beee03f9b4803d02d681b18d2e95ac7591d8f027c463de638e47`。标准八步及第二个外部模块`example.invalid/menu-consumer`的tidy/模块图/test/vet/env五命令均通过；公开检查共五项，包括实现GeometrySurface/DismissInterceptor、实际左侧几何的进入像素、100ms完成、保留退出/AX排除/最终一次关闭回调，以及三路Select关闭触发器的10×6着墨与精确居中。

两轮缓存均从空开始并已删除，GOWORK关闭、依赖图无replace、SDK确来自同一ZIP。Root核对ZIP内十个产品文件与该提交原文、完整收据/哈希，并查看恢复消费者及Gallery明暗图。证据`build/consumer-verification/run-1641041268/menu-distribution-receipt.json`。首次自定义探针把OnDismiss误断言为退场开始时调用；按既有完成退场契约纠正测试后，最终五命令在另一套全新缓存完整重跑，产品未因此修改。并行Combobox候选未混入制品；公共发布、OS放置/实窗及完整平台矩阵仍未完成。

Material Symbols与Combobox最终代码提交`344311e962ce5717946a7fe89986d7ba0ff25125`以
`v0.0.0-20261109153319-344311e962ce`复验，ZIP SHA-256为
`c51916a6b0ec19943dccef507d69d0506d2dde965f4cf949762642b88124897a`。
标准八步，以及第二个真正外部模块的15命令、4项Combobox检查和13项Material公开API断言均通过。
后者验证合法空值、标签/query、无active Enter及多选IME/disabled；四个真实应用验证原始SVG、空导入、
按需Home与全量名称入口。全部应用像素和PNG相同，按需/空导入与baseline均4271650字节，
全量15637314字节；完整资源ZIP仅出现在显式全量应用中。

两轮模块/构建缓存均初始为空、GOWORK关闭、完整依赖图无replace；527份SDK ZIP文件与第二轮模块缓存
逐字节匹配，结束后独占缓存均已删除。Root核对同ref/版本/ZIP、15条命令及日志哈希、外部Go源码、
四份实际二进制/PNG/全ZIP包含关系。新标准四图与本轮此前实际查看的相应图片逐字节相同；Owner和Root
查看外部Home透明PNG，另三图与它字节相同。该PNG是raw SVG.Image默认黑色透明底，不代表全目录或
ui.Icon颜色/布局验收。

完整证据为`build/consumer-verification/run-1850267044/material-combobox-distribution-receipt.json`，
第二模块的源码、环境、图像、二进制、完整依赖图与复制再运行脚本在同目录的`external-icons-combobox`。
SDK来自本地标准模块代理；此验证未执行公开发布或替代原生平台/AT验收。


Carousel多项组合与Combobox源动画阶段完成后，实际提交
`4d39f7a6f2fab11152be115fe51d435a342da7fc`以
`v0.0.0-20261109161542-4d39f7a6f2fa`通过标准八步，ZIP SHA-256为
`c9d38a7f8252631b1f52a9877472a436fd776b57df2da0a2c6f8f6a33c10a454`。
另建真正外部模块`example.invalid/carousel-consumer`，另一套全新缓存从同一ZIP执行
八步：tidy、完整模块图、公开native测试、vet、build、明暗渲染及环境输出。
四个叶检查（含父项共五条通过结果）覆盖水平LTR/RTL和垂直Carousel、前置controls/snapshot、
五slides与四snaps、真实选择与禁用边界、独立slice，以及Combobox实际CSS-ease进入和100ms保留退出/AX排除。

两轮GOWORK均关闭、无replace、SDK来自新模块缓存；第二轮全部574份ZIP文件与实际缓存逐字节匹配，
结束后两套独占缓存均已确认删除。Root核对16条命令完成状态、完整收据、ZIP/源码/日志/二进制及PNG哈希。
标准四图与此前已实际查看的`run-1850267044`对应图逐字节一致；新增外部Carousel明暗两图实际查看。
静态导图关闭动画，独立时钟检查保留真实动画；两者的证据范围分开记录。

完整证据为`build/consumer-verification/run-1157201957/carousel-combobox-distribution-receipt.json`，
第二模块源码、日志和可重现脚本见同目录`external-carousel/README.audit.md`。
此次不包含在途Combobox Base交互修正，也不把兼容EaseOut计作Embla物理、完整controller/Autoplay，
更不代表公共发布或全平台GUI/AT验收。

Carousel控制器与Combobox组合API提交`1d50ec226dd1c4246d1e4e73319b1bff614f7dff`
以`v0.0.0-20261109172000-1d50ec226dd1`完成标准八步。ZIP SHA-256为
`6f04af0ab5af92b7c0e3de87ea34c2f8fd38be4ccb0c991475484e6b99a6833e`。
第二个真正外部模块使用另一套空缓存，从同一ZIP完成八步及一个公开集成流程：
对象型Combobox选值、实际Editor.OnSubmit/FormValues、分组Carousel控制器事件，
以及部分ReInit与下一条导航的FIFO处理。两个依赖图均无replace，GOWORK关闭；
605份SDK文件同时匹配已提交原文、ZIP和实际外部模块缓存，两套独占缓存已删除。

证据见`build/consumer-verification/run-1326536727/controller-compound-distribution-receipt.json`，
外部Go源码与复制再运行脚本见其`external-native-controls/README.audit.md`。
标准四图与此前实际查看的`run-1157201957`对应图逐字节相同；Root查看新的外部明暗两图，
其中实际显示分组跳转后的Project03/04。静态图及本次集成流程使用减少动态效果，
不充当新的源动画验收；本轮没有发布SDK，也未包含在途物理运动/菜单预设/Chart改动或替代平台GUI/AT验收。
