# Slot 04 eighth batch

Three separate named helpers: walk.a, walk_nodes.a and write_value_css.a. tree_view.a contains only flat CSS/value data views; main.a is the differential entry. This package and claims/04.md are the only changed territory.

CSS nodes use stable arena indices as pointer identities, tag for Go Kind and edges for Go Nodes. The arena and edge lists must remain immutable during traversal, matching Go's documented contract. Valid indices identify non-nil nodes; aliasing and repeated roots preserve identity. Walk forwards the unchanged root list and visitor to walkNodes and discards its boolean result. walkNodes visits depth-first, stops globally on action 2, skips children on action 1, descends only on action 0, and ignores unknown actions without descending. Exactly rule, at-rule, context and at-root are containers. Empty/nil child lists are both represented by empty edges.

Value nodes use tag, data and edges. data is an immutable raw byte string represented as integer bytes 0..255, so malformed UTF-8 survives unchanged. writeValueCss appends to a caller-owned byte buffer, copies word/separator data verbatim, wraps recursive function children in parentheses, and ignores unknown kinds including their children. The trees must be finite and have valid arena indices. The buffer is a byte-view adapter for the observable append behavior of Go strings.Builder; copying-builder internals and nil builder pointers are outside this view.

The field names deliberately differ from the JSON reader's OptionValue fields. Using kind/children allowed structural overlap between that owned JSON arena and the CSS arena, and the compiler refused the first profile under its cycle-capable ownership check. The separate tag/edges/data view passes without weakening ownership or editing shared files.

After sourcing /workspace/adamic-tools/env.sh:

```sh
python3 stage1/cohere/lint/helpers/slot04_wave8/testdata/capture.py > /tmp/slot04-wave8-capture.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave8/testdata/readiness.py > /tmp/slot04-wave8-readiness.log 2>&1
go test ./stage1/cohere/lint/helpers/slot04_wave8 -count=1 -v > /tmp/slot04-wave8-tests-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot04_wave8 > /tmp/slot04-wave8-vet.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot04_wave8/helpers_test.go stage1/cohere/lint/helpers/slot04_wave8/testdata/oracle.go stage1/cohere/lint/helpers/slot04_wave8/testdata/exports.go > /tmp/slot04-wave8-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/slot04-wave8-oracle.log 2>&1
```

Temporary Go overlays invoke actual private/public methods and capture actual tree geometry, preserving CSS pointer identity. The source corpus uses actual Go ParseCSS and ParseValue, not an Adamic parser. Capture installs tailwindcss@4.3.3 into a temporary directory and substitutes two existing developer-path fixtures. Exact callback traces, stop results and byte outputs compare on real Go, source Node, sanitized native and emitted JavaScript. Every semantic mutant must compile and execute successfully before being credited as caught. See REPORT.md and evidence/ for exact counts, rules and limits.
