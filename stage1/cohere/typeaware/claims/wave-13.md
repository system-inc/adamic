# Type-aware wave 13

Branch: codex/typeaware-wave-13.
Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.

Selected from VOLUME_REPORT.md's linked compiler-all.counts and
repository-all.counts, descending combined volume with lexical ties, excluding
the 26 rules already ported on the base:

| Remaining position | Rule | Combined findings |
| --- | --- | ---: |
| 37 | no-unassigned-vars | 2 |
| 38 | preserve-caught-error | 2 |
| 39 | @typescript-eslint/consistent-type-exports | 1 |

Checked implementation filenames and Markdown claims under stage1/cohere on
all 273 fetched origin refs. No matching ports or claims were found; no rule
is skipped. Claim pushed before implementation. New Adamic files use .a.

## Continuation claim

Original three ports are tested and pushed through 7ca1f20e. Fetched all origin
heads and scanned 325 refs for claim Markdown. Excluded the 25 checker-dependent
ports found on origin/main and origin/codex/tsgo-c-library and 96 ranked rules
named in origin claims. The first three remaining candidates, by descending
combined volume and lexical ties, are:

- nexus/correctness-no-process-exit-after-output (0)
- nexus/correctness-no-uncleared-race-timeout (0)
- nexus/correctness-require-blocking-standard-streams (0)

Claim update is committed and pushed before implementation. Shared generators
and test harnesses will not be edited. New executable Adamic files use .a.

Continuation status: completed by implementation 7f0647dd. All three ports have
normal and sanitized byte agreement, rule mutants, bridge mutants and lifetime
checks. See wave_13_next/REPORT.md and its validation/ evidence. The previous
BLOCKED.md interpretation is superseded. No further rules were claimed yet.


## Second continuation claim

All six previous claims are implemented, tested and pushed through 60e1618c.
Fetched all origin heads and scanned 347 refs after that push. The 197-rule
ranking has 25 existing checker ports on main/base and 120 claimed ranked rules
on origin; 52 remain. The next three by combined volume and lexical ties are:

- no-obj-calls (0)
- no-object-constructor (0)
- no-promise-executor-return (0)

This claim update is pushed before any implementation. New native files use .a,
and shared registration generators and harnesses remain unchanged.


Second continuation status: implementation cc7b89d0. no-obj-calls and
no-promise-executor-return agree on all exported upstream controls;
no-object-constructor remains partially blocked on three shared-parser refusals.
Normal and sanitized runs agree on 430/433 controls and both full frozen corpora.
All three rule mutants, the raw shorthand mutant, question guards and released
handle checks are recorded in wave_13_more/REPORT.md. No further rules were
claimed because this claim is not fully complete.
