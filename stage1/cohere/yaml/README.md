# YAML formatting slice

Build and format a file to stdout:

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic build stage1/cohere/yaml/main.ts -o /tmp/adamic-yaml
/tmp/adamic-yaml input.yaml
```

For source Node, use `node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/cohere/yaml/main.ts input.yaml`.
The parser and printer use Go cohere's default file formatting options. See
[GAPS.md](GAPS.md) for exact coverage, proving programs, mutants and validation.

[PERFORMANCE.md](PERFORMANCE.md) records the original full formatter throughput.
[SPEED.md](SPEED.md) records profiles, optimizations and the final before/after comparison.
