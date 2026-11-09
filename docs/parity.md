# Native port parity and acceptance gates

This project is a native Go implementation of the shadcn design system. The
current API inventory is a baseline, not evidence of complete behavioral or
visual parity. **Full shadcn parity remains incomplete.**

The inventory below follows the 66 entries under **All Components** in the
[official component index](https://ui.shadcn.com/docs/components), inspected on
2026-11-09. The separate AI Elements collection, site templates, documentation
website and React-specific adapter libraries are distinct from this component
inventory. This does not waive requirements of the requested native UI kit.

Architecture references inspected:

- [Lobe MyGo](https://github.com/lobehub/lobe-mygo/tree/40d9ea87a3c037531a83cdb1578b5e6102b74d03):
  semantic theme tokens, headless native base controls and component packages.
- [MyGo Music](https://github.com/egoist/mygo-music/tree/1af0d8efd77fdf630c27cfc635f12e9cb089ce10):
  main-thread state updates, application-owned state, stable widget identity,
  appearance preferences and rendering on demand. Its AGPL application source is
  a design reference, not copied implementation code.

## Component inventory

"Implemented" below describes the current callable native API. The acceptance
column names material remaining work or the verification still needed. No row
claims every upstream variant, example, accessibility contract or transition has
been reproduced. Callback content and returned `ui.Element` values allow native
composition; they do not automatically supply missing upstream behavior.

| Component | Implemented native API | Remaining acceptance work |
| --- | --- | --- |
| Accordion | `display.Accordion` with single/multiple controlled sections | Verify every size, disabled state, disclosure association and exit motion in both directions. |
| Alert | `display.Alert` | Verify alert/live-region semantics, leading icons and full content layouts. |
| Alert Dialog | `overlay.AlertDialog` | Backdrop protection, alert role and Cancel are tested; verify nested dialogs, initial focus, asynchronous confirmation and reduced-motion exit. |
| Aspect Ratio | `display.AspectRatio` | Verify nested images, extreme aspect ratios and constrained resizing visually. |
| Attachment | `conversation.Attachment` | Current callbacks/status/progress are basic primitives; complete upload/cancel/retry, media previews and drag-and-drop composition examples. |
| Avatar | `display.Avatar` | Verify delayed/error image fallback, avatar groups and badge/size variants. |
| Badge | `display.Badge` | Verify all variants, interactive/link composition, icon spacing and focus states. |
| Breadcrumb | `navigation.Breadcrumb` | Current route actions and current-page semantics are tested; add overflow/ellipsis menus and linked-route examples. |
| Bubble | `conversation.Bubble` | Controlled expansion/reactions exist; verify all tails, sizes, streaming content and reaction layouts against current upstream examples. |
| Button | `control.Button` / `StyleButton` | Vega-aligned variants, loading/busy state, invalid state and all standard/icon sizes are implemented; verify full platform/theme screenshot and screen-reader matrices. |
| Button Group | `control.ButtonGroup` | Joined horizontal/vertical geometry, RTL, nested groups, text/separators and native roving focus are tested; verify all compound compositions visually. |
| Calendar | `display.Calendar` | Currently MyGo's single-date calendar; date ranges, multiple selection, disabled/matcher days, month counts and complete navigation variants remain. |
| Card | `display.Card`, `CardHeader`, `CardContent`, `CardFooter` | Verify action slots, responsive layouts and all size variants. |
| Carousel | `display.Carousel` | Boundaries are tested; add swipe/drag, orientation, looping options, keyboard arrows, slide transitions and carousel announcements. |
| Chart | `display.Chart` | Bar/line/area and missing-data handling exist; pie/radar/radial/stacked examples, tooltip/legend APIs, keyboard inspection and full chart examples remain. |
| Checkbox | `control.Checkbox` | Native checked/mixed/disabled behavior exists; complete screenshot and mixed-state keyboard acceptance. |
| Code Block | `display.CodeBlock` | Native syntax editor, selection and copy exist; verify copy feedback, supported language mapping, themes and large-file behavior. |
| Collapsible | `display.Collapsible` | Verify focus/controls linkage, interrupted/reversed animation and content preservation. |
| Combobox | `control.Combobox` | Grouped/custom single-select and multi-select chips with filtering/removal/limits exist; verify loading/async sources, full variant examples and platform IME behavior. |
| Command | `overlay.Command` / `CommandDialog` | Inline filtering/Enter is tested; verify all disabled-navigation cases, scroll-to-highlight, shortcuts, accessibility and large result sets. |
| Context Menu | `overlay.ContextMenu` | Native secondary-click and actions are tested; verify keyboard context-key/Shift+F10, all submenu directions, touch and collision edges. |
| Data Table | `display.DataTable` / `DataRows` | Filtering/sorting/paging exist; add column visibility, facets, row-selection toolbar, pinned columns and complete editable/custom-cell examples. |
| Date Picker | `display.DatePicker` | Single-date browse/commit and empty value are tested; ranges, disabled dates, date-time, timezone and formatting variants remain. |
| Dialog | `overlay.Dialog` | Escape, accessible metadata, exit and focus restoration are tested; verify nested focus traps, long scroll content and all interrupted transitions. |
| Direction | `display.Direction` | Native direction scope exists; run all directional components in RTL, including popup flipping, keyboard arrows and physical/logical padding. |
| Drawer | `overlay.Drawer` | Bottom modal with draggable dismiss exists; verify touch gestures, velocity/snap thresholds, nested scrolling, focus and motion interruptions. |
| Dropdown Menu | `overlay.DropdownMenu` | Checkbox, radio and nested item API exists; keyboard disabled skipping/submenus/focus restoration are tested. Verify typeahead, RTL, collision and custom content. |
| Empty | `display.Empty` | Verify icon/media/content/action compositions and constrained widths. |
| Field | `control.Field` | Vertical/horizontal/responsive layout, fieldset/group, required state and compound editor label/error linkage are tested; complete IME/screen-reader and validation timing examples. |
| Hover Card | `overlay.HoverCard` | Open/close timers and pointer-gap persistence exist; verify timer cancellation, Escape suppression, interactive focus and placement on every side. |
| Input | `control.Input` | Native text editing/readonly/disabled exists; verify IME, password, selection, clipboard, invalid/focus geometry and all sizes. |
| Input Group | `control.InputGroup` | Inline/block arbitrary addons, buttons, text and compound editor focus/error propagation are implemented; complete textarea/menu/action examples and platform screenshots. |
| Input OTP | `control.InputOTP` | Native hidden editor + visible cells, filtering and completion exist; verify paste, selection/caret position, IME, grouping and pattern variants. |
| Item | `display.Item` | Basic leading/body/trailing slots exist; verify interactive/link, media, separators and all variants/sizes. |
| Kbd | `display.Kbd` | Verify platform modifier mapping, key groups, icons, nested buttons and contrast. |
| Label | `control.Label` | Standalone LabelFor and Field associate visible labels with native editors and focus them on click; verify all disabled and screen-reader combinations. |
| Marker | `conversation.Marker` | Inline/bordered/separator status exists; verify icons, busy states and complete upstream compositions. |
| Menubar | `overlay.Menubar` | Native menu triggers and horizontal focus grouping exist; sibling-menu switching with Left/Right while open, hover switching and all close/restore paths need tests. |
| Message | `conversation.Message` | Sender/avatar/header/footer layout exists; complete action/tool/error/branching examples and verify streaming accessibility. |
| Message Scroller | `conversation.MessageScroller` | Keyed native virtual list, place holding and live-edge following exist; turn anchoring, stream pause heuristics and upstream jump controls remain. |
| Native Select | `control.NativeSelect` | Deliberate MyGo system-menu adapter replaces the alias; native disabled/group/selection and no parent-tree menu painting are tested. Verify real out-of-window menus on every platform. |
| Navigation Menu | `navigation.NavigationMenu` | Native grouped dropdown route menus exist; shadcn content mega-panels, viewport/indicator motion, hover intent and sibling keyboard switching remain. |
| Pagination | `navigation.Pagination` | One-based bounds, ellipsis and current page are tested; verify large page counts, responsiveness and full link composition. |
| Popover | `overlay.Popover` | Native outside/Escape dismissal, attachment and styling exist; verify initial focus, focus restoration, nested overlays, collision and all align variants. |
| Progress | `display.Progress` | Determinate native progress exists; verify accessible value text, indeterminate composition and reduced motion. |
| Questionnaire | `conversation.Questionnaire` | Controlled answers/validation/step navigation exist; verify all keyboard/focus, custom question renderer, review/edit and conditional-flow examples. |
| Radio Group | `control.RadioGroup` | Native selection/arrow behavior exists; verify every orientation, RTL and group/description association. |
| Resizable | `display.Resizable` | Native two-pane split exists; verify min/max during drag, keyboard resizing, persistence, nested and multi-pane layouts. |
| Scroll Area | `display.ScrollArea` | Native scrolling exists; verify horizontal/both axes, scrollbar variants, overscroll and nested scroll chaining. |
| Select | `control.Select` | Grouped options, custom item/trigger content, keyboard and disabled behavior exist; complete align/collision/scroll examples and screenshots. |
| Separator | `display.Separator` | Horizontal/vertical rule exists; verify decorative versus semantic separator roles. |
| Sheet | `overlay.Sheet` | Four-edge modal panel exists; verify every edge, constrained sizing, nested scroll, focus/exit and RTL logical-side behavior. |
| Sidebar | `navigation.Sidebar` | Current route, grouping, collapse and separate sidebar semantic colors exist; responsive/off-canvas, nested collapsibles, icon tooltips, rail/inset/floating variants remain. |
| Skeleton | `display.Skeleton` | Static/circle/pulse exists; verify contrast, reduced motion and loading accessibility without idle redraw. |
| Slider | `control.Slider` | Native single-thumb pointer/keyboard/step and RTL draw exist; range/multiple thumbs, vertical orientation and accessible value formatting remain. |
| Spinner | `display.Spinner` | Native labeled spinner exists; verify reduced motion, sizes and compound loading examples. |
| Switch | `control.Switch` | Native toggling and thumb transition exist; verify RTL thumb geometry, focus/labels and disabled states visually. |
| Table | `display.Table` | Native virtual table with captions exists; verify header/footer/row-header, selection, overflow and empty-state compositions. |
| Tabs | `navigation.Tabs` | Native tab roles/disabled skipping/Home/selection are tested; explicit trigger-to-panel associations, activation modes and moving active indicator remain. |
| Textarea | `control.Textarea` | Native multiline editing and row sizing exist; verify user resize, autogrow limits, IME and field associations. |
| Toast | `overlay.Toast` / `Toaster` | Native queue, action, close and pause semantics are used; action dismissal is tested. Verify timed expiry, hover/focus pause, swipe and every placement. |
| Toggle | `control.Toggle` | Native pressed state exists; verify icon-only labels, every variant/size and focus/disabled screenshots. |
| Toggle Group | `control.ToggleGroup` | Controlled single/multiple values and horizontal keyboard grouping exist; verify vertical groups, RTL and group accessibility. |
| Tooltip | `overlay.Tooltip` | Native delay/focus and trigger description exist; verify all placements, skip-delay groups, Escape and reduced-motion presence. |
| Trace | `conversation.Trace` | Recursive controlled status/disclosure exists; verify running/failed transitions, unique nested IDs and all upstream detail/action compositions. |
| Typography | `display.Typography` | Heading/paragraph/lead/quote/code styles exist; verify semantic heading levels, lists/tables/links, font metrics and complete typesetting examples. |

## Cross-cutting acceptance gates

1. **Source fidelity:** pin and record the shadcn revision used for styles,
   tokens and every example; compare the catalog with that revision. A changing
   website inventory alone does not establish parity.
2. **Visual fidelity:** render every component and state beside its upstream
   example in light/dark themes. Check typography, radius, spacing, borders,
   shadows, focus rings, clipping and disabled/invalid colors at multiple scales.
   A gallery showing a subset cannot prove all-component fidelity.
3. **Theme/color completeness:** verify all upstream base and accent palettes,
   chart/sidebar pairs, nested custom themes, alpha compositing, gamut mapping,
   contrast, live system appearance changes and saved explicit modes. The 24 registry presets and 26 full Tailwind scales are now generated
   from archived, SHA-256-pinned upstream sources; partial brand overrides merge
   onto Neutral exactly as recorded in the source manifest. Full per-component
   visual checks and preference restoration after native app restart remain.
4. **Motion completeness:** verify enter/exit, state changes, reversal,
   interruption, matched indicators, layout/drag animations and zero animation
   under reduced motion for every affected component. The existence of motion
   presets is not evidence that every component applies them correctly.
5. **Native accessibility and input:** verify role/name/state/relationships,
   keyboard-only use, disabled skipping, focus visibility/trapping/restoration,
   screen-reader announcements, IME/clipboard, text scaling/high contrast, RTL
   and platform key conventions. Go role tests supplement real screen readers.
6. **Platform runtime:** compile and launch on supported macOS, Windows and
   Linux targets. Confirm GPU output, system appearance, window resizing,
   display scale changes, clipboard, menus, gestures and idle power behavior.
   Headless native renders prove neither a launched window nor cross-platform
   runtime correctness.
7. **Component examples and regression coverage:** provide runnable examples
   for every public component and every material variant above. Add interaction
   tests around behavior, not tests merely checking exported function names.
   Exported API/catalog counts are coverage bookkeeping, not acceptance tests.
8. **Distribution:** verify a clean consumer module can import the kit, build
   and run the gallery, customize a scoped theme, switch persisted appearance
   and use documented APIs without local module replacements or cached files.

Tests currently present cover selected interactions in `control`, `display`,
`navigation`, `overlay`, `theme`, `motion`, `conversation` and the Gallery. Their names and assertions in the
source are the evidence for what they verify. Unexecuted, missing or narrowly
scoped tests leave the corresponding gate open.
