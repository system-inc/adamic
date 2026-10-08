# GraphQL printer throughput

Observed on base main 5d4c8012, cohere 715ba94f, five visible processors with a
four-CPU cgroup quota, Go 1.27.1, clang 20.1.8 and Node 24.19.0. Setup took 73s,
including 73s warming the build cache. Native was compiled without sanitizers for
this measurement; correctness uses sanitized builds separately.

The benchmark selects the 2,513 successfully formatted corpus texts. All four
sides parse and print, read the same escaped input file and write the same escaped
answers to stdout. The Go executable uses cohere's actual native.Formatter file
wrapper. Node runs the Adamic source through oracle/node.mjs. Prettier is npm
3.9.6, the version cohere embeds. Compilation is excluded; process startup, input
and output are included. There is one untimed warm-up and five measured rounds,
rotating which side runs first. Every warm-up and measured output is checked byte
for byte against Go before its time is accepted. Refusals are not counted as
formatted texts.

| Side | Best seconds | Texts per second | Five measured seconds |
|---|---:|---:|---|
| Native | 0.514950 | 4,880.1 | 0.563946, 0.553598, 0.529672, 0.578816, 0.514950 |
| Node source | 0.399160 | 6,295.7 | 0.450006, 0.399160, 0.430377, 0.452160, 0.461600 |
| Go cohere | 0.260354 | 9,652.2 | 0.276892, 0.271497, 0.273652, 0.283923, 0.260354 |
| Prettier 3.9.6 | 1.139676 | 2,205.0 | 1.210696, 1.175393, 1.196615, 1.196949, 1.139676 |

Native is 0.78x Node, 0.51x Go and 2.21x Prettier in this batch. These are observed
whole-process times, not parser-only or printer-only timings. No profiler was run
in this unit, so these results do not attribute the difference to allocations,
reference counting or any particular compiler operation. The embedded fork is
checked for bytes separately; its speed was not measured.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GRAPHQL_PRETTIER=/tmp/graphql-printer-prettier ADAMIC_GRAPHQL_PRINTER_BENCH=1 go test -v -count=1 ./stage1/cohere/graphql/printer -run TestPrinterThroughput > /tmp/graphql-printer-performance.log 2>&1
```

The recorded measurements are in results/printer-tests.log. New measurements
should retain every timing and output check rather than comparing unchecked work.
