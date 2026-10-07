Built: rebased the existing fifteen-rule wave-03 branch onto current fetched origin/main b8fb957aa; no new claims.
Commits: pre-rebase pushed af8e4496d; rebased implementation 54064cb75; this report accompanies the own-branch landing update.
Checks: four full owned suites passed for twelve rules in an overall 567.780s run; the three-rule React suite explicitly skipped for its blocker.
Mutants: all twelve rule message mutants and the full suites' guard/retained-registry mutants caught; interpolation whitespace mutant caught only by comparison.
Blocked: native new RegExp(pattern, 'u') remains unsupported, preventing the combined React driver and its full oracle.

Origin/main advanced from c01907a70 to b8fb957aa839a9e8cb0b54279dd9864fa317bd30. Rebase completed cleanly and retained the inherited static field implementation. No shared harness, generator or compiler file was edited. No new rules were claimed because the existing branch cannot be fully oracle-green under the mandatory regex contract yet. The three React claims remain reserved; the blocker is regex lowering, not high-level IR, SSA or capture analysis.

After sourcing /workspace/adamic-tools/env.sh, this exact full command ran with output redirected to /workspace/wave-03/resume-oracles.log:

```sh
ADAMIC_WAVE03_FINAL_ARTIFACTS=/workspace/wave-03/resume-final ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus go test ./stage1/cohere/typeaware -run '^TestWave03(AgreementAndMutants|ContinuationAgreementAndMutants|MoreAgreementAndMutants|ReactAgreementAndMutants|FinalAgreementAndMutants)$' -timeout 40m -count=1 -v
```

The new final suite passed in 137.61s, more/core in 132.38s, Nexus continuation in 153.32s, and original in 141.63s. React explicitly printed BLOCKED and SKIP in 2.82s. The package's PASS does not certify the skipped React rules. The twelve active rules repeat exact controls and corpus comparisons, native sanitizers, message mutants and released-handle checks on this main baseline. Final controls: 136 parseable, 100 findings and 71 suggestions, 64880 identical bytes. More/core: 566 parseable controls and 318 findings. Nexus: 94 controls, 96 findings plus ordered 12 and imported one. Original controls: 46 findings plus loose-this four; repository 86, compiler 178. Compiler and repository manifests retain the frozen 77 and 287 roots.

For the newest three rules, single whole-process timings from this run: repository native 0.998926190s versus Go 0.221703053s; compiler native 6.641717299s versus Go 0.599730607s. Native remains slower; these are individual observations, not a statistical benchmark. Complete timing lines for all active suites are preserved in the log.

`go test ./bridge/tsgo/checker -count=1` passed in 1.196s. `go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware` passed. The literal interpolation helper was rebuilt using this rebased compiler in native, sanitizer and emitted-JavaScript profiles. Fifteen controls again match independent Go and Node, with zero exits and empty stderr. The compiled whitespace-star-to-plus mutant still runs normally and is caught only by output comparison. Exact inputs, outputs and logs are in resume_validation/. This helper evidence does not substitute for the blocked full React oracle.

The exact refusal remains: `stage 0 can't lower RegExp with a nonconstant pattern yet` at wave_03_react/gaps/general_regex.a:3:24. The rule itself has the same required constructor in boolean_pattern.a:4:23, documented with a full-driver refusal in REGEX_MIGRATION.md. There is no handwritten matcher fallback. Current main retains this explicit NotYet in internal/lower/regexp.go. No new compiler support was imported from other branches to bypass file ownership.

Not covered: full React findings/fixes/suggestions under runtime options, whole repository gate, shared emitted-JavaScript external-checker execution, fresh timing of the blocked driver. The previous fifteen-rule full green preceded the mandatory regex migration and is historical evidence only. Source files are unchanged by this landing unit apart from reports and evidence. Only codex/typeaware-wave-03 is pushed; integration owns main and area branches.
