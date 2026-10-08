# base/security-require-context-access

Proof consumer for the ecmascript/decorators syntax helpers. Requires configured
protector decorators when a nonoptional method parameter injects a configured
request-context key. Imports aliases, enclosing-class protectors, downstream
GraphQL resolver exclusions and syntactic optional/nullable parameters retain
Go behavior.

Adapted the partial `rule.ts`, `options.ts`, messages and oracle from
`origin/codex/lint-wave1-11`, replacing decorator and import extraction with
shared helpers and completing the Go argument and alias handling. Registration
is owned by `rule.json`; no central registry files change.

The selected gate requires 54 unique upstream source/rule/options cases, including
four cases asserting results directly, plus an owned alias witness in selected and all modes.
It compares complete Go diagnostics with Node, emitted JavaScript and sanitized
native bytes. `context_access_alias_lost` must run and be caught on Node and native.
The shared lint gate also discovers and runs the rule and its owned mutant.

    python3 stage1/cohere/lint/rules/base-security-require-context-access/testdata/run_selected.py

Helper captures and their Go source pin are in
`../../helpers/decorators/testdata/coverage.json`.

Verified after merging `origin/lint-batch/wave2-03` at `86c5d841` without rebasing
(merge `67c4c034a`). All 54 upstream cases and both alias witnesses in selected
and all modes agree byte-for-byte on Go, source Node, emitted JavaScript and
ASan/UBSan native (22,597 bytes in the recorded run). The alias mutant compiles,
finishes and is caught on all three port backends. No parser file is changed by
this unit; the previously blocked decorated class expression is included.

The merged batch leaves an unused context import in shared_test.go. The owned
runner applies an import-only temporary Go overlay when needed, so the rule can
be verified without editing or committing that shared file. Ordinary lint go test
still needs that batch harness import fixed by its owner. All shared helper
packages passed before the merge; this package's complete fixtures and mutants
were rechecked after it.
