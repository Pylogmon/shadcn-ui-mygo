# 原生 CodeBlock

CodeBlock 使用一个真正的 MyGo 只读代码编辑器处理选择、复制、查找、滚动和语法着色。原 `CodeBlock` 入口保留默认 200 DIP viewport；`CodeBlockWith` 默认按内容高度，并提供复合槽位。

```go
var codeState display.CodeBlockState

display.CodeBlockWith(c, display.CodeBlockProps{
    Title: "main.go", Language: "go", Code: source,
    LineNumbers: true, StartLine: 10, Wrap: true,
    MaxHeight: 300, State: &codeState,
    Lines: []display.CodeBlockLine{
        {Index: 3, Variant: display.CodeLineAdded},
    },
    OnCopy: func(value string) { copied = true },
}, display.CodeBlockSlots{
    Footer: func(child *ui.Context) { ui.Text(child, "Read-only example") },
})
```

支持 Default/Destructive、DefaultSize/Small、固定/内容高度和上下限、隐藏 Header/Copy、折行、起始行号以及 Added/Removed/Highlighted 行。行号与 diff 符号在装饰性 gutter 中，行背景覆盖整个代码内容宽度；原生选择和复制只包含实际源文字。

自定义 Header/Actions/Footer 在对应原生容器构建。复合 helpers 包括 CodeBlockHeader/Title/Actions/Footer/Content/CopyButton；CodeBlockAction 可为自定义原生按钮应用浮动操作背景与环。没有标题时操作浮在逻辑末端，并随 hover/focus 显示。代码本身保持 LTR，RTL 只镜像外围操作布局。

切换 Language、主题、Size、Wrap 或流式追加代码不会重建 Buffer。受控文字按 UTF-8 边界做局部 Edit，原生锚点继续跟踪；调用方可持有 CodeBlockState 保留 owner 移除期间的 editor。行号/折行的默认原生编辑器快捷键不能覆盖受控展示，其他原生和应用 bindings 保留。

CopyButton 等待真正的 WriteClipboardThen 成功后才显示 Copied、播报并调用 OnCopy。错误调用 OnError（整块属性为 OnCopyError）；不会显示伪成功。反馈持续两秒，重复复制不延长已显示反馈，空闲时不持续重绘。异步完成只保留数据，回调与播报在存活的当前构建执行；卸载后不回放旧反馈。

内容选择与查找仍用原生编辑器。语法使用 MyGo 已提供的 lexer、别名及主题；未知语言退为纯文本。Shiki 全语言、ANSI 渲染、完整字体/平台/辅助技术矩阵仍是未完成验收，不能用现有颜色测试推定通过。

33 项正式检查覆盖像素/变体、原生 UTF-8 选择与 clipboard、异步失败/成功/重入、真实折行/横向 reveal、十万行文件、稳定 identity、RTL、源码无行号复制及显示快捷键隔离。Copy/Check 使用固定 Lucide 0.474.0 原始 SVG，来源/哈希在 `pkg/display/assets`，共享许可证随应用打包。
