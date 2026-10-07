Built: rebased existing wave-17 onto area d3a37422 containing current main b6b1538b; no new rules or claims.
Commits: validated source 6e40966e7a2d2010d9c078bd931516c0cf584796, replacing owned prior ec35dbb3.
Checks: eight lint suites PASS 900.057s; sanitizers, handles, checker/registry, named kinds, vet and filtered uncached Node pass.
Mutants: every existing rule/question/handle mutant caught, extra RegExp byte 1134, named Unknown and Unicode byte 1; new typeof mutants caught by Node.
Limits: parked React analysis/configuration and driver adapters remain incomplete; no full gate or 17 required external checks run.

All 36 own commits rebased cleanly onto current area
 d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, containing main
b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and the required model/harness
commits. This integration preserves null and slot presence through typeof.
Its compiler/runtime changes were accepted through rebase. No shared harness,
registry, parser, runtime or protected compiler file was edited by this worker.
The before-push remote check still reports those integration tips and own tip
ec35dbb367a8a55cb9467c05ca4a8b5ba0cda4e5. Only the owned branch is pushed,
using an exact lease against that prior tip, never main or an area branch.

| Selected suite | Time | Result |
| --- | ---: | --- |
| TestCoverageAgreementAndMutants | 161.28s | PASS |
| TestWave17FifthAgreementAndMutants | 166.19s | PASS |
| TestWave17FourthAgreementAndMutants | 147.38s | PASS |
| TestWave17HeadJudgmentsAndNativeJSX | 66.21s | PASS |
| TestWave17NextAgreementAndMutants | 124.22s | PASS |
| TestWave17AgreementMutantAndNativeJSX | 70.82s | PASS |
| TestWave17ThirdAgreementAndMutants | 161.70s | PASS |
| TestWave17UnicodeUpper | 2.24s | PASS |

The complete canonical Go comparisons include ranges, IDs, messages, automatic
fixes and all suggestions. Supported controls and options, repository/compiler
corpora, normal/sanitized runs, rule/raw-question mutants and released-handle
guards pass again. All 40 Next head controls and the original JSX witnesses
use the real native parser. Projected React analysis controls are not claimed as
complete production integration. Unicode agrees on all 1,114,112 code points.

Every mutant and catcher is listed in mutant-evidence.txt. Rule mutants exit 0
with empty sanitizer stderr and are killed by complete Go bytes. Registry mutants
finish cleanly but fail required panic-70 checks. The extra fixed RegExp mutant
changes [A-Za-z_] to [A-Z_], exits 0 with empty sanitizer stderr and is caught at
byte 1134. Named-listener verification passes all 15 declarations and catches
Unknown replacing SourceFile. These staging declarations do not assert newly
registered executable adapters. No shared options guard was relaxed or bypassed.

Full go vet has empty output. Checker/registry packages pass. The filtered
uncached Node oracle passes in 13.229s, including seven typeof fixtures,
functions, closures, generic functions, method closures, the one-byte oracle
mutant and four dedicated typeof mutant tests: null classification across five
fixtures, absent-slot presence, constructor classification and string literals.
All dedicated mutants are caught by the external Node observation. No selected
check was skipped, relaxed or removed. The full repository gate and 17 required
external stage-1 correctness checks were not run; they are not claimed green.

Quiet whole-process medians, three alternating rounds with full output hashes:

| Existing fifth-batch ordinary/default scope | Native | Go |
| --- | ---: | ---: |
| TypeScript src/compiler | 2.366595s | 0.354754s |
| Repository | 0.332475s | 0.146741s |

Timing includes checker loading and traversal and covers only the existing
supported scope, not all inventory rules or full React configuration. Native
remains slower. The original three-rule compiler observation is native
1.861645s versus Go 0.298138s. Per-round phases, queries and equal diagnostic
hashes are preserved in measurements.json; raw streams are archived.

Setup: Go/clang/Node ready 0s, submodules ready 1s, cache warm 114s,
done 114s. nproc 5, cgroup quota 4 CPUs, memory 17.6 GB. Commands sourced
/workspace/adamic-tools/env.sh and test output went directly to files.
Cleanup manifests identify only generated ELF/ar artifacts from old or completed
owned scratch suites. Sources, patches, logs, witnesses and oracle streams remain.

The fresh audit contains 658 origin refs and 33 Markdown claim blobs and finds
zero unclaimed checker-dependent rules after excluding main/library ports and
all origin reservations. No new claim is made. The duplicate sixth batch remains
withdrawn in favor of wave 11, prototypes outside the branch. React HIR/SSA/
capture-analysis scopes, complete BooleanPropNaming component/props/configured
regex behavior and legacy shared-driver conversion remain incomplete as already
reported. Native typeof integration does not establish those missing analyses.
