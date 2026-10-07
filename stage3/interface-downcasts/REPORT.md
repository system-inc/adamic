Built: construction-proof proposal and default-off base-interface cast prototype.
Commits: design f991deb837afea95339b1fae2f9346a930bec2cb, ledger update 18f2f624fe520e99675c82b2f0f38cd3d10efc93, refinement 36cecfd; prototype follows separately.
Commands and outputs: setup 95s, nproc 5; lower/load, selected uncached oracles, counts and five source mutants passed their expected checks; full gate interrupted after its compiler/oracle packages passed; details below.
Mutants: five construction-proof omissions and five runtime changes all caught by independent semantic assertions.
Not covered: unchanged tsc, staged builder/publication proofs, typed-layout fallback, enum syntax, generic/nested/callable targets, or verification of predicate bodies.

The proposal in [escape-hatches.md](../../docs/escape-hatches.md#cast-appendix-base-interfaces-and-construction-invariants) was committed and pushed before the prototype. It remains a proposal for the lead to take to @system_adamic, not an accepted language decision. No PR was opened. Main base: ef3d907ecdc4c771b016f7d9c52372def057a340. Checked-downcasts e816797 was read, not merged. No protected compiler file was edited.

The assertions input is `origin/codex/stage3-fixtures-assertions` 5b173f3920ab2c5b7058f0a9fe4b8e4a91f52523, `stage3/fixtures/assertions/ledger-summary.json`: 4,101 as sites, partitioned as 208 upcasts, 16 as const, 136 tagged union downcasts, 1,178 structural interface downcasts without a tag, 10 as unknown as nodes, 2,553 other. Of other, 1,842 are tagged narrowings outside a partitioned union, a candidate bucket rather than an exact base-interface count. Class downcasts are not separately counted. This unit did not remeasure or reclassify cast forms.

The follow-up refinement of only those 2,553 other sites is reproduced by [refine-other.cjs](refine-other.cjs), with [summary](other-summary.json) and [all locations](other-locations.tsv). Stock TypeScript 6.0.3 reports zero diagnostics; original source hashes, expression offsets/text, type strings (normalizing checkout paths only) and both assignability directions are checked against the shared ledger. Nested assertions are matched by both start and end offsets.

| Other refinement | Sites |
| --- | ---: |
| Tagged declared base-interface downcasts | 1,758 |
| Tagged interface unions without a valid partition | 57 |
| Tagged composite/generic/mapped narrowings | 25 |
| Tagged tuple narrowings | 2 |
| Other structural/union narrowings | 255 |
| Type-parameter narrowings | 3 |
| Nullish removal to interface | 6 |
| Non-assignable assertions | 343 |
| Any involved | 104 |
| Declared class downcasts | 0 |

The 1,758 include 1,497 single-interface and 261 interface-union targets, with 1,757 kind fields and one operator. Declared ancestry is checked through explicit extends chains, including instantiated interfaces; same-declaration generic refinements are not counted as inheritance. Node-to-Identifier occurs 27 times. The 25 tagged composites include Node-to-Mutable<Identifier> and stricter BindableObjectDefinePropertyCall shapes: a kind still does not prove their payload. The 57 union cases expose open members rather than sound closed partitions. This refinement does not reclassify the other original categories; zero class downcasts is scoped to the other bucket. Any erases the information needed to certify subtype identity.

To reproduce, read the original ledger and census files directly from their branches into scratch files, clone TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8 in scratch space, install the assertions branch's pinned API lockfile, symlink its node_modules into that checkout, generate diagnostic information with upstream scripts/processDiagnosticMessages.mjs, and run:

```
NODE_PATH=/tmp/interface-downcasts-api/node_modules node stage3/interface-downcasts/refine-other.cjs /tmp/interface-downcasts-typescript /tmp/interface-downcasts-original-ledger.json /tmp/interface-downcasts-census-files.json stage3/interface-downcasts > /tmp/interface-downcasts-other.log 2>&1
```

The predicates input is `origin/codex/stage3-fixtures-predicates` e42eaf9854563617734ff13987cc27e4e09f78b1, `stage3/fixtures/predicates/ledger.json`: 651 predicate syntax nodes and 11 fixtures. Its categories describe bodies, not verified contracts. Both kinds of narrowing need a construction certificate; negative guard branches additionally need a body test equivalent to target membership. Empty, disabled, bodyless and generic assertions remain separate proof obligations.

Construction observations use the compiler corpus in `cohere/TypeScript/tsc/testdata/fixtures/compiler`, through cohere 715ba94f3608a6500086b1076ce5cb7e51b836db and TypeScript Go 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. No cohere code was copied. [constructions.py](constructions.py) reads that checkout in place; [constructions.json](constructions.json) pins hashes and locations. Its lexical discovery found 108 direct generic base allocations and 203 interfaces explicitly declaring a kind. These are construction observations, not cast counts or completeness certificates.

The Identifier base constructor does not initialize escapedText; createBaseIdentifier initializes it later. Common parent and required symbol are initialized with undefined!, and interface brands are absent. Kind is shared with stricter Identifier variants. Allocators are replaceable and cloneNode copies dynamically. Therefore "every allocation with this kind already has the entire declared interface" is false. The assertions ledger's independent genuine-factory mutation witness prints true and undefined on Node with zero TypeScript diagnostics. Publication conventions may explain ordinary visitors, but proving them needs initialization and escape analysis across callbacks, clones, overrides and writes. This unit does not certify every tsc construction site.

`ADAMIC_INTERFACE_DOWNCASTS=1` enables the prototype. The source and target must have required readonly scalar fields and the target must narrow a field to one literal value. Initializer types prove the payload. A per-cast whole-import-graph scan rejects incomplete matching-tag literals, wrong payload types, uncertain broad tags without payload, possible writes through any alias, generic/bodyless/opaque construction, any, reflection, spreads and unsafe assertions. Unused and unrelated matching-tag literals also participate; this deliberately conservative policy can reject unreachable or independently valid programs. Existing lowering still refuses unsupported operations. The scan is not a general builder verifier or reachability analysis.

The check is option (b): emit the discriminant comparison only after construction proof; otherwise refuse. Runtime proof cost is zero. The existing cast performs a field read, comparison and failure branch; string equality can compare tag bytes, and a field-layout cache miss scans names. No additional allocation or payload copy is introduced. Current compile-time scanning is O(C*N + C*A*F) plus checker relation costs. The document gives shape-check alternatives and their costs; existing layout names/reference flags do not encode complete field types, initialization or alias safety.

[nodes.a](nodes.a) defines Node, four subinterfaces and complete factories. [visitor.a](visitor.a) holds values as Node and checks identity and single operand evaluation. Original source on Node prints:

```
idid
42
texttext
5
true
once2
calls 2
```

Wrong-kind source prints casting and otherother on Node because its unchecked cast does not change the object. Both checked backends instead print casting, panic with `cast failed: this Node is not a Identifier`, and exit 70. [missing-name.a](missing-name.a) prints missing on Node and is refused by the compiler. Numeric and boolean tag success/failure cases are generated as .a files by the oracle test. Successful native executions pass ASan/UBSan and leak checks; failing executions pin the checked contract independently of Node's unchecked semantics.

Mutants and their evidence:

| Mutant | Independent catch |
| --- | --- |
| tag-only, omit construction proof | Missing-name refusal assertion fails. Accepted mutant emits valid C and reaches the runtime missing-field compiler-bug guard after passing the tag. Clang and sanitizers are not the catch. |
| omit required-field presence | Lowering test expects missing-payload refusal, gets nil. |
| omit payload type relation | Lowering test expects wrong-payload refusal, gets nil. |
| ignore alias writes | Lowering test expects write refusal, gets nil. |
| scan only entry module | Imported malformed factory refusal assertion fails. |
| skip runtime tag | Valid C exits 0 and prints otherother where checked contract requires exit 70. Hidden payload keeps this mutant from relying on a missing-field guard. |
| skip numeric runtime tag | Checked exit 70 becomes valid C exit 0 with payloadpayload, matching unchecked Node. |
| skip boolean runtime tag | Checked exit 70 becomes valid C exit 0 with payloadpayload, matching unchecked Node. |
| wrong runtime tag | Visitor exits 70 where Node and checked success require 0. |
| evaluate operand twice | Visitor prints once3 and calls 3 against Node's once2 and calls 2. |

[mutants.py](mutants.py) runs the five source mutants sequentially, restores the compiler in finally, and rejects build/sanitizer failures as catches. Runtime mutants alter IR in the oracle test. Logs preserve all ten catches.

Commands (all test output redirected directly to logs, never piped):

```
bash cloud/setup.sh > /tmp/interface-downcasts-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
python3 stage3/interface-downcasts/constructions.py > stage3/interface-downcasts/constructions.json
go test ./internal/lower -run '^TestInterfaceConstruction' -v -count=1 > /tmp/interface-downcasts-proof.log 2>&1
go test ./internal/lower ./internal/load -count=1 > /tmp/interface-downcasts-packages.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInterfaceCast' -v -count=1 -timeout 30m > /tmp/interface-downcasts-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInterfaceCastScalarTags$' -v -count=1 -timeout 30m > /tmp/interface-downcasts-scalars.log 2>&1
python3 stage3/interface-downcasts/mutants.py > /tmp/interface-downcasts-mutants.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/interface-downcasts-counts.log 2>&1
gofmt -l cmd internal > /tmp/interface-downcasts-format.log
go vet ./... > /tmp/interface-downcasts-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/interface-downcasts-gate.log 2>&1
```

Setup timing: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 95s, done 95s on 5 processors; CPU quota 4. Tools: Go 1.27.1, clang 20.1.8, Node 24.19.0. Targeted proof tests pass 0.418s; lower/load pass 8.390s/0.607s; uncached interface oracle passes 1.882s; scalar-tag oracle passes 4.098s; counted table update passes 14.723s. gofmt, full vet and final lower/oracle vet produced no output. The counted rows use a child compiler process to keep the opt-in flag isolated from parallel ordinary fixtures: visitor 21 allocations/21 frees/33 retains/41 releases/13 peak/0 regions; wrong-kind 2/0/2/1/2/0, with panic intentionally terminating cleanup.

Full-gate result: interrupted deliberately after more than 15 minutes of unrelated adapter execution (process exit 143), not claimed as a complete pass. No package failure was observed before stopping. The log records passes for bridge/tsgo (466.991s), internal/lower (37.807s), internal/load (1.688s), internal/native (219.322s), internal/oracle (204.623s), flow, fresh, fuzz, regexp, unicodeproperties (775.528s), and several upstream adapters. The pending lint/markdown and subsequent adapters are not covered by a completed full gate. The packages touched and the filtered uncached oracle had already passed independently; the additional numeric/boolean omission mutants passed their separately rerun scalar test. Final targeted vet of lower/oracle and formatting/diff checks also pass. All logs are retained in [logs](logs).
