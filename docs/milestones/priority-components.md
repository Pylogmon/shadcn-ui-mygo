# 七组件优先里程碑

用户于2026-11-10指定：优先完善Context Menu、Dialog、Input、Kbd、Label、Item，随后补入Tabs；完成这些后作为一个里程碑，暂时停止后续开发。整体Native Go UI Kit目标保留，里程碑完成不表示整个项目完成。

本次收口包括公开原生组件功能、固定shadcn revision的完整源示例、已证实且可修复的专项问题、真实原生Tester指针／键盘／数据保留，以及相关测试、vet和跨平台编译。Native适配保留真实控件：系统菜单使用OS样式，无对应OS控件时明确使用平台主题内容；Date／Time仍绑定明确time.Time，不把零值、只改AX文字或覆盖假编辑器当作HTML空值。

| 组件 | 当前里程碑状态 | 范围 |
| --- | --- | --- |
| Context Menu | 已提交31ee0b3 | 22配置；逐submenu宽度／RTL／动态与禁用视觉、双路菜单及原生交互；全量2382结果、vet与三平台构建通过 |
| Dialog | 已提交3337b90 | 12完整docs／Create／guide配置、表单、滚动／footer、Chat Settings四面板、焦点与关闭行为；隔离Gallery／cmd288结果、vet与macOS构建通过 |
| Input | 专项与最终联合验证通过 | 25配置与真实文本／数字／日期／时间／文件；17原文与API审查无新增阻断，保留明确的原生Date／Time适配边界 |
| Kbd | 已提交05bdc26 | 17配置；真实按钮／输入与独立Tooltip、窄RTL搜索和跨页状态；Gallery／cmd290结果、vet与macOS构建通过，静态keycap不安装快捷键 |
| Label | 本阶段完成 | 9配置、真实目标关联、leading-none／select-none与disabled peer单层样式；联合全库2386结果、vet与三平台构建通过 |
| Item | 专项与最终联合验证通过 | 35配置、12原文、真实主目标与独立动作／header／footer及窄布局；新增公开链接／禁用／窄边界探针通过 |
| Tabs | 已提交b55927d | 20配置、原生pointer／RTL keys／独立状态、真实header控件和禁用单层透明度；全量2380结果、vet和三平台构建通过 |

每个实施阶段分别commit。全部七项收口后记录准确提交和验收证据，并按用户要求暂停目标；不继续其他组件开发。

真实Windows／Linux窗口、完整屏幕阅读器／IME、多显示器／缩放与全主题字体视觉矩阵仍属于整体交付验收，见[完整差距与验收门槛](../parity.md)。本里程碑不把软件渲染或编译称作这些实机验收。

已移出本次范围的候选保存在`build/parked-outside-priority-milestone/2026-11-10`。其中Input Group尚未完成最终图审／交叉审查／Root接线；disabled一致性候选已通过定点验证但未提交。它们均未计入本次已交付组件，恢复时需重新核对原manifest和当前SDK。

七项当前实现已通过最新联合SDK的全量2386测试结果、vet及三平台构建。尚待提交后以精确版本模块ZIP运行最后一轮普通外部消费者；该结果与最终提交将在收口时记录，随后暂停。
