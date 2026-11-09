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
