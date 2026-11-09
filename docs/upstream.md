# 固定上游对照基线

阶段 5 从 shadcn/ui 提交 [`8ab27715521cbd4f87ae6886cd989cbb678f6de7`](https://github.com/shadcn-ui/ui/tree/8ab27715521cbd4f87ae6886cd989cbb678f6de7) 进行源码对照。组件采用 `apps/v4/registry/bases/base/ui` 的组合与行为约定，首个原生视觉配置对照 `style-vega.css`。其他 primitive 家族和视觉配置仍保留在最终差异审计范围，不能由当前配置的测试推定通过。

[upstream.json](upstream.json) 保存组件路径、固定提交和已读取的组件/样式文件 SHA-256。引用原文时使用该提交的 GitHub 路径；不要将不断变化的 `main` 或官网内容与这个基线混用。Data Table 与 Typography 是文档中的组合方案，Toast 对照 `sonner.tsx`，因此上游文件数量与组件目录数量不同。

原有 `pkg/theme/sources/manifest.json` 独立固定主题与色阶的原始字节、SHA-256 和 MIT 声明。主题源的固定字节与组件源的固定提交分别记录，生成器继续离线运行。

这些记录提供可复查的对照来源，不证明 Go 组件已与上游一致。行为、状态、视觉与平台验收继续记录在 [parity.md](parity.md) 与各阶段验证记录中。
