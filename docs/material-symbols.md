# Material Symbols

图标模块使用 Google 官方 Material Symbols SVG，固定提交为
`10ecc14790612ca8fd164e3f4ffe9c16fc50d5b7`。4157 个名称均提供 Outlined、Rounded、Sharp
三种风格及 FILL0/1，共24942份静态SVG；配置为weight400、grade0、optical size24。
所有资源离线内置，不依赖网络、图标字体或系统字体安装。

## 按需引入

导入所需风格包，直接调用图标函数。函数按Google名称转换为PascalCase；数字开头加`Icon`，例如`Icon3dRotation`。

```go
import (
    "github.com/pylogmon/shadcn-ui-mygo/pkg/control"
    "github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
    "github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
    "github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/rounded"
)

control.Button(c, "Save", control.ButtonProps{Icon: outlined.Save()})
material.Icon(c, rounded.Favorite(material.Filled), material.Props{Size: 20})
```

每个函数只引用自己的两种填充资源，无全局图标表或逐图标初始化。
Go链接器可移除未调用函数及其SVG；无需手工下载或维护图标子集。
不要在按需应用中额外导入`material/all`，因为它明确保留完整目录。

## 全量引入

图标名称来自配置、业务数据或用户选择时，显式使用独立的`all`包。

```go
import (
    "github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
    "github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/all"
)

svg, err := all.Lookup("settings", material.Variant{
    Style: material.Rounded,
    Fill: material.Filled,
})
if err != nil {
    // 由应用处理未知名称或非法配置。
    return
}
material.Icon(c, svg, material.Props{Size: 24, Label: "Settings"})
```

默认变体为Outlined/Unfilled。`all.Names()`返回排序后的独立名称列表；名称精确匹配Google的
snake_case。未知名称返回`ErrNotFound`，非法风格或填充返回`ErrVariant`，不会静默替换。
常量名称可简写为`all.Must("home")`；该入口在参数错误时panic。
SVG只在首次请求时解析并缓存，完整目录不会预先解析全部图标。

## 尺寸、颜色与语义

`material.Icon`默认24DIP，可显式设置任意有限正尺寸；图标布局和字形尺寸一致，禁止压缩并居中。
颜色继承当前文字色，也可对返回元素调用`.TextColor(tokens.Primary)`。
通过`ButtonProps.Icon`传入SVG时，按钮统一管理尺寸及颜色。
无`Label`时图标是装饰，有`Label`时提供原生Image名称；图标本身不拦截父按钮点击。

这批静态资源包含所有名称和上述六种变体。显示尺寸缩放不等于重新选择字体光学轴；
其他weight/grade/optical-size配置及连续可变字体轴动画尚未实现。

## 示例与验证

```sh
go run ./examples/material-icons
make icons-app
```

示例提供按需图标、真实图标按钮、完整目录名称搜索、三风格、填充与明暗主题切换。
`-snapshot path.png`和`-dark`支持静态导出。

独立固定基线的四个真实最小应用使用相同构建参数：原始Home SVG、额外空导入outlined、
`outlined.Home()`均为4271650字节；`all.Must("home")`为15637314字节。
四者的实际SVG像素及PNG哈希相同。前两种按需探针没有完整ZIP或未用SVG字面量，
完整目录只出现在显式全量应用中。这是macOS/arm64该构建探针的结果，不是固定体积承诺。
全量资源ZIP为11186521字节，完整原文及来源哈希可追溯。

全部24942份SVG通过公开MyGo解析，另有11项正式尺寸/颜色/填充/按钮/名称/并发检查和
13项独立公开API审查；独立完整SDK1898项及vet通过。明暗截图与macOS原生示例已查看。
完整平台/字体/AT验收仍继续跟踪。

## 更新资源

来源许可、原始README、完整Git blob/SHA索引及生成物哈希在
[`sources`](../pkg/icons/material/sources/manifest.json)。资源使用Apache-2.0许可，
`make icons-app`会把原始许可复制到应用bundle。

生成器不主动联网。从本地已具有所需Git对象的官方仓库离线再生成：

```sh
go run ./pkg/icons/material/internal/generate -repo /path/to/material-design-icons
```

生成器固定提交，验证原始Git blob、XML和24px尺寸，再按稳定顺序生成按需函数、
固定时间ZIP及压缩来源索引。重复离线生成已逐字节核对。
