# Menu selection tracking source

Pinned public Base UI 1.6.0 component originals supporting the default checkbox/radio `closeOnClick=false` contract. URLs, bytes and SHA-256 are in `manifest.json`; the existing Base UI MIT license is `../npm/base-ui-react/LICENSE`. The surrounding menu sources and upstream npm integrity remain recorded in the parent archive.

Styled desktop menus retain checkbox/radio selection by default. `MenuItem.CloseOnSelect` overrides each row; `KeepOpenOnSelect` overrides closing globally. Ordinary action rows close. Native OS and legacy MyGo MenuBase menus retain platform/SDK tracking. `desktop.Select` explicitly closes its radio-backed internal options when committing a value.
