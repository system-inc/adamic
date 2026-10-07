Built: prefer-promise-reject-errors, prefer-regex-literals and prefer-rest-params in native Adamic .a files; wave 24 now has nine ports.
Commits: prior ports cce6c361; third claim aedfc498; implementation is the commit containing this report.
Checks: 439 cases / 295 findings / 176 suggestions and 364 frozen roots match Go byte for byte, normally and under sanitizers; bridge, checker, refusal and static gates pass.
Mutants: three rule and two raw-question mutants compile and exit 0; only diagnostic bytes catch them. The new query rejects a released handle with exact panic 70.
Uncovered: non-default options, the full repository gate, shared emitted-JavaScript/profile and production CLI .a integration, and arbitrary upstream/runtime populations.

## Selection and claim

All six previously claimed rules were tested and pushed at cce6c361 before selecting more.
Fetched all origin branches, 356 refs, and examined 33 distinct Markdown claim blobs.
Excluded the 26 ports on main/bridge and every claimed checker rule on every origin ref.
The first remaining entries in the combined compiler-plus-repository ranking linked by VOLUME_REPORT were positions 158, 159 and 160:

1. prefer-promise-reject-errors
2. prefer-regex-literals
3. prefer-rest-params

Each had zero recorded findings. The full audit and exclusions are in evidence/selection.json.
The tips are recorded in claims/wave-24.md. Claim commit aedfc49830c8d14010fc7fb42e187383addac7c4 was pushed before writing implementation code. No further rules were claimed.

## Implementation and ownership

Each rule has its own .a file in this directory. Other new .a files hold the regex character walk, reference tracing, and raw direct-symbol decoder.
No shared registration generator, existing shared test harness, protected compiler file, previous rule implementation or submodule was edited.
The only previously existing file changed is wave 24's owned checker router: one question-map entry plus one mode-guard addition. A blank line preserves gofmt alignment of the existing entries.

prefer-promise-reject-errors identifies ambient Promise declarations, static reject property accesses and executor rejection callbacks by live symbol identity. It traverses nested callbacks and parameter defaults, collects duplicate named parameter symbols, preserves shadowing, and decides whether the rejection syntax could be an Error natively.

prefer-rest-params asks the checker for the direct accessor's raw symbol identity and declaration count. Zero declarations on a nonzero symbol identify the implicit object; dotted access on that identifier is exempt. This is deliberately different from value provenance: Go reads the shorthand property's own symbol for {arguments}, so it produces no finding there. The comparator caught using the value symbol, which would have reported it.

prefer-regex-literals follows unmodified global RegExp and the default global objects through members, aliases, assignments, default bindings, object patterns, transparent type wrappers, logical/conditional expressions and comma values. The trace is flow insensitive and retains duplicate findings.
Native static-string extraction distinguishes cooked literals/templates from String.raw templates and preserves shadowing. Suggestion construction implements the production comment and preceding-token gates, printable allowlist, flags, pattern walk, group/quantifier checks, class ranges and v-class boundaries, character escaping and adjacent-token padding.
Go's narrow pattern semantics, including omissions and conservative withholding, are retained; Node's full regex execution grammar would answer a different question.
Call/new argument counts come from parser list metadata, so type arguments and optional-chain tokens do not become rejection reasons or regex patterns.

## Bridge question

raw_symbol_declaration_count.go and raw_symbol_declaration_count.a implement raw-symbol-declaration-count.
After the existing version/question header it carries exactly two naturals: direct GetSymbolAtLocation identity (0 when absent) and declaration count (0 when absent).
It accepts only an exact Identifier and no suffix. Native code decides the rule predicate; Go returns no lint verdict, message or edit.
The direct checker test distinguishes unresolved top-level arguments, implicit arguments, and the direct shorthand-property symbol, and refuses the wrong kind and suffix.
Frames retain the strict existing integer/header/trailing-field guards. Program-owned IDs never outlive the program.
The other rules reuse wave 24's raw declaration provenance; no existing fact implementation was changed.

## Oracle and controls

The independent oracle in testdata/oracle.go loads its own program, imports no bridge implementation, walks the production AST and calls the three unchanged Go core rules with nil/default options.
Diagnostics retain duplicates and compare complete canonical fields: byte spans, rule, message ID, message text, every fix and suggestion and all suggestion edit fields, plus file headers.
Promise and rest propose no fixes or suggestions. Regex proposes suggestions only; no automatic fixes are promoted or applied.

There are 102 explicit controls and 337 source strings extracted from the pinned production test AST (327 distinct imported sources). testdata/extract.go and imported.json preserve the extraction; source strings were not retyped.
Regex corpus rows with nonempty option JSON are excluded by the extractor. Promise/rest source strings are replayed under defaults, not with their original option vectors; this is not a claim of full option-matrix coverage.
The 439 cases produce 68 rejection findings, 17 rest findings and 210 regex findings, with 176 suggestions containing one edit each.
Controls include clean cases, shadows, nested and duplicate executor bindings, shorthand symbols, type/parenthesis wrappers, optional and generic calls, raw/cooked patterns, aliases and destructuring, dynamic/template keys, written globals, empty patterns, comments, invalid/repeated flags, malformed patterns, ranges, set/property escapes, control characters, Unicode and CRLF.

| Population | Findings | Suggestions | Identical bytes, normal and sanitized |
| --- | ---: | ---: | ---: |
| 439 cases | 295 | 176 | 158203 |
| Frozen repository, 287 roots | 0 | 0 | 18485 |
| TypeScript compiler, 77 roots | 0 | 0 | 5010 |

Zero corpus findings are not treated as evidence of an active rule; each rule's positive-control presence is mandatory.
ASan, UBSan and LeakSanitizer runs retain complete bytes with empty native stderr. Declaration roots are retained by both independent loaders.
The frozen manifests come from validation-coverage and do not add the new ports. Evidence stores their portable paths and final source hashes.
Cohere remains 715ba94f3608a6500086b1076ce5cb7e51b836db; typescript-go remains 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Compiler corpus is TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

## Mutants and guards

All five new mutants compiled, exited 0 and had empty stderr. The independent Go diagnostic bytes alone caught them:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| promise | Reverse the could-be-Error exemption | 702 |
| rest | Reverse the direct declaration-count predicate | 7855 |
| regex | Append x to every nonempty suggested literal pattern | 11074 |
| direct-symbol-question | Add one to the direct declaration count | 7855 |
| provenance-question | Mark declaration files as non-declaration files | 50 |

The new raw query after release exits 70 with exact adamic: panic: invalid or released checker handle stderr.
The bridge regression reran its independent C ABI controls: 100 query buffers, outputs survive release, nonreused handles, 162 source positions and 3261 bytes under sanitizers.
It catches input/output length +1 with ASan, retained stale registry handles with the refusal assertion, source-file type-position substitution with byte mismatch, missing link opt-in with refusal, missing C output free with LSan, and a region allocation moved to the heap with LSan.
Request guards reran wrong-kind and unknown-question removals, each caught by required panic 70.
The 19 wire mutants each panic 70: empty, length-negative, length-short, trailing, version, question, not-integer, rounded-integer, noncanonical-integer, negative-natural, bad-boolean, zero-id, negative-tuple, missing-root, missing-constraint, missing-element, duplicate-id, missing-link-17 and missing-link-18.

During development the independent bytes exposed optional-chain child positions, direct versus shorthand-value symbols, and angle-bracket assertion child order. They were corrected before the final run.
The earlier nested-constructor workaround remains: dependencies are constructed before rule constructors. The unsupported lastIndexOf overload was replaced with a source slice and supported one-argument call. Failed builds are not counted as mutant catches.
The initial direct-question test config had no supported config input; adding an explicit declaration control fixed it. A transient executor disconnect on the refusal command was retried successfully.

## Commands and output

Setup succeeded: Go ready 0s, clang ready 0s, Node ready 0s, submodules 1s, cache warm 22s, done 22s. nproc prints 5, cpu.max is 400000 100000 (four-core quota), memory 17.6GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.
All test output went directly to log files, never through a pipe.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave-24-third/validate.py \
  /workspace/wave-24-third-final --stage0 /workspace/wave-24-final-complete/adamic \
  --compiler /workspace/wave-24-corpus > /tmp/wave-24-third-final.log 2>&1
# PASS: controls, both corpora, normal/sanitized, five byte-only mutants, released raw query.
go test ./bridge/tsgo/... -count=1 -v > /tmp/wave-24-third-bridge-final.log 2>&1
# PASS: bridge 96.660s; checker 0.394s.
go test ./stage1/cohere/typeaware \
  -run 'TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags' \
  -count=1 -v > /tmp/wave-24-third-refusals-final.log 2>&1
# PASS 57.718s.
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /tmp/wave-24-third-vet-final.log 2>&1
# PASS, empty log. Touched Go files also have empty gofmt -l output.
```

Stage0 is the previously built compiler. No compiler/runtime implementation changed. It can be rebuilt with go build -o /tmp/adamic ./cmd/adamic and supplied to --stage0.
The owned validator builds normal and sanitized C archives, native suites, scratch mutants and the independent Go oracle through an overlay inside cohere. It changes no production rule or shared harness.

## Native time against Go

Three alternating count runs after all builds and tests stopped:

```sh
python3 bridge/tsgo/profile/volume_bench.py \
  /workspace/wave-24-third-final/native /workspace/wave-24-third-final/oracle \
  /workspace/wave-24-third-timings \
  --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-24-third-final/repository.manifest \
  --corpus compiler /workspace/wave-24-corpus/src/compiler/tsconfig.json /workspace/wave-24-third-final/compiler.manifest \
  --rounds 3 > /tmp/wave-24-third-timings.log 2>&1
```

| Corpus | Native process median | Go process median | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.296785s | 0.089554s | 3.31x |
| Compiler | 2.084926s | 0.385971s | 5.40x |

Observation: native is slower on both zero-finding corpora. Whole processes include program loading, native parsing/indexing, bridge work, walking and teardown; no isolated per-rule/C-call speed claim is made. Raw samples and phase stderr are in evidence.

## Limits

These are default-option ports. allowEmptyReject:true and disallowRedundantWrapping:true are not implemented or validated; the runner exposes no option input. Rest has no option surface.
No full go test ./... gate, full production CLI .a integration, configured suppressions, edit application, shared emitted-JavaScript rule/profile comparison, or complete arbitrary TypeScript/JSX/import/runtime population was run.
The shared harness branch is now available at origin/codex/lint-harness-dot-a (f4d98cab); it was not merged or edited. The owned validator already loads .a modules and compares complete suggestion records against Go, so it did not block the native result.
The earlier native/Node/emitted-JS oracle evidence for unchanged compiler/runtime remains in the original wave report and was not rerun here.
The evidence establishes the frozen corpora, 439 explicit/imported source cases, direct raw facts, five successful byte-only mutants and ownership/refusal guards. It is not a proof over all programs.
No PR was opened. No further rules were claimed.
