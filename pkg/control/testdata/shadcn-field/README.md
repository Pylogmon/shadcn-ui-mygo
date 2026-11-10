# Field error reference

The complete fixed Field wrapper and documentation are copied byte-for-byte
from shadcn commit `8ab27715521cbd4f87ae6886cd989cbb678f6de7`.
`manifest.json` records Git blobs, SHA-256 values and the reused Vega stylesheet.
The existing bundled shadcn MIT license also covers these references.

`field.tsx:162–208` gives custom children precedence, deduplicates messages and
renders one message as text or multiple messages as a bulleted list.
`field.mdx:391–404` documents validator issue arrays. Go callers provide message
strings; empty strings are omitted, rather than exposing JavaScript
placeholder-only lists as empty native alerts.

The native FieldError returns an actual alert/list with stable message order;
Field links its accessible summary to the actual editor, including InputGroup
and Textarea. Custom content needs an explicit label or message summary.
Validation rules and change/blur/submit timing stay application-owned. Native
role/relationship checks do not establish real screen-reader announcements.
