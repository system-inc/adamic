JSON landing onto main e011f8f60899586d6373a5ccb07335ad82cfbf3c.

The original codex/library-json-2 remains at 7e4bc3b4a37dd1eb876b3c93aa1d6c6f7b054c43.
Only its four JSON unit commits were replayed; the predecessor JSON/Number branch is handled by seat 1.
Main's newer RegExp freshness analysis was retained alongside JSON callback handling.
Immediate Number boxing in the JSON lowering now calls main's existing proven primitive conversion
without relying on helpers from the predecessor branch. Number and Math files are unchanged.
The unreachable duplicate-literal check was removed while preserving the new toJSON proofs.
The counts conflict was resolved by regeneration; only the five JSON fixture rows were added.

Validation commands (all test output went directly to log files):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m \
  ./internal/lower ./internal/native ./internal/fresh ./internal/flow \
  ./internal/javascript ./internal/ir > /tmp/library-json-2-land-packages-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle \
  -run 'TestJSONStringify|TestJSON(ParseRefusalBoundaries|RuntimeNodeComparison)|TestNativeAgreesWithNode/internal/oracle/testdata/(library_json_|json_stringify_)|TestCountsAreRecorded' \
  -args -update-counts > /tmp/library-json-2-land-oracle-final.log 2>&1
go vet ./... > /tmp/library-json-2-land-vet-final.log 2>&1
```

The first complete uncached oracle passed in 62.801s with 52 native and 33 Node cache misses.
It includes source-Node comparisons, release and sanitizer observations, refusal boundaries,
the Node-only key-order mutant, and the complete counts regeneration.
Vet passed with no output. The full repository test suite was not run.
The initial landing build failed because the predecessor's two Number slot helpers were not on main;
the JSON-local conversion above resolved that dependency before the successful reruns.

Main also added a runtime field-layout gate. The first package run caught the missing
private adamicJSONParsed carrier layout, before registration in internal/native/fields.go.
That omission is a real failing proof probe: TestRuntimeFieldLayoutsAreIncluded reported
"runtime field adamicJSONParsed at 0 is absent or conflicting in the layout proof".
The layout registration preserves checked lookup whenever another shape uses a different offset.

The final uncached JSON oracle passed in 38.906s; all 52 native and 33 Node observations
were cache misses again. The final vet passed with no output.

Final package gate passed: lower 23.551s, native 115.236s, fresh 42.984s, flow 82.492s;
javascript and ir have no package test files. gofmt -l cmd internal returned no files.
