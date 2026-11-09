# 固定上游对照基线

阶段 5 从 shadcn/ui 提交 [`8ab27715521cbd4f87ae6886cd989cbb678f6de7`](https://github.com/shadcn-ui/ui/tree/8ab27715521cbd4f87ae6886cd989cbb678f6de7) 进行源码对照。组件采用 `apps/v4/registry/bases/base/ui` 的组合与行为约定，首个原生视觉配置对照 `style-vega.css`。其他 primitive 家族和视觉配置仍保留在最终差异审计范围，不能由当前配置的测试推定通过。

[upstream.json](upstream.json) 保存组件路径、固定提交和已读取的组件/样式文件 SHA-256。引用原文时使用该提交的 GitHub 路径；不要将不断变化的 `main` 或官网内容与这个基线混用。Data Table 与 Typography 是文档中的组合方案，Toast 对照 `sonner.tsx`，因此上游文件数量与组件目录数量不同。

原有 `pkg/theme/sources/manifest.json` 独立固定主题与色阶的原始字节、SHA-256 和 MIT 声明。主题源的固定字节与组件源的固定提交分别记录，生成器继续离线运行。

这些记录提供可复查的对照来源，不证明 Go 组件已与上游一致。行为、状态、视觉与平台验收继续记录在 [parity.md](parity.md) 与各阶段验证记录中。

Command 的评分行为另外固定到 shadcn 该提交在 `pnpm-lock.yaml` 锁定的 cmdk 1.1.1：提交 [`fb4ea04e9ec211777fbb39c6104e3c5f2ee107d2`](https://github.com/dip/cmdk/tree/fb4ea04e9ec211777fbb39c6104e3c5f2ee107d2)，原仓库 `pacocoursey/cmdk` 现重定向到 `dip/cmdk`。原始 `command-score.ts`、MIT许可证、来源SHA-256和可离线生成的数值对照在 `pkg/overlay/testdata/cmdk`。Go移植的版权声明见 [第三方声明](../THIRD_PARTY_NOTICES.md)。

Input OTP 的底层行为固定到 input-otp 1.4.2 的发布 gitHead：[`81ccdb48c010d800b24942aa231909f0c971b1ca`](https://github.com/guilhermerodz/input-otp/tree/81ccdb48c010d800b24942aa231909f0c971b1ca)。原始 input.tsx、MIT许可证、SHA-256、发布版本证据和默认字符/完成/粘贴选择的精确行号保存在 `pkg/control/testdata/input-otp`。原生 rune 索引、构建预算、候选窗定位与特殊空剪贴板撤销限制见 [Input OTP](input-otp.md)。

Spinner 的 Loader2 图标对照 shadcn apps/v4 锁定的 Lucide 0.474.0，官方 tag 指向提交 [`97bbe1d6b2adc6767831e52f6b12b2886b6498d5`](https://github.com/lucide-icons/lucide/tree/97bbe1d6b2adc6767831e52f6b12b2886b6498d5)。SVG、ISC许可（包含Feather MIT说明）、来源证据和SHA-256在 `internal/indicator`；原始tag与不可变提交字节相同。Loading样式另外核对本次固定的 spinner.tsx、skeleton.tsx、style-vega.css，以及 globals.css 的比例圆角定义。

Toast 的底层参考固定到 shadcn 同一 `pnpm-lock.yaml` 中的 Sonner 2.0.7：官方 `v2.0.7` tag 指向提交 [`3ba7aa17ab7e8101b9cf4893936f873b0d4769b3`](https://github.com/emilkowalski/sonner/tree/3ba7aa17ab7e8101b9cf4893936f873b0d4769b3)。原始 index/styles/assets/types/state、MIT许可及SHA-256归档在 `pkg/overlay/testdata/sonner`，与发布tag字节核对一致。状态图标使用上述同一Lucide提交的CircleCheck、Info、TriangleAlert、OctagonX；关闭图标保留Sonner的12DIP原图形，来源见 `pkg/overlay/assets/manifest.json`。Vega的 `rounded-2xl` 受固定globals.css的比例定义覆盖，按当前radius×1.8处理。
