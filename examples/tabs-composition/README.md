# Tabs composition

Run the real native example:

```sh
go run ./examples/tabs-composition
```

Choose any of the 13 fixed examples. The input example has a working search field
and action; the dropdown example has a real menu. Panel notes stay in application
state across tab changes. Manual activation, no-loop and RTL controls exercise
the native keyboard behavior.

Render this same view without opening a window:

```sh
go run ./examples/tabs-composition -recipe input -snapshot /tmp/tabs-input.png
go run ./examples/tabs-composition -recipe vertical -dark -snapshot /tmp/tabs-vertical.png
```

TabsRoot owns the scope, TabsList only owns its trigger group, and TabsContent can
sit in a separate header/body layout. The example uses the public in-window menu
adapter; a desktop Flyout or NativeMenu can occupy the same header slot.
Icon-only triggers use HideLabel with a useful native Label. TabsMultiple remains
single-selection. The original int-based Tabs convenience API remains available.
