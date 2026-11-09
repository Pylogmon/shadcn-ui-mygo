# Toast 与 Sonner 通知

`Toaster` 使用固定 shadcn/Sonner 的通知样式：默认中性色与 Popover 语义色，错误通知通过图标表达状态；`RichColors` 才启用状态配色。Vega 圆角跟随当前 radius×1.8，默认右下、宽356 DIP、间隔14 DIP，桌面边距24 DIP、窄窗口边距16 DIP。关闭按钮默认隐藏。

```go
var notifications overlay.ToasterState

func view(c *ui.Context) {
    theme.Apply(c)
    if control.Button(c, "Save", control.ButtonProps{}).Clicked() {
        notifications.Show(ui.Toast{
            Title: "Workspace saved", Type: "success", Action: "Undo",
        }, overlay.ToastOptions{
            Action: func(event *overlay.ToastActionEvent) {
                // Apply the application's undo operation here.
                // event.PreventDefault() keeps the notification open.
            },
        })
    }
    overlay.ToasterWith(c, overlay.ToasterProps{
        State: &notifications, Position: overlay.ToastBottomRight,
        CloseButton: true,
    })
}
```

每个窗口只构建一个 viewport。显式 `ToasterState` 管理完整应用队列，`Show` 返回ID；同ID显示或 `Update` 更新原项，不重排或创建第二份。`Dismiss(id)` 关闭指定项，空ID关闭该状态中的全部项；`Toasts` 返回快照。状态在UI线程修改，异步结果通过窗口的Update调度；状态不保留短生命周期的Context。

loading 通知持续到应用完成或显式关闭；可用同一ID更新为 success/error 后恢复正常寿命。普通通知默认4秒，负Timeout持续显示。可见数量、所有六个物理位置、独立位置覆盖、边距、宽度、间隔和默认时长由ToasterProps配置。位置的left/right不随RTL镜像；通知内部的图标、文字、动作和关闭按钮按逻辑方向排列。

`Expand` 控制常驻展开布局，不暂停寿命。用户交互展开、悬停、按住通知和真实 `Context.WindowHidden()` 暂停计时；窗口仅失焦时继续计时。鼠标离开结束交互展开后恢复计时，即使动作按钮仍有焦点，行为对照固定Sonner。默认Alt+T进入通知，Escape收起堆叠；自定义Hotkey同时发布平台原生KeyShortcuts语义。需要精确回到原生触发器时传入 `ReturnFocus *ui.Handle`。

应用切换页面时，ReturnFocus应指向当前仍挂载且可用的控件。MyGo的Handle.Focus会等待尚未出现的目标，持续传递旧页面Handle可能在返回该页时重放请求；Gallery使用当前预览标签作为跨页fallback，完成按钮禁用后改用Start upload。

Action通过ToastOptions.Action接收可调用PreventDefault的事件。Cancel遵循Dismissible，执行后关闭；程序调用Dismiss可以关闭持久或不可手动关闭的通知。OnDismiss、OnAutoClose与兼容的ui.Toast.OnClose区分手动关闭、到期和所有关闭原因。ToastOptions还可覆盖图标、位置、关闭按钮、状态色及内容。

普通Content替换标题/描述列；`Unstyled: true`时Content接管整个通知正文，不加默认背景、边框、内边距、图标或内部按钮。应用可组合真实原生编辑器、媒体与动作，自己调用Dismiss；队列身份、布局、语义、计时和手势仍由Toaster管理。卸载并重新挂载viewport后，存活通知按固定源开始新寿命；已交付关闭的项不重放，未交付的显式关闭仍执行一次。

滑动采用MyGo原生捕获、轴锁、默认物理方向和反向阻尼，达到45 DIP或速度大于0.11 DIP/ms关闭；子按钮、选中文字、右键、loading与不可手动关闭项不会误触。普通进入为400ms CSS ease、整项高度位移；普通退出按源的400ms曲线绘制，但源在200ms物理移除。滑动退出为200ms ease-out，取消滑动400ms归位；减少动画跳到静态结果，原生定时器的1ms精度可能让200ms移除在下一毫秒唤醒。

旧 `Toast(c, ui.Toast)` 和 `Toaster(c)` 保留MyGo窗口队列，直接调用 `c.AddToast` 的通知也能显示；这条兼容路径受运行时最新三项、后台继续计时、原生Action必关闭及Escape关闭等语义约束。这里的“原生队列”指MyGo窗口内队列，不是系统通知中心。需要Sonner完整队列语义时使用显式ToasterState；无需读取或修改运行时私有状态。

两条队列共用位置堆叠与可见数量。跨构建的到达顺序由实际发现次序统一；同一次构建首次同时出现时，旧队列公开API没有入队时间戳，采用先发现native再发现应用项的确定顺序。原生无live/relevant/atomic完整设置器，当前用Region/List/ListItem及显式polite Announce；不能把Go语义检查当作真实屏幕阅读器验收。

来源、原始源码、图标与许可证见 [upstream.md](upstream.md) 和 [第三方声明](../THIRD_PARTY_NOTICES.md)。真实平台、辅助技术与完整视觉矩阵按 [对齐记录](parity.md) 继续验收。
