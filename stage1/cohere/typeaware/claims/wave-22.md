# Type-aware lint wave 22

Branch: `codex/typeaware-wave-22`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 64, 65 and 66 (one based):

1. `@typescript-eslint/no-non-null-asserted-nullish-coalescing`
2. `@typescript-eslint/no-unnecessary-qualifier`
3. `@typescript-eslint/no-unused-private-class-members`

Ranking uses VOLUME_REPORT.md's referenced compiler-all.counts and
repository-all.counts, combined descending volume with lexical ties, excluding
all 26 rules already ported on the base. The registry count table contains 197
checker-dependent rules; method-signature-style is one of the 26 ports but is
absent from that checker-dependent table. All three selected rules have zero
recorded findings on both corpora.

Before this claim, all origin branch refs were fetched and inspected for claims
and type-aware implementations of these names. None was claimed or ported; no
rule was skipped. New Adamic implementation files will use `.a`.

## Continuation claim

After pushing all original wave 22 work, all origin heads were fetched again.
The first three available checker-dependent rules in the same combined-volume
ranking, excluding ports on origin/codex/tsgo-c-library and origin/main and
active claims on any origin branch, are:

1. `@typescript-eslint/no-misused-promises`
2. `@typescript-eslint/no-misused-spread`
3. `nexus/concurrency-no-lost-update`

The first two reservations were explicitly released in origin/codex/typeaware-wave-21's
claims/wave-21.md. The third was explicitly released in origin/codex/typeaware-wave-30's
claims/wave-30.md. Released reservations are available for reassignment; no
other active claims were found. All three have zero recorded corpus findings.
These three rules are now reserved for codex/typeaware-wave-22. This continuation
claim is pushed before writing any implementation for them.

## Third batch claim

After completing and pushing the six existing claims (590f4fa6), all origin
heads were fetched and every origin claim document and base/main port inventory
was inspected again. The first three unclaimed, unported entries of the
combined by-volume ranking are:

1. `no-global-assign`
2. `no-implicit-globals`
3. `no-implied-eval`

All three have zero recorded compiler and repository findings. They are now
reserved for this branch. This claim is pushed before any implementation code.
The fetched ref SHAs and selection inventory are in
validation-wave-22-third/selection.json.

## Fourth batch claim

After completing, testing and pushing all nine earlier claims through a7e33d0c,
all origin heads were fetched and every origin claim and base/main port
inventory was scanned again. The first three available entries in the combined
by-volume ranking are:

1. `prefer-numeric-literals`
2. `prefer-object-has-own`
3. `prefer-object-spread`

All have zero recorded findings on both frozen corpora. These three are now
reserved for this branch. The claim is pushed before implementation; fetched
ref SHAs and the full selection are in validation-wave-22-fourth/selection.json.

## Fourth batch completion

All three fourth-batch rules are ported as .a, compared byte for byte with Go
on positive controls and both frozen corpora, tested with per-rule repair
mutants and sanitizers, and committed with WAVE_22_FOURTH_REPORT.md. All twelve
claims on this branch are complete. No additional rules were claimed.

## Fifth batch claim

All twelve earlier rules were completed, tested and pushed through 7f3b9262
before fetching every origin head again. After excluding base/main ports and
names in every origin Markdown claim, the first available rules are:

1. `react-hooks/set-state-in-effect`
2. `react-hooks/set-state-in-render`
3. `react-hooks/static-components`

These zero-volume rules are reserved for this branch. This claim is pushed
before implementation. The fetched refs and inventory are recorded in
validation-wave-22-fifth/selection.json.

## Fifth batch status

The fifth claim was pushed as ba8ac629 before implementation. All three React
rules remain pending: the native parser rejects a production JSX shape with
exit 70, and the native React HIR/SSA prerequisites are absent on this branch.
No shared parser or harness files were edited. See WAVE_22_FIFTH_REPORT.md and
validation-wave-22-fifth for the reproduction and production Go test results.
The twelve earlier rules remain complete and pushed; no more rules are claimed.

## Fifth batch parked

Per Ahra's explicit parking instruction, the following reservations are parked
and count as finished for the landing-first cap:

- react-hooks/set-state-in-effect
- react-hooks/set-state-in-render
- react-hooks/static-components

Blocker: native React high-level IR, SSA, capture/value propagation and
dominator analyses. Cohere's analysis modules are being ported under #dnv6f2c;
JSX support is landing on area/stage1-lint. The existing reproduction and Go
tests remain in WAVE_22_FIFTH_REPORT.md. These are not completed native ports.

## Sixth batch claim

The branch includes current main f8013f0ba and its twelve completed rule ports
were re-green and pushed as 0aeeea235 before this selection. All origin heads
were fetched: 529 refs and 33 distinct Markdown claim documents were scanned.
The next eligible rules, now reserved for this branch, are:

1. react/jsx-fragments
2. react/jsx-no-undef
3. react/no-adjacent-inline-elements

react/jsx-no-constructed-context-values was skipped because its memo stability
path requires capture/alias escape analysis, the parked prerequisite. The
selection inventory and inspected refs are in validation-wave-22-sixth/selection.json.
This claim is pushed before implementation. New rule directories will declare
rule.json kinds and use .a listener sources receiving the provided node.

## Sixth batch implementation status

react/jsx-fragments, react/jsx-no-undef and react/no-adjacent-inline-elements
are implemented as .a node listeners with rule.json kinds and numeric kinds.
Their controls and frozen corpora match Go byte for byte with per-rule mutants
and sanitizers. They are validated against the published JSX parser dependency
at a8a62d62ca49db7415e14c3887dd305022b17309 in an isolated source tree.
Current main still blocks the direct parser build with its constructor
escape refusal and has not integrated JSX. Shared parser/harness integration
is pending; those files were not edited. See WAVE_22_SIXTH_REPORT.md.
The earlier three React graph claims remain parked. No further rules claimed.

## Seventh batch claim

All fifteen native implementations and their evidence were pushed as 90fa0cbb4
on current main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. The three earlier
React HIR/SSA claims remain parked. All origin heads were fetched again: 570
refs and 33 distinct Markdown claim blobs were inspected. The first three
remaining entries of the combined by-volume ranking, lexical ties, are:

1. react/no-object-type-as-default-prop
2. react/no-unstable-nested-components
3. react/sort-default-props

All have zero recorded findings on both corpora. They use syntax and checker
binding facts, not the parked HIR/SSA/capture pipeline. These names are now
reserved for this branch before implementation. Exact refs and exclusions are
in validation-wave-22-seventh/selection.json. Native sources use .a, numeric
listener kinds, and handed-node visitors. The shared model ab70f38d4 is on
origin/lint-rules/harness and is not yet on main; shared files stay untouched.
Go regexes will use JS RegExp literals, with the shared translation row when
available; no hand-rolled regex matcher is authorized.

## Seventh batch implementation status

react/no-object-type-as-default-prop and react/sort-default-props are ported
as numeric .a listeners. react/no-unstable-nested-components has its default
behavior and allowAsProps ported; custom propNamePattern translation is staged
in options.a with the required new RegExp(source, 'u') constructor. Native
stage 0 refuses that nonconstant pattern, so configured patterns remain blocked.
Its owned gaps/dynamic-pattern.a records the refusal. No regex matcher fallback
is present. Source Node and Go agree on the configured-pattern controls.

The shared JSX/parser integration remains pending. As with the sixth batch,
validation uses the published native JSX parser a8a62d62c in an isolated source
tree. The current main parser refuses constructor escape before fields are set.
Landing checks on main b8fb957aa pass for all eighteen implemented rules:
controls, mutants, both frozen corpora and sanitizer checks. The raw Unicode
question and its released-handle controls also pass. The three older React
graph claims remain parked. No further claim is made.

## Integrated landing status

Rebased onto origin/area/stage1-lint d65a8f931, including main 39638d9e2. All
eighteen existing implementations passed their byte oracles, mutants and
sanitized frozen corpora again. Current integrated parser builds now; the two
owned validation scripts no longer substitute the older JSX parser. Listener
manifest kinds are AST names, with numeric comparisons inside the visitors.

Configured dynamic RegExp remains blocked by native nonconstant-pattern
lowering. Whole-rule emitted JavaScript and shared-driver integration remain
blocked by the native checker binding, which RuleContext does not expose. The
three React HIR/SSA/capture reservations remain parked. No new rules are claimed:
all remaining ranked entries are already ported or claimed on the fetched 601
origin refs. See rules/wave-22-seventh/LANDING_REPORT.md and its owned evidence.

## Main c7991b900 landing refresh

Rebased onto origin/area/stage1-lint b84a9d931, including main c7991b900. All
eighteen existing implementations passed their byte oracles, mutants, handles
and sanitized frozen corpora again with the freshly rebuilt compiler. The
older twelve-rule gate passed in 1002.893s. No rule or checker source changed.
Dynamic native RegExp and whole-rule JavaScript/shared checker binding remain
blocked; the three React analysis claims remain parked. The refreshed 632-ref
claim inventory has no available ranked rule, so no new reservation is made.
See rules/wave-22-seventh/LANDING_MAIN_C7991B900.md for complete fresh evidence.

## Registry integration b46914832 landing refresh

Rebased onto origin/area/stage1-lint b46914832, including main c7991b900. All
eighteen existing ports passed complete byte oracles, mutants, released handles
and sanitized frozen corpora again; older gate 739.135s. Registry descriptor and
mutation tests pass. Owned rule/checker sources are unchanged. Existing dynamic
RegExp and checker-context/JavaScript gaps remain; the three analysis claims are
parked. No available ranked rule remains and no new claim is made. Fresh evidence
is in rules/wave-22-seventh/LANDING_REGISTRY_B46914832.md.

Landing refresh: rebased onto area d3a37422c including main b6b1538b0. All eighteen existing ports re-green; full evidence in rules/wave-22-seventh/LANDING_TYPEOF_B6B1538B0.md. No new reservation: all 197 ranked rules accounted for across 653 origin refs. Existing dynamic RegExp, shared checker/JavaScript and parked React analysis gaps remain.

Unparked onto the area checker by merge. Fifteen owned unified descriptors added; original defaults and supported React options rechecked. Shared gate and configured checker/RegExp blockers remain, with exact reproducers and evidence in stage1/cohere/typeaware/rules/wave-22-seventh/UNPARK_REPORT.md. Prior HIR/SSA/capture React claims remain parked. No additional claims.
