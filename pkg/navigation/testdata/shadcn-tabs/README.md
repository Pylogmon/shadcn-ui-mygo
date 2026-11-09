# Fixed Tabs source

These nine original files are pinned to shadcn/ui
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. Every original byte was checked against
its Git blob and SHA256 from the complete official source inventory. The manifest
records direct official URLs, hashes and the 13 actual collection examples.
MIT attribution is retained through the repository's third-party notices.

The required gap is independent composition: TabsWithDropdown and
TabsWithInputAndButton put unrelated controls beside the list and put panels in a
separate layout container. TabsMultiple is four ordinary single-selection tabs.
The icon-only source has no accessible label; native HideLabel deliberately
preserves a useful Label and does not claim that source example was accessible.

The new native Root/List/Trigger/Content scope is synchronous and restores sibling
and nested scopes. Persisted state contains only a public Handle; Elements and
Contexts remain inside one build. Native buttons participate in a roving tablist;
orientation, disabled skipping, manual activation, no-loop and RTL retain the
existing native behavior. Active panels retain Controls/LabelledBy and are
focusable native regions. An application-owned string is the source value;
initialize it explicitly. Empty selects none, and an unmatched value can address
a panel with no trigger. The original int-based Tabs/TabsProps is unchanged,
including its existing removed/disabled normalization policy.

Panels can be placed before or after the list; links are published after slots
have been collected. Input read later in the build converges through native
invalidation. Inactive Content closures are not called, preserving native input
and accessibility exclusion. Keep application drafts in application state. This
module adds no retained hidden subtree or generic Base UI option universe.
Native LabelledBy also activates a caption click, so association publishing
retains tab focus on default activation. No private runtime access is used.

The public example `examples/tabs-composition` runs all 13 configurations,
including actual editable input, an action button, a working menu in the header,
independent content, icons and vertical layout. Formal public-native tests cover
layout/AX links, interaction, disabled/manual/RTL, nested/sibling scopes,
content-before-list, keyed focus and reduced/normal indicator settling. The
source primitives' CSS timing and real OS screen-reader behavior are not claimed
by native compilation or Tester checks.
