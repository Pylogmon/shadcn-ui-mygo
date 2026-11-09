# Avatar

`display.Avatar` 接受 MyGo 已解码的 `ui.ImageSource`（Bitmap/SVG），默认直径32 DIP；`AvatarSmall` 为24、`AvatarLarge` 为40。`Size` 可显式覆盖直径。

```go
display.Avatar(c, display.AvatarProps{
    Image: photo, Fallback: "SD", Label: "Sofia Davis",
    Scale: display.AvatarLarge,
    Badge: &display.AvatarBadgeProps{Label: "Online"},
})

display.AvatarGroup(c, display.AvatarGroupProps{
    Label: "Project team", Max: 3, Total: 8,
    Items: members, // each AvatarProps.Key is a stable member ID
})
```

设置 `Loading` / `Error` 表达应用自己的异步加载状态，经 `Window.Update` 在UI线程发布解码图像。组件没有隐式网络请求。无图像或解码尺寸无效时显示首字母或 `FallbackContent`；`OnLoadingStatusChange` 仅在 idle/loading/loaded/error 状态变化时通知。

`FallbackDelay` 从首次挂载计时，仅经过一次；后续图像变更不会重启已完成的延迟。这对应上游 [Base UI Avatar Fallback.delay](https://base-ui.com/react/components/avatar) 的挂载语义。延迟中不会提前画出回退底色。

头像组默认重叠8 DIP，`Max` 限制当前显示数量，`Total` 包括尚未传入的成员；剩余数量显示为计数头像，可用 `CountLabel` 本地化或 `CountContent` 自定义。成员重排时应提供稳定 `Key`，保留延迟等本地状态。混合成员尺寸时，可指定组的 `Size` / `Scale` 统一直径和计数尺寸。

头像根节点提供原生image角色和姓名，角标说明合并到头像Description；装饰性头像设置 `Decorative`。角标是非交互状态，不应放入按钮或编辑器。
