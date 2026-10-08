substr with a length lowers through the existing UTF-16 slice path.
Commits: follows object defaults aa1e2961 on codex/generic-function-value; pinned host next stop was 405:16.
Checks: lower 15.641s; uncached substr/string/default regression oracles 1.180s; new fixture counts 21 allocations/frees, 6 retains, 29 releases, peak 3.
Mutant: interpreting length as an end index changes both backends' stdout against Node; both exit 0 with empty stderr. The valid mutant is string-substr-length-end-mutant-valid.log.
Uncovered: coercive nonnumeric arguments and extra arguments retain NotYet; the checker requires a numeric start, so undefined start/no-argument JavaScript-only calls are not claimed; no full repository gate or full counts update.

Date: 2026-10-08 UTC. Capture receiver, start and count once as ordinary IR
helper parameters. Truncate numeric positions, treating NaN as zero. Resolve a
negative start relative to UTF-16 length and clamp it. Clamp count to the units
remaining; an omitted or explicit undefined count means all remaining units.
Call the existing slice operation from start to start plus count. This adds no
runtime tag, string representation or memory-management path.

The fixture covers positive/negative/out-of-range/fractional starts, NaN and
both infinities, omitted/undefined/negative count, prototype.call, surrogate
halves observed through charCodeAt, and r/s/l side effects in operand order.
An initial fixture used calls outside the checker's substr declaration and
failed loading; it was corrected before the valid oracle and mutant proofs.

The 30 GB Go build cache exhausted the overlay; go clean -cache recovered
29 GB without deleting sources or evidence, and the valid oracle then passed.
Logs in /tmp: string-substr-length-{first,valid2,end-mutant-valid,lower,counts,
final-oracles}.log. Full counts updater remains blocked by the pre-existing
regexp_tree.ts project-root diagnostic recorded in FRESH_MAP_RETURNS.md.
Setup for this unit: 30.366s total, clang .186s, build 30.106s, cache 30.340s;
nproc 5.
