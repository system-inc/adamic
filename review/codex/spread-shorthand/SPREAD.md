Spread arguments into push and unshift use an ordinary IR helper. Its receiver is evaluated first and arguments are packed in order, snapshotting each spread before later arguments run. Rest calls retain their existing lowering.
Base: compiler/area-next-fixtures f0c6e6fc2b493722403976a97758ee6227913e3c. No other worker branch was merged.
The ahra CLI is unavailable in this environment. Witnesses were read from scout/42-settings-resolution-on-train 79c7f132e, stage1/cohere/command/SCOUT.md and settings/gaps/call_spread.a.
The focused TestCallSpreadInsert compares Node, sanitized native and the JavaScript backend, with LeakSanitizer. TestCallSpreadEvaluatedTwiceMutant changes the emitted spread-producing call into two calls, releasing the first result. It must execute successfully without sanitizer errors and disagree with Node stdout.
The original settings/main.a now stops at main.a:28:86 on JSON literal shorthand, after passing the original resolve.a:251:55 spread boundary. The sources were extracted into a scratch directory without changing them.
Counts: only the new call_spread_insert.a row is added: allocations 49, frees 49, retains 56, releases 81, peak 10, regions 0. Existing rows do not change.
Commands: focused uncached oracle with -count=1 -timeout 80s; TestCountsAreRecorded with -update-counts (44.063s); go vet ./internal/lower ./internal/oracle. Logs are beside this report.
Setup: go 0.021s, Node 0.022s, submodules 0.055s, markdown dependencies 0.068s, clang 0.150s, build 41.593s, total 41.818s; nproc 5 (4 CPU quota).
This slice does not add heterogeneous tuple spread storage or custom iterator call expansion. Their existing NotYet diagnostics remain.
