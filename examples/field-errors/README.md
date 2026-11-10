# Field errors

Run `go run ./examples/field-errors`. Edit the project code and check it, or
switch between a deduplicated error list and custom native content. The example
owns its validation rules and timing; the component only presents messages and
links them to the actual InputGroup editor.

`FieldProps.Error` remains the single-message convenience. `Errors []string`
adds messages in order, omitting empty strings and exact duplicates. No messages
means no error container; one message is ordinary text, multiple messages form
a native alert containing a semantic list. Clearing errors preserves the
editor, selection, description and draft.

`ErrorContent` replaces the visual message/list. `ErrorLabel` supplies its
accessible summary, defaulting to the combined Error/Errors. A content builder
without messages must provide a nonempty label. It executes in the current
build and can compose ordinary native text or actions.

For an independent error slot, call
`control.FieldError(c, control.FieldErrorProps{Errors: messages})` and link its
returned element using the real editor's `ErrorMessage` and `DescribedBy`.
An empty result is an invalid Element, suitable for a `Valid()` guard.

The fixed source Field wrapper and documentation, including the multiple-message
contract, are archived in
[shadcn-field](../../pkg/control/testdata/shadcn-field/README.md). Empty Go
strings represent missing messages; the native version omits empty alerts
rather than reproducing JavaScript placeholder-only lists.

```sh
go run ./examples/field-errors -snapshot /tmp/field-errors.png
go run ./examples/field-errors -dark -custom -snapshot /tmp/field-errors-custom.png
```
