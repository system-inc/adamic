Rebased onto a local base containing current main c01907a7 and the named harness ab70f38d4; JSX parsing now works.
Validated source SHA: 25216f50d5e4d3c7471ea6b8969e05591c2a354e; evidence commit follows on wave 20 only.
Checks: all nine rule oracles, sanitizers, checker, Node, registry, batch-8 position/release and fifteen declaration/manifests PASS.
Mutants: nine semantic, six raw-fact, five registry, fifteen compiling declaration and fifteen JSON mutations caught.
Not covered: three parked HIR/SSA analyses, three unfinished JSX analyses, numeric-kind migration, full repository gate and own-rule emitted JavaScript.

The explicit named dependency ab70f38d47de1d4974082b38f84a56af2368b7af
is not yet on origin/main. A detached local merge combined it with current
main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06, and the 21 own commits were
rebased onto that combined base without conflicts. Both commits are ancestors
of the validated branch. Only codex/typeaware-wave-20 is pushed, with an exact
lease against its prior cea877edc878166dd86fa0f3baffd07a205b9ed7 tip. Main
remained c01907a7 in the final remote check. No shared change was reverted.

The valid-JSX probe that previously panicked now exits 0 in both freshly built
Go and native runners, with identical output and empty native stderr. This
closes that observed parser failure. It tests parsing using the prior core
rules, not findings of the three unfinished JSX rules. Old BLOCKED.md parser
failures are historical; the appended dependency update supersedes them.

The shared harness now supports node:true and hands visit(node,index[,parent])
the actual node. The seven-argument report form and wire protocol remain
unchanged; reportNode/reportRange are added. Its registry descriptor Kinds is
still []string, and ParseNode.kind remains string-only. Thus handed-node support
is present, but the numeric SyntaxKind requirement is not implemented in that
shared interface. Modifying the shared parser/driver is forbidden for this unit.
The native type-aware metadata rule.json files are numeric declarations, not
complete shared registry descriptors: they have no factory/class/oracle fields
and must not be represented as registered source analyses. There are no new
placeholder handlers. The three JSX rules remain unfinished; no new claims.

First gate PASS 240.427s: 48 controls and 12,936 identical bytes normal and
ASan/UBSan/LSan. Nexus PASS 121 projects/160 findings; core PASS 682 projects,
502 findings/365 suggestions. Every message, finding range, automatic fix and
suggestion edit byte agrees with unchanged production Go. Frozen corpus checks
again cover 287 repository and 77 compiler roots, normal/sanitized: 18,485 and
5,241 matching bytes with zero findings. Positive control suites provide the
finding coverage. The documented invalid upstream TSX exclusion is unchanged.
Each semantic and raw-fact mutant compiles and exits normally; independent
finding bytes catch it. Released checker questions panic 70; five registry
retention mutations exit 0 and fail that expectation. Fifteen declaration
mutants compile and exit 0 with empty stderr, then fail production-registration
bytes. Fifteen JSON first-kind increments fail separately. Metadata checks do
not prove correctness of the absent React source analyses.

Checker PASS 0.924s; representative Node differential PASS 0.594s using existing
cache observations; vet empty. Shared registry PASS 0.064s; batch-8
TestBatch8PositionsAndRelease PASS 47.697s. The full shared lint gate and full
repository gate were not run. No source algorithm changed in the nine own rules.
Toolchain reused: last setup 121s, Go 1.27.1, clang 20.1.8, Node 24.19.0,
nproc previously 5, quota four cores. No new setup timing is asserted.

Commands source /workspace/adamic-tools/env.sh. All earlier c019 validation
commands in LANDING_C01907A7_REPORT.md were repeated with harness-first,
harness-second, harness-third, harness-listeners, harness-manifests,
harness-jsx-listeners, harness-checker, harness-node and harness-vet artifact
and log names. Additionally:

```sh
go test ./stage1/cohere/lint/registry -count=1 > /workspace/wave20-validation/harness-registry.log 2>&1
go test ./stage1/cohere/lint -run '^TestBatch8PositionsAndRelease$' -count=1 -v -timeout 10m > /workspace/wave20-validation/harness-batch8.log 2>&1
```

The JSX probe invokes harness-third/oracle and harness-third/suite with the
previously recorded fourth/jsx-probe/tsconfig.json and manifest; exit codes,
stdout/stderr are archived. Every test output went to a file. The compressed
archives exclude binaries. Source and commands are reproducible from the prior
landing reports and this dependency commit.

Whole-process observed native/Go seconds: first repository .800979/.318741,
compiler 5.387335/1.079188; Nexus repository .371269/.217872, compiler
2.482431/.523009; core repository .365117/.216893, compiler 2.328990/.516224.
Gates overlapped, so these are observations rather than isolated benchmarks.
Native remains slower; no numeric dispatch improvement is claimed.

Regex policy remains in force. None of these twelve source-rule families has a
Go regexp compile site to translate; no matcher or raw option fallback was added.
The three React compiler-analysis claims stay PARKED on #dnv6f2c. Their JSX
parsing prerequisite is now available here, but HIR/SSA/capture construction
and the full lint comparisons are still absent.
