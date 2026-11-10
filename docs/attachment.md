# Attachment 组合

`Attachment` 提供文件图标/图片、标题、描述和独立操作。`AttachmentWith` 暴露Media、Content、Actions、Title、Description、Action和Trigger槽位；Parts只在当前构建中使用，不保存到后续帧。

```go
conversation.Attachment(c, conversation.AttachmentProps{
    ID: "report", Name: "report.pdf", MediaType: "application/pdf",
    State: conversation.AttachmentUploading, Progress: .64,
    OnCancel: cancelUpload,
    OnRemove: removeAttachment,
})
```

状态为Idle、Uploading、Processing、Error、Done；空值和旧ready值兼容Done。Progress是0到1的显示值。三种Size与水平/竖直Orientation可组合；Image使用真实原生图片、Cover裁剪和ImageAlt。HideMedia可做纯内容组合。上传/处理标题使用共享Shimmer，桌面减少动画下显示静态文字。

`AttachmentProps.AlignSelf` 对 `Attachment` 与 `AttachmentWith` 的外层 Root 生效，零值保留 `ui.Start`；`ui.Center` 在横排中垂直居中、纵排中水平居中。Center/End 不改变附件自然宽度，竖向附件继续保留固定宽度；也可在返回元素或 `Parts.Root` 上链式设置对齐。

OnOpen保留旧的独立Open动作；OpenLabel可重命名。需要整卡触发时，在AttachmentWith中调用Parts.Trigger，子按钮和编辑器仍保留自己的输入、选区、键盘及辅助技术目标。取消、重试、下载和移除都是显式回调；本库不启动上传服务或网络传输。

```go
conversation.AttachmentWith(c, props, func(a *conversation.AttachmentParts) {
    a.Media(conversation.AttachmentIconMedia, func(c *ui.Context) {
        ui.Icon(c, fileIcon).Size(20, 20)
    })
    a.Content(func(c *ui.Context) {
        a.Title("Design tokens")
        a.Description("JSON · 2.4 KB")
    })
    a.Trigger("Open design tokens")
})
```

AttachmentGroup/AttachmentGroupWith提供真实原生横向viewport、snap和边缘渐隐；调用方可持有ScrollState。附件ID保持重排后的编辑器、焦点与选区身份。Home/End仅在viewport自身聚焦时滚至首尾，内层编辑器保留自己的按键。纯水平越界滚动被包含，普通竖直滚动仍交给外层；在边界混合轴事件中，公开HandleInput不能只转发其中一个轴，这一差异明确保留。源的4 DIP scroll-padding目前也没有公开原生对应项。

OnDropFiles是可选的原生文件拖入组合；应用收到路径副本并负责验证、传输与持久状态。源Attachment本身也不是上传引擎或自动dropzone。所有回调都在UI构建线程中处理；异步任务通过窗口Update更改应用模型。

原生Theme没有计算祖先颜色getter，上传标题的扫光以CardForeground为基础；自定义父文字色应通过显式组合控制。完整平台、触控、文件拖入、辅助技术和视觉矩阵继续记录在对齐文档。
