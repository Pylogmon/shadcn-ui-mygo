# Typography 与 Kbd

固定上游的 Typography 是一组排版示例，并不默认安装全局样式。`Typography` 保持单块组合，`TypographyProse` 为本模块的段落、标题、列表与表格管理示例中的文档间距。

```go
display.TypographyProse(c, func(child *ui.Context) {
    display.Typography(child, "A clear interface", display.H2)
    display.TypographyRich(child, display.Paragraph, func(inline *ui.Context) {
        ui.Text(inline, "Build with ")
        ui.RichText(inline, display.TypographyCodeSpan(inline, "ui.Context"))
        ui.Text(inline, " and read the ")
        display.TypographyLink(inline, "guide", "https://example.com/guide")
        ui.Text(inline, ".")
    }).Selectable()
})
```

标题公开原生 heading rank；H2 有实际底部分隔线。段落、Lead、Small、Muted、引用和代码采用固定字体大小、行距、字重与 tracking；引用边框跟随逻辑起始侧。原生字体的真实字形盒可能大于 CSS leading-none 的数值，不用固定高度裁掉字形。

TypographyRich 的回调只构建原生 inline Text/Link/RichText；布局控件应放在外面。TypographyCodeSpan 有真正参与行内布局的背景与 padding。Selectable 是返回 Element 的原生能力；链接仍保留实际 URL/Router 和 Enter 激活。

TypographyList 接受带稳定 Key 的 Text 或 Content，支持无序与有序原生 list/listitem、逻辑侧标记与动态重排。TypographyTable 是静态语义表格，包含 column header、row、cell、交替背景和可选 MinWidth 横向 viewport。列为等宽；它不实现完整 HTML intrinsic/colspan 算法。需要大数据虚拟化、排序和选择时使用独立 Table/DataTable 模块。

Prose 只管理本模块 helper 的流式间距。任意原始 UI sibling、嵌套业务组件的外边距由应用安排；不存在隐含的全局 CSS。

```go
display.KbdShortcut(c, ui.Cmd|ui.Shift, "K")
display.KbdGroup(c, func(child *ui.Context) {
    display.Kbd(child, "Ctrl")
    display.KbdWith(child, "K", display.KbdProps{InTooltip: true})
})
```

Kbd 为 20 DIP 高、20 DIP 最小宽的非交互键帽，无默认边框。分组间距 4 DIP，可放图标或非交互 Content；Label/Description 命名其真实原生文字目标。键帽不注册快捷键，也不增加 Tab 停靠点，不阻挡外围原生按钮的点击。

快捷键按实际平台显示修饰符。InTooltip 采用固定源的逆色样式。字体使用当前 native Theme.Font 的正文字体，避免继承局部 monospace；返回 Element 的 Font 仍可显式覆盖。应用选择字体时应同时设置其原生 scoped Theme.Font，而非仅修改一个父 Element。

21 项聚焦检查覆盖排版像素、间距/作用域、原生链接与选择、列表/表格语义、RTL、稳定身份、键帽组合与原生按钮激活、平台修饰符及字体覆写。完整字体、辅助技术与跨平台视觉矩阵继续验收。固定来源见 [upstream.md](upstream.md)。
