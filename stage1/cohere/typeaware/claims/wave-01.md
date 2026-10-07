# Type-aware wave 01

Branch: codex/typeaware-wave-01.
Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.

Claimed checker-dependent rules, in descending combined volume after excluding
all 26 existing ports:

1. nexus/correctness-no-implicit-return: 251 compiler, 0 repository.
2. @typescript-eslint/no-deprecated: 238 compiler, 0 repository.
3. no-else-return: 202 compiler, 0 repository.

Selection uses VOLUME_REPORT.md's full checker-dependent counts in
validation-volume/compiler-all.counts and repository-all.counts.
All origin branches were fetched and inspected for typeaware claims and existing
stage1 implementations of these rule names. None were found; no candidate skipped.
No rule implementation was written before this claim commit was pushed.

## Continuation after the original three ports

Original rules completed and pushed in 85ad9cf11e2fe77cecd70e919f37b60b7adea675.
Fetched all origin heads again on October 7, 2026, inspecting main at ef3d907e
and tsgo-c-library at 5afbdb83, the source ports and all origin claim documents.
Ranking is descending combined volume with lexical ties, as in the first claim.
The next three available checker-dependent entries are reserved here:

1. nexus/correctness-require-child-process-error-listener (0 compiler, 0 repository).
2. nexus/correctness-require-response-status-check (0 compiler, 0 repository).
3. nexus/performance-no-independent-await-in-loop (0 compiler, 0 repository).

An initial scan found process-exit-after-output, uncleared-race-timeout and
blocking-standard-streams available, but the immediate pre-claim refresh found
new origin claims for them. They were skipped without implementation.
Released older reservations were also checked: their rules now have active
continuation claims on wave 22. No selected rule was ported on main or the
bridge branch or named in an origin claim at the final pre-claim inspection.
This claim update is committed and pushed before implementation begins.

## Third batch

The previous six ports are completed and pushed through 98f1cd3230075b5363f0736395db9c5b88e5fe22.
After refreshing all 356 origin refs, the next unported and unclaimed entries are:

1. react-hooks/globals (0 compiler, 0 repository).
2. react-hooks/immutability (0 compiler, 0 repository).
3. react-hooks/no-deriving-state-in-effects (0 compiler, 0 repository).

The immediate refresh found new claims for prefer-promise-reject-errors,
prefer-regex-literals, prefer-rest-params and react-hooks/exhaustive-deps.
Those candidates were skipped. Main is ef3d907e and the bridge branch is 5afbdb83.
This reservation is pushed before implementation.

## Parked React batch

Status: parked by Ahra's React parking instruction; counts as finished for the
landing-first cap. No native implementation is represented as complete.

- react-hooks/globals: native JSX parsing and component gate integration.
- react-hooks/immutability: native React high-level IR, SSA construction and capture analysis, plus JSX.
- react-hooks/no-deriving-state-in-effects: native React high-level IR, SSA and capture/context identity, memo erasure and callback inlining, plus JSX.

Analysis modules are being ported on #dnv6f2c. JSX support is landing on
area/stage1-lint. Existing blocker controls and partial work are pushed in
wave_01_third. The six completed ports are oracle-green on main f8013f0b and
published through 25f5e73b. These three reservations stay parked until their
shared dependencies land; they are not offered to another worker implicitly.

## Fourth batch after parking React

The six completed ports are landing-ready on origin/main f8013f0b, oracle-green
and pushed. The React batch is parked under the instruction above.
After inspecting 529 origin refs, main and tsgo-c-library ports, and every origin
claim document, reserve these next non-React entries (all zero on both corpora):

1. require-atomic-updates
2. require-await
3. symbol-description

React entries and structure/react-hook-no-any-type are skipped while React is
parked. These three use checker/ordinary CFG facts rather than React HIR/SSA.
The immediate pre-claim refresh found none claimed. This reservation is pushed
before implementation. The latest contract declares typescript-go ast.Kind names in rule.json and
consumes the supplied node. Earlier numeric metadata was superseded by Ahra
and corrected in all nine manifests.

Fourth-batch status: partial, not finished and not parked.
- symbol-description: native supplied-node listener and raw bridge question tested using captured syntax; shared native-driver integration pending.
- require-atomic-updates: native transfer/meet kernel tested; event collection and full-rule analysis pending.
- require-await: native numeric syntax kernel tested; checker-demand analysis, diagnostic/suggestion rendering and full-rule replay pending.

The exact observations and limits are in wave_01_fourth/REPORT.md. No next batch
will be claimed while these three remain incomplete.
