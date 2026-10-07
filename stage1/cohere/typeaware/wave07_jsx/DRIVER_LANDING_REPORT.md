Built: twelve existing ports rebased onto b46914832's generated shared driver; no new claims.
Commits: previous pushed tip 2f20ff56a; fresh tested source and bases are pinned in evidence/metadata.json.
Checks: 19 owned landing steps and 100 external passing events across 16 roots pass, with zero skips or failures.
Mutants: all twelve rule witnesses, dispatch/listener/fact/library/ownership guards, nine record mutants and external comparison witnesses caught.
Not covered: three parked React analyses, four production bridge routes, shared checker context, own emitted-JavaScript parity and the full repository gate.

## Integration and source scope

Rebased cleanly onto origin/area/stage1-lint
b46914832d70e00847d82d5d221ab7bb24040c53, containing current origin/main
c7991b900362796aefd111474e65eb5398e91953. Incoming shared lint migration
moves legacy rules into directories and dispatches through the generated registry.
The context and registration migration, tests and options corpus were retained.
No shared or protected files or rule bodies were edited. Only owned evidence,
collector and reports changed. Only codex/typeaware-wave-07 is pushed.

## Fresh checks and every detector

All 19 steps in evidence/landing/commands.json.gz exit zero. Exact commands,
environment, durations and full outputs are archived alongside it. All twelve
ports compare Go findings, fixes and suggestions on controls, repository and
TypeScript compiler corpora. Sanitizers and released handles pass. The JSX gate
also checks every option combination and named listener declarations.

AREA_LANDING_REPORT.md's Every mutant and its detector table names all twelve
rule mutations. Each compiles and exits zero with empty stderr; independent Go
finding/fix/suggestion bytes catch it. Three JSX dispatch mutations and nine
legacy listener mutations also compile and exit zero; Go output catches each.
Raw question fact mutations change symbol identity, successors, module target
and regex character span; Go findings or suggestions catch each. Registry
retention mutations are caught by the required released-program panic 70.
The library-data mutation exits zero and fails exact reference byte comparison.
All fresh streams are archived in the rule, question, JSX and listener folders.

Native record tests catch indices-in-insertion-order, uint32-max-as-index,
deleted-key-iterated, overwrite-key-leaked, stored-key-freed, own-slot-null-read,
prototype-membership-restored, missing-read-silent and own-read-checked-as-missing.
Their detector names and results are in runtime-integration.log.gz. Decoder
refusals and type flags pass. Node parity passes 36 fixtures, the one-byte
mutant, reverse-search runtime comparison and iterator refusals. Four targeted
lower type-proof tests and runtime release/string checks pass.

Shared finding-model tests verify .a modules, emitted-JavaScript mismatch,
complete suggestions, suggestions alongside fixes, descriptor rejection and
stable regeneration. These establish shared-harness behavior, not emitted
JavaScript parity for the owned standalone rule drivers.

The required external runner supplied the same pinned TypeScript v6.0.3 and
seven library versions described in REQUIRED_INPUTS_REPORT.md. Its fresh receipt
records 100 passing events, 16 roots, 11 packages, zero skips and zero failures.
This includes compiler comparison through the newly migrated shared lint driver.
All full external streams and the exact invocation are archived. No selected
check was skipped, relaxed or removed. This is a filtered gate, not the full gate.

Setup succeeded in 40 seconds on nproc 5: Go 1s, clang 1s, Node 1s,
submodules 1s, build-cache warm 40s, total 40s; CPU quota four cores, memory
17.6 GB. Full setup and rebase logs are driver-setup.log.gz and driver-rebase.log.gz.

## Fresh quiet timing

Three samples per implementation, median seconds including process/checker startup.
All timed outputs match Go completely. No other checks ran during sampling.

| Corpus | Rule | Native | Go |
| --- | --- | ---: | ---: |
| repository | fragments | 0.321 | 0.137 |
| repository | undef | 0.309 | 0.146 |
| repository | adjacent | 0.323 | 0.143 |
| compiler | fragments | 1.957 | 0.295 |
| compiler | undef | 2.012 | 0.309 |
| compiler | adjacent | 1.980 | 0.327 |

Native remains slower; no speedup is claimed. Older reports/timing logs are
historical. The fresh driver logs and timings.json.gz are the current measurements.

## Remaining integration gaps

Shared RuleContext still lacks a checker/program handle. Four private bridge
routes still require production registration; the normal archive refuses them
with panic 70, separately verified. Three hook claims remain parked for native
high-level IR, single-assignment and capture analysis (#dnv6f2c), with dependency
probes and Go findings preserved. Own emitted-JavaScript parity and the full
repository gate were not run. No new claim is made before the green push.
