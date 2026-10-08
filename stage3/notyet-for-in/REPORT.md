Built runtime for-in enumeration for both assigned kinds; the earlier ruling-only result is superseded.
Implementation commits d90f6c61 and 25a9e938; checked non-null c41c0e06 merged in 7a7f3905; current main 5fda2d26 merged in 21773ca4.
Packages pass: lower 89.992s, native 332.005s, IR 67.000s, fresh 145.193s, oracle 185.335s, flow 115.983s, cmd/adamic 22.601s; JS has no package tests. Counts pass 32.869s.
Ten semantic mutants produce wrong stdout; two representation-proof mutants fail their negative assertions.
Ten origin roots and one array root lose these guards; three representative replays reach named next stops. Sparse array construction remains unsupported.

## Results

| Kind | Status | Assigned roots | Replayed next stop |
| --- | --- | ---: | --- |
| Unproven fixed plain-object origin | Lowered | 10 | commandLineParser.ts:2788:24 and :2975:24 reach core.ts:1266:29, `a value of type object` |
| Array enumeration | Lowered | 1 | factory/nodeFactory.ts:7538:23 reaches :7539:9, `an array index that isn't a number` |

These are selected-unit observations on the checker-rejected census project, not a claim that tsc compiles. The other eight origin roots were source-reviewed, not replayed. Replays use the exact adapted source: all 81 sizes and SHA-256 values match the census manifest. The replay command retains the old reason, so exit 1 means that signature no longer occurs; the JSON records the new stops.

The receiver is evaluated and retained once. Both backends snapshot enumerable keys and check presence before each turn. Native snapshots distinguish absent synthetic storage from present undefined, preserve string insertion order, sort canonical indices, include represented array properties and inherited static fields, and handle scalar and string values hidden behind structural object parameters. A conservative all-call-target use proof permits heterogeneous structural slots only when every use forwards to enumeration; field reads do not acquire an unsafe representation cast.

Nullable spread objects receive a tagged hidden native slot containing their actual ordered keys. Copying makes this metadata independent; writes append only newly present names. The slot is not a JavaScript property, including when a real property uses its storage name. Supported dense arrays enumerate live indices and RegExp own fields; sparse literals still reach `an OmittedExpression in an array literal`. No other worker's lower function was changed. Prototype mutation and arbitrary expandos retain their existing language restrictions.

Runtime-owner review: **internal/native/runtime/for_in.c** is a new separate helper file; no existing C runtime file or header was edited. Its helpers provide key snapshots, presence checks, independent metadata copies and metadata writes. The RegExp array storage shape's disabled fields are excluded from enumeration.

## Fixtures and mutants

All new fixtures are .a and registered from for_in_ruling_test.go. for_in_options_refused.a now lowers despite its historical filename; for_in_array_refused.a still holds Node's sparse-array behavior and asserts the next construction stop. for_in_runtime.a covers missing fields, present undefined, copy independence, insertion/numeric order, hidden arrays, popping during iteration, one receiver call and source-binding reassignment. for_in_static.a covers inherited static fields; for_in_primitives.a covers numeric/boolean boxing and UTF-16 string indices; for_in_storage_name.a covers hidden-slot name collisions; for_in_array_properties.a covers ordinary match versus /d indices and indices.groups.

| Mutant | Catcher |
| --- | --- |
| synthetic_keys | Runtime fixture stdout differs from Node |
| shared_presence | Copied object's new key incorrectly appears in original |
| missing_write | Newly added synthetic property absent from snapshot |
| numeric_order | Runtime snapshot has insertion order instead of canonical index order |
| removed_key | Popped array index executes; native and JS stdout differ |
| receiver_twice | Receiver called twice; native and JS stdout differ |
| own-only inherited keys | Static fixture omits inherited base key |
| scalar string boxing | Primitive fixture erroneously enumerates number's string characters |
| metadata exposed through Object.keys | Runtime fixture exposes absent storage and hidden metadata in stdout |
| disabled RegExp indices present | Array-property fixture adds indices without /d; exit 0 and stdout mismatch |
| ignored structural property read | TestForInStructuralSlotProof wrongly accepts field use |
| first virtual target only | TestForInStructuralSlotVirtualTargets wrongly accepts a target with field use |

The first nine semantic mutants are executable tests in for_in_mutant_test.go. The RegExp mutation and two proof mutations use scratch Go overlays; their diffs and logs accompany this report. All semantic mutants build and exit normally with clean sanitizer stderr; stdout alone kills them. Proof mutants fail explicit negative assertions.

## Validation commands

Every shell sources /workspace/adamic-tools/env.sh. Output goes directly to logs, never a pipe.

```sh
go test ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/fresh ./internal/oracle -count=1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
go test ./internal/native -overlay=/tmp/notyet-for-in-proof-field-use/overlay.json -run '^TestForInStructuralSlotProof$' -count=1
go test ./internal/native -overlay=/tmp/notyet-for-in-proof-virtual-targets/overlay.json -run '^TestForInStructuralSlotVirtualTargets$' -count=1
go test ./internal/oracle -overlay=/tmp/notyet-for-in-property-mutant/overlay.json -run 'TestNativeAgreesWithNode/internal/oracle/testdata/for_in_array_properties' -count=1 -v
go run ./stage3/census/latent/replay -project /tmp/notyet-for-in-census-input/src/tsc/tsc.ts -where /tmp/notyet-for-in-census-input/src/compiler/commandLineParser.ts:2788:24 -kind NotYet -reason 'for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly)'
```

Repeat replay with commandLineParser.ts:2975:24, and factory/nodeFactory.ts:7538:23 with the old array reason. Evidence JSON gives exact observations. Fixture oracles compare source Node, JavaScript backend, release native, ASan/UBSan and leak checks. No full repository gate was run.

Setup succeeded: Node 0.027s, Go 0.040s, clang 0.249s, markdown ready 1.077s, submodules 16.447s, Go build 221.748s, done 221.886s; nproc=5, CPU quota=4. Go 1.27.1, clang 20.1.8, Node 24.19.0. GOPROXY used https://proxy.golang.org|direct.

The two kinds share the runtime dispatch and structural-slot proof, so they form one minimal functional implementation group. The largest-kind witness was implemented first. Outside-function production changes are limited to IR, fresh analysis, both emitters and native ownership/copy hooks; the commit body names every changed file.

## Requested checked non-null merge and landing

Merged c41c0e062e99da37820f822968d4df1b48cdaee7 in 7a7f3905, then current origin/main 749a69adbfae2a7bf22c1f0436d9ef73345c3070 in b16699be. Both merges are clean; the implementation group and merges were pushed to codex/notyet-for-in. Replayed all three examples again after both merges. The origin loop bodies still reach core.ts:1266:29 `a value of type object`; the array body still reaches factory/nodeFactory.ts:7539:9 `an array index that isn't a number`. The 2788 unit also has earlier independent stops `a Map of CompilerOptionsValue` and `a BinaryExpression with a value and a value`, and a later `reading result` stop. None of these three representative results is a NonNullExpression stop. All three old-reason replay commands exit 1 because their former for-in signatures are gone.

Final isolation fix: programs using for-in must also keep the new tagged metadata private from existing Object.keys operations. The own-key hook uses tracked real keys and retains class own-only behavior; for-in alone includes inherited static fields. The new Object.keys mutant exits normally under sanitizers and exposes absent fields plus the hidden storage key, caught only by Node stdout. Focused final oracle/mutants pass in 12.944s; counts refresh passes in 32.869s. The sparse refusal witness is excluded explicitly from runnable flow graphs; its omission is not a weakening of the Node refusal oracle. Full flow passes in 115.983s after that registration fix.

Validation history: the package run started before the requested merge timed out its native package at the default 10-minute limit and observed counts while the merge changed their rows. The post-merge package run passes lower (89.992s), IR (67.000s), fresh (145.193s), oracle (412.681s); JavaScript has no package tests. Native again reaches the default 10-minute limit. Command tests pass in 22.601s. A subsequent native run was cancelled because the Object.keys isolation fix superseded its source snapshot. Final native and oracle commands use a 30-minute timeout; their results are recorded below after completion. No timeout is counted as a semantic mutant kill.

```sh
go test ./internal/flow -count=1 -timeout 30m
go test ./internal/native -count=1 -timeout 30m
go test ./internal/oracle -count=1 -timeout 30m
go test ./internal/oracle -run 'TestForIn|TestNativeAgreesWithNode/internal/oracle/testdata/(for_in_|library_for_in)' -count=1 -timeout 30m -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
```

Latest-main landing: merged origin/main 5fda2d266f6d213a781e7eff70df9ab7a80b018b in 21773ca4. Its changes affect only the census lane and velocity record, with no compiler or fixture input changes. The metadata isolation and flow registration fix is 25a9e938. Final full oracle passes in 185.335s with the final metadata helper and fixture counts.

Final full native passes in 332.005s with the final compiler and runtime source. Final full oracle passes in 185.335s. Both commands exit 0; all earlier timeout and registration failures are superseded by these successful checks and the successful full flow run. git diff --check is clean. Final report and evidence are committed and pushed to the own branch; no PR is opened.

Rechecked scratch-overlay mutants on final source: RegExp property presence fails only stdout (both processes exit 0, clean stderr), 11.259s; ignored field-use proof fails its assertion, 0.008s; truncated virtual targets fails its assertion, 0.009s. These expected exit-1 results are saved alongside the passing final packages.
