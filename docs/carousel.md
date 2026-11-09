# Carousel

`CarouselContainer`、`CarouselContent`、`CarouselItem`和独立的Previous/Next控件
提供原生多项布局。`Carousel`和`CarouselWith`继续保留原有无间距分页入口。

```go
// selectedSnap和slides由应用持有。
var snapshot display.CarouselSnapshot
display.CarouselContainer(c, &selectedSnap, display.CarouselContainerProps{
    Key: "projects",
    Label: "Projects",
    Count: len(slides),
    Height: 160,
    Align: display.CarouselAlignStart,
}, func(c *ui.Context) {
    display.CarouselContent(c, display.CarouselContentProps{}, func(c *ui.Context) {
        for i, slide := range slides {
            display.CarouselItem(c, i, display.CarouselItemProps{
                Key: slide.ID,
                Basis: display.CarouselFraction(.5),
            }, func(c *ui.Context) {
                ui.Text(c, slide.Title).Padding(24)
            })
        }
    })
    display.CarouselPrevious(c, display.CarouselControlProps{})
    display.CarouselNext(c, display.CarouselControlProps{})
    snapshot = display.CarouselSnapshotOf(c)
})
```

调用方需为外侧按钮保留至少48DIP空间。默认不附加状态文字、圆点或按钮；
可以省略控件，也可通过返回的原生Element和`CarouselControlProps`定制位置、内容、按钮样式。
默认使用按需Material Symbols chevron、Outline/IconSmall和圆形外观。

## 布局与吸附位置

`selectedSnap`与`OnSelect`使用吸附位置索引，不是slide索引。
例如五项在start对齐、trimSnaps下，full、half、third分别产生5、4、3个吸附位置。
`CarouselSnapshot`提供独立的SlideCount、Snaps、Visible、选中位置和前后可滚动状态；
返回的slice是值快照，应用不能通过修改它改变组件内部状态。
Content前后都可读取快照：前置使用最近解析的item/spacing布局并采样当前动画，
Content收集新规格和原生拖动值后更新本帧结果；布局改变会请求后续帧收敛。
它是当前同步构建的值快照。需要在回调外查询或导航时，使用下文的应用持有controller。

- Basis零值填满一屏；`CarouselFraction(.5)`为半屏，`CarouselFixed(180)`为固定DIP。
- 默认Spacing为16DIP，包含补偿的前置padding；指向0的`Spacing`显式取消间距。
  fraction依据实测viewport加前置补偿计算，Basis包含该项的前置padding。
  实际outer尺寸至少为前置Spacing，与原生padding的最小尺寸一致。
- Align默认Center，另有Start、End；多个源示例显式选择Start。
- Contain默认TrimSnaps，另有KeepSnaps和Uncontained；Loop空间不足时按源规则回退为非循环。
- Vertical使用Height作为viewport高度，未指定时为240DIP；RTL使用逻辑方向。
- 正在进入或离开的部分可见项仍保留原生输入和可访问性；完全离屏项排除输入/AX。

Content内必须按索引顺序构建Count个Item，每项一次。
可比较、唯一且自反的Item.Key保留本地状态、原生编辑器和选区；非法或重复key回退到独立位置身份。
改变Count、Basis、Spacing、方向、尺寸或Loop会重新解析几何并校正选中位置，不触发用户OnSelect。

## 原生输入与阶段边界

组合函数在Container的同步children回调内使用；不要替换返回Element的Children来重建作用域。
可嵌套或并列，不保留跨帧Context、Element或应用回调。
编辑器内的方向键继续编辑文字，Carousel根或控件取得焦点时才导航；
垂直方向使用Up/Down，水平RTL反转左右键。这是对网页capture处理的原生适配。

未传Controller和Options的组合入口保留原生分页的250ms EaseOut与兼容拖动阈值；
ContainerProps上的正Duration、DragThreshold属于这一兼容路径，不是Embla同名数值选项。
控制器/Options路径使用下文的源物理运动。减少动态效果及负Props.Duration使运动即时收敛；
控制器隐藏时暂停并重置累积器，恢复后按源规则重新暖机，不追赶隐藏期间的墙钟时间。
Autoplay及插件仍需后续阶段实现；真实平台读屏和完整视觉矩阵尚未验收。

固定shadcn、Embla8.5.2、许可和可重现的源函数几何数据见
[`testdata/shadcn-carousel`](../pkg/display/testdata/shadcn-carousel/README.md)。

## 应用持有的控制器

`NewCarouselController`创建可保存在应用状态中的控制器，通过ContainerProps.Controller绑定。
创建和配置会返回错误，调用方应处理错误后再使用控制器：

```go
api, err := display.NewCarouselController(display.CarouselOptions{
    Loop: display.CarouselValue(true),
    SlidesToScroll: &display.CarouselSlideGroup{Count: 2},
})
if err != nil {
    return err
}
// api和selectedSnap保存在应用状态中；在当前view中传入：
display.CarouselContainer(c, &selectedSnap, display.CarouselContainerProps{
    Count: len(slides),
    Controller: api,
}, buildSlides)
// 可从应用操作或按钮处理逻辑发出指令，true表示立即跳转。
api.ScrollNext(false)
api.ScrollTo(0, true)
```

ScrollPrev/Next/To、Configure和ReInit按调用顺序排队，在下一次owner构建时生效，
通过MyGo公开的弱Services唤醒窗口；Snapshot只返回已提交的值。
这些排队操作和快照查询可从后台调用；Emit、Destroy和监听器在UI线程运行，
异步应用逻辑应通过Window.Update调用后者。
同一controller只能绑定一个存活的Container。离开视图会销毁对应绑定，
Destroy后不会自动复活；再次挂载应创建新controller。

`CarouselOptions`是部分更新：nil字段保留原值，`CarouselValue(false)`和`CarouselValue(0)`
表达明确的false/零值。支持align及应用持有的纯值Aligner、axis/direction、contain、
固定或Auto分组、startIndex、active/loop、四种watch开关、in-view阈值和有序Breakpoints。
Breakpoints针对原生窗口尺寸、方向与外观偏好按顺序合并；相同Key保留规则身份。
ContainerProps.Options还可传入当前构建的选项。无效patch原子拒绝，
不会悄悄应用其中的有效字段。

`Nodes`以公开Handle选择已声明并可测量的track/items，不接受其他组件的节点。
Slides子集保留原始矩形位置及调用方数组顺序。Configure/ReInit复制Handle值，
不会保留调用方的slice或Handle指针。`Hooks`只在当前Container构建中调用；
BeforeDrag/Resize/Slides/Focus返回false可否决对应默认动作。
关闭WatchResize/Slides会保留引擎几何，原生布局仍正常测量与显示。

`SelectedScrollSnap`、`PreviousScrollSnap`、`ScrollSnapList`、`ScrollProgress`、
`SlidesInView`、`SlidesNotInView`和`CanScrollPrev/Next`提供控制器查询。
ScrollSnapList是源定义的归一化位置，区别于组合快照Snaps的DIP偏移；
Progress在越界拖动时可能超出0..1。RootHandle指viewport，RegionHandle指外围具名区域，
ContainerHandle指track。快照中的slice均独立复制。

使用`NewCarouselListener`创建有稳定身份的应用监听器，通过On/Off订阅init、select、scroll、
settle、reInit、resize、slidesChanged、slidesInView、pointerDown/Up、slideFocusStart/Focus和destroy。
重复On会重复通知，Off删除该监听器的全部相同注册；一次Emit使用当时的监听列表快照。
监听器不能捕获跨帧Context或Element。Destroy会清空当时的订阅；显式On/Emit仍可用于应用通知，
但不恢复引擎。立即跳转的新动画按源顺序先scroll、后select，不额外产生settle。

## 物理运动与原生拖动

控制器路径实现固定Embla8.5.2的吸引力、摩擦、边界阻力、循环目标选择与60Hz累积器，
首次动画包含源定义的两次暖机seek，绘制在前后物理状态间插值。中途改目标、即时跳转、
循环换坐标及事件中间状态共用同一求解器。

- Options.Duration默认25，是吸引力参数而非毫秒；0使程序导航即时完成。
- DragFree控制释放后是否吸附，SkipSnaps控制大速度及目标已变化时的跨吸附位置行为。
- 释放速度取原始指针点与时间戳的170ms窗口，鼠标/触摸分别使用源速度增益，
  释放运动采用源25/43速度参数，独立于程序导航的Duration。
- DragThreshold默认10DIP，记录点击抑制策略；导航的力阈值另按viewport的20%夹到50..225DIP。
  EngineSnapshot.ClickSuppressed是当前捕获手势的策略状态，不能取消其他原生控件的点击。

这些选项支持部分更新、显式false/0、深拷贝及断点合并。EngineSnapshot还提供Velocity、
当前Duration/Friction、PointerDown和ClickSuppressed值，供应用查询。
Content之前和之后的组合快照均按当前物理姿态计算可见项目。

背景区域通过公开Dragged声明捕获，非消费HandleInput取得越过viewport边界的move/up；
求解器负责全部平移，不叠加MyGo Draggable的独立动画。原生子按钮正常点击，文字编辑器继续编辑。
取消或窗口失去输入时返回选中吸附位置，不提交甩动；触摸跨轴移动占主导时取消本次捕获。

目前MyGo公开接口无法从已经取得指针的任意子按钮接管拖动，也不能由祖先取消该按钮的点击；
从这种子控件开始拖动并按源阈值阻止其动作仍是明确缺口。多触点数量及cancelable等完整触摸元数据
同样尚未提供，单指mouse/touch轨迹验收不代表完整触摸平台验收。


## 插件与自动播放

`NewCarouselAutoplay(options)` 创建应用持有的插件身份，构造错误需由调用方处理。
将其传给 `CarouselContainerProps.Plugins`；例如使用
`Delay: display.CarouselValue(2*time.Second)` 配置两秒间隔。默认间隔四秒，
默认播放，用户交互或焦点进入后停止。Play（可选jump）、Stop、Reset与控制器共享命令FIFO；
IsPlaying读取已提交状态。`TimeUntilNext(c.Now())` 显式使用当前UI时钟，
零及亚毫秒间隔按原生调度器的最小1ms异步安排，不在初始化或重置构建中同步切换。
返回剩余时间及是否有timer；它按当前选中snap的delay计算，可能区别于先前已安排的deadline。

选项支持Jump、PlayOnInit、StopOnFocusIn、StopOnInteraction、StopOnMouseEnter、
StopOnLastSnap、Active与有序媒体断点。纯DelayResolver为每个snap返回非负duration；
Root Handle或纯RootResolver选择原生hover区域，返回零Handle使用默认viewport，
无法解析的非零Handle报告typed error。Resolver显式由应用持有，只能捕获模型与Handle值。

`NewCarouselPlugin` 提供typed Init/Destroy/Event/Frame钩子。每次调用收到时间、
环境、快照、Root Handle及当前回调有效的Actions，可同步导航、发事件、控制Autoplay、
设置下次deadline。回调返回后Actions失效并清除帧引用；不能保留Context、Element或旧帧。
外部指令继续排队，插件内同步操作保留源事件的中间状态与注册顺序。

Props.Plugins为nil时保留配置，非nil空slice清空；显式SetPlugins/ReInitWithPlugins替换列表，
其nil也表示清空。原ReInit保留插件。插件按slice顺序初始化/销毁；重名实例都运行，
Plugins返回每个名称最后的实例。重复同一身份保留源重复订阅行为。一个身份不能同时绑定两个
存活的原生owner；更换、销毁及视图卸载都会移除订阅和timer。媒体变化在新布局测量后重建。

自动播放接通实际物理引擎，支持hover、pointer、Tab焦点、隐藏暂停与恢复、末项停止和loop。
重新初始化保留源mouseIsOver/playOnDocumentVisible纯状态，同时清除旧timer；
单snap时保持静止，包括手动Play。公开SDK不能精确观测同一slide内部任意两个editor间的
pointer focusout，使用track/slide变化和Tab进行适配。完整平台输入与AT仍需验收。

## Gallery

Data & display页提供可操作的Carousel示例：单项、双项、三项、响应式和固定宽度，
可切换间距、循环、集合长度及项目顺序。卡片内的真实原生编辑器把备注保留在对应项目上，
导航、重排和切换预览页后仍保留。另有270DIP垂直半屏布局及独立RTL示例，
导航按钮位于viewport外侧，小窗口通过页面滚动到达所有示例。
搜索Carousel、Slides或Slideshow可进入该页。
示例已使用实际controller/源物理；可操作jump、duration、dragFree、skipSnaps与alignment，
另有两秒Autoplay及播放/停止/重置。八份固定官方例子映射到这些实际配置，具体见source manifest。
离页销毁绑定，返回时创建新controller，项目选择、备注与插件源交互carry保留。
