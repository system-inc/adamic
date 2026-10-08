# Census review for the landed cohere pin

Merged area/stage1-lint d45a323be353add8ed1c593f51e781d3f67ade18, which contains
main 48c05d09 and cohere 7945d102a6c18dd36adf9114a758ce646e8b2359.
Compared all standard-library regexp sites against 715ba94f3608a6500086b1076ce5cb7e51b836db
before updating sites.json. census-review.json retains the old/new rows, the exact
introducing/removing diffs, and all 66 line-only moves with each affected file's
commits in the requested range. The recorded corpus was retained unchanged.

| Change | Site | Cohere commit | Review |
| --- | --- | --- | --- |
| Added | nexus/localization_no_untranslated_value.go:408 | bb39e2dd | Locale stem is recognized with `^[a-z]{2,3}(?:[-_][A-Za-z0-9]+)*$`. Translate the absolute Go end anchor to `(?![\s\S])`, observe with gu. New positive/negative fixture included. |
| Removed from Go regexp census | tailwind/class_literals.go:167 at old pin | 0f2797dd | NewClassLiteralReader now delegates to newNamePatterns. Current line 200 compiles through esregexp.Compile(pattern, "") for JavaScript semantics. This is an explained engine migration, not lost matching behavior. |
| Edited pattern texts | None | None | All 106 surviving source expressions and Go patterns are identical; their JS translations and flags are also identical. |
| Line-only moves | 66 sites | See census-review.json file history | IDs and provenance comments updated; no pattern change is inferred from a line move. |

No unexplained disappeared row. The census is specifically Go's standard-library
regexp Compile/MustCompile sites, so Tailwind's esregexp replacement is named here
but not relabeled as a Go regexp site. Its old fixture remains recoverable in the
previous commit, and the engine change is backed by the recorded upstream diff.

Total remains 107: 89 MustCompile (six dynamic), 18 Compile, 83 fixed, 24 dynamic.
Shapes: plain 77, (?i) 4, (?s) 1, (?m) 1, dynamic 24. No selected instance is
outside the port's shapes. Fixture inputs and Node observations were regenerated;
every selected instance still has a positive Go match. Source line movement also
required the named helper exporter to locate by source pattern instead of old lines.

TestShapeFixtures is now a real package _test.go entry point. It builds gate.go,
checks its logged count against the reviewed table, executes all fixtures on source
Node and area constant-string emitted JavaScript, and attempts every runtime native
leg. Only the named nonconstant-pattern refusal is accepted, labeled
`awaits codex/regex-runtime-compiler`; when lowering succeeds sanitized native and
runtime emitted JavaScript are required. Its subtests prove the one-character
translation mutant is killed by comparison and the forced native control exposes
the pending refusal. Set ADAMIC_REGEX_TRANSLATION_MUTANT=1 to make the ordinary
Go test itself fail on the planted translation, without modifying tracked fixtures.

Regeneration after the review:

```
go run stage1/cohere/lint/regex/testdata/inventory.go cohere/internal/lint/rules > stage1/cohere/lint/regex/sites.json
python3 stage1/cohere/lint/regex/testdata/generate.py --table-only > /tmp/regex-seat-table-generation.log 2>&1
go run stage1/cohere/lint/regex/testdata/shapes/build.go stage1/cohere/lint/regex/testdata/shapes > /tmp/regex-seat-generation.log 2>&1
node stage1/cohere/lint/regex/testdata/shapes/generate.mjs > /tmp/regex-seat-node-generation.log 2>&1
go test ./stage1/cohere/lint/regex -count=1 -v -timeout=20m > /tmp/regex-seat-tests.log 2>&1
ADAMIC_REGEX_TRANSLATION_MUTANT=1 go test ./stage1/cohere/lint/regex -run '^TestShapeFixtures$' -count=1 -v -timeout=20m > /tmp/regex-seat-go-mutant.log 2>&1
```

The last command must exit 1 and name the source Node fixture mismatch at
base/consistency_require_pagination_argument_name.go:27. Compiler/type failures
do not count as a killed translation mutant.

Observed final results:

- Full `go test ./stage1/cohere/lint/regex -count=1 -v -timeout=20m` passed in
  64.454s. TestShapeFixtures took 51.43s and explicitly logged 107 fixtures on
  source Node and emitted JavaScript, 107 runtime Node companions, and 107 native
  legs labeled `awaits codex/regex-runtime-compiler`.
- Its ordinary one-character-mutant subtest passed by catching the comparison,
  and its forced-native subtest passed by exposing the precise named refusal.
- The separate ADAMIC_REGEX_TRANSLATION_MUTANT=1 Go test exited 1, failed
  TestShapeFixtures, and named the intended source Node mismatch. Node executed
  cleanly; no type error, compiler refusal or warning killed that mutant.
- The inventory test passed at the actual cohere 7945d102 checkout. The 83 fixed
  patterns also passed Go/Node/emitted-JS/sanitized-native comparison: 125,809 match
  records and 2,355,726 identical bytes. Their original mutant remains caught on
  all three execution backends.
- go vet ./stage1/cohere/lint/regex exited 0 with no diagnostics.

No unexplained census change. No runtime/compiler code or rule migration was edited.
Native runtime compilation remains pending, and the full repository gate was not run.
The older library-specific JavaScript evidence is historical; the normal package
gate requires no external checkout or prebuilt library compiler.
