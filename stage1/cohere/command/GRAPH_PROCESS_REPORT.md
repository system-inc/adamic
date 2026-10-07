Import-graph scoping and real child execution, continuing codex/stage1-command
from 526836c40ff5e4343f510657596cb76c7f6fb3f7. This is a partial gap closure:
compiler graph construction and the multi-project scheduler remain open.

Built graph.ts: reverse resolved imports within the project population, duplicate
suppression, transitive consumers, cycles, compiler-population output order,
500-file limit with whole-tree fallback, exact note text and dependent counts.
Scope values are copied; the caller's write scope is preserved. The overlay captures
Go's actual fallback fmt.Fprintf output in a per-call builder, avoiding mirrored
expected wording or a global stderr race during parallel original tests. Rebuild transfer
matches names in rebuilt order and reports missing names sorted by Go UTF-8 order.
The production API consumes compiler files/edges, not tsconfig roots. Go graphs
supply test inputs only; expected scopes, closures and process output never enter
the TypeScript implementation. There is no Go answer bridge in production.

Built process.ts and the typed runProcess primitive across prelude, lowering, IR,
freshness, native emission, native C and JavaScript backend/Node runtime. Actual
fork/execve launches an absolute executable with closed stdin, cwd, patched
inherited environment and one descriptor capturing both streams. A close-on-exec
handshake distinguishes launch failure from a child returning 127. The parent
waits, decodes owned output, and returns exit status, signal and launch error.
Temporary capture avoids pipe-capacity deadlock; parent buffers, descriptors and
owned runtime values are released. The Node binding uses real spawnSync.
No protected compiler file was edited. Existing CLI host adapters are unchanged.

Go's own TestNarrowToClosureAddsConsumersAndCountsThem,
TestNarrowToClosureFallsBackRatherThanTruncating and
TestNarrowToClosureLeavesAWholeTreeScopeAlone run with their original assertions.
The overlay compares every helper call across native, source Node and backend,
using real Go Program.GetResolvedModules and ordered ProjectFiles as inputs.
Additional fixtures cover cycles, empty scopes, rebuilt/lost files, exact 500/501
closure sizes, and the graph-construction gap witness. That witness observes
roots a.ts,b.ts but compiler ProjectFiles b.ts,a.ts after a imports b.

Seven child cases call Go's production runProject with a real stand-in executable,
then run the port independently while fixtures exist: successful capture, exit 7,
self-SIGTERM, missing/denied executable and missing/non-directory cwd. Ordered combined bytes, error
strings and verdict agree, alongside cwd, spaced arguments, alternate tsconfig,
engine/label environment, cleared inherited verdict/yield and supplied yield path.
Canonical stdout is compared byte for byte; stderr must be empty and each probe
must exit 0. This does not claim the full upstream mixed-project CLI tests pass
in Adamic. Child output has no clock text; no normalization hides differences.

Eight primitive cases use independent direct Node spawnSync, not oracle/adamic.mjs
as the reference. They cover success, ordinary exits 19 and 127, self-SIGTERM, environment removal
and repeated replacement, empty stdin, actual malformed UTF-8/emoji/CR/NUL and
192 KiB output. Observed Node text/status/signal is canonically quoted by Go and
compared to native, Node source and backend. Native builds use ASan/UBSan and run
with default Linux LeakSanitizer enabled. The first fixture draft over-escaped
shell octal/newline sequences; it was corrected before final validation. An
initial typed-array/context inference and multi-push refusal were corrected in
source; compiler refusals are not mutant evidence.

Four new executable mutants:

| Mutant | Catch |
| --- | --- |
| Reverse map records target instead of importer | Original Go consumer/dependent-count boundary, all three execution paths |
| Raise closure limit to 50,000 | Original Go fallback and generated exact-boundary comparisons, all three paths |
| Discard the child's nonzero verdict | Go runProject exit-7 fixture, all three paths |
| Native runtime replaces the actual child status with zero | Independent Node spawnSync control; native output says 0 instead of 19 |

Every mutant checks a passing baseline and must execute cleanly with exit 0 and
empty stderr before an output disagreement counts. No compiler or sanitizer
failure counts as a kill. Previous unit's five mutants remain in the full suite.

Setup succeeded, Go 1.27.1, clang 20.1.8, Node 24.19.0. nproc printed 5:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (47s)
setup: done in 47s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Commands source /workspace/adamic-tools/env.sh; test output goes directly to logs:

```sh
bash cloud/setup.sh > /tmp/stage1-graph-setup.log 2>&1
go test -count=1 -timeout 15m -v ./stage1/cohere/command > /tmp/stage1-graph-release.log 2>&1
go test -count=1 -timeout 15m ./internal/ir ./internal/lower ./internal/fresh ./internal/native ./internal/javascript ./internal/load > /tmp/stage1-graph-compiler.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 15m -v ./internal/oracle -run '^(TestInput|TestRealPathAgreesWithNode|TestCountsAreRecorded)$' > /tmp/stage1-graph-oracle.log 2>&1
go vet ./... > /tmp/stage1-graph-vet.log 2>&1
go -C cohere run ./command/cohere --format-only --no-cache --directory /workspace/adamic stage1/cohere/command/graph.ts stage1/cohere/command/graph_probe.ts stage1/cohere/command/process.ts stage1/cohere/command/process_probe.ts > /tmp/stage1-graph-format.log 2>&1
```

Scoped compiler packages passed: lower 17.631 s, fresh 26.304 s, native 92.197 s,
load 1.140 s; IR and JavaScript have no package tests. Uncached oracle/counts gate
passed in 29.504 s, including 263 native cache misses and two realpath mutants.
No full repository gate was run. No new fixture was added to the oracle fixture
catalog; command probes are held by their independent package tests. Files/second
for full graph construction is unmeasured because there is no production builder.
Prior enumeration timing remains in REPORT.md and is not graph throughput.

Remaining work: TypeScript parser/resolver/compiler program construction and its
ProjectFiles validation/yielding, checker integration, actual post-fix rebuilding,
parallel project scheduling, worst-exit aggregation, parent signal forwarding,
Swift engine dispatch, yield-file lifecycle, and full CLI engine parity. Native
processes do not expose arbitrary Node child_process or ambient process.exitCode.
Unix errors beyond held missing/denied-executable/non-directory cases, platform paths/case folding, arbitrary
binary Go output and capture resource failure are unheld. See GAPS.md for limits.

The added non-directory cwd case caught a real translation defect: the adapter
said chdir <cwd>, but Go says fork/exec <executable> because its preliminary Stat
succeeds on a file. process.ts now performs that preliminary stat and preserves
Go's exact launch-error split. Missing cwd retains its chdir wording. This was
an oracle disagreement corrected before commit, not accepted as mutant evidence.

Validation results: complete command package PASS 200.559 s, including the prior
36 CLI cases, 74 argument cases, 24 original helper tests and five old mutants.
After strengthening actual Go note capture and adding process error/exit edges,
the final graph/process package filter PASS 56.007 s, including all three Go
narrowing tests, generated graph/rebuild/500-boundary fixtures, seven Go child
cases, eight independent Node primitive cases and all four new executable mutants.
The final filter was:

```sh
go test -count=1 -timeout 15m -v ./stage1/cohere/command -run '^(TestOriginalGraphAndProcessBoundaries|TestGraphAndProcessExecutableMutants|TestProcessPrimitiveAgreesWithNode|TestProcessRuntimeVerdictMutant)$' > /tmp/stage1-graph-process-release.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 15m -v ./internal/oracle -run '^TestInputAgreesWithNode$' > /tmp/stage1-graph-input.log 2>&1
```

The separate exact-name input oracle PASS 1.001 s, seven input fixtures, seven
uncached probe misses. The first combined oracle filter used TestInput, which is
not an existing test name; it ran counts and realpath only. The corrected command
above ran the input comparisons. Full-suite and final narrowed logs are respectively
/tmp/stage1-graph-release-final.log and /tmp/stage1-graph-process-release.log.
Formatting, vet and git diff --check have no findings. Full mixed-engine CLI and
termination tests remain outside the native parity claim; capture-file transport
also differs from Go pipes (seekability and inherited-writer lifetime are unheld).
