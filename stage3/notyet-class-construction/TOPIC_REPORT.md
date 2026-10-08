# Topic-only class construction delivery

Base: origin/main 6998ebc24ae353193cb1495d3d51308131a4b5c7, resolved by fetch for this delivery.
Census: codex/stage3-notyet-table e8c283b5, stage3/notyet-table/rerun-0730/after/roots.csv.
Branch: codex/notyet-class-construction-topic.

This branch carries only this worker's non-merge commits: 19943ae5 (from 31af91e6), aa3aaa38 (from 490aadcd), 7ed6acac (from 4e0d9b0a), e3c42245 (from 6353ee98), plus this delivery evidence commit. There are zero merge commits beyond the pinned main. No area/compiler, replay, checked non-null, binary, or statics dependency commits were carried. The older REPORT.md, CLASS_INTERFACE_RULING.md and their logs are historical evidence, not validation on this topic base.

## Surviving original kinds, largest first

The morning census lists 49 field roots and one base root for the original cluster. The replay worker was built against this topic tree through a scratch Go overlay, not merged or committed. All 81 frozen source manifest files matched byte counts and SHA-256 hashes. Replays observe checker-rejected entry-root programs and are not proof that TypeScript's compiler builds.

| Morning roots | Kind | Status |
| ---: | --- | --- |
| 24 | a field of type string \| NodeArray<JSDocComment> \| undefined | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 7 | a field of type string \| number \| undefined | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 4 | a field of type NodeArray<ParameterDeclaration> \| readonly JSDocParameterTag[] | Skipped: first example stops at PrefixUnaryExpression on a value; second stops at NonNullExpression. Earlier merged-tree attribution was property, outside this worker's functions. |
| 4 | a field of type "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 2 | a field of type boolean \| (() => boolean) \| undefined | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 1 | a field of type AnyBuildOrder \| undefined | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 1 | a field of type false \| string[] \| undefined | Skipped: selected site stops earlier at PrefixUnaryExpression on boolean | undefined; same replay reaches the identical field kind at moduleNameResolver.ts:2277:12. |
| 1 | a field of type "boolean" \| "list" \| "number" \| "object" \| "string" \| Map<string, string \| number> | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 1 | a field of type "boolean" \| "number" \| "object" \| "string" \| Map<string, string \| number> | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 1 | a field of type 0 \| boolean \| undefined | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 1 | a field of type boolean \| (() => boolean) | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 1 | a field of type string \| false | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 1 | a field of type string \| false \| undefined | Skipped: exact signature reproduces in object.go::property, owned by the object-property worker. |
| 1 | a base that isn't a declared class | Skipped: signature did not reproduce; next stops are EmitNode intersection union at utilities.ts:423:34 and reading autoGenerate at 424:18. No new base lowering rule justified. |

The string/number kind's first example stopped earlier; its second example checker.ts:51074:20 reproduced in object.go::property. Ownership was checked with git log origin/codex/notyet-object-property -- internal/lower/object.go; latest observed commit 5a870515 owns the property's work. No other worker's lowering function was edited in this pass. No new original-cluster kind is claimed lowered or refused for a ruling. None was newly implemented with checked views; the blocked intersection/union representation is left with its owner. The topic-only instruction prevents importing the checked non-null dependency to remove the secondary example's earlier stop.

## Carried ruling and verification

The fresh class/interface literal remains structural, property presence preserves its origin, calls dispatch by the runtime value, and a class downcast checks its nominal tag. The exact ruled fixture prints `A literal literal` and `A a woof`, matching source Node in both backends. Four carried .a fixtures pass the oracle, including sanitizer/leak checks and two checked-stop fixtures. All other nominal-class refusals remain tested. No new fixture was added for a stop belonging to another worker or an unreproduced signature.

Commands on this topic tree (all test outputs saved directly to logs):

```sh
export GOPROXY='https://proxy.golang.org|direct'
ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/class_interface_' -count=1 -v -timeout 10m
python3 cloud/notyet-class-interface-mutants.py /tmp/class-topic-mutants
go test ./internal/lower -run '^(TestClassInterfaceRuling|TestInheritance|TestUncheckableCastsStayRefused|TestCheckedCastProofAndElision)' -count=1 -timeout 10m
go test ./internal/ir -run '^Test(CallTargetsIncludeEveryDescendant|ClosureTargetsBoundOnlyProvenValues|CallTargetReaders)$' -count=1 -timeout 10m
go test ./internal/native -run '^Test(PassThroughsAreNotConsumers|RuntimeFieldLayoutsAreIncluded|UniformFieldsMatchNode)$' -count=1 -timeout 10m
go test ./internal/javascript -run '^TestClassInterface' -count=1 -timeout 10m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 20m -args -update-counts
(cd cohere && go test ./internal/lint/rules/adamic -run '^TestNominalClass' -count=1 -timeout 10m)
```

Setup timing: markdown ready 0.367s, clang ready 0.537s, Go build ready 71.193s, test binaries deferred 71.524s, cache warm 71.526s, done 71.573s. nproc=5, cgroup quota=4 CPUs. Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup succeeded. Oracle PASS 17.061s; lower PASS 1.469s; IR PASS 27.431s; native PASS 1.290s. JavaScript has no package test files and is covered by the differential oracle. Counts PASS 33.859s, with no counts.md changes required. Nominal-class PASS 0.093s. Results are saved with this delivery's evidence.

Every mutant ran and was killed for its intended reason:

| Mutant | What caught it |
| --- | --- |
| erase-inferred-origin | presence fixture stdout mismatch |
| static-value-dispatch | exact ruled fixture stdout mismatch |
| static-statement-dispatch | presence fixture stdout mismatch |
| refuse-structural-literal | nominal-class refusal on admitted literal |
| unchecked-contextual-class | checked fixture failed expected exit-70 check |
| unchecked-explicit-class | cast fixture failed expected exit-70 check |
| own-only-presence | presence fixture stdout mismatch |
| invoke-getter-on-presence | presence fixture stdout mismatch |

The carried runtime addition internal/native/runtime/class_interface_property.c is a new separate helper and still needs runtime-owner review. No existing runtime C file was edited. The implementation commit aa3aaa38 names every file outside the owned function. This pass adds only evidence and refreshes counts if needed. No full package suite or full gate was run.
