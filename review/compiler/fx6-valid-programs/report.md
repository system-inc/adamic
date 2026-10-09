Built receiver-type/member checked-read registration and destructured contract IDs for items 134 and 136.
Implementation commit: dd5712e2; current-main merge: fd3cf831 (main a7448d73).
Final checks: internal/lower 59.228s; TestCallTargetReaders 0.697s; oracle fixture/review slice 11.325s; counts refresh 82.311s; vet exit 0.
Both revert mutants fail Node agreement: receiver registration loses p37 output; missing destructuring ViewTypeID loses p53 output.
Item 137 remains open: original p70 prints 2 on Node, JavaScript exits 70, native emission still panics.

The checked read registry now uses the receiver checker type identity and member name. Nested contracts register their own receiver identities. Record-storage representation checks retain their separate bare-name registration. Destructuring records the receiver identity, field type identity, and source location. The lazy-contract consumer uses the same receiver key. The predicate admission test now checks registration against the exact target member type identities.

Original task witnesses became available in the fetched main and replaced the reconstructed p35, p36, p37, p38, p49 and p53 programs. p38_number additionally holds the original number | undefined witness. Seven valid lower fixtures use lowersAndAgreesWithNode, with separate top-level parallel tests. The oracle runs release native, ASan/UBSan native, and JavaScript against Node; the standard review slice also checks leaks. Fixed review witnesses no longer have pending sidecars.

Original p51 is named callable_union_destructure_wrong and gives a function taking string to a number-parameter callable union. Node prints undefined. All three compiler backends now stop before the call, with the same incompatible-parameter diagnostic. I chose the sound checked stop rather than reproducing the unchecked result. A supplemental valid scalar-union p51-shaped fixture agrees with Node. p49 likewise stops at its exact union-membership diagnostic; both wrong originals have intended sidecars pinned to item 136's checked-stop requirement.

The original p70 is retained as p70.a here, and its existing review pending sidecar remains. Receiver registration does not cover it: Sized is an object contract while the actual value is a tuple. JavaScript rejects the array at viewed.value; native still reaches the incomplete checked-field adapter at view_fields.go:26. Supporting the tuple/object view and its length projection requires additional representation/backend work, beyond the two completed fixes. See p70-observation.log for the failing observation. No refusal was substituted for this valid program.

Final commands (output files in this directory):

- go test ./internal/lower -count=1 -timeout 90s: PASS, lower-final.log.
- go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s: PASS, call-targets-final.log.
- go test ./internal/oracle -run '^TestFX6|^TestReviewProgramsAgreeWithNode/fxspptb_oct9_views_p(35|36|37|38|49|51|53)' -count=1 -v -timeout 90s: PASS, oracle-final.log.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 3m -args -update-counts: PASS, counts-final.log. The initial 90s run timed out; the completed refresh records the final fixture sources.
- go test ./internal/lower -run '^TestFX6' -count=1 -v -timeout 90s after restoring mutants: PASS, lower-fixtures-final.log.
- go vet ./internal/lower ./internal/oracle: exit 0, vet.log.
- Standard oracle fixture slice also passed, fixtures-gate.log.

Mutants were applied and then reversed with git apply. revert-receiver.diff restores name-only registration/readiness/lazy demand, retaining the key helper so the updated test file still compiles. TestFX6P37 fails with JavaScript stdout "3\n" versus Node "3\nnumber 5\n" (0.14s). revert-destructure.diff removes only ViewTypeID. TestFX6P53 fails with empty JavaScript stdout versus Node "n5\n" (0.15s). These are behavior failures, not build failures. A preliminary complete destructuring revert also removed receiver identity and survived the valid scalar fixture; the final precise mutant isolates the contract-ID fix.

Final lower fixture seconds: P35 0.29, P36 0.29, P37 0.14, P38 0.14, P38Number 0.27, P51 0.28, P53 0.28. Oracle fixture and review leaves all finish below 2s in the final run (oracle-final.log).

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh. Timing lines: Go ready 0.064s, Node ready 0.085s, clang ready 0.514s, markdown ready 1.103s, submodules ready 26.074s, Go build ready 305.162s, done 305.370s. nproc=5; cgroup quota=4 CPUs. The printed environment path was /workspace/adamic-tools/env.sh; /opt/adamic-tools/env.sh does not exist here. The required full lower run initially found missing @types/node; npm ci --prefix stage3/api installed the pinned dependencies, after which it passed. No dependency manifests changed.

Delivery advances task #1fk58py's valid-program view correctness work, items 134 and 136. Item 137 and Node agreement for the deliberately wrong original p51 are not claimed. No full repository gate was run.

Integration lane checks passed: lane checks 1.2 s: gofmt and tools on 7 Go files, t.Parallel on 2 test packages; vet 2 packages. The checkout initially fetched only main; explicit tracking ref fetches made origin/devtools/fast-gate and origin/cloud/merge-tree available, then the required literal lane-check command passed.
