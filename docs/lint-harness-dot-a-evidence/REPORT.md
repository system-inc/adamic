Built: `.a` registration, four-way lint certification, complete suggestion serialization and shared profile snapshots.
Commits: `2650ad5` (early push), `593ced9`, `0efa34b` (latest implementation push); evidence commit follows.
Commands/results: broad parity 759,147 bytes; worker non-null parity 60 cases / 52,752 bytes; details below.
Mutants: emitted-JS-only mismatch and second suggestion edit caught; inherited and registry mutants enumerated below.
Not covered: full repository gate, pinned TypeScript compiler corpus, JSX parsing or timing certification.

## Changes for rule authors

Rename `rules/<slug>/rule.ts` to `rule.a`, keeping its contents unchanged. Remove the old filename. No `rule.json` change is needed. If `mutant.json` specifies `file`, update it; omitted `file` follows the discovered entry. Existing `.ts` entries keep working. Owned helper modules may also use `.a`; update their explicit imports when renamed. Run `go run ./cmd/lint-registry` before manual builds.

For general suggestions, import `Suggestion` and `SuggestionEdit` from `../../suggestions.a` and push suggestions onto the `Finding` returned by `context.report(...)`. Use UTF-16 offsets, upstream ids/descriptions and ordered edits. Existing legacy single suggestions and wave-05's `suggestion-edits:<id>` encoding still work. Multiple alternatives, empty alternatives, distinct edit ranges and delimiter-bearing text use complete suggestion rows. Suggestions stay unapplied even when a finding also carries an automatic fix. No author edits to shared serialization or profiling are needed.

Raw witnesses may end in `.ts.txt`, `.tsx.txt`, `.js.txt` or `.jsx.txt`; captured filenames/script kinds are retained. This transports extensions, not new parser syntax.

## Observations

Started from fetched `origin/main` `ef3d907`, merging registration (`efeb3f6`) and helpers (`3705b64`). Read the wave-01 proposal and wave-05/wave-09 failure evidence. Main's thirty-rule scanner and generated/upstream corpora are retained. The initial five registered rules replace their legacy dispatch. Tied no-var/vars-on-top findings retain Go ordering.

`no-var/rule.a` is a byte-identical rename. `TestDotARename` compares the `.a` port and a scratch `.ts` rename against the same Go result: 646 bytes each, including source Node, Adamic emitted JavaScript and sanitized native. Normal certification independently checks all three executions against Go; execution failure or stderr fails the harness. Go bodies and the upstream submodule were not changed.

The profile builder copies the same recursive module graph, including `.a` and `.generated`, then regenerates registration and loads the copied main. It builds release, counted and profiling binaries. `TestProfileCompilation` exercises that exact path on an owned witness; full external-corpus artifacts remain opt-in.

The scratch wave-05 overlay ([test source](wave05-test.go.txt)) uses the worker's actual two non-null ports (entry/helpers renamed to `.a`), independently unchanged upstream Go adapters and all 60 relevant upstream source/options cases: 52,752 bytes identical across Go/source Node/emitted JS/sanitized native, no exclusions. Worker copies are not part of this branch. The checked-in synthetic suggestion fixture independently exercises three alternatives, multiple edits, zero edits, tabs/newlines/emoji/backslashes and `|`/`:` text. A separate mixed-fix probe checks fixed source while leaving suggestions unapplied.

## Commands and logs

All test output was redirected to files. Toolchain commands began with `source /workspace/adamic-tools/env.sh`.

| Command | Observation / log |
|---|---|
| `bash cloud/setup.sh` | Initial warm-build failed on an unresolved merge marker; rerun after resolution succeeded. [setup-final.txt](setup-final.txt): Go ready 0s; clang ready 0s; Node ready 0s; submodules ready 0s; warm cache 79s; total 79s. `nproc`: 5, cgroup CPU 400000/100000. |
| `go test ./stage1/cohere/lint ./stage1/cohere/lint/registry -count=1 -v -timeout=15m` | [package.txt](package.txt): registry passed; real ordering regression found, then timeout during last volume mutant. This is NOT a passing package run. |
| `go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestVolumeMutants/interface_accepts_type_suffix|TestDecorationOptionMutant|TestCountGuardMutant|TestCommentFoldMutant|TestPositionIndexMutant)$' -count=1 -v -timeout=10m` | [remainder.txt](remainder.txt): corrected broad parity 759,147 bytes (1,273 captured upstream combinations plus generated cases) plus four final parallel checks passed. Slash expression did not select volume parent; separate command below does. |
| `go test ./stage1/cohere/lint -run '^TestVolumeMutants$/^interface_accepts_type_suffix$' -count=1 -v -timeout=3m` | [volume-last.txt](volume-last.txt): final volume mutant passed. |
| `go test ./stage1/cohere/lint -run '^TestLegacyMutants$' -count=1 -v -timeout=10m` | [legacy-final.txt](legacy-final.txt): rerun after fixing baseline ordering; earlier catches against that broken baseline are not credited. |
| `go test ./stage1/cohere/lint -run '^TestProfileCompilation$' -count=1 -v -timeout=5m` | [profile-final.txt](profile-final.txt): PASS 30.45s, release/count/profile built, 654-byte parity, counted output matches Go, 107 allocations / 107 frees. Initial probe rejected expected instrumented stderr; test now checks that separately. |
| `go test ./stage1/cohere/lint -run "^(TestEmittedJavaScriptMismatch|TestDotARename|TestOwnedWitnesses|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix|TestWitnessScriptKind)$" -count=1 -v` (run in focused groups) | [proofs.txt](proofs.txt), [complete-suggestions.txt](complete-suggestions.txt), [mixed-suggestions.txt](mixed-suggestions.txt), [witness-kind.txt](witness-kind.txt). |
| `go test -overlay /tmp/lint-dot-a-wave05-overlay.json ./stage1/cohere/lint -run "^TestWave05DefaultHarness$" -count=1 -v` | [wave05.txt](wave05.txt): 60 cases / 52,752 bytes, PASS. |
| `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v` | [oracle.txt](oracle.txt): PASS, one-byte mismatch caught by Node/native oracle. |
| `go vet ./...`; final `go vet ./stage1/cohere/lint/...` | Both exit 0. `gofmt -l cmd internal` and touched Go files clean. |
| `go run ./cmd/adamic types stage1/cohere/lint/testdata/serialization/rule.a`; types main.ts | Both exit 0. Scoped cohere formatting applied. |

## Mutants and independent checks

The emitted-artifact mutant appends `console.log('planted emitted JavaScript mismatch')` only to Adamic's `.mjs`; the subprocess must fail the real comparison with the emitted-JavaScript label. Source Node/native remain baseline. The second-edit mutant changes suggestion range 9..10 to 9..11: Go disagreement caught on source Node, emitted JS and native. Registry rejects stale `.ts` beside `.a`.

Inherited checks are intentionally described with the sides actually exercised; ordinary parity includes all three execution paths even where an older mutant test only exercises two.

| Mutant | Check that caught it | Evidence |
|---|---|---|
| second suggestion edit mutant | Go disagreement on Node, emitted JavaScript, native | [package.txt](package.txt) |
| debugger fix suppressed | Go disagreement on Node, emitted JavaScript, native | [package.txt](package.txt) |
| empty function body reported | Go disagreement on Node, emitted JavaScript, native | [package.txt](package.txt) |
| suggestion applied as fix | Go disagreement on Node, emitted JavaScript, native | [package.txt](package.txt) |
| var declaration suppressed | Go disagreement on Node, emitted JavaScript, native | [package.txt](package.txt) |
| duplicate case suppressed | Go disagreement on Node, emitted JavaScript, native | [package.txt](package.txt) |
| wrong descriptor kind | Go disagreement on Node, native | [package.txt](package.txt) |
| finish-hook omission | Go disagreement on Node, native | [package.txt](package.txt) |
| ignored decoded-option mutant | Go disagreement on Node, native | [package.txt](package.txt) |
| predicate kind omitted | Go disagreement on Node, native | [package.txt](package.txt) |
| console receiver widened | Go disagreement on Node, native | [package.txt](package.txt) |
| for update option ignored | Go disagreement on Node, native | [package.txt](package.txt) |
| method listener lost | Go disagreement on Node, native | [package.txt](package.txt) |
| wrapper Number ignored | Go disagreement on Node, native | [package.txt](package.txt) |
| enum bitwise option ignored | Go disagreement on Node, native | [package.txt](package.txt) |
| enum declaration listener lost | Go disagreement on Node, native | [package.txt](package.txt) |
| condition polarity reversed | Go disagreement on Node, native | [package.txt](package.txt) |
| return parentheses option reversed | Go disagreement on Node, native | [package.txt](package.txt) |
| decoration range collapsed to one character | Go disagreement on Node, native | [remainder.txt](remainder.txt) |
| fold mutant | Go disagreement on Node, emitted JavaScript, native | [remainder.txt](remainder.txt) |
| position mutant | Go disagreement on Node, emitted JavaScript, native | [remainder.txt](remainder.txt) |
| interface accepts type suffix | Go disagreement on Node, native | [volume-last.txt](volume-last.txt) |
| finish hook omitted | Node/native findings against Go | package.txt |
| root-only import rewrite | Node/native module-loading rejection | package.txt |
| count guard +1 | Node/native count mismatch; ordinary output still identical | remainder.txt |
| unknown field, duplicate public name, unsafe public name, missing named export, invalid AST kind, missing oracle export, missing listener, missing factory, missing class, missing finish hook | Registry validation rejects each descriptor | package.txt |
| duplicate oracle adapter function | Registry validation rejects | package.txt |
| stale `.ts` beside `.a` | Ambiguous-entry rejection | proofs.txt, package.txt |
| emitted-artifact extra line | Real emitted-JavaScript comparison fails subprocess | proofs.txt |
| second suggestion edit range | Go mismatch on Node/emitted JS/native | complete-suggestions.txt |
| one byte changed in oracle fixture | Independent Node/native oracle misses | oracle.txt |

Legacy mutants after baseline repair are appended below with final outcomes. The original ordering regression itself was caught by `TestRulesAgree` on source Node (package.txt); the corrected broad comparison includes all three execution paths (remainder.txt).

## Limits and delivery

No full-repository runtime gate or TypeScript v6.0.3 compiler corpus certification is claimed. Fetching the pinned compiler source failed repeatedly with Git authentication errors; the normal suite skips those opt-in tests. An actual JSX probe `<span/>` failed explicitly in the existing stage 1 parser (`expected GreaterThanToken, got SlashToken`); see [jsx-parser-limit.txt](jsx-parser-limit.txt). The final script-kind witness uses supported syntax in a `.tsx` file. Parser-recovery refusals remain explicit and visible in the broad parity log. Throughput/artifact timing benchmarks are not claimed.

Push `2650ad5` succeeded early. Later pushes briefly failed twice with `could not read Username for https://github.com`; the required fallback was `/tmp/lint-harness-dot-a-593ced9.patch`. A later retry successfully pushed both `593ced9` and `0efa34b`, so that patch is superseded. No pull request was opened. Compiler implementation files reserved by the unit were not edited.

## Final legacy mutant outcomes

`TestLegacyMutants` passed in 268.354s after the baseline repair. Every row below disagrees with independent Go on all three execution paths.

| Mutant | Caught by |
|---|---|
| control statement omitted | Node, emitted JavaScript, native |
| destructuring hole reported | Node, emitted JavaScript, native |
| nested generator owns outer yield | Node, emitted JavaScript, native |
| await crosses function boundary | Node, emitted JavaScript, native |
| option ignored | Node, emitted JavaScript, native |
| regex fix eats extra byte | Node, emitted JavaScript, native |
| boolean inverse changed | Node, emitted JavaScript, native |
| comment self directive exemption removed | Node, emitted JavaScript, native |
| bom removes two marks | Node, emitted JavaScript, native |
| label option widened | Node, emitted JavaScript, native |
| directive prefix ignored | Node, emitted JavaScript, native |
| comma chain reports inner | Node, emitted JavaScript, native |
| empty placeholder reported | Node, emitted JavaScript, native |
| overlap winner misreported | Node, emitted JavaScript, native |

See [legacy-final.txt](legacy-final.txt) for exact differences. Profile test and final registry/vet checks passed; the initial package timeout was resolved with the focused commands above.
