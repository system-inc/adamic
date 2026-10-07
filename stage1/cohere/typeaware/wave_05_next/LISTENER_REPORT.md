Built: numeric listenerKinds declarations added to all six completed wave-05 rule modules; no new claims.
Commits: implementation follows published landing tip 40399354 on current main e8ba3d5d; the enclosing commit records these declarations and evidence.
Commands and outputs: production Go listener/enum comparison PASS; both drivers compile; both frozen corpora match Go; normal/sanitized positive controls PASS.
Mutants: six changed numeric-kind metadata mutants caught by declaration comparison; existing rule verdict and CFG mutants remain recorded in the latest full landing validation.
Not covered: numeric dispatch, elimination of string-kind reads/per-rule node fetches, three blocked React ports, fresh full corpus sanitizer/mutant reruns, full gate or new timing benchmark.

# Numeric listener declarations

Each rule exports `listenerKinds: readonly number[]`. Values are the pinned typescript-go parser's `ast.Kind` numbers, rather than stock TypeScript enum values. They match the unchanged production Go listener registrations:

| Rule | Numeric kinds |
| --- | --- |
| no-unnecessary-type-parameters | 220, 263, 219, 180, 181, 186, 185, 174, 175, 264, 232 |
| no-useless-assignment | 307 (SourceFile) |
| restrict-template-expressions | 229 (TemplateExpression) |
| no-uncleared-race-timeout | 214 (CallExpression) |
| no-process-exit-after-output | 307 (SourceFile) |
| require-blocking-standard-streams | 307 (SourceFile) |

`python3 stage1/cohere/typeaware/wave_05_next/check_listener_kinds.py` compares every declaration with the production Go rule's listener map and the Go enum's generated numeric assertions. Changing the first kind in each of the six declarations by one is rejected by that comparison. These are metadata mutants, not replacements for native rule-verdict mutants.

No dispatch behavior changed. The shared native `ParseNode` at `stage1/typescript/parser/nodes.ts` exposes only `readonly kind: string`. There is no numeric kind field for a handed-node callback to consume. The existing drivers still call `run()` and the established implementations still read string kinds and fetch nodes. The user explicitly requested declarations even while today's driver ignores them; those declarations are now present. The complete speed rule remains unmet until the shared parser and kind-indexed driver provide the numeric handed-node contract. No shared parser, registration generator or harness was edited under Ahra's restriction to own rule directories.

# Validation

All subprocess output was written directly to files. Summary/build logs are retained in [listener_evidence](listener_evidence/); individual canonical streams are in `/workspace/wave-05-listener-validation`.

- The scratch cohere formatter with .a registration, previously documented in OUTPUT_REPORT.md, formats all six changed files: exit 0. Its output explicitly says type/lint checks were not run.
- Fresh `go build ./cmd/adamic`, then both drivers built with the pinned checker archive: exit 0. Both also rebuilt with the existing ASan/UBSan checker archive and `--sanitize`: exit 0.
- Updated first driver: controls 20 findings / 6,741 identical bytes; repository 13 findings / 24,608 bytes; compiler 70 findings / 23,501 bytes. All release outputs equal the independent production Go oracle. Sanitized controls equal Go, with empty native stderr.
- Updated continuation driver: output controls 28 findings / 17,718 identical bytes; repository zero / 18,485 bytes; compiler zero / 5,318 bytes. Release equals Go; sanitized controls equal Go with empty native stderr.
- DOM and Node timer-positive controls: 12 findings / 7,800 identical bytes each on Go, native and sanitized native. The first timer-check invocation incorrectly rejected the Go oracle's timing-only stderr despite exit 0; the corrected check permits its `cohere: load_ns=` timing line and still requires empty native stderr. No finding output was normalized.
- `git diff --check`: clean.

Full corpus sanitizers, rule-verdict/CFG mutants, released handles, bridge tests and filtered Node oracle last passed on the same current main base in LANDING_2_REPORT.md. This metadata-only change repeats compilation, normal corpus parity and sanitizer-positive controls; it does not claim a fresh full gate. No new speed measurement is claimed. Previous uncontended native/Go medians remain in the original reports: first three compiler 14.029s/0.588s and repository 0.532s/0.184s; continuation compiler 2.380s/0.348s and repository 0.276s/0.125s.

The three existing React claims remain blocked on shared JSX parser integration, as documented in LANDING_2_REPORT.md. No additional rules were claimed. Publication stays on codex/typeaware-wave-05 only.
