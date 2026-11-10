# 七组件优先里程碑

状态：2026-11-10完成本次里程碑，按用户要求暂停后续开发。

范围由用户指定的Context Menu、Dialog、Input、Kbd、Label、Item及随后补入的Tabs组成。整体Native Go UI Kit目标保留；本次验收范围为公开原生组件、固定shadcn revision的完整源示例、已证实且可修复的专项问题、原生Tester交互与状态，以及相关测试、vet和跨平台编译。

| 组件 | 本次状态 | 配置数 | 实施提交与主要结果 |
| --- | --- | --- | --- |
| Context Menu | 完成 | 22 | `0a7485f`／`31ee0b3`；双路菜单、逐子菜单宽度、RTL与动态几何、禁用单层透明度 |
| Dialog | 完成 | 12 | `3337b90`；完整Profile、滚动／固定footer、两种无关闭契约、四面板Chat Settings与焦点恢复 |
| Input | 完成 | 25 | `bf3ab71`／复核`7fdfcf9`；实际原生编辑器、数字空值与0、类型化输入及独立草稿 |
| Kbd | 完成 | 17 | `05bdc26`；静态键帽、实际父按钮／输入、独立富Tooltip与RTL，不自动安装快捷键 |
| Label | 完成 | 9 | `270aa8b`；真实目标关联、leading-none／select-none、disabled peer样式与完整关联Field付款表单 |
| Item | 完成 | 35 | `c562a6b`／复核`7fdfcf9`；真实主链接、独立action／header／footer、窄布局和禁用行为 |
| Tabs | 完成 | 20 | `b55927d`；完整内容、独立Root／面板、真实header控件、RTL键盘与禁用透明度 |

共140个配置，阶段分别commit。最终SDK代码提交为`270aa8bf9823cd088e4b0631b3e5a87240f8e024`；后续验收文档提交不改变SDK代码、依赖或资源。

最终联合版本在新鲜隔离中通过全库2386个测试结果、`go vet ./...`、macOS应用打包及Linux／Windows编译。Gallery的精确搜索实际定位源配置区，原生点击／编辑／键盘与跨页返回保留独立模型。代码与源示例均完成独立审查；代表图片已逐张检查。

普通模块消费者验证使用版本`v0.0.0-20261110055946-270aa8bf9823`、新空模块／构建缓存、`GOWORK=off`、无`replace`。标准8命令与最后外部6命令全部成功；一个七组件原生集成流程通过，1344个SDK文件的Git、模块ZIP与实际下载缓存原文逐字节一致。独占缓存已删除，两个最终PNG分别检查通过；菜单子窗使用显式500×240软件画布，不作为OS FitContent证明。

本地证据保留在`build/priority-milestone-2026-11-10/manifest.json`与`build/consumer-verification/run-1140721527/priority-public-run-5pxb9c7y/receipt.json`。外部收据SHA256为`8b9850f23a24c952d74d08e70c49801ad268071bf341cf7373c6cd54029437ab`。本地`build/Shadcn Gallery.app`已替换为精确最终代码对应的已验证bundle，并保留旧版本备份；本次未宣称启动该bundle作实窗验收。

Native适配保留真实控件：系统菜单使用OS样式，无对应OS控件时明确使用平台主题内容。Date／Time仍绑定明确time.Time，不能表示HTML空值；不把零值、只改AX文字或覆盖假编辑器作为空值实现。

真实Windows／Linux窗口、完整屏幕阅读器／IME、多显示器／缩放与全主题字体视觉矩阵仍属于整体交付验收，见[完整差距与验收门槛](../parity.md)。这些待办保留，后续组件开发按本次请求暂停。

范围外候选保存在`build/parked-outside-priority-milestone/2026-11-10`。Input Group候选尚未最终验收；其他组件的disabled一致性候选也未提交或计入本次交付，恢复时需重新核对manifest与当前SDK。
