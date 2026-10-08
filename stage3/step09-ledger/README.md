Step 09 ledger of TypeScript 6.0.3's tsc source closure, measured with the compiler API.
Merged topic base: origin/main 45487a809f89885a3fc651cd590e7dabf31362dc; extra adaptation 43 stays in scratch.
Stock inventory: 210 explicit any tokens, 214 any-typed declarations and 6,231 as expressions.
The one-of-each fixture passes; the mutant which leaves rewritten sites open is caught.
Main plus adaptation 43 is measured; the shared setTimeout return site is counted once.

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

A safe upcast needs both the stock compiler API's assignability on the adapted site and Adamic's actual castProof
acceptance without a runtime-check plan. as const uses Adamic's explicit literal
assertion exemption and is identified separately by proof_kind. Runtime-checked
casts remain source-open: the probe reports that a check is available, without
claiming that the whole tsc program builds or that code was emitted. A panic or
other unresolved probe stays open with an unknown unchecked-refusal result.

The cast probe is a Go overlay, not a compiler change. It disables the ordinary
loader output path, preserves the populated checker despite diagnostics, and calls
main's castProof independently at every adapted as expression. Its checker uses
main's Adamic defaults plus Node declarations. Those options differ from the compiler API's
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
The branch's return-contract changes can also change inferred declarations;
those receive immediate semantic evidence rather than token-only credit. Its two
new object-view assertions are counted, rather than assumed safe from the rewrite.
Main at 45487a80 contains 41-explicit-any-remaining, including scanner rules 1
and 4. There is no separate 42-scanner-any directory in that tree; scanner sites
are credited to the adaptation which actually edits them, 41. No missing adapter
is invented or credited.

41 rule 13 and 43's setTimeout rule replace the same stock return token.
Their complete declarations differ: main uses a parameterless handler and
NodeJS.Timeout, while 43 uses any[] and ReturnType<typeof globalThis.setTimeout>.
Preparation retains main's reviewed declaration and removes only the overlapping
43 timer rule from the extracted scratch rules. It checks that 41 has already
applied before running the remaining 43 rules. The original branch files stay
unchanged. composition.json records both alternatives. The API-based overlap
fixture finds the stock function's return token, requires exactly one ledger
entry credited to 41, and catches a mutant which appends a second credit to 43.

71-writable-views also matches the two JSON converters' complete signatures,
including their old any return annotations. Applying 43 first makes those exact
anchors fail. The initial run caught that mismatch before 71 wrote any files.
Preparation copies 71 into scratch and updates only its before/after return
annotations to 43's concrete types. Its parameter edits, audits and runtime
fingerprint checks are unchanged. The compiler API verifies both final return
contracts and both parameter types; a mutant restoring convertToJson's any
return is caught by that contract check. The two composed patterns are recorded in
composition.json; the repository's adaptation files remain untouched.

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
python3 stage3/step09-ledger/run.py /tmp/step09-new /tmp/step09-output > /tmp/step09-run.txt 2>&1
go vet ./... > /tmp/step09-vet.txt 2>&1
go test ./internal/load -count=1 > /tmp/step09-loader-tests.txt 2>&1
go test ./internal/oracle -count=1 -timeout 30m -run '^TestNativeAgreesWithNode/dedication/dedication.a$' > /tmp/step09-oracle.txt 2>&1
```

The fixture's single authored main.a has exactly one explicit any, one declaration
of type any, and one as cast. Its temporary adapted copy changes the any annotation
to number. The known answer is two rewritten sites and one safe upcast. Node's
transpiled source runs and reports value=1 widened=1. The real classification mutant
suppresses rewrite dispositions; it reports two open sites and is caught by that
independent expected answer, rather than a parser or build failure.

Refresh setup: Node 0.052s, Go 0.038s, submodules 0.182s, markdown 0.130s, clang 0.465s,
build 39.785s, cache 39.967s, total 40.052s. nproc is 5; CPU quota is 4.
The initial full scratch copy exhausted disk space; only newly created upstream
test copies were removed and the run resumed with source-only snapshots and the
single API baseline required by adaptation 40. The complete Go gate, native tsc
execution, and oracle/lane comparisons of the adaptations are not claimed by this
ledger unit. No production compiler or adaptation file is changed.

Current measured counts against main plus 43

| Kind | Stock sites | Rewritten | Safe upcast | Open |
| --- | ---: | ---: | ---: | ---: |
| Explicit any | 210 | 98 | 0 | 112 |
| Any-typed declarations | 214 | 55 | 0 | 159 |
| As casts | 6,231 | 8 | 2,266 | 3,957 |
| Total | 6,655 | 161 | 2,266 | 4,228 |

Adaptation 40 clears 125 stock ledger entries, 41 clears 24, 43 clears 11,
and 32 clears one. These are overlapping syntactic/declaration inventories,
not 161 independent runtime obligations. Relative to the earlier main-plus-43
measurement, 23 more stock sites are rewritten, six more are safe, and 29 fewer
are open. Scanner sites are credited to 41's actual rules, not to a missing
42 directory. The shared timer return contributes one entry to 41's 24.

The final tree adds 61 casts: 30 safe upcasts and 31 open. It also adds two
explicit-any tokens in 65's local ambient require return declarations; these
are open. open.json therefore contains 4,261 current open obligations.
All 3,988 open casts have a full-span Adamic probe: 3,757 receive the unchecked
cast refusal, 118 receive another refusal, 108 have a runtime-check plan,
two differ from the adapted API assignability result and remain open, and
three panic and remain unresolved. The complete adapted cast probe covers
6,284 casts with no missing spans.

Both tsc files outside compiler have zero sites of all three kinds. Per-file
counts including the added helper are in FILES.md and SUMMARY.json; complete stock
entries are in ledger.json; all current open obligations are in open.json.

The nested regression fixture gives one safe upcast and one unchecked refusal
for 1 as unknown as number. Their expression starts are identical; their end
positions differ. Probe matching therefore uses both start and end, and duplicate
proof spans fail the ledger rather than overwrite a result. The complete runner
passes on the refreshed full closure. The touched measurement fixture, rewrite
mutant, duplicate-credit mutant, JSON-return mutant, overlay vet, repository vet,
internal/load tests and the filtered
TestNativeAgreesWithNode/dedication/dedication.a oracle pass. Output is in evidence/. runner-fixture.txt retains the pre-refresh miniature
runner receipt; run.txt records the current full-closure run.
The canonical preparation CLI ran through 70, then caught the original 71
signature mismatch. After the documented scratch-only composition, preparation
resumed from that atomic rejection; the complete run.py measures the resulting
full closure. The preparation CLI includes that composition for future fresh
runs. Its immutable snapshots hard-link identical contents to save scratch disk.

Source excerpts in the JSON data come from Microsoft TypeScript 6.0.3,
copyright Microsoft Corporation, under Apache-2.0. Upstream trees are external
scratch inputs, rather than vendored source.
