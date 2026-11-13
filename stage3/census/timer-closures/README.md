# Timer closures checkpoint

Adaptation: 108e464716912d16359d542d65f830360c7283d6 on main 0aba99914be812ef7a0676c5ddd93024cefc81b5.

The refreshed adapted lane passed all 398 shards: 106,366 passing tests, one expected API failure, zero pending. The refreshed main control passed 360/398 shards. No lane process remains running. CHECKPOINT.json lists all 38 unrun main shards exactly: project-reference-cases-02 through -23 and symlink-case-00 through -15. The complete refreshed aggregate comparison is unrun.

Both refreshed full applies passed. Adamic types on 79 roots, stock diagnostics on those roots and the stock full source solution all passed with zero diagnostics for both trees. Source-scope checks passed, and the new stage replay made zero edits. The earlier-base replay of 41, 93, axolotl-x and crypto passed with source byte/mtime equality; those four earlier-stage replays have not been repeated on the refreshed tip.

All required refreshed witnesses passed: poller mutation via setCustomPollingValues, request-field mutation, WatchHost path, and the separate original blocker probe. Before/after witness stdout is byte-identical. All four mutants exited 1: poller-schedule-capture, request-schedule-capture, timer-label-baseline, and restored-variadic-slot. No required witness or mutant is unrun.

Both refreshed full-mode sys.ts census runs and their comparison passed. Main has 233 records (33 Refused, 90 NotYet, 110 Boundary); adapted has 236 (35 Refused, 89 NotYet, 112 Boundary). There are 225 retained, 8 gone and 11 new records, all preserved in census-delta-tip.json. The captured Function.call road introduces an inherited-library-member refusal. None of the ten timer calls is optional; calls.json in the adaptation records their exact spellings.

The initial e3f59bf55 controls completed all 398 shards on both sides and their aggregate comparison passed: matching identities, counts and every non-API baseline byte, with exactly six sanctioned API members changed. Those completed initial runs are archived separately from refreshed runs.

Every status JSON records the actual child exit. Raw setup failures are retained too: missing source environment, a scratch parser resolving upstream LKG rather than pinned 6.0.3, and an unbuilt declaration target were corrected before the successful definitive runs. The final suffixed records distinguish successful proofs from those setup attempts. No evidence is stored in the adaptation directory.
