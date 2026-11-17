# Readonly narrowing filter freeze measurement

Report-only task #ev14ch9 on source base 49f1d643. The result and per-site table
are under Measurement 2 in [REPORT.md](../filter-narrowing-copy/REPORT.md).
Measurement 1 was pushed at 2f473173f before this measurement began.

Tools originate at df886e40, stage3/census/filter-freeze. The observer, control,
resolver and aggregator are unchanged; tool-provenance.json records exact hashes.
The instrumenter selects original overloads 3 and 7 (declarations at 265 and 273),
compares them with the previously stored independent 713-root baseline census,
requires exactly 23 sites, and checks every source span before replacing it.
Instrumentation retains all 157 call result/argument types and overload choices.
Other 134 call expressions and core filter are unchanged.

The complete lane finished with 106363 passing, 4 failing, 0 pending: the expected
API baseline failure and three observed __pos writes. No timeout or observer error.
All sites were called; four never returned an array identity. The unchanged lane
verdict is FAIL, including its worker-policy error. No gate or expectation changed.

See evidence/commands.json for commands and paths. Use a fresh scratch copy of the
unchanged completed adapted tree, resolve.cjs before wrapping, instrument.cjs,
then resolve.cjs and validate-types.cjs after wrapping. Install/build the private
tree before running control.cjs or built/local/tsc.js with observe.cjs preloaded.
run-lane.py wraps the original lane in memory, instrumenting only after apply exits
0. FILTER_FREEZE_WORKERS=4 matches df886e40's canonical full-corpus run; all runners
and the 40000 ms test timeout remain unchanged. The lane's guard still expects 8.

aggregate.py retains all raw candidate classifications. classify.py separately
identifies source-proven runCode rethrows with matching process, case and full
stack; it never discards raw observations or treats propagation as a new write.
The 12 lane candidates are three assignments, three rethrows and six unrelated
explicit decorator throws. Counter archives and logs include original process
exits and exact stacks. Bootstrap and preliminary overlapping self counts are
separate; only self-final-counts is canonical self evidence.

Identity remains a premise. Zero direct comparisons were found; arbitrary later
comparisons through slots are not seen. These are counts under freezing, whose
blocked writes can truncate individual test bodies. No copying behavior, landing
adaptation, ledger acceptance or internal compiler change is claimed.
