# Type-aware wave 24

Branch: `codex/typeaware-wave-24`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed checker-dependent rules, positions 70, 71 and 72 after excluding
all 26 rules already ported on the base:

1. `@typescript-eslint/prefer-return-this-type`
2. `@typescript-eslint/return-await`
3. `@typescript-eslint/use-unknown-in-catch-callback-variable`

Selection uses the combined compiler and repository counts linked from
VOLUME_REPORT.md in `validation-volume/*-all.counts`, descending combined
volume with lexical ties. Each selected rule has zero recorded findings in
both populations. The inventory ranking includes non-checker rules and is
not the checker-only selection used here.

Checked 268 origin refs after fetching all branch refs on October 7, 2026.
No competing claim or named type-aware port for these rules was found.
This claim precedes implementation; it does not assert completed ports.

## Next batch after the original ports were pushed

Original three completed and pushed in implementation commit
`406e7546846aeb58fb918ad4668f09f131a78e1e`.

Next claimed rules, first eligible checker-dependent entries in descending
combined VOLUME_REPORT ranking after excluding ports on main and the bridge
branch and every claim on every fetched origin branch:

1. `nexus/correctness-no-process-exit-after-output`
2. `nexus/correctness-no-uncleared-race-timeout`
3. `nexus/correctness-require-blocking-standard-streams`

Fetched 325 origin refs on October 7, 2026; examined 33 distinct Markdown
claim blobs, the 26 existing ports and main's stage1/cohere implementation.
Main tip: `ef3d907e`; bridge tip: `5afbdb83`. These are full-ranking positions
122, 123 and 124, with zero findings in both frozen populations. None was
claimed or ported on the checked refs. This update precedes implementation.

No shared registration generator or existing shared test harness will be edited.

## Third batch after all six prior ports were pushed

Previous batch completed, tested and pushed in `cce6c361a94ed730f5a743ce84deab9f91142b58`.

Next three eligible checker-dependent rules:

1. `prefer-promise-reject-errors`
2. `prefer-regex-literals`
3. `prefer-rest-params`

Fetched 356 origin refs on October 7, 2026, inspected 33 distinct Markdown
claim blobs, and excluded ports on main and the bridge branch and all claims
on every origin branch. Full-ranking positions 158, 159 and 160, each zero
findings in both frozen populations. Main: `ef3d907ecdc4c771b016f7d9c52372def057a340`.
Bridge: `5afbdb83da2ed7ad9815657cd3f6ececd5294bf6`. This update precedes implementation.

No shared registration generator or existing shared test harness will be edited.

## Fourth batch after all nine prior ports were pushed

Previous batch completed, tested and pushed in `09bf947bb040786384d3ef055feb376ea4c0832c`.

Next three eligible checker-dependent rules:

1. `react-hooks/set-state-in-effect`
2. `react-hooks/set-state-in-render`
3. `react-hooks/static-components`

Fetched 389 origin refs on October 7, 2026; inspected 33 distinct Markdown
claim blobs and excluded ports on main/bridge and all claims on all origin
branches. Full-ranking positions 168, 169 and 170, each zero findings.
Main: `e011f8f60899586d6373a5ccb07335ad82cfbf3c`.
Bridge: `5afbdb83da2ed7ad9815657cd3f6ececd5294bf6`.
This update precedes implementation. No shared harness or generator will be edited.

## Fourth batch parked

Status: parked, counted finished for the landing-first cap by Ahra's instruction.

- `react-hooks/set-state-in-effect`: blocked on native high-level IR and capture analysis.
- `react-hooks/set-state-in-render`: blocked on native high-level IR, single-assignment construction and control dominance.
- `react-hooks/static-components`: blocked on native high-level IR; JSX parser support is now integrated on the lint-area base.

Cohere analysis modules are being ported to Adamic on #dnv6f2c; JSX support is landing on `area/stage1-lint`. These are not completed native rule ports. The Go oracle, native JSX refusal probe, listener metadata and evidence already pushed remain in `wave-24-fourth/`. No shared source changes are made to bypass these dependencies. Nine earlier native ports remain oracle-green on current main `f8013f0baac41ddc340d76f83bddde38536a8f07`.

## Fifth batch after React parking

Next claimed checker-dependent rules without the parked native analysis dependencies:

1. `require-await`
2. `symbol-description`
3. `valid-typeof`

Fetched 529 origin refs and inspected 33 distinct claim blobs: 154 rules claimed, 18 remaining before dependency exclusions. Ranking uses the same frozen combined VOLUME_REPORT counts. Remaining React/JSX rules (including `structure/react-hook-no-any-type`) are skipped under the user's React parking instruction; `require-atomic-updates` is skipped because its Go implementation requires capture-sensitive control-flow graph construction and solving. The selected three are the first eligible rules after these exclusions. None is claimed on any origin branch or ported on main/bridge. Main is `f8013f0baac41ddc340d76f83bddde38536a8f07`. The branch is rebased, oracle-green and pushed. This claim precedes source implementation; no shared harness or generator will be edited. New rule implementations will declare numeric `rule.json` kinds and receive only their relevant nodes.

## Fifth batch completed and landing ready

`require-await`, `symbol-description` and `valid-typeof` have native `.a` default-option ports, AST-name `rule.json` and compiled listener declarations, and handed-node dispatch. Implementation commit `875865e99208e7106c60254cabacc96a36cbfa75` is rebased onto `c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06`. Complete bytes agree for 217 controls (125 findings) and the 287-root repository/77-root compiler populations in normal and sanitized modes. Six compiled byte mutants, the numeric metadata mutant and all three new released-question refusals pass. All nine earlier ports are re-green after this rebase. The report and reproduction evidence are in `wave-24-fifth/README.md` and `wave-24-fifth/evidence/`. Nondefault valid-typeof options and the full repository gate are not covered. Three React claims remain parked with their blockers above. No additional claim is made.

## AST-name listener correction

Ahra's corrected registry contract is applied to all twelve completed ports: rule.json kinds and native listener exports are AST names, not numeric values. The independent Go names are normalized exactly as the shared registry validates them. All four rule suites and both listener checks were rebuilt and re-green on main b8fb957a, with full-byte corpora, sanitizer modes, all mutants and released-handle checks. No new rule is claimed; the three React analysis claims remain parked. Latest evidence is wave-24-fifth/KIND_NAMES_REPORT.md.

## Integrated lint-area landing refresh

Rebased onto origin/area/stage1-lint 7481e0324e34a2537aafa9db7eeacda50405611b as explicitly requested. It contains main 39638d9e and harness 41eb6eab2. All twelve completed ports, named listener declarations, full-byte normal/sanitized corpora, 23 byte mutants, registry/native/JSON mutations and released-handle checks pass again with a fresh area compiler. Shared sources and developer-tools changes are preserved. JSX support is integrated; the three React claims remain parked on native analysis dependencies. No new claim is made. See wave-24-fifth/AREA_LANDING_REPORT.md.

Final requested area base advanced to d65a8f931c98655936ae04c6899f38f14862b73e with allocator/search optimizations. Rebased again and re-green on a fresh compiler: all twelve ports, named declarations, full-byte corpora, sanitizers, all mutants and released handles pass; Node runtime search oracle passes. Main 39638d9e is contained. Post-push audit finds no unclaimed rule among 610 origin refs/33 claim blobs. No new claim is made. Final evidence is in wave-24-fifth/evidence/area-landing/final/.


## Checker-area merge, October 8

Merged (not rebased) area c4bdc23fa86d55cf7e579989201c11258f4d3a62 and preserved both sets of checks. Wave fact calls now go through the area's RuleContext.checker. symbol-description and valid-typeof are migrated to unified typed descriptor directories. Remaining standalone suites are blocked by the fresh compiler refusal at stage1/cohere/lint/checker.a:66:20 (lowering.useOfThis); this is not a parity result. React claims remain parked on the named high-level lowering/capture services; a standalone SSA module is present but does not supply them. No new rules are claimed. Current verification, reproducer and logs: wave-24-fifth/UNPARK_REPORT.md and wave-24-fifth/evidence/unpark/.
