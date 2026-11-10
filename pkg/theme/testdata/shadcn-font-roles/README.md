# Fixed Create font roles (stage one)

The twelve byte-exact originals in `manifest.json` match official shadcn commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`; each Git blob and SHA256 is recorded.
The original shadcn MIT license is included. No font binary is included here.

The complete Base UI `cn-font-heading` marker set is CardTitle (card:36),
DialogTitle (dialog:126), AlertDialogTitle (alert-dialog:113), DrawerTitle
(drawer:196), SheetTitle (sheet:112), EmptyTitle (empty:64), and
QuestionnaireTitle (questionnaire:63). The native Dialog header also serves
Drawer and Sheet. The legacy Questionnaire keeps the same title role as its
compound counterpart. Ordinary body text, navigation labels, and Typography's
unmarked headings are not reclassified by this stage.

Create's `config.ts` selects Inter/inherit by default and Base Vega, and
Noto Sans/Playfair Display for Base Sera. Its picker offers separate body and
heading choices. `search-params.ts` normalizes equal body/heading choices to
the explicit inherit sentinel, so a later body change continues to propagate.

The native role API uses actual MyGo font families/fallback lists. An empty body
is the system font; empty/inherit heading follows the body. `system` explicitly
requests the interface font for either role. Unconfigured titles retain native
ancestor inheritance, and Native window surfaces retain their platform theme.
Scopes and styled independent windows copy only role values; they never retain
a parent Context or Element for font lookup. A local ancestor `.Font` has no
public computed-family getter, so cross-window inheritance uses Theme.Font or
FontScope explicitly.

This stage provides roles and inheritance, not the 26 named-font assets. The
next stage must obtain the exact family definitions, pin official TTF/OTF assets
and their original font licenses, and connect catalog identifiers to registered
native families. The shadcn MIT license does not license those font files.
