# Input OTP

InputOTP 使用一个真实原生编辑器提供选择、键盘、剪贴板和 IME，再在外层绘制非交互槽位。返回值与可选 Handle 都指向该编辑器，可继续用于 Field、标签、焦点和原生编辑器 API。

```go
control.Field(c, control.FieldProps{
    Label: "Verification code", Description: "Enter your six-digit code.",
}, func() ui.Element {
    return control.InputOTP(c, &code, control.InputOTPProps{
        Length: 6, Groups: []int{3, 3}, DigitsOnly: true,
        OnComplete: func(value string) { verifyCode(value) },
    })
})
```

默认不限制字符类型，与 input-otp 1.4.2 保持一致。数字验证码应显式设置 DigitsOnly；旧 Alphanumeric 兼容 ASCII 数字与 Unicode 字母。AllowRune 覆盖预设，Pattern 验证经过粘贴转换、清理和长度截取后的候选值；需要整串匹配时为正则加首尾锚点。长度与选择使用原生 rune 索引，区别于网页的 UTF-16 code-unit 索引。

Groups 控制分组，未设置 Length 时可由组大小推导总长。默认槽位 36×36 DIP，组间 Gap 为 8 DIP，可覆盖分隔符、SlotContent、宽高和间距。自定义槽位收到实际字符、活动选择、caret、错误与组首末状态；视觉槽位不创建额外键盘停靠点或可访问目标。

OnChange 报告接受的原生文字变化，应用直接替换值不报告；OnComplete 仅在变化前不满、变化后满长时触发。初始满值及满值替换不重复完成；应用把不满值改为满值也符合上游完成契约。回调在当前父帧执行，回调自己替换的受控值会在下一帧同步。

```go
control.InputOTP(c, &invite, control.InputOTPProps{
    Groups: []int{2, 4}, PasteTransform: strings.ToUpper,
})
```

PasteTransform 在原生粘贴后应用，并保留普通非空粘贴的原生撤销。上游在转换粘贴后将选择移动到末尾槽位，粘贴到中间也如此；这与普通键入后继续选择下一槽的行为不同。空剪贴板被转换为非空文字时，值、选择和回调仍正常，但当前 MyGo 不创建原生撤销事务，也没有公开的事务替换 API，该特殊变换无法保证一次 Undo 回到原值。

读取/拖动选择、全选、覆盖、RTL、Backspace、复制粘贴、撤销/重做与 IME 使用同一个原生编辑器。选区内开始组字再取消或先清标记后提交，不会向调用方发布被临时删除的底文。ReadOnly 保留选择与复制，Disabled 阻止编辑。

原生构建预算最多 4096 个槽位；负数、溢出、超预算或无法表示为有限几何的参数标记 Invalid 并使用有限的默认布局，避免巨量分配和无穷 AX 范围。该预算是原生适配限制，未作为网页源行为宣称。

真实输入法候选窗与辅助技术仍待平台验收。当前公开 API 无法把 IME 候选窗精确定位到每个自绘槽位；已验证原生客户端与整控件范围的 caret 几何，未把逐格候选定位算作完成。DOM 密码管理器角标、no-script CSS 和浏览器专属粘贴适配不映射为桌面 UI 功能。

行为固定到 input-otp 1.4.2 提交 `81ccdb48c010d800b24942aa231909f0c971b1ca`。源码、MIT 许可、SHA-256 与精确行号归档在 [testdata](../pkg/control/testdata/input-otp/manifest.json)，参见 [上游基线](upstream.md) 与 [第三方声明](../THIRD_PARTY_NOTICES.md)。
