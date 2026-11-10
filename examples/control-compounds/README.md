# Button groups and Avatar dropdown

This Gallery application module composes the existing public controls; it adds
no new button, input, menu or avatar SDK. Its application-owned state exposes
two entry points for the Gallery owner:

- `(*controlCompoundsPreviewState).buttonGroups(c, host, native)` for Forms.
- `(*controlCompoundsPreviewState).avatar(c, host, native)` for Data.

Retain one state in the Gallery. Existing Profile, email, table, portrait and
InputGroup models remain independent. This directory is documentation and
source mapping, not a standalone executable.

The configuration picker contains the fixed source's **12 ButtonGroup example
files plus its inline ButtonGroupText example**, for 13 arrangements: mail
actions with nested radio labels, vertical controls, three sizes, nested input,
copy/paste separator, secondary split button, search input, pill voice input
group, Follow/dropdown, currency/input/action, task Popover, text/label and
translated RTL actions. The split example is a primary button and a separate
Plus action; it does not pretend to contain a menu.

Editors, selected currency, voice UI mode, task draft and radio label survive
configuration/page switches. Voice mode disables the actual editor and keeps
its text; this follows the source UI toggle and does not record audio. Clipboard
copy reports its actual completion or refusal, Paste reads the real SDK
clipboard, and the joined amount action reads both current controls into a
local preview. Task content is built in its independent child Context and
keeps an application draft. It does not call a Copilot backend.

The Avatar trigger uses the Gallery's existing local portrait. Profile, Billing,
Settings and destructive Log out are the exact source menu items, with a bounded
local action readout. They do not sign in or change a real account. Menu,
Select, Tooltip and Popover calls follow the supplied Gallery route: styled
Flyout or OS menu/platform-themed native container. This distinction is about
the supported platform capability, not a claim that an OS has a native task
editor widget.

Original Git blobs, SHA-256 values and source/native adaptations are recorded
in [source-map.json](source-map.json), against shadcn commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. The kit retains its Vega styling
and Material Symbols; no additional source icon assets are copied.

One concentrated native workflow covers real joined editing and action reads,
voice disable/restore, clipboard refusal/success, separate task-child editing,
label focus, RTL language/radio state, page drafts and both Avatar menu routes.
The test's owner/page host is deliberately local; the Gallery owner separately
verifies actual `State.View` integration. Its native close callback is injected
because Tester does not emulate NSPopover Escape/outside tracking. The three
PNG exports are native render evidence, not OS-window, screen-reader or complete
13-configuration visual acceptance.
