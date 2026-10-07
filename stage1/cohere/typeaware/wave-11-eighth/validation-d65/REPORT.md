Rebased wave 11 onto origin/area/stage1-lint d65a8f93, including current main 39638d9e and runtime profile changes.
The evidence commit following this report is pushed only to codex/typeaware-wave-11.
All eight owned suites, bridge tests and vet PASS; filtered compiler retry PASS 12.279s.
All 46 worker mutant observations revalidated; semantic byte and required released-handle panic checks caught them.
No unclaimed rules remain: 605 origin refs, 592 distinct trees, 34 reservation documents; no new claims or code.

The first compiler run could not create its runtime-cache build directory:
no space left on device. Its failed log and exit code remain in results.json.
Retry sourced /workspace/adamic-tools/env.sh and set
XDG_CACHE_HOME=/tmp/wave11-latest-cache, then ran go test ./internal/oracle
with the same fixture families, -count=1 -timeout 10m -v, output directly
to compiler-retry.log. The explicit [.]a$ filter avoids shell escaping.
Retry PASS covers original Node, native sanitizer execution, emitted JavaScript
and the one-byte mutant. Only named obsolete worker binaries were deleted;
source and evidence logs were retained. Bridge process 79.937s, seventh suite
86.297s, eighth 106.644s. Complete commands and outputs are adjacent.

Every existing production-default rule was checked on its controls and frozen
repository/compiler populations, including full findings, fixes and suggestions,
sanitizers and released handles. Shared helper/runtime changes were retained.
No shared harness/generator or protected compiler implementation was edited.
No new node-kind declaration was needed. Older standalone numeric metadata
is not claimed compatible with the shared name-based registry. Four analysis
claims remain parked. Existing nondefault-option limits, shared emitted-JavaScript
lint execution and full repository gate remain uncovered. Original quiet
native/Go times remain repository 0.9227/0.1380s and compiler 7.0781/0.3421s
on c01907a7; current concurrent verification timing lines are observations
in d65-eighth.log, not quiet benchmark medians.
