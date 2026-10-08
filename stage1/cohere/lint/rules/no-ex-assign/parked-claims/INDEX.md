# Wave 01 checker audit

Area base: c4bdc23fa86d55cf7e579989201c11258f4d3a62.

The checker context has landed. Every claim below stops at an unanswered
question or missing shared helper. These are blockers, not rule ports.
No descriptor was disabled or installed. No private checker/helper was added.

| Claim | Exact call and reproducer |
| --- | --- |
| nexus/correctness-no-implicit-return | [nexus-correctness-no-implicit-return/BLOCKED.md](./nexus-correctness-no-implicit-return/BLOCKED.md) |
| nexus/correctness-require-child-process-error-listener | [nexus-correctness-require-child-process-error-listener/BLOCKED.md](./nexus-correctness-require-child-process-error-listener/BLOCKED.md) |
| nexus/correctness-require-response-status-check | [nexus-correctness-require-response-status-check/BLOCKED.md](./nexus-correctness-require-response-status-check/BLOCKED.md) |
| nexus/performance-no-independent-await-in-loop | [nexus-performance-no-independent-await-in-loop/BLOCKED.md](./nexus-performance-no-independent-await-in-loop/BLOCKED.md) |
| no-else-return | [no-else-return/BLOCKED.md](./no-else-return/BLOCKED.md) |
| react-hooks/globals | [react-hooks-globals/BLOCKED.md](./react-hooks-globals/BLOCKED.md) |
| react-hooks/immutability | [react-hooks-immutability/BLOCKED.md](./react-hooks-immutability/BLOCKED.md) |
| react-hooks/no-deriving-state-in-effects | [react-hooks-no-deriving-state-in-effects/BLOCKED.md](./react-hooks-no-deriving-state-in-effects/BLOCKED.md) |
| require-atomic-updates | [require-atomic-updates/BLOCKED.md](./require-atomic-updates/BLOCKED.md) |
| require-await | [require-await/BLOCKED.md](./require-await/BLOCKED.md) |
| symbol-description | [symbol-description/BLOCKED.md](./symbol-description/BLOCKED.md) |
| @typescript-eslint/no-deprecated | [typescript-no-deprecated/BLOCKED.md](./typescript-no-deprecated/BLOCKED.md) |

For each listed claim: upstream cases certified 0; new mutant executions 0.
Earlier captured-fact proofs on the old branch are not counted here.

Landing first: the area was merged, without rebasing, into both previous
local worker branches. The old no-ex-assign port is already an ancestor of
the area. The wave-01 merge stays local pending a package gate on that
distinct branch; this new review branch starts from the area and carries
only these owned rule-directory notes.
