Built: rebased existing wave-17 onto area b84a9d93 containing current main c7991b90; no new rules or claims.
Commits: validated rebased source 00deb49fa0e9d6484dca8acbd39f03be84a9afbc, replacing owned prior tip d945f82e.
Checks: all eight lint suites PASS 897.690s; sanitizer, handles, checker/registry, named listeners, vet and filtered uncached Node checks pass.
Mutants: all existing rule/question/handle mutants caught; fixed RegExp byte 1134, named-kind Unknown and Unicode byte 1 caught.
Limits: parked React analysis/configuration and driver adapters remain incomplete; no full gate or 17 required stage-1 correctness checks run.

The requested landing-first refresh found current main
c7991b900362796aefd111474e65eb5398e91953 and lint area
b84a9d9314b65d3d0261ee017e233287b4f071da ahead of the owned branch.
All 34 own commits rebased cleanly onto the area, which contains current main.
The required shared model 50a5f105 and harness 41eb6eab2 remain ancestors.
The integration update contains compiler predicate/proven-relation changes and
record runtime support. They were accepted through rebase; no protected compiler,
shared harness, runtime, generator or parser file was edited by this worker.
The before-push remote check still reports these main/area tips and the owned
lease d945f82e084b8424d751b1366cbb94080cd0f9d8. Only the owned branch
codex/typeaware-wave-17 is pushed, never main or any area branch.

| Selected suite | Time | Result |
| --- | ---: | --- |
| TestCoverageAgreementAndMutants | 159.18s | PASS |
| TestWave17FifthAgreementAndMutants | 161.58s | PASS |
| TestWave17FourthAgreementAndMutants | 147.01s | PASS |
| TestWave17HeadJudgmentsAndNativeJSX | 66.75s | PASS |
| TestWave17NextAgreementAndMutants | 125.77s | PASS |
| TestWave17AgreementMutantAndNativeJSX | 71.38s | PASS |
| TestWave17ThirdAgreementAndMutants | 163.69s | PASS |
| TestWave17UnicodeUpper | 2.34s | PASS |

The complete Go comparison includes finding ranges, IDs, messages, automatic
fixes and all suggestions. Supported upstream/edge controls, option variants,
frozen repository and compiler corpora, ordinary/sanitized runs, rule and raw
checker-question mutants, and released-handle checks all pass on the rebased
compiler. The Next head suite uses the real native JSX parser for all 40 controls;
the original three rules also retain their exact-byte JSX witnesses. Existing
React projected-analysis experiments are not relabeled as complete production
integration. Unicode matches Go on all 1,114,112 code points.

Every mutant result is preserved in mutant-evidence.txt. Decision mutants exit
successfully and are caught by the complete byte comparison. Released-registry
mutants finish cleanly but are caught by the required panic-70 check. The extra
fixed RegExp mutant changes [A-Za-z_] to [A-Z_], exits 0 with empty sanitizer
stderr and is caught by Go at byte 1134. The independent named-listener verifier
accepts all 15 declarations and catches Unknown replacing SourceFile. These
staging metadata declarations do not assert newly registered driver adapters.

Auxiliary results: go vet ./... has empty output; checker/registry packages pass;
filtered uncached Node oracle PASS 11.723s, including all five new predicate and
proven-relation fixtures, functions, closures, generic functions, method closures
and the one-byte comparison mutant; lower proof checks PASS 1.020s. No selected
check was skipped, relaxed or deleted. The full repository gate and the 17
required external stage-1 correctness checks were not executed; this report does
not claim those checks are green or that the branch passed the full gate.

Quiet whole-process medians, three alternating rounds with full output hashes:

| Existing fifth-batch ordinary/default scope | Native | Go |
| --- | ---: | ---: |
| TypeScript src/compiler | 2.513276s | 0.359296s |
| Repository | 0.339412s | 0.149793s |

These observations include loading and traversal, not just rule execution, and
are limited to the already supported default scope. Native remains slower.
The original three-rule compiler observation is native 1.768054s versus Go
0.312235s. measurements.json retains phases, queries, per-round timing and
identical diagnostic SHA-256; the raw streams are archived.

Setup passed: Go/clang/Node/submodules ready 0s, cache warm 111s, done 111s;
nproc 5, cgroup quota 4 CPUs and memory 17.6 GB. Commands sourced
/workspace/adamic-tools/env.sh. All test output went directly to files.
Only identified generated ELF/ar files in owned scratch runs were removed for
space, including completed current suites; cleanup manifests list every path.
Sources, mutants, logs, stdout/stderr, controls and corpus manifests remain.

The refreshed ranking audit has 636 origin refs and 33 distinct Markdown claim
blobs, with zero unclaimed entries after excluding main/library ports and every
origin reservation. No new claim is made. The duplicate sixth batch remains
withdrawn in favor of wave 11; its prototypes are outside the branch. React
HIR/SSA/capture-analysis scopes, complete BooleanPropNaming component/props and
configured-regex behavior, and legacy driver conversion remain parked or
incomplete as described in previous reports. Nothing about the new compiler
proofs establishes those missing analyses.
