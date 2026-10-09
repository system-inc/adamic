# Receiver-independent method destructuring

Built ordinary closure storage for proven receiver-independent literal methods, preserving extracted identity and arguments.
Delivery branch: `codex/hidden-method-destructuring`, based on `4885cec50290686df487b62aac47c85d871ed40c`; delivery SHA is reported with the push.
Focused lowering, Node/native/JavaScript differential oracles and counts refresh passed; exact head replay moves to an overload result refusal.
A mutant dropping the extraction proof admitted five unsafe witnesses and failed `TestMethodDestructuringSafety`; an omitted nested census attempt failed scope validation.
The selected region remains 7,102 hidden bytes before and after, revealing 0 bytes; class extraction, dynamic receivers and neighboring owned stops remain unsupported.

## Change and proof

`receiverIndependentMethod` scans a literal method's runtime syntax, including parameter defaults and nested arrows. A method must have a body and an object-literal parent; explicit receiver parameters, runtime `this`, and `super` prevent conversion. Nested dynamic receiver syntax is conservatively rejected too. Absence of a TypeScript `this` annotation is not a proof.

Proven methods use the existing ordinary closure ABI and keep their lexical captures. Receiver-dependent methods retain their existing receiver argument. The iterator member invocation helper uses the same proof to select that ABI; no iterator protocol feature was added. Neither backend nor runtime changed.

`destructureFrom` accepts these own literal methods and structural method signatures holding ordinary callbacks. A structural signature alone does not prove a body independent: the existing whole-program `erasedMethodField` scan still rejects every compatible receiver-dependent literal or class prototype origin. Library prototype guards remain. Actual class methods remain refused, even when their bodies do not read `this`. Direct property extraction's separate refusal remains; this unit supports destructuring.

The field stores the original closure; extraction reads it rather than allocating a binding wrapper. The fixture verifies repeated-extraction identity, callback argument positions, captured string survival after the holder dies, a single source evaluation, and a saved callable surviving field replacement. The source Node output is:

```text
5 9 true
2,3
7
8
kept-alive:3
1 3 12
```

Stderr is empty; exit 0. The original short positive witness is included in the larger fixture. Receiver-sensitive behavior is pinned by `internal/lower/testdata/method_destructuring_receiver.a` and checker-valid negative lowering tests instead of enabling an unimplemented dynamic receiver ABI. The new negative fixture actually invokes the extracted method: source Node observes a TypeError reading `value` from undefined, empty stdout and exit 1 (the observation driver prints the error name/message; `/tmp/hidden-method/negative-node.stderr`). Its pinned test requires lowering to refuse it. The earlier Node-only explicit-this witness's unbound call is not asserted to be checker-accepted here.

The two old receiver-independent refusal probes in `iteration_test.go` became positive tests in `method_destructuring_test.go`. Class/prototype and inherited-library refusal probes remain. No neighboring views, overloads, computed keys, defaults or binding patterns were implemented. The ownership ledger remains `codex/hidden-wave-2` at `e2c6e8df`, `stage3/hidden-briefs-2/TABLE.md`; the pending destructuring unit shares `destructureFrom` and must reconcile that function after landing.

## Exact replay and hidden intersection

Source ranking: `ec0b16c04f3bbdfbaf3932ab01307f90b08156fd`, `stage3/census/hidden/RESULT.json`. All 82 adapted source hashes match that manifest. Adapted source was prepared from `388096e6` and kept unchanged. Whole-project checker rejection remains explicit: these are latent lowering observations, not accepted whole-project compilation.

Before change, the inherited full-project replay reproduced `Refused: a method in object destructuring` at `utilities.ts:11316:35`, exit 0. After change, the same selector exits 1 because the old diagnostic no longer reproduces; the selected attempt now stops at `utilities.ts:11320:5`:

```text
overload 1 of evaluate result EvaluatorResult<string | undefined> cannot be served by implementation result EvaluatorResult<string | number | undefined>
```

Owner: `codex/overload-results`. No change to that check was made. Logs: `/tmp/hidden-method/before-replay.log` and `/tmp/hidden-method/after-replay.log`.

The historical manifest's actual hidden intersection with `utilities.ts [455532, 462634)` is 7,102 bytes. A scoped independent census on the assigned base and final compiler gives:

| Region | Historical pinned intersection | Assigned-base intersection | Final intersection | Revealed difference |
| --- | ---: | ---: | ---: | ---: |
| utilities.ts [455532, 462634) | 7,102 | 7,102 | 7,102 | 0 |

The scope retains full project loading, registration, checker diagnostics, rollback, ancestor binding context and independent attempts. Only attempt traversal is filtered to this file and nodes intersecting the query interval. The syntax-only refusal scan is omitted because it records no successful coverage and contributes no hidden boundaries. The independent stock AST catalogue proves that both intersecting units were attempted: `createEvaluator` at `11316:1` and its nested `evaluate` implementation at `11322:5`. Omitting the nested unit fails the scope assertion. The nested implementation hits the same overload result refusal before and after.

The unchanged corrected `hidden.py` from the ranking pin unions boundaries and subtracts independent coverage. Its stock input is the original hash-pinned `stock.json.gz` at census pin `388096e6`; every source hash was verified. The calculation is restricted to the query intersection; outside-file and outside-query totals are not reported. Initial whole-corpus runs were stopped after this complete region measurement; their partial output is not evidence of a corpus result. Initial scratch filters were discarded and are not used for the reported numbers.

The historical 34,480 credited bytes across 47 regions are not a measured group improvement. No whole-corpus total or other region's revealed difference is claimed. [Recorded region result](evidence/region.log.txt).

## Reproduction and checks

Environment: Go 1.27.1, Node 24.19.0, clang 20.1.8. `nproc` 5; cgroup CPU quota 4. Required setup succeeded with `GOPROXY='https://proxy.golang.org|direct'`; sourced `/workspace/adamic-tools/env.sh`. Timing lines: node 0.030s; go 0.034s; submodules 0.081s; markdown dependency step 0.007s, ready 0.102s; clang 0.217s; Go build 9.975s; test binaries deferred 10.147s; cache warm 10.149s; done 10.175s. Log `/tmp/hidden-method-setup.log`.

Replay builds used `go build -buildvcs=false -overlay=<scratch-overlay> ./stage3/census/latent/replay/worker`. Each replay used `GOMAXPROCS=1`, full `-project /tmp/hidden-adapted/src/compiler`, `-where /tmp/hidden-adapted/src/compiler/utilities.ts:11316:35`, `-kind Refused`, and `-reason 'a method in object destructuring'`, with stdout/stderr directed to its log. The baseline executable was built before production edits.

For region census reproduction, generate the standard overlay with `stage3/census/latent/make_overlay.py`. In its scratch `internal_lower_latent_units.go`, restrict the file loop in `latentFullSelected` to the full utilities.ts path and restrict its independent candidate loop to `node.End() > 455532 && node.Pos() < 462634`. Do not filter `latentRegisterNested`: every project declaration must remain registered. Omit its syntax-only `selection == nil` refusal scan. Build `./stage3/census/latent/tool` with `-buildvcs=false -overlay=<region-overlay>`. Run both versions with `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 GOMAXPROCS=1`, source directory and output JSONL as positional arguments. The baseline overlay additionally replaces `collections.go`, `iteration.go` and `iteration_origin.go` with their exact assigned-base bytes. Feed the resulting rows and utilities.ts stock metadata to the pinned `hidden.calculate`, assert the observed units equal every stock unit intersecting the query, and intersect its residual ranges with the query. Scratch recipe and logs: `/tmp/hidden-method/measure_region.py`, `baseline-region-v2-overlay.json`, `after-region-v2-overlay.json`, and `region-result.log`.

Final focused selectors (all output written directly to logs):

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestMethodDestructuringSafety|TestReceiverIndependentMethodDestructuring|TestDestructuredMethodsCannotLoadOwnSlots|TestIterator|TestLiteralMethod|TestNativeAgreesWithNode/internal/oracle/testdata/(hidden_wave2_method_destructuring|user_iterators|user_iterator)' -count=1 -timeout 10m > /tmp/hidden-method/regression-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(hidden_wave2_method_destructuring|census_optional_values_(system|forms|padding)|census_unary_numeric|closure_convention_receiver_rest|host_never_branches|host_void_method|user_iterators)' -count=1 -timeout 10m > /tmp/hidden-method/changed-fixtures.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts > /tmp/hidden-method/counts.log 2>&1
```

Observed final lowering 0.821s, oracle 4.889s; changed-fixture oracle 11.271s; counts 43.005s, all passed. Oracle execution includes source Node, JavaScript backend, release native, ASan/UBSan and leak checking. [Every counts-row change](COUNTS.md) is explained separately. Earlier test drafts failed TypeScript checking or an existing assignment refusal; they were corrected, not counted as mutant catches.

Independent compiler mutant: use a scratch Go overlay adding unconditional `return true` at the beginning of `destructurableMethod`, leaving `receiverIndependentMethod`, the actual receiver ABI and class-origin checks intact. Run `go test -overlay=/tmp/hidden-method/mutant-overlay.json ./internal/lower -run '^TestMethodDestructuringSafety$' -count=1 -timeout 10m`, log `/tmp/hidden-method/mutant-final.log`. Observed exit 1: five negative witnesses were actually admitted, causing the intended `receiver or prototype origin was erased` assertions to fail. These are direct this, arrow receiver, default receiver, explicit receiver and structural receiver. Class-origin witnesses remain refused by the independent origin guard. This is the brief's allowed negative-test alternative to a receiver-binding semantic mutant; no dynamic-receiver acceptance is claimed. [Raw mutant failure](evidence/receiver-proof-mutant.log.txt).

No whole packages or full repository gate were run. No sanitizer or semantic result is claimed for refused programs. No speed comparison or calendar target is claimed.
