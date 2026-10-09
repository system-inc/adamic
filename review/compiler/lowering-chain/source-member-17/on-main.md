Step 21's five exception rulings are rebuilt on `d72728e5`, with terminal checks preserved.
Eight scout own commits replayed; no scout merge or unlanded branch was merged.
Lower, IR, flow, the whole uncached oracle and both stage3 runs pass against the same-machine base.
All 36 scout mutants and four replay-preservation mutants are caught by their intended checks.
Error stack/options/reflection, unrepresented library failures, allocation exhaustion and whole-tsc compilation remain outside this proof.

# Exceptions on pinned main

Delivery branch: `compiler/exceptions-21`. Base: `d72728e570fe09d91cf55b37b564dbad0fefc24d`. Scout: `3ce869e508790c09891edb855772eec119784527`. This serves roadmap step 21, task #cxr5x2v. Main and area branches were not pushed or merged into. The newer origin/main was not substituted for the requested comparison pin.

The implementation admits owned tagged throws, real narrowing of unknown catch bindings, identity-preserving saved Error/rethrow paths, nominal Error subclasses with the shared prefix, and the scout's represented catchable library failures. Uncaught language payloads flush output, release ownership and exit 1 without backend rendering or conversion. Compiler-inserted checks exit 70 before catch/finally.

## Replay

| Scout own commit | Replay commit |
|---|---|
| `2ea38a94` | `128bb5bc` |
| `e9978e0e` | `5507aeb7` |
| `dd217f32` | `6cb994c9` |
| `d59cb6a8` | `a0bc73d3` |
| `48f01bea` | `6762e024` |
| `df121a79` | `f9546223` |
| `0418ff7e` | `eec5fd92` |
| `3ce869e5` | `2d7c2612` |

The scout merge `258a8f2c` was excluded. All eight own commits replayed without manual conflicts. Main's class Definition metadata and current interface-view behavior remain in the merged source; the full IR guard and view tests pass.

The replay exposed stale NotYet tests for catchable library calls. Their negative cases moved to positive admission coverage, held by the scout's library fixtures. Source Node's old global uncaught-to-panic normalization also hid the difference between language exceptions and terminal checks. It is now explicit for 35 unchanged legacy witnesses, with root source hashes and all static-import hashes for the five cyclic entries. A changed entry or imported source keeps ordinary Node behavior. This catalog preserves the base observations; it is not a proof that arbitrary thrown values are terminal.

The source runner uses `--terminal-source` only when the test supplies its original entry path. IR Program.Source remains a display basename. The entry path and source graph hash participate in the generated JavaScript observation's cache key. The lower ready-check seam uses explicit `--terminal-check` for its fixed terminal controls. Emitted readiness checks call panic directly, so catch/finally cannot take them. Three catalog/ready mutants and the dropped-entry-path mutant prove those boundaries.

## Per-check comparison

The counts below include Go parent tests and subtests, not just leaves. The baseline IR package was rerun after repairing the local submodule checkout. [comparison.json](on-main-evidence/comparison.json) lists every added and removed test.

| Check | Base | Replay | Difference |
|---|---|---|---|
| internal/lower | 1,226 pass, 2 skip | 1,220 pass, 2 skip | Seven obsolete refusal subtests removed; one positive test covers five newly admitted library probes |
| internal/ir | 5 pass | 5 pass | None |
| internal/flow | 2,034 pass | 2,082 pass | 48 checks for the 16 scout fixtures |
| full internal/oracle | 2,034 pass, 5 skip | 2,091 pass, 5 skip | 57 scout fixture/proof checks, including three preservation tests |
| stage3/fixtures | 608 pass, 1 skip | 608 pass, 1 skip | No changed record or test result |
| stage3/fixtures, platform guard lifted | 608 pass, 1 skip | 608 pass, 1 skip | No changed record or test result |

Every common test result is unchanged. The stage3 skip is TestTransformedNodeRunnerGuardHook, an intentionally dormant subprocess hook. The platform-tagged records are host/09_realpath.a, host/14_getCurrentDirectory.a and host/24_useCaseSensitiveFileNames.a, all tagged linux and checked on this Linux host in both runs. The platform guard was lifted through a Go overlay and was not committed. No stage3 status.json changed.

The seven removed lower subtests covered repeat under try, a call reaching toFixed under try, stored Error throws, string throws, number throws, frozen-object writes under try and Hash calls under try. The first five are scout changes; the last two were stale base expectations reached by the replay. Regexp-global and typed-array try refusals inside existing table tests also moved to positive admission coverage. Their other negative cases remain refused.

The scout has sixteen fixtures. Six already lower on the base, and ten stop there; the exact base diagnostics are in [base-scout-fixtures.json](on-main-evidence/base-scout-fixtures.json). The restored oracle holds all sixteen to source Node, the JavaScript backend, sanitizer native and release native. The twelve normal fixtures finish with matching stdout/stderr/exit and clean leak checks. Two uncaught fixtures agree on stdout and exit 1; Node's engine renderer is excluded by the adopted ruling, while both backends must stay quiet and LeakSanitizer checks their cleanup. Two inserted-check fixtures exit 70 in both backends while Node runs the source catch/finally path. Unknown assertion/read and unrepresented Error observation refusals remain pinned.

| Scout fixture | Base lowering | Replay oracle |
|---|---|---|
| `step21_builtin_narrow_terminal.a` | NotYet: stage 0 can't lower new an Identifier yet | Pass |
| `step21_catch_callback.a` | Compiles | Pass |
| `step21_dynamic.a` | Refused: Adamic 0.1 refuses throwing a undefined; throw an Error: throw new Error(String(value)); what a catch takes is unknown, and an Error is what it can be sure of | Pass |
| `step21_dynamic_uncaught.a` | Refused: Adamic 0.1 refuses throwing a undefined; throw an Error: throw new Error(String(value)); what a catch takes is unknown, and an Error is what it can be sure of | Pass |
| `step21_error_subclasses.a` | Refused: Adamic 0.1 refuses throwing a Other; throw an Error: throw new Error(String(value)); what a catch takes is unknown, and an Error is what it can be sure of | Pass |
| `step21_finally_callback.a` | Compiles | Pass |
| `step21_finally_completion.a` | Compiles | Pass |
| `step21_library_failures.a` | NotYet: stage 0 can't lower instanceof against a value that isn't a declared class yet | Pass |
| `step21_library_host.a` | NotYet: stage 0 can't lower instanceof against a value that isn't a declared class yet | Pass |
| `step21_library_types.a` | NotYet: stage 0 can't lower instanceof against a value that isn't a declared class yet | Pass |
| `step21_liveness.a` | Compiles | Pass |
| `step21_object_uncaught.a` | Refused: Adamic 0.1 refuses throwing a Payload; throw an Error: throw new Error(String(value)); what a catch takes is unknown, and an Error is what it can be sure of | Pass |
| `step21_region_payload.a` | Refused: Adamic 0.1 refuses throwing a Failure; throw an Error: throw new Error(String(value)); what a catch takes is unknown, and an Error is what it can be sure of | Pass |
| `step21_rethrow.a` | Compiles | Pass |
| `step21_saved_error.a` | NotYet: stage 0 can't lower throwing an Error that isn't made where it's thrown or caught by the catch around it yet | Pass |
| `step21_soundness_terminal.a` | Compiles | Pass |

## Commands and observations

Toolchain: `export GOPROXY='https://proxy.golang.org|direct'`, `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Node is 24.19.0, Go is 1.27.1 and clang is 20.1.8. `nproc` prints 5; cgroup cpu.max is 400000 100000. Setup timing lines: Go 0.022s, Node 0.023s, submodules 0.068s, markdown 0.071s, clang 0.236s, build 325.587s, test binaries deferred 326.173s, warm cache 326.189s, done 326.372s.

The initial checkout exhausted the workspace filesystem because its reproducible Go build cache occupied about 30 GiB. Automatic review rejected a forced checkout because it could discard the partial tree. The partial checkout was preserved in a stash, the build cache cleared, and ordinary checkout/setup succeeded. The stash remains. The baseline worktree initially used a submodule symlink; Git rejected it during the IR guard. The same pinned submodule checkout was materialized locally with no source edits, and the entire IR package passed on retry. Neither incident is counted as a compiler result.

The command below ran in the exact base worktree and in the restored feature checkout, with output redirected to the corresponding logs:

```sh
ADAMIC_GATE_UNCACHED=1 go test -json ./internal/lower ./internal/ir ./internal/flow ./internal/oracle -count=1 -timeout=30m > /tmp/exceptions-21-restored-packages.json.log 2>&1
```

The baseline lower, flow and oracle passed on the first full run. The baseline IR retry used `ADAMIC_GATE_UNCACHED=1 go test -json ./internal/ir -count=1 -timeout=30m`. All four restored packages pass together, exit 0.

Both checkouts first ran `npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund`, adding three packages and exiting 0. Each then ran:

```sh
ADAMIC_GATE_UNCACHED=1 go test -json ./stage3/fixtures -count=1 -timeout=30m > /tmp/exceptions-21-restored-stage3.json.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -overlay /tmp/exceptions-21-restored-guard-overlay.json -json ./stage3/fixtures -count=1 -timeout=30m > /tmp/exceptions-21-restored-stage3-all-platforms.json.log 2>&1
```

Both commands exit 0 on base and replay. The overlay's only change is replacing the platform condition with `false &&` that same condition. The tracked guard and npm manifests are unchanged.

Counts: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m -args -update-counts`, exit 0. [Every changed row is explained here](on-main-counts.md): 16 new and 88 changed existing rows, zero outside the scout's own count commits. Predicate direction counts are byte-identical. A restored full oracle rerun confirms the refreshed table.

The first replay runs found stale refusals and the uncaught-handler regressions. A later full run found eight JavaScript comparisons using a display basename instead of the entry path. The eight focused source/JavaScript/sanitizer/release comparisons pass after the explicit-path correction. Every scout production mutant was then rerun serially with restoration, followed by the complete restored comparison above. The compressed logs preserve the observations and exclude historical scout logs from the new proof.

## Mutants

All 40 are caught. Compiler/build errors are excluded as catchers. The runners restore each changed production file; the original scout evidence files are also restored, with this base's raw logs stored separately. The following 33 production/harness mutants ran through `run-mutants.py`, `run-adopted-mutants.py`, `run-library-mutants.py`, `run-subclass-mutants.py`, `run-saved-mutants.py`, `prove-uncaught-sanitizer.py` and `run-main-mutants.py`:

| Mutant | Intended observed catcher |
|---|---|
| `missing-exception-edge` | runtime error: |
| `assignment-kills-handler-value` | old text is dead |
| `throw-temporaries-leak` | leaks: |
| `pending-finally-leak` | leaks: |
| `error-message-not-retained` | AddressSanitizer: heap-use-after-free |
| `unknown-authorizes-assertion` | admission check |
| `implicit-string-conversion` | uncaught lifetime check: exit 70 |
| `pending-undefined-lost` | stdout differs |
| `catch-assumes-error` | exit codes differ |
| `typeof-boolean-is-number` | stdout differs |
| `null-is-undefined` | stdout differs |
| `finally-releases-twice` | AddressSanitizer: heap-use-after-free |
| `region-payload-freed` | AddressSanitizer: heap-use-after-free |
| `uncaught-is-panic` | uncaught lifetime check: exit 70 |
| `range-error-is-error` | stdout differs |
| `wrong-range-message` | stdout differs |
| `type-error-is-error` | stdout differs |
| `hash-failure-is-panic` | exit codes differ |
| `host-type-error-is-error` | stdout differs |
| `soundness-check-is-catchable` | want the inserted check to fire |
| `error-ancestry-missing` | stdout differs |
| `error-default-name-wrong` | stdout differs |
| `error-prefix-released-twice` | AddressSanitizer: heap-use-after-free |
| `error-fields-out-of-order` | stdout differs |
| `rethrow-wrapped` | stdout differs |
| `builtin-narrow-forgets-subtype` | want the inserted check to fire |
| `saved-error-wrapped` | stdout differs |
| `backend-renders-stack` | uncaught renderer wrote stderr |
| `uncaught-sanitizer-visible` | sanitizer failure despite matching stdout and exit |
| `changed-source-keeps-terminal-mode` | changed source borrowed terminal convention |
| `changed-dependency-keeps-terminal-mode` | changed dependency borrowed terminal convention |
| `backend-drops-entry-source` | JavaScript backend: exit codes differ |
| `readiness-runs-handler` | JavaScript readiness guard ran catch/finally |
| IR output inserted in `catch_callback` | Source Node stdout comparison in JavaScript, sanitizer native and release native; sanitizer/leak runs stay clean |
| IR output inserted in `finally_callback` | Source Node stdout comparison in JavaScript, sanitizer native and release native; sanitizer/leak runs stay clean |
| IR output inserted in `rethrow` | Source Node stdout comparison in JavaScript, sanitizer native and release native; sanitizer/leak runs stay clean |
| IR output inserted in `finally_completion` | Source Node stdout comparison in JavaScript, sanitizer native and release native; sanitizer/leak runs stay clean |
| IR output inserted in `liveness` | Source Node stdout comparison in JavaScript, sanitizer native and release native; sanitizer/leak runs stay clean |
| Census selector includes readingException | Assertion on the historical selector count |
| Census diagnostic accounting reduced by one | Assertion on the historical diagnostic count |

The two census mutants are independent script mutations and each exits nonzero at its intended assertion. The five IR mutants run in the full oracle. The uncaught sanitizer mutant proves that the exit-1 stderr exception cannot hide ASan failure. [Raw mutant summary](on-main-evidence/restored-mutants.log.gz) and the individual compressed logs identify each failure.

## Limits

This is the scout's represented exception surface, not all ECMAScript library failure paths. Error stack, constructor options/cause, unsupported reflection, captured unknown cells, allocation exhaustion and oversized concatenation/builders retain their existing boundaries. There is no new cycle/escape proof or collector. The old historical census credit remains separate; this unit claims no hidden-byte reveal and does not recompile all TypeScript compiler sources. The proof is on Linux with ASan, UBSan, LeakSanitizer and counted builds; no macOS runtime run or whole `./...` gate is claimed.
