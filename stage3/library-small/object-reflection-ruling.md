Built: checked Object.entries and Object.assign for plain scalar object slots under the October 8 ruling.
Commits: this topic starts at origin/main c966d4c4509ecaf37ae275dbaafae33291367ac8; no other worker branch merged.
Commands and outputs: focused oracle, lower, IR, JavaScript, native and CLI checks pass; counts refreshed (details below).
Mutants: entries check removal, assign check removal, physical key order and literal-domain widening are caught; visible-count and scalar-layout source mutants fail their tests.
Not covered: full scanner census replay and general runtime objects, reference-valued members, undefined/null discrimination or growable targets.

The scanner fixture reproduces the minimal count({x:number}) program from codex/stage3-scanner-native-3 a3a8c599, witnesses/13-entries.a. It now returns 1 in temporary .ts input on Node and both backends. Its original .a spelling retains the unproven-shape refusal.

The 17 ruling fixtures cover the scanner shape, hidden string and boolean fields with passing controls, literal string membership, mixed primitives, integer-index ordering and its boundaries, embedded NUL names, nested entries pairs, proven Record/index-signature producers, Object.assign hidden source fields, and a source key absent from the target declaration. Five proven .a fixtures also enter the ordinary oracle registry. Source fixtures are stored as .a; the ruling harness makes temporary .ts copies.

Assumption: Object.assign checks both the source slot domain and the declared target slot domain. A source key absent from the target declaration stops loudly, rather than creating storage. This is the conservative interpretation stated in the pending clarification. Present scalar target slots are required; optional, union-representation or index-signature targets still produce named NotYet stops. Ordinary fresh .a assign retains its prior fixed-shape behavior.

Visible checks: the explicit ObjectReflection IR operation is never removed from apparent type information. Program.ReflectionChecks counts admitted calls, and adamic build prints `adamic: checks: object reflection N`. The baseline had no general inserted-check report to extend. Static shape metadata now distinguishes number and boolean storage even when their physical layout is identical, and carries complete key byte lengths. Unsupported compatible runtime producers keep a named metadata stop. Existing checked RegExp groups retain their protocol.

Runtime owner review: only new separate helpers were added, internal/native/runtime/object_reflection.c and object_reflection.h. Existing runtime C files were not edited. Native reflection sorts canonical array-index keys before other keys in creation order, uses actual slots and generated primitive metadata, and checks before materializing or assigning each value. JavaScript uses Object.keys in the same sequence. Native runs include sanitizers, leak checks for successful runs, and a sanitized split-output build.

Validation commands (all output redirected to logs, no full package/gate run):

- `go test ./internal/oracle -run '^TestObjectReflection|TestNativeAgreesWithNode/internal/oracle/testdata/object_reflection_ruling' -count=1 -v` -> PASS, 49.381s; /tmp/object-ruling-oracle-certified.log.
- `go test ./internal/lower ./cmd/adamic ./internal/ir ./internal/javascript ./internal/native -run '^TestObject|^TestNumericEnum|^TestCallTarget|^TestRuntimeKeyIncludesEveryInput$|^TestUnitsPreserveSharedState$|^TestUniformFieldsMatchNode$|^TestRuntimeFieldLayoutsAreIncluded$' -count=1 -v` -> PASS for all selected packages; /tmp/object-ruling-packages-certified.log.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts` -> PASS, 67.615s; /tmp/object-ruling-counts-final2.log.
- `go test -overlay /tmp/object-reflection-count-overlay.json ./cmd/adamic -run '^TestObjectReflectionBuildReportsChecks$' -count=1 -v`: mutant removes ReflectionChecks increment; fails because successful build diagnostics lack the count.
- `go test -overlay /tmp/object-reflection-layout-overlay.json ./internal/oracle -run '^TestObjectReflectionRuling$/boolean_layout$' -count=1 -v`: mutant omits scalar types from native shape identity; fails because native incorrectly exits 0 while JavaScript checks and exits 70.

The permanent IR mutant tests remove the entries or assign proof (checked fixture then exits 0 and matches Node instead of stopping), select physical field order (integer-key stdout differs from Node), or erase literal membership (hidden literal fixture then exits 0). All mutated native binaries compile and run under sanitizers; failures are the intended oracle disagreement, not compiler errors.

Setup: GOPROXY=https://proxy.golang.org|direct; cloud/setup.sh succeeded; env=/workspace/adamic-tools/env.sh; node 24.19.0, go 1.27.1, clang 20.1.8; nproc=5 (quota 4). Timing seconds: node .039, Go .042, markdown skip .021/ready .130, submodules .187, clang .383, build 65.190, test-binary defer 65.354, cache 65.356, done 65.406. Setup log: /tmp/library-small-object-ruling-setup.log.

The mandatory broader count check found an existing RegExp named-group access blocked by the initial guard; the guard now delegates that established checked dictionary protocol. The focused native layout audit found a nonstandard shape declaration spelling; the new declaration now follows the audited spelling. Both were fixed before commit.
