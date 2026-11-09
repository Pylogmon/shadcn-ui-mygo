# Flyout 菜单动画

`desktop.Menu`、`DropdownMenu`、`ContextMenu`和`Menubar`使用独立原生Flyout，并对实际菜单面板应用Vega的100ms CSS ease、95%缩放和透明度变化。进入时按实际面向锚点的方向移动8DIP；退出目标不带平移。进行中的打开/关闭反转从当前可见姿态继续。系统减少动态效果、Instant及隐藏窗口立即收敛到目标。

连续反转是原生交互适配：中途关闭时，尚未完成的进入位移会从当前值连续归零。它遵守退出目标不平移的样式，不宣称任意中断帧与CSS重启动画逐像素相同。

此处沿用项目既有的固定`bases/base`组件基线。Base MenubarContent通过DropdownMenuContent组合，继承其关闭动画；同一shadcn提交的Radix直接MenubarContent使用另一组class。实现按组合后的实际class核对，不能只读单独的`.cn-menubar-content`规则。

子菜单按完整父路径区分身份，同层替换在同一列重叠进退。关闭期间保留面板和测量空间，输入、焦点与无障碍内容立即退出，动画完成后释放。重新打开同一路径接续当前姿态。保留状态只保存路径和数值，不保存Context、Element、旧菜单数据或旧ItemContent回调；每帧从当前Items解析内容，失效路径及时清理。通用Popover使用自己的预设，NativeMenu系列的呈现和动画由系统负责。

可选`GeometrySurface`让自定义Backend报告`Geometry() (anchor, content ui.Rect, known bool)`。两个矩形使用同一屏幕DIP坐标系，分别代表实际传入原生窗口的锚点及观察到的子窗口内容边界。菜单只在这些矩形足以证明实际side时应用方向位移，并按Base UI的虚拟箭头规则计算缩放原点：投影锚点中心，按箭头padding限制交叉轴，主轴可位于面板外的gap处。

缺少几何能力、重叠不能确定side、窗口不可用或坐标尺度不能安全比较时，保留中心缩放与透明度，不把请求的placement当作实际碰撞结果。原生提供器在Windows跨显示器尺度不一致时返回unknown；Linux若环境表明可能使用Wayland也返回unknown。环境信号只用于保守排除，不能证明实际窗口后端。完整混合DPI、碰撞、缩放及Wayland实窗矩阵继续在[验收清单](parity.md)跟踪。

来源位于`pkg/desktop/testdata/shadcn-menu`：固定shadcn原文与Git blob、Base UI 1.6.0/工具0.3.1的退出完成和定位逻辑、tw-animate-css 1.4.0，以及Radix菜单/Presence对比资料。五份官方npm包的SHA-512与固定lockfile一致；原生应用附带对应MIT许可，不引入这些JavaScript或CSS运行时。
