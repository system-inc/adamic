# Batch 6 native Linux profile

Branch `codex/stage1-lint-batch6-perf`, based on batch 6 commit
`c5d128ab4158197a3b3a9f6452b9224e946ac462`. Production changes are confined to
`registry.ts` and its single caller in `lint.ts`. Compiler, parser and runtime
sources are unchanged. Native release (`-O2`, no sanitizers, no counting build) throughput improved
from 301.01 to 392.38
findings/s on this worker, 30.35%. The previous unit's 294.69 findings/s was a
separate measurement; the matched comparison below uses preserved binaries.

## Build flags audit and matched rerun

The original headline binaries used `adamic build` with `native.Options{}`:
**release `-O2`, no sanitizers, no `-DADAMIC_COUNT`, no `-g` or profiling flags**.
The original build's ephemeral temporary filename was not recorded. An audited
rebuild captured the following actual clang command and produced exactly the
original optimized binary, SHA-256
`9aaa6ec5b3de380091b3a5a924a90735d676b5aba7f016ee06b0ef069e744417`:

```bash
/workspace/adamic-tools/llvm/bin/clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -I /home/agent/.cache/adamic/runtime/bc4c6117d4c8bc77c88124c1fe31880cd08703ecd34b86045df008eca6d98f38 -o /workspace/scratch/batch6-perf/flags-rerun/adamic-release /tmp/adamic-gate/adamic-build-3783009302/main.c -Xlinker --whole-archive /home/agent/.cache/adamic/runtime/bc4c6117d4c8bc77c88124c1fe31880cd08703ecd34b86045df008eca6d98f38/runtime.a -Xlinker --no-whole-archive -lm
```

The runtime archive was also compiled with the same release flags. All archive
compile invocations and the final link command are in
[flags-adamic-build-commands.sh](batch6_perf_evidence/flags-adamic-build-commands.sh).
These captured temporary paths document the invocation; C and runtime sources
in scratch/the repository are the persistent reproduction inputs.

For the requested matched rerun, the exact preserved baseline and optimized C
were freshly compiled together with the unchanged runtime source files. The
only flag difference between release and sanitized builds is:

- Release: `-O2`.
- Sanitized correctness: `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`.

Both use all twelve common flags shown in the captured command above. Neither
uses `-DADAMIC_COUNT`. All four expanded, exact build commands are recorded in
[flags-build-commands.sh](batch6_perf_evidence/flags-build-commands.sh) and
[flags-build-commands.json](batch6_perf_evidence/flags-build-commands.json).
For example, the optimized builds are reproduced from the repository root by:

```bash
/workspace/adamic-tools/llvm/bin/clang \
  -std=c11 -Wall -Wextra -Werror -pedantic \
  -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function \
  -Wno-unused-parameter -Wno-self-assign -ffp-contract=off \
  -fno-optimize-sibling-calls -O2 -I internal/native/runtime \
  /workspace/scratch/batch6-perf/optimized.c internal/native/runtime/*.c -lm \
  -o /workspace/scratch/batch6-perf/flags-rerun/optimized-release
/workspace/adamic-tools/llvm/bin/clang \
  -std=c11 -Wall -Wextra -Werror -pedantic \
  -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function \
  -Wno-unused-parameter -Wno-self-assign -ffp-contract=off \
  -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined \
  -fno-sanitize-recover=all -I internal/native/runtime \
  /workspace/scratch/batch6-perf/optimized.c internal/native/runtime/*.c -lm \
  -o /workspace/scratch/batch6-perf/flags-rerun/optimized-sanitized
```

The same sorted 77-file manifest has SHA-256
`bc54951e3474e72273433954b6fc4c801f22333dee00246ea04a1b237fed6211`.
All fifteen rules run and produce 481 findings. Five rounds rotate the order of
the four native binaries and Go. Timing excludes compilation. Ordinary findings
and repairs were first compared on every binary: 11,631,597 identical bytes,
exit zero, empty stderr. Sanitized executions use
`ASAN_OPTIONS=detect_leaks=1:halt_on_error=1` and
`UBSAN_OPTIONS=halt_on_error=1`, with no sanitizer or leak reports.

| Implementation   | Release best / median seconds (`-O2`, no sanitizers, no counting build) | Sanitized best / median seconds (`-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, no counting build) |
| ---------------- | ----------------------------------------------------------------------: | -------------------------------------------------------------------------------------------------------------------: |
| Baseline native  |                                                     1.589939 / 1.603882 |                                                                                                  6.757687 / 6.773706 |
| Optimized native |                                                     1.205443 / 1.215393 |                                                                                                  4.655801 / 4.675865 |

Go cohere, built with the unchanged ordinary `go build` oracle command, takes
0.297717 seconds best / 0.302968 median. Release native (`-O2`, no sanitizers,
no counting build) improves from 302.53 to 399.02 findings/s; Go is 1,615.63
findings/s. The release improvement is 31.90%. Sanitized timing is reported only
as the requested diagnostic comparison; performance claims use release builds.

Timed invocations, full flags, five raw timing observations, binary hashes and
manifest hash are in [flags-rerun.json](batch6_perf_evidence/flags-rerun.json).
The exact script is [flags-rerun.py](batch6_perf_evidence/flags-rerun.py):

```bash
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/batch6_perf_evidence/flags-rerun.py \
  > /workspace/scratch/batch6-perf/flags-rerun/run.log 2>&1
# The measured optimized runner invocation:
/workspace/scratch/batch6-perf/flags-rerun/optimized-release \
  --manifest /workspace/scratch/batch6-perf/compiler.txt --count
```

The runner's `--count` argument selects findings-only output and skips formatting
and fixing. It is independent of the compiler's allocation-counting build option;
none of these measured binaries contains `-DADAMIC_COUNT`. Sanitized builds remain
correctness artifacts. Every native timing emitted by `TestThroughput` now prints
the complete `native.Flags(native.Options{})` list, explicitly saying no
sanitizers and no counting build.

The resumed environment lacked `/tmp/adamic-gate`, which the toolchain sets as
`TMPDIR`; the first clang attempt failed with `unable to make temporary file:
No such file or directory`. Recreating that directory with mode 1777 fixed it.
No compiler/runtime or rule behavior changed in this audit.

The changed benchmark harness was executed with:

```bash
ADAMIC_LINT_BENCH=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/batch4-typescript \
  go test -count=1 -timeout 5m -v -run '^TestThroughput$' ./stage1/cohere/lint \
  > /workspace/scratch/batch6-perf/flags-rerun/throughput-test.log 2>&1
```

PASS, 20.050 seconds; release native (`-O2`, no sanitizers, no counting build)
1.204029 seconds / 399.49 findings/s; Go 0.311466 seconds / 1,544.31 findings/s.
This separate harness verification is not substituted for the matched table.
The full ordinary output agrees before timing and each native timing line shows
all flags. The prior unit's eighteen lint mutants and external oracle mutant
remain recorded above/below; this audit changes only reporting and evidence.

## Top three costs

Linux perf sampled `cpu-clock:u` at 997 Hz with frame-pointer call stacks.
Twenty passes over the pinned 77 TypeScript compiler files produce 9,620
findings. Baseline: 32,662 samples, approximately 32.76028 sampled CPU seconds,
zero lost samples. Percentages are self costs across the entire runner.

| Baseline symbol  | Self CPU | Approximate CPU seconds | Self CPU under batch 6 dispatch, percentage points of whole run |
| ---------------- | -------: | ----------------------: | --------------------------------------------------------------: |
| `adamic_release` |   15.17% |                 4.96991 |                                                            5.77 |
| `adamic_retain`  |    8.46% |                 2.77232 |                                                            1.65 |
| `Parser.node`    |    6.44% |                 2.10832 |                                                            4.55 |

The runtime functions implement reference ownership. The indexed AST getter
executes compiler-generated ownership, access checks and virtual calls. These
are the three hottest symbols; their remaining implementation costs belong to
the compiler/runtime. The actionable rule-code cause is repeated invocation of
inapplicable visitors: `visitRules` accounts for 26.88% inclusive CPU, about
8.80742 seconds. Its descendants include the costs in the last column.
Self and inclusive percentages overlap and must not be added.

## Rule-code fix

The traversal already holds the node kind. Pass that kind to the registry and
call only listeners whose syntax kinds apply. Every rule retains its own
configuration and semantic checks. Constructor dispatch preserves the original
order: default-param-last, then parameter-property assignment. No finding,
fix, suggestion or sorting logic changed. Formatting removed the redundant
final return and empty default arm; the final release binary is byte-identical
to the measured optimized binary (SHA-256
`9aaa6ec5b3de380091b3a5a924a90735d676b5aba7f016ee06b0ef069e744417`). This avoids irrelevant node getters,
enabled checks and fresh membership arrays while retaining each rule file.

Afterward: 25,079 samples, approximately 25.15446 CPU seconds, zero lost samples,
23.22% less sampled CPU. `release` falls to 3.41424 seconds (13.57%), `retain`
to 2.20160 seconds (8.75%), and `Parser.node` to 0.74724 seconds (2.97%). The
third hottest symbol after the fix is `Scanner.code`, 7.60% / 1.91174 seconds;
its absolute baseline cost was 1.92979 seconds. Its larger percentage does not
mean a regression. Parsing remains the largest remaining inclusive phase.

Clang inlines the optimized dispatcher into `Linter.walk`; an absent dispatch
frame is not zero cost. The stable walk boundary falls from 41.42% of 32.76028
seconds (about 13.57 seconds) to 23.94% of 25.15446 (about 6.02 seconds).
These are sampled observations, not exact phase timers.

One 77-file pass, using `--count` instrumentation:

| Counter                |    Baseline |  Optimized |
| ---------------------- | ----------: | ---------: |
| Allocations, all freed |   9,308,253 |  4,914,403 |
| Retains                | 108,304,329 | 67,893,139 |
| Releases               |  93,746,790 | 61,232,306 |
| Peak live allocations  |     730,903 |    730,903 |
| Regions                |           0 |          0 |

## Remaining compiler/runtime reproducer

[perf_probes/getter/main.ts](perf_probes/getter/main.ts) is the complete minimal
program reproducing the indexed getter and ownership costs without lint rules:

```typescript
/* cohere-disable max-classes-per-file -- keep the indexed getter reproducer in one standalone program */
// An indexed object getter makes the compiler's ownership traffic observable in perf.
import { panic, programArguments } from 'adamic';
class Entry {
    readonly kind: string;
    constructor(kind: string) {
        this.kind = kind;
    }
}
class IndexedNodes {
    readonly nodes: readonly Entry[];
    constructor(nodes: readonly Entry[]) {
        this.nodes = nodes;
    }
    node(index: number): Entry {
        return this.nodes[index] ?? panic('missing node');
    }
}
function matches(nodes: IndexedNodes, index: number): boolean {
    return nodes.node(index).kind === 'Identifier';
}
const args = programArguments();
const nodes = new IndexedNodes([new Entry(args[0] ?? panic('pass Identifier'))]);
const iterations = Number.parseInt(args[1] ?? '3000000', 10);
let findings = 0;
for(let index = 0; index < iterations; index++) {
    if(matches(nodes, 0)) findings++;
}
console.log(`${findings}`);
```

Build with `adamic build stage1/cohere/lint/perf_probes/getter/main.ts -o getter`.
Run `getter Identifier 3000000`. Node and sanitized native both print
`3000000`, with no sanitizer or leak report. A counted build makes eight
allocations, frees all eight, and performs 9,000,008 retains and 9,000,013
releases: approximately three ownership pairs per loop, without loop allocation.
The dynamic argument makes the compared kind a real heap string.

At 100,000,000 iterations, perf records 1,801 samples / approximately 1.80642
CPU seconds, zero lost samples: `IndexedNodes.node` 39.53%, `adamic_release`
19.27%, `adamic_retain` 5.77%, `adamic_virtual` 3.66%. This one program
reproduces all three baseline hotspot families. Ownership elimination or
getter devirtualization would require compiler/runtime work outside this unit;
no such edit is included. Unresolved addresses are not assigned a speculative
cause.

## Throughput and parity

Five fresh-process count rounds, rotating baseline/optimized/Go/Node order;
77 files, 481 findings, ordinary output first compared byte for byte on all four
executables: 11,631,597 identical bytes. Release builds use normal stage 0
`-O2` flags, without sanitizers, `-DADAMIC_COUNT`, or profiling's
debug/frame-pointer additions. Full timings
are in [benchmark.json](batch6_perf_evidence/benchmark.json).

| Backend                                                    | Best seconds | Best findings/s | Median findings/s |
| ---------------------------------------------------------- | -----------: | --------------: | ----------------: |
| Native baseline (`-O2`, no sanitizers, no counting build)  |     1.597941 |          301.01 |            299.87 |
| Native optimized (`-O2`, no sanitizers, no counting build) |     1.225853 |          392.38 |            389.47 |
| Go cohere                                                  |     0.308521 |        1,559.05 |          1,501.88 |
| Node optimized source                                      |     0.781033 |          615.85 |            597.14 |

Pinned TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8`; cohere:
`715ba94`. Machine: Linux 6.18.44 x86_64, `nproc` 5, cgroup CPU quota 4.
The corpus and binaries reside in `/workspace/scratch/batch6-perf` and
`/workspace/scratch/batch4-typescript`.

## Validation

Each command writes output directly to a log under
`/workspace/scratch/batch6-perf`; no test-output pipeline was used.

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/batch4-typescript \
  go test -count=1 -timeout 20m -v \
  -run 'TestRulesAgree|TestCompilerAndStage1Agree|TestBatch6DispatchMutants|TestBatch6GetterProbe' \
  ./stage1/cohere/lint > /workspace/scratch/batch6-perf/final-parity.log 2>&1
go test -count=1 -timeout 20m -v \
  -run 'TestBatch6Mutants|TestBatch6CountCheck|TestMutants' \
  ./stage1/cohere/lint > /workspace/scratch/batch6-perf/mutants.log 2>&1
go vet ./... > /workspace/scratch/batch6-perf/vet.log 2>&1
go test -count=1 -timeout 5m -v -run TestTheOracleCatchesOneByte \
  ./internal/oracle > /workspace/scratch/batch6-perf/oracle.log 2>&1
go test -count=1 -timeout 5m -v -run TestNestedConstructorGap \
  ./stage1/cohere/lint > /workspace/scratch/batch6-perf/nested.log 2>&1
```

Final parity PASS, 65.952 seconds: 761 upstream distinct cases, one exact documented
parser refusal checked on both backends, plus 34 generated cases = 794 supported
runs / 327,071 identical bytes. Compiler and stage1 corpus PASS: 196 files /
12,520,199 identical bytes. The additional corpus file is the getter probe.
Go, Node and sanitized native agree on findings, fixes and ordered suggestions.

All eighteen lint mutants were caught:

| Mutant                                             | Catch                                                   |
| -------------------------------------------------- | ------------------------------------------------------- |
| Constructor dispatch moved to CaseBlock            | Go byte oracle differs on Node and sanitized native     |
| VariableStatement dispatch moved to EmptyStatement | Go byte oracle differs on Node and sanitized native     |
| Module identifier suppressed                       | Same byte oracle                                        |
| Defaulted parameter suppressed                     | Same byte oracle                                        |
| Confusing operator suppressed                      | Same byte oracle                                        |
| Enum anchors on later duplicate                    | Same byte oracle                                        |
| Dynamic identifier accepted                        | Same byte oracle                                        |
| Redundant bang fix empty                           | Same byte oracle                                        |
| Construct span widened                             | Same byte oracle                                        |
| Assignment suggestion drops semicolon              | Same byte oracle                                        |
| Const insertion changes type                       | Same byte oracle                                        |
| Default position reversed                          | Same byte oracle                                        |
| Second suggestion omitted                          | Same byte oracle                                        |
| Wrap closing edit wrong                            | Same byte oracle                                        |
| Count-only suppression                             | Ordinary bytes agree first; count differs, 43 versus 33 |
| Suggestion applied as fix                          | Go byte oracle differs on Node and sanitized native     |
| Empty function body reported                       | Same byte oracle                                        |
| Duplicate case suppressed                          | Same byte oracle                                        |

Existing sixteen-mutant run PASS, 191.736 seconds; new dispatch mutants PASS,
23.91 seconds. External one-byte oracle mutant PASS (caught), 0.062 seconds.
`go vet ./...` exits zero with empty log. Getter probe PASS. The inherited nested
constructor compiler-gap test also passes. No full repository test gate was run;
validation covers the touched package, external oracle mutant and repository vet.
The inherited exact parser refusal remains a limitation, not a skipped success.
Scoped cohere check exits zero: no findings, no type errors, 100% Adamic-ready
(four checked program files). `gofmt` and `git diff --check` are clean.

## Profiling reproduction and setup

`bash cloud/setup.sh` succeeded; toolchain Go 1.27.1, clang 20.1.8, Node 24.19.0:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (15s)
setup: done in 15s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Perf was absent and `apt-get download linux-perf` failed because package lists
were empty. Debian linux-perf 6.12.107-1 and its five shared-library dependencies
were downloaded and extracted into scratch without changing system packages.
Hardware `cycles:u` reports unsupported on this virtual machine; software
CPU-clock works. Consequently no cycle, IPC or cache-miss conclusions are made.

Generate C with `adamic c stage1/cohere/lint/main.ts > profile.c`, then compile
with standard native release flags plus debug info and frame pointers:

```bash
clang -std=c11 -Wall -Wextra -Werror -pedantic \
  -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function \
  -Wno-unused-parameter -Wno-self-assign -ffp-contract=off \
  -fno-optimize-sibling-calls -O2 -g -fno-omit-frame-pointer \
  -I internal/native/runtime profile.c internal/native/runtime/*.c -lm -o profile
export LD_LIBRARY_PATH=/workspace/scratch/batch6-perf/perf-tools/usr/lib/x86_64-linux-gnu
python3 stage1/cohere/lint/batch6_perf.py \
  --perf /workspace/scratch/batch6-perf/perf-tools/usr/bin/perf \
  --binary /absolute/path/profile \
  --manifest /workspace/scratch/batch6-perf/compiler-20.txt \
  --output /absolute/path/profile-results
```

`compiler-20.txt` repeats the sorted compiler manifest twenty times. The script
records perf data, self/inclusive reports, stacks, sample-weight summaries and
binary/manifest hashes. Before profiling the optimized tree, baseline C and
all baseline binaries were preserved from the exact branch base. To reproduce
the baseline, use that base in a separate checkout with the same compiler flags.

Small profile tables, counters and summaries are committed under
[batch6_perf_evidence](batch6_perf_evidence/). Raw data and logs remain in scratch:
`baseline.perf.data` (6.279 MB), `optimized/perf.data` (4.981 MB),
`getter.perf.data`. Sampling and worker load introduce variance; throughput
includes file I/O, parsing and process startup. The improvement is measured for
this all-rule compiler workload, not every configuration or full cohere CLI.
