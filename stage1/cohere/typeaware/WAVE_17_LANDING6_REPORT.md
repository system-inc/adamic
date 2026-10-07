Built: rebased existing wave-17 onto area b4691483 after shared legacy-rule registry migration; no new claims.
Commits: validated rebased source 9513a11b4a11f1f8c7dc24ec27014ec8783b9241, replacing owned prior bf5df4afd.
Checks: eight lint suites PASS 887.957s; sanitizers, handles, checker/registry, named listeners, vet and filtered uncached Node pass.
Mutants: every existing rule/question/handle mutant caught; extra RegExp byte 1134, named Unknown and Unicode byte 1 caught.
Limits: parked React analysis/configuration and driver adapters remain incomplete; no full gate or 17 required external checks run.

Current main is c7991b900362796aefd111474e65eb5398e91953. The updated lint
area b46914832d70e00847d82d5d221ab7bb24040c53 contains it, plus the legacy
rule migration onto the shared registry. All 35 own commits rebased cleanly onto
that area. Shared model 50a5f105 and harness 41eb6eab2 remain ancestors. The
before-push remote check reports the same integration tips and owned old tip
bf5df4afd2c9c8220940fc46730e0ae7e59697f2. Only the owned branch is pushed,
with an exact lease against that tip. No main or area branch is pushed.

The shared migration was accepted through rebase. This worker did not edit
shared harness, registry, parser, runtime or protected compiler files. Wave-17's
shared finding-model import remains valid. No rule implementation was changed;
this landing adds only evidence and a current status entry to the owned claim.

| Selected suite | Time | Result |
| --- | ---: | --- |
| TestCoverageAgreementAndMutants | 154.99s | PASS |
| TestWave17FifthAgreementAndMutants | 161.66s | PASS |
| TestWave17FourthAgreementAndMutants | 144.30s | PASS |
| TestWave17HeadJudgmentsAndNativeJSX | 68.77s | PASS |
| TestWave17NextAgreementAndMutants | 123.68s | PASS |
| TestWave17AgreementMutantAndNativeJSX | 68.98s | PASS |
| TestWave17ThirdAgreementAndMutants | 163.14s | PASS |
| TestWave17UnicodeUpper | 2.42s | PASS |

The complete Go comparison includes ranges, IDs, messages, automatic fixes and
all suggestions. Supported upstream/edge controls, option variants, compiler
and repository corpora, ordinary/sanitized runs, rule and checker-question
mutants, and released-handle guards passed on this exact rebased source. All
40 Next JSX head controls use the real native parser, as do the original three
JSX witnesses. Existing projected React analysis controls are not claimed as
complete production integration. Unicode agrees with Go on all 1,114,112 code
points. Every mutant and its catcher is preserved in mutant-evidence.txt.

Decision mutants exit successfully with empty sanitizer stderr and are caught
by complete Go bytes. Retained registry mutants finish cleanly but fail the
required panic-70 comparison. The extra fixed RegExp mutant changes [A-Za-z_]
to [A-Z_], exits 0 with empty sanitizer stderr and is caught at byte 1134.
The independent production-listener verifier accepts all 15 named declarations
and catches Unknown replacing SourceFile. These declarations remain staging
metadata, not newly registered executable adapters.

Auxiliary checks: full go vet has empty output; checker and registry packages
pass, including deterministic regeneration and descriptor mutants. Filtered
uncached Node oracle PASS 1.062s with functions, closures, generic functions,
method closures and the one-byte oracle mutant. No selected test was skipped,
relaxed or removed. The full repository gate and the 17 required external
stage-1 correctness checks were not run. This report does not claim those
checks or the full gate are green.

Quiet three-round alternating whole-process medians with complete output hashes:

| Existing fifth-batch ordinary/default scope | Native | Go |
| --- | ---: | ---: |
| TypeScript src/compiler | 2.431382s | 0.386125s |
| Repository | 0.337386s | 0.144250s |

These observations include loading/traversal and only the supported default
scope, not all lint rules or complete projected React configuration. Native
remains slower. The original three-rule compiler observation is native
1.637050s versus Go 0.276426s. measurements.json preserves every round,
phase, query count and equal diagnostic SHA-256; raw streams are archived.
The repository observation uses the current sources after the shared registry
migration, so its query count is not assumed equal to the previous landing.

Setup passed: Go/clang/Node/submodules ready 0s, cache warm 35s, done 35s;
nproc 5, cgroup quota 4 CPUs and memory 17.6 GB. Every build sourced
/workspace/adamic-tools/env.sh. Tests wrote directly to logs. Cleanup manifests
list only identified generated ELF/ar artifacts removed from completed owned
scratch suites for space. Source, controls, manifests and oracle streams remain.

The fresh audit finds 644 origin refs, 33 Markdown claim blobs and zero unclaimed
checker-dependent rules after excluding main/library ports and all origin
reservations. No new claim is made. The duplicate sixth reservation stays
withdrawn in favor of wave 11; prototypes remain outside the branch. React
HIR/SSA/capture-analysis scopes, complete BooleanPropNaming component/props and
configured regex behavior, and legacy driver conversion remain incomplete as
previously reported. Registry migration does not establish those analyses.
