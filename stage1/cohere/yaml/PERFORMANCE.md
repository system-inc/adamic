# YAML formatter throughput

The speed follow-up measures native 11,504, source Node 13,819 and Go 15,190
texts/s. Native clears the original Node target; the optimized Node port still
takes less time. [SPEED.md](SPEED.md) contains profiles, proving programs,
optimizations and the complete final five-round comparison. This page retains
the original checkpoint for comparison.

Measured the complete file formatter at checkpoint `3803cbd`, not just its lexer
or scalar resolver. Five interleaved fresh-process rounds used the exact 10,026-case
correctness corpus, including all 36 repository/submodule YAML files. Every timed
native, source Node and Go run exited zero, wrote no stderr and reproduced all
992,282 saved Go answer bytes.

| Driver | Median texts per second | Median elapsed seconds |
| --- | ---: | ---: |
| Adamic native, release | 2,768.29 | 3.621730 |
| Same Adamic TypeScript source on Node | 7,300.26 | 1.373376 |
| Go cohere public file formatter | 14,894.45 | 0.673137 |

At the original checkpoint, native took 5.38 times Go's elapsed time and 2.64 times Node's.
This is an observed speed deficit; the port had not been profiled or optimized then.
Node here means the same port source through `oracle/node.mjs`, not published
Prettier. Published Prettier 3.9.6 is independently checked for correctness, with
42 proved upstream differences described in [GAPS.md](GAPS.md).

Timing includes process startup, file input, parsing, printing and protocol output.
The corpus has 4,241 successful formatting answers and 5,785 syntax/error answers;
parser chunk variants repeat some texts and count as separate formatting calls.
The escaped batch input is 1,411,737 bytes. These are this mixed edge-case corpus's
texts per second, not a representative production-workload claim. Timed stdout
and stderr go to scratch files; verification happens outside the timed interval.
No test jobs ran concurrently. Native release is unsanitized; correctness tests
separately run ASan, UBSan and leak detection. Memory use was not measured.

Reproduce after setting up the pinned scratch libraries as described in GAPS:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library ADAMIC_YAML_ARTIFACTS=/tmp/stage1-yaml-format-artifacts go test -v -count=1 -timeout=20m ./stage1/cohere/yaml -run 'TestFormatter|TestFileDriver|TestBundledParser' > /tmp/stage1-yaml-format-final.log 2>&1
go run ./cmd/adamic build stage1/cohere/yaml/main.ts -o /tmp/stage1-yaml-format-artifacts/native-format > /tmp/stage1-yaml-format-release-build.log 2>&1
python3 stage1/cohere/yaml/benchmark_lexer.py /tmp/stage1-yaml-format-artifacts /tmp/stage1-yaml-library --layer format --rounds 5 > /tmp/stage1-yaml-format-timing.log 2>&1
```

The test artifact directory saves the unmodified public Go formatter executable,
escaped input and expected answers. The benchmark reuses those exact artifacts.
Full samples and verification lines: [format-timing.log](audit/format-timing.log).

Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0, cohere
`715ba94f3608a6500086b1076ce5cb7e51b836db`. `nproc` prints 5; cgroup
`cpu.max` is `400000 100000` (four CPUs), memory is 17.6 GB.
`bash cloud/setup.sh` completed successfully and printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (78s)
setup: done in 78s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Earlier parser-only measurements remain in GAPS and their audit logs; they are
not full formatter throughput. No full repository gate was run: the full YAML
package passed in 384.721s, followed by the corrected formatter suite in 139.079s;
the uncached filtered compiler oracle passed in 2.421s. All 24 port mutants are
caught by successful-execution byte comparisons. The latest three are root
newline loss (first difference 28293), colon separation loss (1694) and broken
flow trailing comma loss (783469), each on both native and source Node.

Green implementation checkpoints, each pushed on `codex/stage1-yaml`:

| Commit | Checkpoint |
| --- | --- |
| `67dc261` | Lexer |
| `9d4a5d7` | CST parser |
| `87e92ca` | Scalar resolution |
| `8afa696` | Property resolution |
| `a77218f` | Schema patterns |
| `8f43a77` | Document composition |
| `066e3f9` | Printer tree and comments |
| `3803cbd` | Printer, layout and file driver |

All implementation is under `stage1/cohere/yaml/`; no compiler/runtime file changed.
