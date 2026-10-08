# Verify optional parity

Landing base: `ad7bd06632f119abc7680719ad3a7d3b71100f58`. Uses the area's `context.checker` and `raw-type` question. The descriptor subscribes to `PropertyDeclaration` and `Parameter` and visits the node handed to it.

All 22 upstream cases under `TestVerifyOptionalParity` match Go cohere byte for byte on Node, emitted JavaScript and sanitized native. The registered witness fires. The `Only root nullability accepted` mutant compiles and is caught by the wire comparison on all three runtimes. Messages are verbatim; upstream supplies no fixes or suggestions.

Fresh area certification: `/tmp/wave28-area-rule-certification.jsonl`; four tests/subtests pass. Across the 22 upstream projects, Go program creation and lint totaled 10.267932s; native program creation, lint and transcript recording totaled 13.586442s. These timings include checker startup and are not steady-state throughput.

Registry generation, `gofmt -l` on `oracle.go`, and `go vet ./stage1/cohere/lint` pass. Full package logs: `/tmp/wave28-area-held-full.jsonl` and `/tmp/wave28-area-landing-full.jsonl`. Exclusion details and checker reproducers: `/tmp/wave28-area-exclusions.json`.
