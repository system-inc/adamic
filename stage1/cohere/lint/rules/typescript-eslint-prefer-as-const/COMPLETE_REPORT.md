Built: six previously claimed .a rule ports with complete fix/suggestion data and owned comparison runners.
Commits: claims e9a72247, 6dce77a4 and 4d5ef284 preceded implementation; ports are separate branch commits.
Validation: all original Go assertions pass; 277 supported fixture/witness files and 77 compiler plus 152 stage1 sources compare byte for byte on all four paths.
Mutants: six rule mutations and two refusal mutations compile and run; byte/runtime comparisons catch all on Node, emitted JavaScript and sanitized native.
Not covered: shared integration, one modified-destructuring parser fixture, split-UTF-8 repairs and the full repository gate.

# Complete repair handoff

The six owned rule directories contain .a listeners, descriptors, real upstream Go adapters, raw witnesses and semantic mutants. `repairs.a` carries independent diagnostic ranges, arrays of fix edits and ordered suggestion arrays. `complete_runner.a` exercises the current three; `backlog_runner.a` also exercises the earlier three blocked claims. Neither changes or overlays an existing Adamic registry, harness, parser, linter or compiler file.

The upstream capture uses virtual Go test copies for the specific six rule test files and a new virtual helper. It runs their real assertions first, then records the source and semantic filename. The independent Go serializer directly invokes the real upstream rules. It preserves every edit range/text and ordered suggestion, applies each suggestion individually, and uses Go's normal fix convergence for safe fixes. Adamic serializes the same data and rewritten source. Tests require exit 0 and empty stderr, including sanitizers; compiler or sanitizer failures do not count as semantic mutant kills.

| Rule | Fixture files | Findings | Compared bytes |
|---|---:|---:|---:|
| no-unnecessary-type-constraint | 44 | 27 | 15302 |
| prefer-as-const | 70 | 27 | 13934 |
| prefer-enum-initializers | 22 | 23 | 13658 |
| no-extra-non-null-assertion | 20 | 12 | 5447 |
| no-confusing-non-null-assertion | 29 | 12 | 4975 |
| no-unnecessary-parameter-property-assignment | 92 | 63 | 56815 |

Corpus: the first three compare 225 files, including all 77 pinned compiler sources and 148 stage1 sources. `validate_delta.py` closes the four .a files added while that run was in progress. The earlier three compare all 229 files directly. Enum initialization produces 680 natural findings, with **1,412,444,566 identical serialized bytes** including all three full suggested rewrites per finding. Other rules have zero natural findings.

| Rule | Native findings/s | Node findings/s | Go findings/s | Input |
|---|---:|---:|---:|---|
| no-unnecessary-type-constraint | 428.76 | 593.66 | 2336.07 | corpus plus 500 witness copies |
| prefer-as-const | 428.51 | 621.73 | 2496.54 | corpus plus 500 witness copies |
| prefer-enum-initializers | 590.68 | 780.85 | 3377.72 | natural corpus, 680 findings |
| no-extra-non-null-assertion | 840.64 | 1124.90 | 4345.51 | corpus plus 500 witness copies |
| no-confusing-non-null-assertion | 413.55 | 583.67 | 2355.71 | corpus plus 500 witness copies |
| no-unnecessary-parameter-property-assignment | 427.22 | 591.76 | 2159.53 | corpus plus 500 witness copies |

Measurements are best of five interleaved complete count-mode invocations, including startup and parsing. Native timing uses an unsanitized release artifact; correctness uses ASan/UBSan and default leak checks. Synthetic supplements are labeled above. These measure the owned runner, not an integrated shared-driver profile.

Reproduce after `source /workspace/adamic-tools/env.sh`, with the compiler checkout at the recorded pin:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-14-typescript-pinned python3 stage1/cohere/lint/rules/typescript-eslint-prefer-as-const/validate_complete.py > /tmp/wave14-complete-run.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-14-typescript-pinned python3 stage1/cohere/lint/rules/typescript-eslint-prefer-as-const/validate_backlog.py > /tmp/wave14-backlog-run.log 2>&1
python3 stage1/cohere/lint/rules/typescript-eslint-prefer-as-const/validate_boundaries.py > /tmp/wave14-boundaries-run.log 2>&1
```

`validate_delta.py <complete scratch directory>` is needed only when .a sources were added after the first corpus froze. The scratch directory is recorded at the top of `evidence/complete.log`. Both complete validators and the boundary validator end in PASS. `evidence/upstream-complete.log` and `upstream-backlog.log` retain every original Go assertion; `complete.log`, `backlog.log`, `delta.log` and `boundaries.log` record the four-way comparisons, mutants and measurements.

Every rule mutant is in its own descriptor. Constraint comma omission, `as never` insertion, one-based enum suggestion changed to position+2, bang deletion shifted one character, wrong wrap closer and retained trailing semicolon all compile, exit 0 and have empty stderr on Node source, emitted JavaScript and sanitized native. Go byte comparison catches each. Two additional tests prove explicit refusals: a mutant bypassing the shared-contract refusal cleanly prints accepted instead of refusing, and a mutant permitting a split-UTF-8 edit cleanly produces byte ranges/rewrites different from Go.

The modified destructuring parameter fixture is the only excluded upstream case. All 93 parameter-property cases were preflighted; only `constructor(public { a }: { a: string })` fails, at offset 33 with `parser slice expected CloseParenToken, got OpenBraceToken`. Its raw source is kept in that rule's gaps directory. The Go assertion passes. A suggestion that blindly eats the first byte of a non-ASCII trailing character is explicitly refused, rather than silently deleting the whole UTF-16 character.

Shared integration remains blocked: the current registry hardcodes .ts in `registry/registry.go:95` and `:187`; the shared Finding lacks independent fix/suggestion arrays and ranges; the shared fixer uses the diagnostic range. The six factories explicitly refuse the unsupported contract. A future handoff must bridge the owned diagnostics into the shared Finding once its contract lands. `origin/codex/lint-harness-dot-a` was absent when origin was fetched. No shared registration generator or test harness was edited.

`bash cloud/setup.sh` exited 1 during package warming because `profile_test.go:32` ranges over the `portFiles` function rather than calling it. Its emitted timing lines were Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s; no completion timing line was emitted. `nproc` is 5. Tools are Go 1.27.1, clang 20.1.8, Node 24.19.0. Workaround: independent owned builders and original Go rule packages work without the broken lint profile test. The full setup output is `evidence/setup.log`.

`gofmt -l cmd internal` is empty. `go vet ./...` exits 1 on the same pre-existing profile compile error, recorded in `evidence/vet.log`. `go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestCountsAreRecorded' -count=1 -timeout 30m` passes in 13.498s (`evidence/filtered-oracle.log`). The complete repository gate and shared lint package gate were not run past that compile blocker.

Automatic approval review rejected the registry generator because it writes shared .generated files. It was not retried or bypassed; read-only inspection records the exact loading contract in `evidence/registry-readonly.log`.

After the six ports were pushed, a fresh fetch found the harness branch at
2650ad595b82220c368631ea13139fad4b306ed6. It now provides .a registration,
emitted-JavaScript comparison, profile compilation and complete suggestion arrays.
The published dependency was inspected read-only; this worker does not edit or
merge shared files. Its fixer still uses diagnostic ranges and its Go serializer
still rejects multi-fix arrays, so independent single fixes and as-const multi-edits
remain blocked even after that dependency is integrated. Gap sources were moved
outside testdata so they are not mistaken for positive conformance witnesses.
