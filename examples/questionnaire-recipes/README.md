# Questionnaire workflow recipes

The Gallery **Questionnaire workflows** card uses one application-owned
`questionnaireRecipesState.view(c)` with three independent drafts. Its host
retains that state and calls `view` from the existing questionnaire area;
changing configuration or leaving the page keeps the answers and checkpoint.
The existing workspace preferences review/edit flow remains separate.

- **Conditional execution plan:** Local workspace skips the cloud environment.
  Cloud workspace adds its required environment question and changes progress
  from two to three items. Switching back to Local excludes a previously
  answered environment from successful submission. The status shows the actual
  current submitted runtime, environment and approval choice.
- **Resume saved migration:** open at Verification with an Incremental migration,
  tests/typecheck and a note supplied as saved defaults. Edit them, then use
  Reset changes to return to those saved values and that second checkpoint.
  Update draft receives the actual native answers. This is hydration from
  application-provided defaults, not a disk persistence service.
- **Validate response policy:** Public audience requires a Complete answer.
  Submitting Public + Concise returns to the earlier Detail question and its
  native error/focus target. Correcting it and submitting reports current values.

These use delivered `QuestionnaireWith`, `DefaultAnswer`, `DefaultItem`,
controlled Item, Disabled and ValidateAll APIs. Native Radio/Checkbox/Input,
required validation, collection navigation, focus and error linkage remain the
SDK's responsibility. No navigation state machine, schema/NLP dependency or
network backend is added. Only pure application data is retained; callbacks
and scoped Parts run during the current build.

The source is shadcn commit `8ab27715521cbd4f87ae6886cd989cbb678f6de7`.
Original Git blobs, hashes and exact lines are in [source-map.json](source-map.json).
The configurations preserve `questionnaire-conditional`, `questionnaire-resume`
and `questionnaire-validation` data and rules; readable native status text
replaces their browser toast presentation. UI tests use one concentrated flow
through the three configurations and a page return, not an expanded family or
platform/AT matrix.
