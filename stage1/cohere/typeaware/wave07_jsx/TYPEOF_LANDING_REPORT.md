Built: twelve existing ports rebased onto current main and lint integration with incoming typeof fixes; no new claims.
Commits: prior pushed tip deb8aa1bc; fresh tested source and bases are pinned in evidence/metadata.json.
Checks: 19 landing steps, 43 Node fixtures and 100 external passing events across 16 roots pass; zero skips or failures.
Mutants: twelve rule witnesses and all guard mutants, nine record mutants, eight typeof executions and external comparison mutants caught.
Not covered: three parked hook analyses, four production bridge routes, shared checker context, own emitted-JavaScript parity and the full repository gate.

## Rebase and owned scope

Rebased cleanly onto origin/area/stage1-lint
d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, containing current origin/main
b6b1538b0cebc4ba6741ac34f1aedb60293c1d06. Incoming typeof lowering,
IR, native classification and slot-presence fixes and their oracle tests were
retained. No shared or protected source was changed. Rule bodies remain unchanged.
Only owned validation filters, evidence collector, reports and evidence changed.
Only codex/typeaware-wave-07 is pushed.

## Every fresh check and detector

All 19 owned steps in evidence/landing/commands.json.gz exit zero, with exact
arguments, environments, durations and adjacent full output streams. All twelve
rules match Go on findings, fixes and suggestions over controls, repository and
TypeScript compiler corpora. Sanitizers and released handles pass. The JSX gate
includes every option combination. No options adapter guard was relaxed.

The original and continuation rule mutations and their comparison-only detectors
are listed individually in AREA_LANDING_REPORT.md's Every mutant and its detector
section and the fresh rule streams. All twelve compile, exit zero and have empty
stderr; only independent Go finding/fix/suggestion bytes catch them. Three JSX
named-dispatch and nine listener mutations also exit zero and disagree with Go.
Raw bridge facts mutate symbol identity, graph successors, module targets and regex
character spans; complete Go finding/suggestion comparisons catch each. Registry
retention mutants fail the required released-handle panic 70. The library-data
mutation exits zero and fails exact byte comparison. All fresh streams are archived.

The owned Node filter now includes seven incoming typeof fixtures, alongside
36 previous fixtures. All 43 pass source Node, emitted JavaScript and sanitized
native parity. TestTypeOfNullMutant catches five clean-exit executions:
typeof_null.a, typeof_null_compare.a, typeof_null_switch.a, typeof_null_slots.a,
and typeof_null_roll.a. TestTypeOfNullSlotPresenceMutant catches presence forced
true; TestTypeOfConstructorMutant catches constructors classified as objects;
TestTypeOfStringLiteralMutant catches strings classified as undefined. Every
mutant builds under sanitizers, exits zero without stderr/leaks and is caught
only by Node stdout disagreement. Exact outputs are in node-oracle.log.gz.
The existing one-byte oracle mutant, reverse-search comparison and iterator
refusals also pass.

Runtime tests catch all nine record mutants named in DRIVER_LANDING_REPORT.md;
full named detectors are in runtime-integration.log.gz. Four lower proof checks,
decoder guards, bridge suite, vet and gofmt pass. Shared finding-model tests cover
.a modules, emitted-JavaScript mismatch, complete suggestions, suggestions beside
fixes and registry guards. They certify the shared harness, not emitted-JavaScript
parity for these standalone owned rule drivers.

The pinned external runner uses TypeScript v6.0.3 and the seven versions in
REQUIRED_INPUTS_REPORT.md. Its fresh receipt records 100 passing events, 16 roots
across 11 packages, zero skips and zero failures. It includes compiler-corpus
comparison through the shared generated driver. Full external logs and exact
invocation are archived. No selected check was skipped, relaxed or deleted.
This remains a filtered worker gate, not the complete repository gate.

Setup succeeded in 110 seconds: Go 0s, clang 0s, Node 0s, submodules 0s,
build-cache warm 110s, total 110s; nproc 5, CPU quota four cores, 17.6 GB memory.
Before the JSX proofs, go build -o /workspace/wave-07-next-rest/adamic ./cmd/adamic
rebuilt their scratch compiler from this source; its log is archived. Setup and
rebase logs are typeof-setup.log.gz and typeof-rebase.log.gz.

## Fresh native versus Go time

Three samples per implementation, median seconds including process/checker startup.
No other owned checks ran during measurement; all timed output bytes match Go.

| Corpus | Rule | Native | Go |
| --- | --- | ---: | ---: |
| repository | fragments | 0.341 | 0.149 |
| repository | undef | 0.327 | 0.140 |
| repository | adjacent | 0.328 | 0.147 |
| compiler | fragments | 2.064 | 0.297 |
| compiler | undef | 2.024 | 0.318 |
| compiler | adjacent | 2.026 | 0.306 |

Native remains slower. The fresh typeof logs and timings.json.gz are this unit's
measurements; earlier reports and timing logs are historical.

## Concrete remaining blockers

Shared RuleContext still lacks a checker/program handle, and four private bridge
routes need production registration. The normal archive explicitly refuses them
with panic 70, checked separately. Three React hook claims remain parked for
native high-level IR, single-assignment and capture analysis (#dnv6f2c). Dependency
probes preserve their current Go findings and parser observations. Own emitted
JavaScript parity and the full repository gate were not run. No new claims are
made before the green push; the refreshed all-origin selection audit follows it.
