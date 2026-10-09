# Deletion-set replay

Main: 7b9d4272c28f59530ab13daa5c49067e47933b06

77 records, 66 distinct standalone patches; zero stale diffs or unresolved build failures. Baseline passed in 277.272 seconds. Two bounds-diagnostic mutants passed the remaining package completely. Their two guards are retained jointly. A separate interface-dispatch mutant completed with only a witness failure after 25 panic exclusions; its two candidate guards are retained conservatively. The other 15 candidates meet the gathered-mutant deletion criterion.

See report.json for the decision, mutant-list.json for the inventory made before runs, matrix.json for exact commands and observed outcomes, and compressed logs for original Go JSON output. No production source or test was changed.

Replay elapsed including interruption: 7249.502 seconds. Sum of replay wall times: 11711.988 seconds with two workers. Witness-only initial proof and disk-aborted/interrupted attempts are labeled separately.

Scope corrections, default skips, environment recovery, and the rejected additional cache cleanup are recorded in notes.md and report.json.
