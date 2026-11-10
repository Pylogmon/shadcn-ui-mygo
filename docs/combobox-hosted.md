# 独立窗口 Combobox

`desktop.Combobox[T]` 与 `desktop.NativeCombobox[T]` 把触发器和可编辑弹层分别构建在父窗口、独立 MyGo Flyout 的 Context 中。弹层可超出父窗口，包含真实搜索编辑器、空状态和普通单列选项集合；单选和多选共享 `control.ComboboxState[T]`、Adapter、Filter 及可取消事件。

普通版本使用统一 shadcn/Vega 外观和退出动画。Native 版本使用 plain Flyout、自身平台 Theme、`ui.TextInput`/`ui.Button` 和平台颜色、字号、间距。MyGo 没有公开的系统可编辑 ComboBox 控件；Native 版本属于平台主题的受约束 Go 内容，不是系统 ComboBox，也没有用静态系统菜单代替搜索。

```go
// model、host 和 nextInput 是应用持有的长期状态，host 绑定父窗口。
// 在父窗口每次构建中调用 BeginFrame / EndFrame。
provider := func(window *ui.Context) control.ComboboxRootProps[Country] {
    return control.ComboboxRootProps[Country]{
        Items: model.Countries,
        Adapter: control.ComboboxAdapter[Country]{
            Key: func(v Country) string { return v.Code },
            Label: func(v Country) string { return v.Name },
        },
        AutoHighlight: control.ComboboxHighlightOnInput,
        Locale: model.Locale,
        Multiple: model.Multiple,
        OnValuesChange: func(values []Country, event *control.ComboboxChange) {
            if len(values) > model.SelectionLimit {
                event.Cancel()
            }
        },
    }
}

trigger, err := desktop.Combobox(c, host, "country", &model.Selection,
    desktop.ComboboxProps{
        Label: "Country",
        Placeholder: "Choose a country",
        SearchPlaceholder: "Search countries…",
        EmptyText: "No countries found.",
        OnTab: func(backward bool) { nextInput.Focus(c) },
    }, provider)
// 用相同参数调用 desktop.NativeCombobox 即选择平台主题版本。
```

Provider 是与 Popover 子 View 类似的显式窗口构建工厂。只捕获长期应用模型，在传入的当前 `window` 上求主题、方向或构建处理器；不要捕获父 Context、Element、Scope 或父帧的 RootProps。组件分别在父、子当前帧调用 provider，返回的适配器、Filter 和事件处理器只在该次 Root 构建同步求值，不保存在 Host 的持久模型中。跨窗口的 Key/选值含义必须保持一致。

搜索提交在子窗口直接调用当前 Filter，不等待父窗口重建，也不会把旧结果冒充新 query 的匹配结果。预编辑文本可存在真实编辑器中；确认组合输入后才过滤和发布 query，候选确认 Enter 不同时选择建议。`OnQueryChange`、`OnValueChange`、`OnValuesChange`、`OnHighlighted` 和内部关闭的 `OnOpenChange` 都在当前子构建执行；事件 `Cancel()` 同步保留原值。父触发器继续负责关闭态 typeahead；展开期间不重复筛选、高亮、清 query 或发布打开完成事件。

接受的应用值当场写入 `ComboboxState[T]`，父触发器在下一次构建报告一次 `Changed()`。`OnOpenChangeComplete(true)` 在子窗口进入完成时执行；关闭后由当前父构建执行一次 `OnOpenChangeComplete(false)` 和可取消 query 清理。单选关闭回到触发器，多选继续编辑；退出期间同一 Surface 保持存在，完成后才清理搜索。`SearchHandle` 可绑定子窗口真实编辑器，只能在该窗口的当前 Context 中解析。

标准 MyGo plain Flyout 提供 `DismissInterceptor`：外点关闭先排队一个纯值请求，当前父构建用最新 `OnOpenChange` 决定接受或 Cancel。Cancel 保留同一个子窗口、编辑器、IME 和选区；接受后沿用 Host 的退出与回焦。若自定义 Backend 不支持该公开拦截能力，原生关闭已经发生后才能处理通知；Cancel 可以重新打开，但不能恢复已销毁窗口的 IME/光标。这是 Backend 能力边界。

MyGo 没有公开的跨窗口 Tab 顺序操作。提供 `OnTab(backward)` 时，组件在接受关闭、子窗口释放后的当前父构建调用它；应用用相邻控件的公开 Handle 选定前后目标。未提供时回到触发器。取消关闭时不会调用 OnTab。

父构建会比较 Items/Groups 的公开描述、选择相等性、Locale、Disabled、ReadOnly、查询和内容配置，刷新同一子窗口。只有处理器或 Filter 策略改变、描述值不变时，递增 `Popup.Revision`，或显式调用 `Host.Refresh(key)`。数据加载与请求调度由应用负责。这个入口构建普通 Collection；Grid/Virtualized 会明确返回错误，需要保留在已有 `control.ComboboxRoot` 的显式布局中。未提供跨窗口任意槽位或富渲染器协议。

本模块不修改 MyGo 依赖、module cache 或其他 SDK 内部。它复用已交付 Root 的 typed 选择、查询、过滤、IME 和导航，增加 `HostedTrigger`、`HostedContent` 与 Host 关闭桥接。来源仍是 [固定 Combobox 档案](../pkg/control/testdata/shadcn-combobox/README.md) 的 shadcn `8ab27715521cbd4f87ae6886cd989cbb678f6de7`、Base UI 1.6.0；重点对应 `combobox-popup.tsx` 内部搜索和 `base/combobox.tsx` Content 的 6 DIP gap、普通集合、输入组与 Vega 100ms 动画。独立 Context、provider、平台主题及取消关闭协议是 Go/MyGo 适配，不是上游网页 API。

正式证据为 `TestHostedComboboxTypedIndependentWindowWorkflow`：同一条操作流程创建两个独立 `ui.Tester` Context，验证子 Handle 不在父解析、父内无选项、IME 确认、即时自定义 Filter、query/选值 Cancel、对象单选/多选、最新 handler、集合刷新、保留同 Surface 的外部 Cancel、平台自有字体和原生语义、tr/en Locale 切换的即时刷新、显式 Tab 邻接 Handle，以及非 reduced 退出的关闭完成/清理不重复。设置 `SHADCN_COMBOBOX_SNAPSHOTS` 可导出该流程两个子窗口画面。这个正式 native UI 测试使用公开 MyGo 测试驱动和可拦截 Backend，不等于所有真实操作系统、输入法和屏幕阅读器矩阵均已验收。
