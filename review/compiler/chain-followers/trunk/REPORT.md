Built: independent arrow receiver capture and host compiler residuals on main 5e33a17b; no lowering chain merged.
Commits: #rh16bxg source 33764e70 rebuilt as 8b6b0097b; #4my3dqg source 6c892bee is this commit.
Checks: Node, JavaScript, release native, ASan/UBSan native and leaks; main admission delta; lower shards; reader guard; counts; lane checks.
Mutants: both arrow mutants, twelve host overlays, and seven host IR output mutants caught; individual commands and results are preserved.
Not covered: skipped namespace, optional boolean and stricter options tasks; full repository gate; excluded host memoize runtime.

Task order and decisions:

| Task | Source | Result | Dependency or proof |
| --- | --- | --- | --- |
| #rh16bxg | 33764e70 | kept | Three receiver fixtures pass; constructor-context and dropped-capture mutants fail with undeclared C identifiers. Leaves 12.13, 12.14 and 12.13 seconds. |
| #agwccbw | ae6ea1d9 | skipped | Needs initialization-throws slice 71972e180, internal/lower/namespaces.go:444. Main emits terminal ir.Panic; the source guard and added sanitized witness fail because Node catches TypeError and continues while both backends exit 70. namespace-proof.log records this. Main's cache closure bound remains at internal/lower/new_class_value.go:167. |
| #p6a00tt | 1a03ba61 | skipped | Needs optional-presence slice a774d316 (rebuilt b215cedbd): internal/lower/optional_fields.go:11 optionalLiteralSlots, internal/ir/ir.go:399 ObjectLiteral.Missing, internal/native/runtime/object.c:178 adamic_object_absent. None exists on main. |
| #k881crd | 9128a35f | skipped | Its own strictness diff modifies the absent optional-presence member internal/lower/optional_fields.go:95 and assumes own-presence runtime storage from a774d316, object.c:178. Lands after that slice. Earlier chain candidate also exposed a contract decision: strict:true without noUncheckedIndexedAccess admits an absent indexed read whose source prints undefined but native exits 70; no language choice made. |
| #4my3dqg | 6c892bee | kept | Own net diff 031a1259..6c892bee, reconciled against main. Existing phantom helpers reused. Focused fixtures, mutants and host matrix recorded. |

The host source already excludes memoize graph runtime d70bc6b1, and no runtime dependency is imported. Main already has non-null checked semantics, nested-empty literals and debugger support; their focused proofs and mutants confirm these equivalent implementations. The kept residuals supply optional scalar concatenation, phantom brands, proven-predicate guard fixes, generic empty arrays and contextually specialized generic function values. Main's guards are retained. The three stage3 status files carry only the source task's seven changes; Node observations stay intact.

Runtime files changed: none. Roadmap contribution: compiler support and independent proof toward the host/compiler lowering work named by these tasks; no numbered step was supplied in the brief.

Commands are recorded in admission.py, host-matrix.py, lower-shards.json, host-mutants/results.json and proof logs. Every go test command is bounded. TestCallTargetReaders passed in 36.983 seconds. The host focused run passed lower in 1.887 seconds and oracle in 3.868 seconds; all changed leaves remain below 60 seconds. Counts passed in 62.343 seconds as the existing aggregate harness. Counts attribution names all fifteen new rows, with no changed old row. The initial counts attempt failed on missing @types/node and wrote no table; npm ci installed pinned dependencies before the one successful regeneration.

Setup timings (seconds): Go 0.064; Node 0.091; clang 0.501; Markdown dependencies 1.006; submodules 21.323; shared cache 28.023; go build 360.030; cache warm 360.126; done 360.152. nproc 5, CPU quota 4. Setup succeeded with GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh.

Final host matrix: all 25 host sources observed on Node and checked in both backends; 4/25 admitted and agree exactly, remaining stops preserved. Seven changed stage3 status rows proved independently: five admissions agree exactly with Node, including the negative import-order case (all exit 70 with the same ReferenceError); two retain their named refusals. All 294 lower tests pass in nine shards. Ten new standard oracle admissions and five new stage3 admissions agree with Node against main. Lane initially found missing top-level Parallel in the scanner mutant group; fixed and its three mutant leaves pass in 0.41, 0.42 and 0.44 seconds. Final reader guard passes in 1.785 seconds.

Final lane output: lane checks 11.1 s: gofmt and tools on 33 Go files, t.Parallel on 2 test packages; a-check 2 .a files; vet 2 packages.
