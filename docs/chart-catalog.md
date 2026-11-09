# Chart 源码目录与 Tooltip 数据

`go run ./examples/chart-catalog` 提供固定 shadcn 源的70个 Chart 配置、8个文档示例及 Create 的6张图表卡片。Gallery 的 Data & display 页面使用同一份目录；原有图表探索器、Cartesian 与 Polar 独立应用继续保留。固定源与来源哈希见 [目录归档](../pkg/display/testdata/shadcn-chart/catalogue/manifest.json)，数据与运行方式见 [示例说明](../examples/chart-catalog/README.md)。

高层 `Chart` 的 `Labels`、`Series.Values` 继续决定分类和绘图数值。可选 `ChartProps.Data` 按相同分类索引提供原始行，供 Tooltip formatter 读取日期、分组或其它字段。某个系列的非 nil `ChartSeries.Data` 覆盖该系列使用的原始行，包括显式空切片；适用于各自持有数据集的多环 Pie。它不替代该系列的 `Values`，也不会从原始行隐式推导数值。

`ChartSeries.DataKey` 保留原始字段名或数值键，nil 时回退到系列名称；`TooltipName` 可独立覆盖 payload 的 Name，nil 时先使用 DataKey，再回退到系列名称。这样可分别持有可读的系列名称、数据字段和配置查找键。`ChartTooltipData` 保留已有的 Label、Items、FormatValue，并增加逻辑分类 `Index` 与 `Payload []ChartPayloadItem`。Payload 提供 Name、DataKey、当前原始 Value、颜色及原始行，可直接交给现有 `ChartTooltipContent`。

```go
func activityChart(c *ui.Context, selected *int) {
    rows := []map[string]any{
        {"date": "2024-07-16", "running": 380.0, "swimming": 420.0},
        {"date": "2024-07-17", "running": 520.0, "swimming": 120.0},
    }
    props := display.ChartProps{
        Kind: display.BarChart, Height: 260, Selected: selected,
        Labels: []string{"2024-07-16", "2024-07-17"}, Data: rows,
        TooltipFollowInspection: true,
        Series: []display.ChartSeries{
            {Name: "Running", DataKey: "running", TooltipName: "activity",
                Values: []float64{380, 520}},
        },
        TooltipContent: func(d display.ChartTooltipData) {
            display.ChartTooltipContent(c, display.ChartTooltipContentProps{
                Active: true, Payload: d.Payload,
                Label: &display.ChartContentLabel{Text: d.Label},
            })
        },
    }
    display.ChartContainer(c, display.ChartContainerProps{
        Config: display.ChartConfig{"activity": {Label: "Running"}},
    }, func(c *ui.Context) {
        ui.Column(c).FillWidth().Children(func() { display.Chart(c, props) })
    }).AspectRatio(0)
}
```

此片段使用公开的 `ui` 与 `display` 包。`ChartContainer` 安装的配置作用域只在其同步 Children 构建期间有效；Tooltip、Label 和 Legend 的内容回调同样在当前构建中执行，不能保存 Context、Element 或帧对象。自由绘图和内容组件的独立用法见 [Chart 组合](chart-composition.md)。

原始行须为无环数据。快照递归复制 `map[string]any`、`[]any`、`[]map[string]any`，并复制 `[]string`、`[]float64`、`[]int`；这些值在 formatter 中的修改不会写回调用方数据。其它应用对象保持不透明并借用其原值，不承诺任意 Go 对象的深复制。Payload 的 Value 与已有 Items 始终使用当前 `Series.Values`，不会变成动画中间值。无效数值仍按已有规则过滤；`ChartTooltipData.Index` 是分类索引，而 `ChartTooltipFormatData.Index` 是过滤后 payload 的行索引。高级源例由 formatter 的 `Payload` 读取 running 与 swimming 的原始字段计算合计。

`TooltipFollowInspection: true` 启用源配置的定位：跟随检查类别、扇区或指针，使用实际原生内容尺寸、10DIP间距与视口边界翻转。默认检查位置的非类别坐标包含图表外框与坐标轴预算。零值保留原有左上角固定位置；`HideTooltip` 仍可单独隐藏视觉内容而保留键盘检查。`ChartTooltipContent` 自身只负责内容与样式，独立使用时仍由调用方安排位置和交互。

`ChartProps.Cartesian.ReversedCategory` 显式反转类别轴位置，保留原始数据索引与柱组内系列顺序。指针和方向键沿实际位置检查，Home / End 仍指向逻辑首末分类。零值保留原来的物理类别顺序；文字方向与类别反转是分别配置的能力。文档里的 Arabic / Hebrew 示例同时应用 RTL 文字方向和反转类别，English 则使用正常顺序，翻译轴、Tooltip 与 Legend 的标签。

目录可按家族筛选，或搜索标题与源码 ID。各项操作作用于实际数据和图形：

- Area Interactive 的 Date range 在固定数据中筛选日期；Bar、Line 与文档 Demo 的 Desktop / Mobile 总数按钮切换真实系列。文档 Demo 保留独立的30条 April 数据。
- Pie Interactive 的 Active month 更新中心数值与活动扇区；瞬时指针或键盘检查不改变该月份选择。
- 9个 Tooltip 配置展示 line / 无指示器、隐藏 / 自定义 / 日期格式标题、formatter 单位、图标及原始行合计。第二项的初始检查只初始化一次，Escape 清除后不会在重建时重新选中。
- 8个文档项包括逐步组合 marks、grid、axis、tooltip、legend，以及交互 Demo、RTL 和 Tooltip anatomy。Anatomy 使用归档的箭头路径与现有内容组件。
- Create 显示 Area、Bar、Line、Pie、Radar、Radial 六张卡片，使用其各自的源数据、标题和配置，并继承宿主当前的配色与外观。

```sh
make catalog-app # build/Shadcn Catalogue.app
go run ./examples/chart-catalog -example chart-tooltip-advanced
go run ./examples/chart-catalog -collection docs -example docs-chart-rtl
go run ./examples/chart-catalog -collection create
go run ./examples/chart-catalog -collection create -export-height 4000 -export /tmp/create.png
```

`internal/chartcatalog` 是本模块内 Gallery 与三个目录应用共用的示例构建器。应用须从此模块运行，不能把含 internal 导入的整个 main.go 当作外部模块示例直接复制；公开 `display` API 不依赖该内部目录。需要独立组合入口时使用 [examples/chart-composition](../examples/chart-composition/README.md)。Cartesian、Polar 的能力分别见 [对应文档](chart-cartesian.md)与 [Polar 文档](chart-polar.md)。

字体度量、文字换行、图标与边缘栅格化遵循原生平台。Tooltip 图标示例按需使用 Material Symbols Footprint / Waves。84项 inventory 是一次配置构建覆盖，不表示84项独立视觉或真实辅助技术验收；固定源未使用的 Brush、ReferenceLine 或其它 Recharts API 不由此目录隐含提供。
