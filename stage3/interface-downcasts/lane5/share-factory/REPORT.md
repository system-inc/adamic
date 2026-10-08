Built: resolved callable arguments and declared overload-set checks; 3 pairs / 368 candidate reads certified toward roadmap step 09.
Commits: base a62778110b958f54d52622da01bf7758bce629dc; the first delivery commit carries this report on codex/views-callables-factory.
Commands/results: uncached focused lower/native/JavaScript/oracle checks pass; ten fixture count rows and declaration checks pass; vet passes; global counts remains red outside this unit.
Mutants: three last-overload omissions, three undefined-packing omissions, both backend result guards, resolved declaration membership, and overloaded-union admission are caught by semantic assertions.
Uncovered: five pairs / 130 reads remain for subsequent groups; full TypeScript execution, complete adjacent carrier types and exact runtime reachability are not measured.

The first group certifies NodeFactory.createIdentifier (206 reads),
createStringLiteral (104), and createUniqueName (58). These are original member
signature fixture certificates against the existing candidate-read inventory.
The original good fixtures are byte-for-byte copies of share a's pinned controls.
Each additional fixture retains both original member declarations and the original
member read. Adjacent carriers remain reduced. SyntaxKind and
GeneratedIdentifierFlags use number representation; enum membership is not certified.

Calls use GetResolvedSignature on the call node. For an overload set, the
resolved declaration must belong to that set. Each argument is converted to
its resolved parameter representation, including explicit undefined. The callable
member read independently checks every declared signature against immutable
producer metadata selected by generated code identity. Arguments compare
contravariantly and results covariantly. Extra producer-ignored arguments still
execute; omitted producer parameters must admit undefined. The closure ABI,
receiver evaluation and callable identity are preserved.

Map stores remain on single signatures. The existing callable-union selector
also retains its single-signature alternatives; conjunctive overload sets cannot
borrow its disjunctive proof. Both boundaries retain refusals rather than supplying
an unsupported certificate. No protected compiler file is edited. No other lane
branch is merged, and no cohere source is copied.

Every positive control matches source Node in native release, native ASan/UBSan,
and generated JavaScript. Finishing positives pass the independent leak checker.
Wrong-overload producers satisfy the first declaration and fail the second.
All three modes pin exit 70, empty stdout, and the complete field/type/signature
message. Dropping the second declaration then prints 7 with exit 0 in all three
modes, with finishing leak checks. That independently proves full-set coverage.

The conversion fixtures cover the longest signature, omitted arguments, and
explicit undefined. Removing only undefined packing produces clean native output
9/110/0 instead of 9/110/110, 3/110/0 instead of 3/110/110, and 11/100/3 instead
of 11/100/103. Release and sanitizer runs catch all three; no sanitizer failure
or invalid C is used as evidence. Generated JavaScript retains source semantics.

The isolated wrong-result control calls a void producer and discards the
advertised result. Native and JavaScript result-guard omissions print read with
exit 0; the pinned signature-1 failure catches each. Native sanitizer also
finishes with that forbidden output. Bypassing resolved-declaration membership
admits an excluded third declaration and fails the dedicated lower assertion.
Admitting an overloaded alternative through the union selector fails the
unsupported-read assertion. The probe supplies the registry a real cast supplies;
an earlier nil-map failure was rejected as mutant evidence. All sources are restored.

The unit adds ten rows to internal/oracle/counts.md, changing no existing row.
The required global updater was run before and after installing the pinned
stage3/api dependencies. Missing Node typings are resolved. The latter run exits
1 in 81.418s on fixtures outside this unit. Four representative failures reproduce
on an unchanged a6277811 worktree: census_overload_contracts, census_small_boolean,
maybe_number_slots, and graph_regions_regression_07. Other global failures,
including errno-field input cases, are not independently triaged here. The global
table is not certified green. The unit updater and subsequent recorded-row check
pass.

Commands, with test output written directly to logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/views-callables-factory-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --ignore-scripts # in stage3/api; output captured in its setup log
python3 stage3/interface-downcasts/lane5/share-factory/run-mutants.py > /tmp/views-callables-factory-guard-mutants.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/views-callables-factory-counts-global-with-api.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableFactoryCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/views-callables-factory-counts-owned.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run '^TestResolvedCallableDeclarationMembership$|^TestPrepareViewCallableRead$|^TestViewCallableShapeContract$|^TestViewCallableProducerCertificate.*$|^TestViewCallableShape.*$|^TestCheckedViewCallableFactory.*$|^TestCheckedViewCallableShareA$|^TestCheckedViewCallableShareAFrontiers$' -count=1 -v -timeout 10m > /tmp/views-callables-factory-first-group-final.log 2>&1
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle > /tmp/views-callables-factory-vet.log 2>&1
git diff --check
```

Final package results: lower 1.025s, native 8.077s, JavaScript 0.933s,
oracle 108.148s, all PASS. The final oracle includes share a's existing 30 pairs,
its 60 arity-mutant executions, the eight original frontier controls, and this
unit's fixtures, mutants, declarations and recorded counts. No whole-package
test or full repository gate is run.

Setup timing: Node 0.167s, Go 0.168s, clang 0.875s, markdown 1.491s,
submodules 217.572s, build 615.586s, cache 615.765s, done 615.813s.
nproc is 5; cpu.max is 400000 100000. The environment file is
/workspace/adamic-tools/env.sh. Original-source provenance and candidate reads
remain in certificates.json; executable observations are in logs/.
