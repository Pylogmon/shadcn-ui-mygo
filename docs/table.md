# Table 组合与原生虚拟表格

`TableWith` 对应固定源的静态表格组合：标题在底部，Header、Body、Footer共享列宽；行高随原生内容，保持单行内容并允许水平滚动。页脚支持跨列，自定义单元格构建真实原生控件。

```go
display.TableWith(c, display.TableOptions{
    Label: "Invoices",
    Columns: []display.TableColumn{
        {Key: "invoice", Width: 120},
        {Key: "status"},
        {Key: "amount", Align: ui.End},
    },
},
    display.TableHeader(display.TableRow(display.TableRowProps{Key: "head"},
        display.TableHead("Invoice"),
        display.TableHead("Status"),
        display.TableHead("Amount"),
    )),
    display.TableBody(display.TableRow(display.TableRowProps{Key: "inv-1"},
        display.TableRowHead("INV001"),
        display.TableCell("Paid"),
        display.TableCell("$250.00"),
    )),
    display.TableFooter(display.TableRow(display.TableRowProps{Key: "total"},
        display.TableCell("Total", display.TableCellProps{ColSpan: 2}),
        display.TableCell("$250.00"),
    )),
    display.TableCaption("Recent invoices"),
)
```

Header、Body、Footer、Row、Head、RowHead、Cell、Caption构造的是描述符；TableWith一次性构建实际UI，使各区段共享列尺寸。Key保持单元格内编辑器、选区和焦点与记录绑定。Content回调替换Text，直接构建Input、Select、Checkbox、Badge或独立动作；回调不为测量而重复执行。动态列也应提供稳定Key。

Row与Cell的Style回调接收当前构建中真实的ui.Element，可继续设置颜色、字号、行高或原生描述等；它在默认样式后、子内容前执行一次。不要保留这个Element。修改较大的布局内边距时，应同时提供合适的列尺寸约束；共享轨道的自然测量以源默认单元格间距为基础。

默认按实际原生内容测量自然宽度。Columns可指定首选Width、MinWidth、MaxWidth和对齐，Cell可覆盖Align或指定ColSpan。内容不换行，较宽内容通过真实横向viewport查看；调用方可持有ScrollState保持位置。这里采用共享原生轨道，不实现完整HTML intrinsic/rowspan排版算法。

Selected只标记应用已选择的行；Expanded为true时应用源中后代aria-expanded=true的背景。Checkbox为true时移除该cell的物理右内边距，RTL仍是right。这两项需要调用方标记，因为MyGo没有公开后代角色/属性查询；组件不会读取私有UI树，也不额外注入选择、展开、排序或过滤行为。没有数据时不自动插入提示，可自行提供跨全部列的空态内容。

旧 `Table(c, TableProps)` 保持MyGo真实虚拟化入口，适合大数据。State接受调用方ListState（含Key、原生列尺寸与选择），RowCount配合CellContent按可见行构建，无需分配全部Rows；Height限制viewport。它的原生固定表头、选择和列交互保留运行时样式，不能将其视作静态CSS表格的完整视觉对照。更高阶过滤、排序和分页使用独立DataTable。

TableWith提供真实Table/Row/Cell/ColumnHeader/RowHeader、索引与跨列语义；单元格中的控件各自保留键盘和辅助技术行为。完整平台、屏幕阅读器、字体与视觉矩阵继续记录在 [parity.md](parity.md)。
