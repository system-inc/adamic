Built: twelve existing ports rebased and re-green on current main and lint integration; nothing eligible remains.
Commits: tested source f8af20264, green pushed proof 25d7bc704; this audit commit changes evidence only.
Checks: 19 landing steps, 43 Node fixtures and 100 required external passing events, zero skips; 664 origin refs audited.
Mutants: all twelve rule witnesses, guard witnesses, nine record mutants, eight typeof executions and external comparison mutants caught.
Not covered: three parked hooks, four production bridge routes, shared checker context, own emitted-JavaScript parity and the full repository gate.

The ranking audit ran after the green proof push to codex/typeaware-wave-07.
Both fetched origin/main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and
origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898 are
ancestors of this branch. No main or area branch was pushed.

All 197 ranked names were checked against original bridge ports, main sources
and rule descriptors, and all claim files on 664 origin refs (33 distinct blobs).
No eligible unclaimed names remain, so no new claims were made. The complete
reference pins, ranking and exclusions are in evidence/typeof-final-selection.json.gz,
with length and SHA-256 in its metadata file.

TYPEOF_LANDING_REPORT.md records fresh commands and outputs, every mutant
detector, setup measurements, native versus Go medians and concrete blockers.
Earlier reports and selections are historical. The source and proof pins above
certify the current landing refresh.
