Built: rebased twelve existing ports onto b84a9d931; no new claims.
Commits: prior pushed tip e5c3bea8a; tested source is pinned in evidence/metadata.json.
Checks: 19 landing steps and 16 external root checks pass; 100 external passing events, zero skips or failures.
Mutants: twelve rule mutants, dispatch/listener/fact/library/ownership guards, nine record mutants and external comparison witnesses caught.
Not covered: three parked React analyses, four production bridge routes, shared checker context, own emitted-JavaScript parity and the full repository gate.

Superseded landing refresh: [DRIVER_LANDING_REPORT.md](DRIVER_LANDING_REPORT.md).

## Landing and checks

Rebased onto origin/area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da,
containing origin/main c7991b900362796aefd111474e65eb5398e91953.
Incoming runtime record ownership and type-proof changes and their tests were retained.
Rule bodies and shared source were untouched. Only owned validation, evidence and documentation changed.
Only codex/typeaware-wave-07 is pushed.

All 19 steps in evidence/landing/commands.json.gz exit zero; it records exact arguments,
environment and duration, with adjacent full output streams. This reruns all twelve
ports against Go on controls, repository and TypeScript compiler corpora, findings,
fixes and suggestions; sanitizer and released-handle checks pass.
The shared harness emitted-JavaScript and suggestion tests pass, but do not establish
emitted-JavaScript parity for these standalone rule drivers.

The Node oracle passes 36 fixtures, including all five incoming proven guards,
class guards, assertions, satisfies and upcasts fixtures, plus reverse string search,
the one-byte detector and iterator refusals. Native record, release and string tests
and four lower type-proof tests pass. The latter reject unproven predicates and
unsafe relations and prove accepted relations erase. Nine record mutants are caught:
indices-in-insertion-order, uint32-max-as-index, deleted-key-iterated,
overwrite-key-leaked, stored-key-freed, own-slot-null-read,
prototype-membership-restored, missing-read-silent, own-read-checked-as-missing.
Their detecting tests and output are in runtime-integration.log.gz.

All twelve rule mutants and their individual comparison witnesses remain as listed
in AREA_LANDING_REPORT.md and the refreshed rule streams. Three named-dispatch and
nine listener mutants, raw question-fact mutants, library-data mutation and live versus
released-handle checks pass. Required external comparison streams name every extra
mutant and its detector. No checks were skipped, relaxed or removed.

The external-input runner uses the seven pinned packages and TypeScript v6.0.3 documented
in REQUIRED_INPUTS_REPORT.md. Its receipt records 100 passing events, 16 roots in
11 packages, zero skips and zero failures. This is the selected external correctness
profile, not the full repository gate.

Setup completed in 121 seconds: Go 0, clang 0, Node 0, submodules 0, cache 121,
total 121; nproc 5 (CPU quota four cores). Setup and rebase logs are archived.

## Quiet timings

Three samples per implementation; medians include process and checker startup.
Each timing run compares complete native and Go bytes. Seconds:

| Corpus | Rule | Native | Go |
| --- | --- | ---: | ---: |
| repository | fragments | 0.359 | 0.140 |
| repository | undef | 0.314 | 0.133 |
| repository | adjacent | 0.313 | 0.135 |
| compiler | fragments | 1.887 | 0.300 |
| compiler | undef | 2.088 | 0.309 |
| compiler | adjacent | 2.024 | 0.300 |

Production integration still needs four private bridge routes and a shared checker
context. Three React hook claims remain parked for native high-level IR,
single-assignment and capture analysis (#dnv6f2c). No new rules are claimed before
the green landing push. The subsequent fetched claim audit is reported separately.
