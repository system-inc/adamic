Step 09 ledger of TypeScript 6.0.3's tsc source closure, measured with the compiler API.
Topic base: origin/main 3ffb1a835184713998a34874e86326cd21db971f; no compiler or adaptation merges.
Stock inventory: 210 explicit any tokens, 214 any-typed declarations and 6,231 as expressions.
The one-of-each fixture passes; the mutant which leaves rewritten sites open is caught.
Publication is pending the branch references for adaptations 41 and 42; adaptation 43 is included.

Coverage and definitions

The stock entry reaches 80 source files, not 81. Adaptation 47 adds
src/compiler/hostErrors.ts; the adapted closure reaches 81. The two files outside
src/compiler are src/tsc/tsc.ts and src/tsc/_namespaces/ts.ts. The generated helper
has no stock counterpart and is reported separately. The stock commit is
050880ce59e30b356b686bd3144efe24f875ebc8, tag v6.0.3.

An explicit any is an AST AnyKeyword, including nested type arguments, return
annotations, constraints and casts. An any-typed declaration has a resolved type
whose flags include TypeFlags.Any; containers such as any[] and functions returning
any are not themselves type any, but their explicit tokens are counted. Inferred
any declarations, parameters, properties and binding elements are included.
Type-only import aliases are resolved to their declared types: asking for their
value-position types otherwise returns an API fallback any and produces thousands
of false entries. Namespace labels with no value/type do not declare any values.
Named tuple labels use their slot type. The declaration and explicit-token counts
overlap deliberately. Each as expression is counted, including nested as expressions
and as const. Angle-bracket assertions are not as expressions, although any tokens
inside them are included. Library declarations and files outside the source closure
are checker dependencies, not ledger files.

The compiler API uses the compiler project's inherited options and installed Node
types, loads tsc from source without project references, and records semantic
diagnostics. It reports zero semantic diagnostics on stock and the measured final
tree. AST enumeration and resolved types use typescript@6.0.3, not regex.

Disposition evidence

The ledger tracks exact UTF-16 AST anchors through each adaptation's immediate
before/after source. Line-ending changes retain position correspondence. Changed
blocks are refined only when bounded; an ambiguous large block stops measurement.
An eliminated token or assertion is credited to its actual adaptation. A declaration
whose inferred type changes is credited only with immediate before/after compiler-API
snapshots proving the type changed from any. An edit that leaves a cast in source
does not by itself clear that cast. New adapted casts are inventoried separately
and are included in open.json when unresolved.

A safe upcast needs both stock checker assignability and Adamic's actual castProof
acceptance without a runtime-check plan. as const uses Adamic's explicit literal
assertion exemption and is identified separately by proof_kind. Runtime-checked
casts remain source-open: the probe reports that a check is available, without
claiming that the whole tsc program builds or that code was emitted. A panic or
other unresolved probe stays open with an unknown unchecked-refusal result.

The cast probe is a Go overlay, not a compiler change. It disables the ordinary
loader output path, preserves the populated checker despite diagnostics, and calls
main's castProof independently at every adapted as expression. Its checker uses
main's Adamic defaults plus Node declarations. Those options differ from stock's
project options; the safe category therefore requires both observations. Generic
casts are observed in their uninstantiated source context. Future concrete generic
instantiations are outside this census's claim. Every open cast carries its observed
refusal/check state and message; adamic_refuses_unchecked is true only for the
actual 'a cast the runtime cannot check' refusal, false for another observed result,
and null when unresolved. Other refusal messages are retained verbatim.

Adaptation provenance

Main supplies 40-explicit-any and the rest of its numeric pipeline. Adaptation
43-any-returns comes from codex/stage3-real-any at
643639ea30051875d2ba1a133866081bc8999e2a and runs before 45 in numeric order.
Its README and rules were read; its source is extracted into scratch only.
The branch's six return-contract changes can also change inferred declarations;
those receive immediate semantic evidence rather than token-only credit. Its three
new object-view assertions are counted, rather than assumed safe from the rewrite.
No branch containing adaptations 41 or 42 was found in the inspected remote refs.
Their references have been requested; they are required before publication.

Reproduction

Use Node 24.19.0 and the stock compiler API installed from stage3/api's lockfile.
Supply upstream node_modules from npm ci matching the pinned upstream lockfile.
A sparse stock checkout avoids copying upstream's large test-baseline corpus.
All commands below write their output to files. Use new scratch/output directories.

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step09-setup.txt 2>&1
source /workspace/adamic-tools/env.sh
export NODE_PATH=/path/to/stock-typescript-6.0.3/node_modules
python3 stage3/step09-ledger/prepare.py /tmp/step09-new --node-modules /path/to/upstream/node_modules --extra 643639ea:stage3/adapt/43-any-returns > /tmp/step09-prepare.txt 2>&1
node stage3/step09-ledger/enumerate.cjs /tmp/step09-new/stock /tmp/step09-new/stock.json > /tmp/step09-stock.txt 2>&1
node stage3/step09-ledger/enumerate.cjs /tmp/step09-new/adapted /tmp/step09-new/adapted.json > /tmp/step09-adapted.txt 2>&1
python3 stage3/step09-ledger/make_overlay.py /tmp/step09-overlay > /tmp/step09-overlay.txt 2>&1
go build -overlay /tmp/step09-overlay/overlay.json -o /tmp/step09-probe ./stage3/step09-ledger/probe > /tmp/step09-build.txt 2>&1
```

Pass all adapted closure paths and node_modules/@types/node/index.d.ts as arguments
to the probe, storing its stdout JSON and stderr in separate files. Then:

```
python3 stage3/step09-ledger/ledger.py /tmp/step09-new/stock.json /tmp/step09-new/adapted.json /tmp/step09-new/transitions.json /tmp/step09-casts.json /tmp/step09-result > /tmp/step09-summary.txt 2>&1
STEP09_CAST_PROBE=/tmp/step09-probe python3 stage3/step09-ledger/test_fixture.py > /tmp/step09-fixture.txt 2>&1
go vet -overlay /tmp/step09-overlay/overlay.json ./stage3/step09-ledger/probe > /tmp/step09-probe-vet.txt 2>&1
go vet ./... > /tmp/step09-vet.txt 2>&1
```

The fixture's single authored main.a has exactly one explicit any, one declaration
of type any, and one as cast. Its temporary adapted copy changes the any annotation
to number. The known answer is two rewritten sites and one safe upcast. Node's
transpiled source runs and reports value=1 widened=1. The real classification mutant
suppresses rewrite dispositions; it reports two open sites and is caught by that
independent expected answer, rather than a parser or build failure.

Setup: Node 0.033s, Go 0.034s, submodules 0.090s, markdown 0.092s, clang 0.171s,
build 11.900s, cache 12.076s, total 12.157s. nproc is 5; CPU quota is 4.
The initial full scratch copy exhausted disk space; only newly created upstream
test copies were removed and the run resumed with source-only snapshots and the
single API baseline required by adaptation 40. The complete Go gate, native tsc
execution, and oracle/lane comparisons of the adaptations are not claimed by this
ledger unit. No production compiler or adaptation file is changed.

Current measured counts, pending adaptations 41 and 42

| Kind | Stock sites | Rewritten | Safe upcast | Open |
| --- | ---: | ---: | ---: | ---: |
| Explicit any | 210 | 81 | 0 | 129 |
| Any-typed declarations | 214 | 53 | 0 | 161 |
| As casts | 6,231 | 4 | 2,260 | 3,967 |
| Total | 6,655 | 138 | 2,260 | 4,257 |

Adaptation 40 clears 125 ledger entries, 43 clears 12, and 32 clears one.
These are overlapping syntactic/declaration inventories, not 138 independent
runtime obligations. Thirty additional adapted casts are open. All 3,997 open
casts have a full-span Adamic probe: 3,764 receive the unchecked-cast refusal,
120 receive another refusal, 108 have a runtime-check proof, two differ from the
stock assignability result and remain open, and three panic and remain unresolved.
Both tsc files outside compiler have zero sites of all three kinds. Per-file
counts including the added helper are in FILES.md and SUMMARY.json; complete stock
entries are in ledger.json; all current open obligations are in open.json.

The nested regression fixture gives one safe upcast and one unchecked refusal
for 1 as unknown as number. Their expression starts are identical; their end
positions differ. Probe matching therefore uses both start and end, and duplicate
proof spans fail the ledger rather than overwrite a result. The complete runner
also passes on the miniature fixture. The touched measurement fixture, rewrite
mutant, overlay vet, repository vet, internal/load tests and the filtered
TestNativeAgreesWithNode/dedication/dedication.a oracle pass. Output is in evidence/.
The preparation CLI is a packaged equivalent of the source-only preparation
commands used for this run; a second full preparation was not run. Use run.py to
execute the measurement and fixture over its prepared tree in one fail-fast call.

Source excerpts in the JSON data come from Microsoft TypeScript 6.0.3,
copyright Microsoft Corporation, under Apache-2.0. Upstream trees are external
scratch inputs, rather than vendored source.
