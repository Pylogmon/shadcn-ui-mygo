# Independent native input answers

This is a native adaptation note, not an upstream original. The byte-exact
fixed originals and their Git/SHA-256 provenance remain in `manifest.json`.

The fixed examples each demonstrate one freeform input per item. The primitive
itself registers every Input instance independently:

- `use-questionnaire-input.ts:33,45–66`: each instance creates an answer ID and
  registers its own default selection and actual input control.
- `use-questionnaire-item.ts:120–137,152–168`: registration is a collection;
  a single item replaces the selected ID and a multiple item accumulates IDs.
- `use-questionnaire-input.ts:98–111,128–143`: accepted filled edits select that
  instance, and only selected controls receive the same item name. Ordinary
  native FormData preserves declaration order and duplicate text values.
- `use-questionnaire-item.ts:191–218`: unregistration removes that answer's
  default and selected registration. Disabling keeps its state while native
  successful-form controls exclude it.

The Go API keeps `Answer.Values` for fixed choices, `Answer.Text` for the first
unnamed input, and the existing aggregate `Mode`. With additional/explicit
slots, `Answer.Inputs` contains `QuestionnaireInputValue{Key, Text, Selected}`.
State retains drafts; `Result` returns only successful registered input values.
The native label comes from current `QuestionnaireInputProps.Label` and is not
included in same-name form values.

```go
parts.Item("context", func(item *conversation.QuestionnaireItemParts) {
    item.Title("Execution context")
    item.Choices(func(*ui.Context) {
        item.Input(conversation.QuestionnaireInputProps{
            Key: "target", Label: "Target environment",
        })
        item.Input(conversation.QuestionnaireInputProps{
            Key: "review", Label: "Review policy",
        })
    })
    item.Error()
})
```

Declare the containing Question with `Multiple: true` to submit both successful
values. `FormValues(questions)["context"]` is the ordered same-name value list;
equal text from two separate inputs remains two values. A single item selects
the last edited filled input and retains the other drafts. Selecting a fixed
radio excludes all input drafts; multiple checkbox selection remains additive.
Fixed choices are deduplicated by their own identity, so an earlier input with
text equal to a later choice value does not suppress that successful choice.

Key is optional. Unnamed inputs use declaration order, with the first retaining
the existing default slot. Subsequent implicit identities have a reserved NUL
prefix. Explicit keys cannot contain NUL, must be unique within the item and
preserve identity across reordering or insertion. Repeated explicit keys panic
with a descriptive error. Dynamic removal disposes that slot's registration
and draft; remounting applies its current default. Disabled inputs keep drafts
but do not submit. Skipping keeps drafts while clearing successful selection;
reset restores the independently registered defaults.
Explicit `Question.DefaultAnswer.Inputs` selection flags survive reset,
including nonempty unselected drafts. A nonempty slot `DefaultValue` declares
that slot's own default. Per-input validation retains only the failing control
key between builds and focuses its current enabled native editor; this mirrors
`use-questionnaire-item.ts:289–302`, without retaining a frame Element/Context.

The contract is exercised by one actual native flow in
`questionnaire_inputs_test.go`. It starts with two unnamed editors (the old
implementation panicked because both reused `qInputSlot`), follows single and
multiple selection and duplicate text submission, then uses keyed editors for
reordering, disabling, removal/remount, current callbacks, skip and reset. No
schema engine, independent form field names, persistence service or new
navigation state machine is introduced.
