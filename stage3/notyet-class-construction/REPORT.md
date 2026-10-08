Investigated all 14 kinds; no production compiler change or oracle fixture is retained.
Base is b410340dc8f889b5799c3bc519117c63def3aa24; replay tooling merge is 1bf805f81f6a4f1f539ff20bfd7db5a7c5d45cba.
All 13 field kinds reproduce at property in object.go, using the second example for string | number | undefined; focused inheritance tests pass.
The withdrawn construction experiment's initializer mutant failed stdout; its reference-bitmap mutant survived, and the later planned narrowed-read mutant was not run.
Covered roots: 0 lowered, 0 refused for a ruling, 47 skipped; 46 belong to another worker's function and the remaining example does not reproduce its old stop.

## Ownership blocker, including the clarified territory rule

The same diagnostic text occurs in both inheritanceConstructor and property. Grouping by text attributed these examples to the wrong function. A measurement-only diagnostic overlay uses runtime.Caller at notYet to record the actual caller for the exact requested source position. All reproduced field-type examples name `internal/lower/object.go::property`, at line 484 on the requested compiler base. The second checker example confirms the thirteenth kind. The first checker example instead reports `reading result`, preceded by a non-null stop and `reading initializer`.

The user clarified that IR, backends, flow walkers and called lowering helpers are available, while another worker's internal/lower function stays theirs. The required branch check found `origin/codex/notyet-object-property` at `5a8705151a7f03c5d03113eba8bf3705f32b494d`. Its commit `Lower own-property optional chains with early returns` edits property itself. Thus expanding backend scope does not remove this ownership conflict. No further production edits were made after that check. Reassignment or a coordinated ownership transfer is required; no request to relax runtime or backend permissions is needed.

Observed source nodes include `node.comment` in binder and `option.type` in commandLineParser, not class declarations or constructions. A class-free scratch `.a` reduction printed `text1`, `comment2`, `missing` on source Node; production lowering stopped at its interface-field read with `a field of type string | readonly Comment[] | undefined`. Native, generated JavaScript and sanitizer validation cannot pass that reduction until property learns the operation.

## Per-kind disposition, largest first

All field-type rows below are skipped for the same verified ownership conflict. Counts are the supplied table's root counts, not a claim to have replayed every individual root.

| Table roots | Kind after `a field of type` | Status and evidence |
| ---: | --- | --- |
| 23 | string \| NodeArray<JSDocComment> \| undefined | Skipped: property; binder 2121:20 reproduced |
| 4 | "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | Skipped: property; commandLineParser 2599:13 reproduced |
| 4 | NodeArray<ParameterDeclaration> \| readonly JSDocParameterTag[] | Skipped: property; checker 16202:62 reproduced |
| 3 | string \| number \| undefined | Skipped: property; checker 51074:20 reproduced; 47945:13 now reaches reading result |
| 2 | AnyBuildOrder \| undefined | Skipped: property; tsbuildPublic 656:12 reproduced |
| 2 | boolean \| (() => boolean) \| undefined | Skipped: property; moduleNameResolver 512:46 reproduced |
| 2 | false \| string[] \| undefined | Skipped: property; moduleNameResolver 2235:23 reproduced |
| 1 | "boolean" \| "list" \| "number" \| "object" \| "string" \| Map<string, string \| number> | Skipped: property; executeCommandLine 375:21 reproduced |
| 1 | "boolean" \| "number" \| "object" \| "string" \| Map<string, string \| number> | Skipped: property; commandLineParser 1896:13 reproduced |
| 1 | 0 \| boolean \| undefined | Skipped: property; checker 39371:13 reproduced |
| 1 | boolean \| (() => boolean) | Skipped: property; watchUtilities 741:23 reproduced |
| 1 | string \| false | Skipped: property; moduleNameResolver 2418:12 reproduced |
| 1 | string \| false \| undefined | Skipped: property; moduleNameResolver 2415:9 reproduced |
| 1 | a base that isn't a declared class | Skipped: exact signature does not reproduce at transformers/utilities 441:33. Selected class reaches the emit-node union stop at 423:34 and reading autoGenerate at 424:18 instead. No independent base failure or design refusal was established. |

## Pins, input and commands

Resolved the requested area/compiler head with `git ls-remote origin refs/heads/area/compiler`: `b410340dc8f889b5799c3bc519117c63def3aa24`. Created only `codex/notyet-class-construction` from that explicit unit base and merged `9a1f14c5d994aa855625e7cfa295677060348fec`. Initial origin/main `d65e2d5e65197969480c2417832ca58fa007c933` was already an ancestor. For delivery, fetched current origin/main `ef3141e9b1152ab51b51497f8ce3a2799449c8a3` and merged it as `f9229127485f162bd0b8225c79ba89fd45705ccf`. Its changes leave internal/lower, IR and both backends identical to the replayed compiler. The same focused tests passed again in 0.277s; evidence/delivery-tests.log.txt preserves the output. No area or main branch was modified.

Setup ran with `GOPROXY='https://proxy.golang.org|direct'`. The initial run's package enumeration overlapped the branch checkout and failed with missing typed-array, parameter-property and census helper symbols. A fresh package enumeration confirmed the files and symbols exist. Retrying on the stable branch passed: Go ready 0.138s, Node ready 0.148s, submodules 0.338s, markdown dependencies 0.343s, clang ready 1.943s, Go build ready 167.130s, test binaries deferred 167.523s, cache warm 167.530s, done 167.843s. `nproc=5`, cgroup CPU quota `400000 100000`. Used the printed `/workspace/adamic-tools/env.sh` thereafter.

Reconstructed the original adaptation snapshot with `git archive 9d534d3a31814f1a192a528e701f6c2ea7c910bc stage3`, then ran that snapshot's `stage3/apply.sh /tmp/class-construction-census-source`. After completion, all 81 file byte lengths and SHA-256 hashes matched `stage3/meter/runs/20261008T035244Z.latent-full/tsc/source-manifest.json`. Earlier exploratory replays overlapped adaptation and were superseded by the stable-input baseline and caller-trace runs. `npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund` supplied the locked Node types.

Built the guarded worker using `make_overlay.py` and `go build -buildvcs=false -overlay=... ./stage3/census/latent/replay/worker`, as documented by replay's README. Each invocation used `-project /tmp/class-construction-census-source/src/tsc/tsc.ts`, exact `-where`, `-kind NotYet`, and exact `-reason`. [Argument ledger](evidence/replay-arguments.py), [selected findings](evidence/replays.json), and individual `replay-*.log.txt` files preserve positions, reasons and caller traces. The diagnostics tracing source exists only in `/tmp/class-construction-trace-diagnostics.go`, mapped through `/tmp/class-construction-trace-overlay.json`; it adds runtime.Caller logging gated by `LATENT_TRACE_WHERE`. Ordinary Load/Lower output guards remain enabled. No backend is invoked by measurement replay.

The final unchanged compiler passed:

```text
go test ./internal/lower -run '^(TestInheritanceKeepsCheckerConstructorRules|TestInheritanceRefusesUnsoundOverrides)$' -count=1 -timeout 10m
ok github.com/system-inc/adamic/internal/lower 0.251s
```

Test output was written to logs. No whole-package test or full gate ran. No new fixture remains, so counts.md was not regenerated. No native runtime C helper was added.

## Withdrawn construction experiment and mutants

Before the caller attribution was settled, an experiment allowed boxed union construction and added a class-only property read hook. Its `.a` fixture exercised initializer order, string/array/undefined storage, source Node, generated JavaScript, native and sanitizers. The original compiler stopped at construction; the experiment's focused oracle passed in 22.671s. However, replay of the requested largest example still produced its original field-read stop. This established that the experiment did not resolve the assigned root.

An initializer-execution mutant changed the constructor's initialized body to an empty slice. It failed against Node in both backends: expected `text0`, `comment1`, `missing` plus field/constructor order; mutant printed three missing values and only constructor events. The reference-bitmap mutant excluded Union from the initial layout metadata, but the fixture passed: it did not prove that check. The runner stopped there; the planned narrowed-read mutant was not run. A separate direct-narrowed-field boundary test had passed before mutation. All mutant edits were restored, then the entire experiment, tests and fixture were removed. These are historical experiment observations, not verification of a shipped change.

The evidence commit retains only this report, argument/observation ledgers and logs. It establishes the scope mismatch and ownership blocker; it does not claim any lowering lesson is complete or any source site is a proven census echo.
