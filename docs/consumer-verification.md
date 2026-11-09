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

```sh
go run ./cmd/verify-consumer -ref <commit> -out build/consumer-verification
```

该流程证明本地模块制品可被独立消费者使用并运行无窗口原生渲染，没有发布到公共Go代理，也没有替代真正窗口的启动、系统主题变化、三平台运行或完整组件验收。新模块实现提交后可再次使用同一流程验证该提交。

会话SDK与Gallery阶段完成后，于2026-11-09再次验证提交`09a56667eac94250dc89a8fbc42534c5458faf77`，伪版本为`v0.0.0-20261109111738-09a56667eac9`。空模块/构建缓存、关闭workspace、无replace的外部构建、跨进程主题恢复像素及Gallery明暗渲染均通过。证据保存在`build/consumer-verification/run-1723552001`，模块ZIP的SHA-256为`1f1e7069991f4a38803b4bec0d4630f42a3a28cdc378b9821e4c8593934a0ba4`；消费者源码SHA-256为`b678ffb61c408baf655f9a26dfc89effa00a3c54041f25949d347a4c85efa91f`。Root另外核对receipt、实际依赖图、ZIP哈希，并查看恢复后的消费者与模块Gallery深色PNG。
