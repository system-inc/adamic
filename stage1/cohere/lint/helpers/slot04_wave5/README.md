# Slot 04 fifth batch

One named helper per Adamic file: breakpoint_bucket.a, comment.a, declaration.a. css_leaf.a defines the constructors' leaf view; main.a is the differential entry. Only slot-owned files and claims/04.md change.

Inputs are immutable byte strings represented by readonly arrays of integer bytes 0..255. This preserves arbitrary Go string bytes, including invalid UTF-8, without decoding. Consumers must keep the backing arrays immutable. Constructors allocate fresh mutable nodes, retain immutable property/value data, preserve empty values, and initialize every other leaf field to its Go zero value. contextPresent and nodesPresent are explicit absence flags for Go's nil map/slice; both are always false here. This leaf view is not a replacement for container-node representation. Stage 0 cannot lower the null fields, so no shared compiler change was made.

Run after sourcing /workspace/adamic-tools/env.sh:

```sh
python3 stage1/cohere/lint/helpers/slot04_wave5/testdata/capture.py > /tmp/slot04-wave5-capture.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave5/testdata/readiness.py > /tmp/slot04-wave5-readiness.log 2>&1
go test ./stage1/cohere/lint/helpers/slot04_wave5 -count=1 -v > /tmp/slot04-wave5-tests.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot04_wave5 > /tmp/slot04-wave5-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/slot04-wave5-oracle.log 2>&1
```

Go overlays call the pinned Cohere implementation. Source Node, sanitized native and emitted JavaScript compare exact output against it. Temporary overlays instrument upstream capture and helper calls without editing shared files. Capture installs tailwindcss@4.3.3 into a temporary directory and substitutes the two existing developer-path fixtures. The six-rule corpus is derived from readiness.json and checked with a consumer-omission mutant. All output is written directly to log files.
