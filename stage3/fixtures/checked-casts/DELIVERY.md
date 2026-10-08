Current correction: see [PROVENANCE.md](PROVENANCE.md) for separate read-stop diagnostics, cast-site provenance, the authorized verifier update and complete observer controls. Earlier delivery below is historical.

Built: step 09's 18 formerly blocked runtime contracts now pass, with all 20 acceptance fixtures held to Node and both backends.
Commits: tagged group ce60a068; scalar mechanism 675951e7 and evidence 26556166; remaining view/generic/index mechanism 1af953e6; final evidence is the delivery branch tip.
Commands and outputs: unchanged strict verifier PASS, 20 Node goldens and runtime contracts PASS, 30 observer controls and 40 header mutants caught; focused oracles/adapters/vet/counts PASS.
Mutants: emitted tag checks, scalar guards, consumed view reads, source writes and receiver checks are caught; exact outcomes and logs follow.
Not covered: all 3,957 individual cast sites, general generic assertions, unsupported nullable array consumers, Error facts, full scanner execution or full gate; global counts retain 39 inherited failures.

# Acceptance delivery for step 09, #b5w3ycg

The inputs are the complete checked-casts subtree from ea1b2359. No peer branch was merged. The compiler remains on codex/scanner-cast-checks, based on the requested checked-views integration 432d4913. The population ledger is the supplied immutable 855bcfaa. The verifier, classification, source sites, source Node goldens, runtime stdout/exit requirements and failing twins are unchanged. Removing obsolete refusal headers permits the existing a-check gate to require lowering. The observer now recognizes the actual emitted tag helpers and scalar/view failure calls, records all contracts instead of aborting, and removes the actual emitted checks. It does not turn compiler refusals or runtime mismatches into passes.

## Which of the 18 pass

Initial measurement on 561166d1: three of the original 18 passed (both 01 fixtures and 03 positive), five of 20 overall including the preexisting 02 pair. After the scalar group, nine of 18 passed. On committed compiler 1af953e6, all 18 pass, and the preexisting 02 pair still passes. Each row below passes source Node, sanitized native (ASan and UBSan), native release and emitted JavaScript. Positive executions are leak checked. Failing twins stop 70 at the pinned cast/read with exact stdout and exact diagnostic in the focused oracle.

| Originally blocked fixture | Mechanism | Result |
| --- | --- | --- |
| `01_enum_declaration.a` | tag check at cast | PASS |
| `01_enum_declaration_fails.a` | tag check at cast | PASS |
| `03_chain_info.a` | lazy checked field view | PASS |
| `03_chain_info_fails.a` | lazy checked field view | PASS |
| `04_literal_payload.a` | lazy checked field view, source slot certificate | PASS |
| `04_literal_payload_fails.a` | lazy checked field view, source slot certificate | PASS |
| `05_extension.a` | owned finite string/enum guard | PASS |
| `05_extension_fails.a` | owned finite string/enum guard | PASS |
| `06_parser_keyword.a` | owned finite number/enum guard | PASS |
| `06_parser_keyword_fails.a` | owned finite number/enum guard | PASS |
| `07_generic_next.a` | concrete generic checked array view | PASS |
| `07_generic_next_fails.a` | concrete generic checked array view | PASS |
| `08_generic_node.a` | concrete generic checked object view | PASS |
| `08_generic_node_fails.a` | concrete generic checked object view | PASS |
| `09_primitive_string.a` | owned primitive kind guard | PASS |
| `09_primitive_string_fails.a` | owned primitive kind guard | PASS |
| `10_generator_label.a` | owned readonly index snapshot and scalar guard | PASS |
| `10_generator_label_fails.a` | owned readonly index snapshot and scalar guard | PASS |

The unchanged strict verifier reports `ledger_casts:3957, node_goldens:20, runtime_passed:20, runtime_blocked:0, observed_mutant_controls:30` and exits 0 with `--require-runtime`. Ledger coverage is a classification assertion, not evidence that every cast site lowers.

## Missing mechanisms built locally

Tagged downcasts were first, reflecting their 2,051 sites. Both tagged pairs already had executable cast-point checks; the first delivery added their independent Node/backend oracles, pins and omission mutants. Owned scalar/finite domains followed in 675951e7. These evaluate the operand once, guard runtime kind before narrowing and enforce finite members. Numeric enum domains are open unless the target is a finite subset. Any/unknown and unrelated brand/intersection/double assertions remain refused.

The remaining missing pieces were ours, rather than library Error facts, and are in the separate mechanism commit 1af953e6:

| Lane | Missing piece | Implementation |
| --- | --- | --- |
| Lane 1, codex/interface-downcasts | A backing string/number union field was not interned before the conservative view use scan; source slot certificates could not compare complete primitive unions. | `internal/lower/view_primitive_reads.go:87`, `internal/lower/view_writes.go:122`, and `internal/ir/view_writes.go:25` intern complete declarations and check union members. Checked fields keep boxed storage, so later source mutation remains visible. |
| Lane 2, codex/views-arrays-callables | A homogeneous `[12]` initializer used number storage despite a declared `(number|string)[]`; reference writes only dispatched object certificates. | `internal/lower/view_array_context.go:12`, `internal/lower/view_array_writes.go:31`, `internal/native/view_array_writes.go:20`, and `internal/javascript/view_array_primitive_writes.go:10` preserve boxed storage and check incoming values against the original declared primitive domain. |
| Crosscut between admission lanes | Generic preflight rejected `as T` and `as T[]` before concrete instantiation could prove the cast. | `internal/lower/view_generic_cast.go:13` defers the generic candidate; it grants no proof. Each concrete body runs the ordinary cast proof again, and unrelated instantiations re-refuse. |
| Lane 2 array storage / owned scalar bridge | `readonly (number|Expression|undefined)[]` could not provide a safe physical indexed snapshot for an immediate scalar cast. | `internal/lower/view_scalar_index.go:11` and `internal/native/view_scalar_index.go:8` evaluate array/index once and use the producer-certified physical snapshot. The scalar helper checks the selected runtime kind before conversion. Unread object alternatives are not asserted or scanned. |
| Lane 2 receiver consumers | A newly admitted nullable writable view could otherwise dereference an absent array root. | JavaScript indexed adapters check `Array.isArray`; the native adapters already validate the root. `internal/lower/view_nullable_array_consumers.go:8` conservatively refuses array property/method reads, iteration, destructuring, spreads and array call arguments across the closed program for these new views, until they have receiver adapters. |

This is the smallest sound scope exercised by the reductions. Untagged views (1,292 sites in the ledger) remain lazy: the live field/element is checked on each read, including after backing mutation. To satisfy the newer strict acceptance contract, checked field/element failure diagnostics now begin `cast failed:` while preserving their field, expected type and found type details. Existing diagnostic pins were migrated with that change; authored panic strings are not rewritten.

Error remains library's #ddwcejg on library/area-on-next. The typed reduction [error-library-cast.a](../../drivers/scanner/cast-checks/error-library-cast.a:3) stops at line 3, column 1 with exact `*lower.NotYet.What == "node:globals.ErrorConstructor.captureStackTrace"`, emitted by `internal/lower/library_node.go:93-94`; the real declaration is `stage3/api/node_modules/@types/node/globals.d.ts:51`. Its Node golden is `ok`. `TestScannerErrorLibraryCastNotYet` intentionally turns red as soon as that frontier closes, requiring replacement by runtime execution and counts. No Error host facts are fabricated. The String constructor's demanded `fromCodePoint` view and producer certificate from b52fbc43 remain green; current read pins use the migrated diagnostic prefix.

## Every mutant and what caught it

| Mutant | Observed catch |
| --- | --- |
| Remove one cast-point tag test in each 01/02 failing twin, native release and JavaScript | A later read still stops 70, but its diagnostic differs. The exact cast-point stop pin catches each omission. |
| Remove actual emitted JavaScript failure checks in all ten failing twins | Each finishes exit 0 with erased-source output and no stderr; exit/stdout/check-location contracts catch all ten. |
| Remove the cast and later tag-domain read check from 02's actual emitted C | Valid ASan/UBSan C finishes exit 0, leak clean, exactly matching erased source on Node. The required stop-70 contract catches it. |
| Remove finite-domain guards, 05/06, native release and JavaScript | Exit 0 prints `.wrong` / token 80. Stop pins catch both. |
| Remove runtime kind guard, 09/10, native release and JavaScript | Surviving representation narrowing stops with a generic union mismatch. Exact expression/expected/found pins catch the lost cast guard. |
| Remove one consumed checked read in 03/04/07/08, sanitized native and JavaScript | 03/07 native return wrongly decoded data; 04/08 native trigger ASan faults after unsafe decoding. All four JavaScript mutants finish exit 0 with wrong data. Runtime stop pins catch all; no clean native continuation is claimed for 04/08. Unmutated executions are sanitizer clean. |
| Forge an incoming Boolean array value, then omit each backend's primitive source-write guard independently through source overlays | Unmutated writes stop at `<array write>` naming `string | number` and Boolean. Omission reaches the later `result[0]` read instead; the exact write-site pin catches both. |
| Forge a Boolean object slot write, then set WriteProven to omit source-slot checks in native release and JavaScript | Original declared slot certificate stops at the write. Omission moves the failure to a later checked Boolean read; exact write pin catches both. |
| Omit each new JavaScript array receiver guard independently through source overlays | An absent-root fixture/IR liar reaches TypeError instead of the expected typed receiver message. Exact receiver pins catch both. Native absent roots also stop 70 cleanly. |
| String demanded-view admission, native producer certificate, JavaScript producer certificate omissions | The established named lower/oracle tests each fail; refreshed constructor mutant logs record all three. Existing whole-read omission test also remains green. |
| Existing interface tag omission, wrong tag, twice-operand mutants | TestInterfaceCastRuntimeMutants passes all three. The twice-operand mutation now targets the second actual `next()` IR call even when proof erased the cast. Extra calls change the oracle output and are caught. |
| Observer output-marker and erased-source controls | Ten of each are caught; together with the ten emitted-JavaScript omissions these are the 30 observer controls. Source erasure is a negative control, not an implementation mutant. |
| Drop a ledger row; reclassify a tagged row while keeping category counts consistent | Both coverage mutants are rejected by the unchanged strict verifier. |
| False `refused adamic/no-unchecked-cast` and unrelated refusal headers on all 20 fixtures | Gate.aCheck rejects all 40 header mutants. Only its a-check method runs; the full gate was not invoked. |
| Existing forged physical array producer metadata / storage guard omission | Focused native primitive snapshot tests pass, including the independent guard omission mutant. |

## Exact commands and saved outputs

Commands ran from the repository with `source /workspace/adamic-tools/env.sh`; all test output went to log files. The complete fixtures/oracles are in `internal/oracle/step09_acceptance_test.go`; source-overlay scripts are under worker-evidence. Commands below are focused selections, not whole-package runs.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step09-worker-setup.log 2>&1
go build -o /tmp/step09-delivery-adamic ./cmd/adamic > /tmp/step09-delivery-build.log 2>&1
NODE_PATH=$PWD/stage3/api/node_modules node stage3/fixtures/checked-casts/observe.cjs /tmp/step09-delivery-adamic /tmp/step09-delivery-observed > /tmp/step09-delivery-observe.log 2>&1
node stage3/fixtures/checked-casts/verify.cjs --require-runtime > /tmp/step09-delivery-verify.log 2>&1
node stage3/fixtures/checked-casts/native-mutant.cjs /tmp/step09-delivery-observed /home/agent/.cache/adamic/runtime/5e2744e1412ef8b76fe44db28ae0b6a05786ed70ee8559eee1d676c7a7750085 > /tmp/step09-delivery-native-mutant.log 2>&1
python3 stage3/fixtures/checked-casts/a-check.py /tmp/step09-gate.py /tmp/step09-delivery-a-check > /tmp/step09-delivery-a-check.log 2>&1
go test ./internal/oracle -run 'TestStep09|TestScannerCastDiagnostic|TestScannerStringConstructor|TestScannerErrorLibraryCastNotYet|TestScannerCastCounts|TestCheckedViewObjects|TestCheckedViewArrays$|TestCheckedViewNativeArrays|TestCheckedViewPrimitiveArrayPairs|TestCheckedViewPrimitiveArraySafety|TestInterfaceCastOracle|TestInterfaceCastChecksMalformedRead|TestInterfaceCastRuntimeMutants' -count=1 -v > /tmp/step09-delivery-oracles.log 2>&1
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'TestPrimitiveUnionSlotWriteDomains|TestGenericCast|TestScalarIndex|TestNullableWritableArray|TestPrimitiveArraySnapshotStorageAndLifetime|TestViewArraysNode|TestViewArrayFieldPresence' -count=1 -v > /tmp/step09-delivery-adapters.log 2>&1
go test ./internal/lower -run 'TestNullableWritableArray|TestGenericCast|TestScalarIndex' -count=1 -v > /tmp/step09-nullable-admission.log 2>&1
python3 stage3/drivers/scanner/cast-checks/run-constructor-mutants.py > /tmp/step09-constructor-mutants.log 2>&1
python3 stage3/fixtures/checked-casts/worker-evidence/write-mutants.py /tmp/step09-write-mutants > /tmp/step09-write-mutants.log 2>&1
python3 stage3/fixtures/checked-casts/worker-evidence/receiver-mutants.py /tmp/step09-receiver-mutants > /tmp/step09-receiver-mutants.log 2>&1
go test ./internal/oracle -run '^TestScannerCastCounts$' -count=1 -v -args -update-counts > /tmp/step09-delivery-counts.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/step09-complete-global-counts.log 2>&1
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle > /tmp/step09-delivery-vet.log 2>&1
git diff --check
```

Observer, strict verifier, sanitized C mutant and a-check: PASS. Final focused oracle selection: PASS, 20.365s. Focused IR/lower/native/JavaScript adapters: PASS; targeted vet has empty output and exits 0. Dedicated scanner count refresh: PASS, 3.969s, records all 20 new acceptance rows in `internal/oracle/counts.md` and preserves earlier scanner rows. The required global refresh fails on 39 inherited fixtures (57.593s), so no global-green claim is made. Historical failed intermediate runs caught the formatter error before correction; the final focused rerun and fresh committed-compiler observer pass. No full package suite or full gate was run.

Setup: nproc 5, CPU quota 4. Timing lines: Node ready 0.030s, Go ready 0.030s, markdown ready 0.091s (validated dependencies, skipped install step 0.008s), submodules ready 0.099s, clang ready 0.306s, Go build ready 49.034s, test binaries deferred 49.161s, cache warm 49.163s, done 49.190s. These are observed setup measurements, not estimates.

No protected orchestration files or cohere files were edited, no peer lane was merged and no PR was opened. Only the worker branch is pushed. New program files are `.a`. The delivery adds executable runtime obligations and check-failure evidence toward roadmap step 09, without claiming complete step-wide cast coverage.

Delivery staging was initially rejected by automatic approval review because its broad oracle scope could include unrelated changes. A read-only audit proved all 84 existing oracle edits are literal checked-read prefix substitutions only, with no logic changes; the four acceptance/count/mutant files were separately reviewed. Automatic review then accepted tracked-only staging with explicitly named new files and logs. The archived global counts log strips trailing spaces on blank lines; the raw output remains at `/tmp/step09-complete-global-counts.log`.
