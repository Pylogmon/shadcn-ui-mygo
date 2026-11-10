# Independent questionnaire inputs

Run `go run ./examples/questionnaire-inputs`. One multiple-answer question
contains two real text editors and a native checkbox. Its current successful
values follow the declaration order, so the initial two `team` values remain
separate successful controls. The application supplies the complete initial
answer through `DefaultAnswer.Values/Inputs`. Save takes an independent snapshot;
Reset restores defaults. Add `!` to the secondary destination and submit to
exercise application validation and return focus to that editor.

`QuestionnaireInputProps.Key` keeps an editor stable through reordered or
conditional content. Without keys, inputs use independent declaration slots;
the first unnamed input keeps the existing `Answer.Text` contract. Additional
drafts and their selection live in `Answer.Inputs`. A single-answer question
selects one input or fixed choice; a multiple-answer question retains each
selected control. Identical text does not merge distinct successful controls.

Validation rules and saving belong to this example. It sends no report or
network request. The SDK owns the actual editor, selection, input method and
questionnaire navigation. See the [source contract](../../pkg/conversation/testdata/questionnaire/native-inputs-contract.md).

```sh
go run ./examples/questionnaire-inputs -snapshot /tmp/questionnaire-inputs.png
go run ./examples/questionnaire-inputs -dark -snapshot /tmp/questionnaire-inputs-dark.png
```
