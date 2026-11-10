# Payments Gallery composition

The Gallery's Payments card composes the existing public `DataRows`, `TableWith`, `CheckboxState`, Input, buttons and styled/native menus. It preserves Recent invoices and the editable Table composition. It is a Gallery application recipe; this directory does not contain a separate executable or a new table SDK.

Its application-owned model keeps the email query, ascending/descending email sort, hidden data columns, stable payment-ID selection, page size/page, horizontal scroll and last copy acknowledgment across Gallery page changes. The header checkbox derives Off/Mixed/On from the current page and toggles only that page. Selection/actions remain visible when every data column is hidden. Selection counts use the filtered source set, including selected records on other pages. The delivered `DataRows` model performs filtering/sorting/paging; the example does not duplicate those algorithms.

The five payment fixtures and default ten-row page match the fixed source. Next being disabled on the initial page is expected. The explicit **2 per page (demo)** and 5-per-page options let users exercise pagination with these five records; they are application demonstration choices, not a claim that the source demo defaults to two rows.

Copy payment ID calls the actual asynchronous `Context.WriteClipboardThen`; only its completion reports success. Its callback keeps application values, never a Context or Element. View customer/details display only locally available email/payment fields; they do not contact a customer backend. Column, page-size and row menus follow Gallery's styled independent Flyout / Native OS menu route. The layout mirrors in RTL; the fixture uses USD formatting and English labels, without claiming the source's optional Arabic/Hebrew translation bundle.

## Fixed source map

All references are to shadcn/ui commit `8ab27715521cbd4f87ae6886cd989cbb678f6de7`, already verified in the local source corpus.
Exact original blobs, SHA-256 values and native adaptations are recorded in [source-map.json](source-map.json).

- `apps/v4/examples/base/data-table-demo.tsx:64–101`: exact original five payment records and Payment shape.
- `data-table-demo.tsx:105–125`: native current-page all/mixed selection and protected selection column.
- `data-table-demo.tsx:127–159`: capitalized status, lowercase/sortable email and USD amount.
- `data-table-demo.tsx:161–188`: protected actions column and actual Copy payment ID handler. The source customer/details items have no handlers; this recipe provides a bounded local readout only.
- `data-table-demo.tsx:224–259`: email-only filter and hideable-column checkbox menu.
- `data-table-demo.tsx:278–325`: selected rows, empty results, filtered selection count and previous/next paging.
- `apps/v4/content/docs/components/base/data-table.mdx:20–26`: tables are application compositions; `842–884` gives reusable headers, pagination (including page size) and visibility controls.
- `apps/v4/examples/base/data-table-rtl.tsx`: actual RTL composition; the native table supplies mirrored layout.

The concentrated native workflow covers email filtering/sort, mixed page selection, stable keys across pages, protected visibility, clipboard refusal then success, local details, OS-menu switching, page-size changes, RTL, empty rows and remount persistence. It uses actual Tester actions and optionally exports light/dark RTL PNGs through `GALLERY_PAYMENTS_SNAPSHOT_DIR`. It does not add sticky pinning, editable-cell policies, browser DOM emulation or platform acceptance matrices.
