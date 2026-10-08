# First slice verification

Base 1f9e223d; cohere 7945d102a6c18dd36adf9114a758ce646e8b2359.
Bootstrap: GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh, completed successfully.

Command: `go test -v -count=1 -timeout 20m ./stage1/cohere/high_level_intermediate_representation`

Result:

```text
=== RUN   TestStraightLineOracle
12 functions match Go on Node and natively
return-store-to-nil mutant caught on Node and natively
--- PASS: TestStraightLineOracle (17.84s)
PASS
```

Seven corpus functions plus five probes; provenance in testdata/manifest.json. Native output and Node output each matched the tagged Go adapter byte for byte. Mutants executed successfully but differed from the oracle. The Go adapter was installed only through an overlay; cohere remained clean. The native build compiles the parser and imported generic SSA implementation with this HIR adapter. This is slice coverage, not whole-package or full-rule conformance.
