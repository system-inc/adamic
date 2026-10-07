Built: rebased wave-17 onto area d65a8f93, named listener declarations and native Next JSX comparisons; no new claims.
Commits: validated source f4bba2554e6c86effb45f3ac956ebf84f2c687ea, based on main 39638d9e and area d65a8f93.
Checks: eight lint suites PASS 911.697s; sanitizer, released-handle, checker/registry, runtime, filtered uncached Node and vet checks pass.
Mutants: every existing rule/raw-question/handle mutant caught; extra literal-regex byte 1138 and named-kind Unknown mutant caught.
Limits: React analysis/configuration and legacy driver integration remain incomplete; no full repository gate; no unclaimed rules remain.

The owned branch was rebased onto origin/area/stage1-lint, first at
7481e0324e34a2537aafa9db7eeacda50405611b and then at
 d65a8f931c98655936ae04c6899f38f14862b73e after its runtime changed.
Current main 39638d9e278d38bb5aeae887f46d55a70e47aaad, model 50a5f105
and harness 41eb6eab2 are ancestors of the validated source. The remote check
before push still reports those main/area tips and own prior tip
0a83e18d1c8092c744b6998f18781cf86f2e11f9. Only the owned branch is pushed,
with an exact force-with-lease against that prior tip. No shared generator,
harness, parser, protected compiler or runtime source was edited by this worker.
The shared leak-check/runtime changes were accepted through the rebase.

All 15 own listener manifests now declare registry-valid ast.Kind names without
Kind prefixes, rather than numeric values. The independent production-listener
verifier passes and catches Unknown replacing SourceFile. These remain staging
metadata, not newly registered shared-driver adapters; historical node-handler
integration limits are not relabeled as complete.

The area parser now supports JSX. The first run retained two obsolete assertions
expecting parser refusal, so those tests failed because parsing succeeded.
That initial failure log is preserved. Only the two owned wave-17 tests changed:
the head suite now uses the ordinary native parser for all 40 controls, including
its two comparison-only mutants, and the original suite requires exact Go bytes
for three native JSX witnesses. A focused rerun passed in 137.719s on the first
area base. The entire selected gate then passed again on the final runtime base.
The historical Next JSX parser blocker is therefore closed.

| Suite | Time | Result |
| --- | ---: | --- |
| TestCoverageAgreementAndMutants | 159.05s | PASS |
| TestWave17FifthAgreementAndMutants | 176.83s | PASS |
| TestWave17FourthAgreementAndMutants | 146.14s | PASS |
| TestWave17HeadJudgmentsAndNativeJSX | 67.30s | PASS |
| TestWave17NextAgreementAndMutants | 130.50s | PASS |
| TestWave17AgreementMutantAndNativeJSX | 68.66s | PASS |
| TestWave17ThirdAgreementAndMutants | 160.81s | PASS |
| TestWave17UnicodeUpper | 2.41s | PASS |

Complete canonical finding streams include positions, IDs, messages, fixes and
suggestions. The supported controls/options, frozen repository and compiler
corpora, normal and sanitized runs, comparison-only rule/question mutants and
released-handle guards passed again. Every kill and first differing byte is in
validation-wave-17-area/mutant-evidence.txt; compiler failures are not counted
as mutant kills. The extra fixed RegExp mutant changes its identifier-start
class from [A-Za-z_] to [A-Z_], exits 0 with empty sanitizer stderr and is caught
by complete Go bytes at byte 1138. Unicode covers all 1,114,112 code points.

Final auxiliary checks: go vet ./... has empty output; checker and registry
packages pass; runtime release and string-equality tests pass in 7.799s; filtered
uncached external Node checks pass in 2.396s, including the newly changed
lastIndexOf runtime behavior and its one-byte oracle mutant. The .a rename,
complete suggestions, combined automatic-fix/suggestions and emitted-JavaScript
mutant shared-harness tests passed in 134.268s on the first area base 7481e032;
that initial-base result is preserved separately and is not represented as a
rerun on d65a8f93. The final base received its own eight lint suites and runtime,
Node, checker/registry, named-listener and vet checks.

Quiet three-round alternating whole-process medians for the existing fifth-batch
ordinary-parser/default-option suite, with complete diagnostic SHA-256 agreement:

| Corpus | Native | Go |
| --- | ---: | ---: |
| TypeScript src/compiler | 2.552346s | 0.369444s |
| Repository | 0.350299s | 0.145629s |

These measurements include checker loading and traversal. They do not establish
whole-inventory performance or projected React/complete configuration coverage.
Native remains slower. The original three-rule compiler observation was native
1.682922s versus Go 0.283608s. All timed streams and phase counters are retained.

The refreshed audit checks 597 origin refs and 33 distinct Markdown claim blobs:
zero unclaimed rules among the 197 checker-dependent entries after excluding
main/library ports and all origin reservations. No new reservation is made.
The duplicate sixth batch remains withdrawn in favor of wave 11; its prototypes
stay outside the branch. React HIR/single-assignment/capture-analysis scopes,
complete BooleanPropNaming component/props/configured regex behavior, and legacy
shared-driver adapter conversion remain parked or incomplete as previously
reported. JSX parser success does not prove those analyses. No full repository
gate was run.

Setup: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
cache warm 49s, done 49s. nproc 5; cgroup quota 4 CPUs, memory 17.6 GB.
All builds sourced /workspace/adamic-tools/env.sh. Test output was redirected
to log files; command-output.tar.gz preserves individual stdout/stderr and
manifests. Commands and results are recorded beside this report.
