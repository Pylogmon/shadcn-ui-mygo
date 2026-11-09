# 加载指示

Spinner 使用固定上游的 Lucide Loader2 轮廓，16 DIP 默认大小，一秒匀速旋转一周。图形继承实际当前文字颜色，因此可放入按钮、Badge、InputGroup 或 Empty；在返回 Element 上设置 TextColor/Size 同样生效。

```go
display.Spinner(c, display.SpinnerProps{
    Size: 24, Label: "Refreshing workspace",
    Description: "Your changes are being synchronized",
})

display.Spinner(c, display.SpinnerProps{Decorative: true})
```

默认向原生可访问性公开有名称的未知进度，保留原有 Native Go 语义。已有父控件播报加载状态时可设置 Decorative，避免重复目标。Spinner 不增加键盘停靠点。Static 禁用旋转；桌面 ReduceMotion 和 instant frame 也始终显示同一个静止图形，重绘其他控件时不会偷偷改变角度。

Skeleton 默认 16 DIP 高且占满可用宽度，颜色为 Muted；Circle 用高度作为直径。Width、Height 为零使用默认；非法非有限尺寸回退为有限布局并标记 Invalid。返回值可继续设置百分比宽度、圆角、背景或布局。

```go
display.Skeleton(c, display.SkeletonProps{Circle: true, Height: 40})
display.Skeleton(c, display.SkeletonProps{}).WidthPercent(75)
```

脉冲按固定 Tailwind 默认的两秒周期、cubic-bezier(.4,0,.6,1) 在 opacity 1 与 .5 之间变化。圆角遵循本次基线 globals.css 的 radius×.8；没有采用旧版本 radius−2px 的规则。Static、减少动画或 instant frame 均跳过循环，恢复普通偏好后重新参与运动；移除控件后不保持空闲重绘。

Busy Button 与公开 Spinner 共用同一图形与动画；真实按钮继续承担 Busy、Disabled 和原生激活。图标字节、许可证和哈希在 `internal/indicator`，macOS 包携带 Lucide 许可证。源映射见 [upstream.md](upstream.md)。

像素、匀速周期、动态偏好、原生未知进度、尺寸/颜色继承及移除后的空闲行为已有检查。真实屏幕阅读器与完整平台/字体/视觉矩阵仍需验收。
