Built: labelsOf in its own .a file, with supplied parent facts and numeric labeled-statement kind 257.
Commits: claim 3658a9863b9387e939518269050e75b600e2f81e; preceding truthiness delivery d6c77911d0cea8f28c4d3d52badc447ef065157d.
Checks: actual Go, source Node, emitted JavaScript and ASan/UBSan native agree on 30,228 observations and 64,096 bytes; focused suite PASS 52.153s.
Mutants: omit-outer-labels row 29,647, ignore-child-identity row 24,218 and wrong-kind row 24,219, all compiling and exiting cleanly on each port backend before Go comparison catches them.
Uncovered: four CFG consumers lose one prerequisite each, zero final blockers; full CFG/findings/fix integration and repository-wide gating remain uncovered.

API: labelsOf(node: LabelsNode): string[]. LabelsNode supplies presence, numeric kind, label text, a Statement/current-child identity bit and a zero-or-one parent array. Each successive identity bit comes from actual pointer equality on the Go AST, not the expected labels result. The helper must independently check the numeric kind and identity bit, collect labels and advance through parents. Parent facts include the first non-label parent; a missing/nested non-label boundary cannot silently turn into a label. No parser lookup or string-kind relevance test occurs in Adamic.

Go expectations call the actual private helper through an oracle-only export overlay, leaving cohere unchanged. The adapter parses all 2,119 captured fixture sources and visits every node, preserving file extension/script kind. It adds 42 parsed control sources combining six label chains and seven statement bodies, including nested loops, switch, blocks, if, try/catch and inner labels, Unicode/BMP/supplementary identifiers, repeated labels and eight-level nesting. Duplicate/semantically invalid labels are parser controls; no claim is made that they pass TypeScript semantic diagnostics. Output records label count and every label's UTF-16 code units, in order. Complete successful observation bytes match on all four runtimes. Go nil slices and empty Adamic arrays are equal for these observation operations; Go slice nilness/capacity introspection is outside the API.

The actual Go helper panics on nil node before reading Parent. The port also refuses nil. A separate test checks all four runtimes for nonzero exit, diagnostics and no successful stdout. Stack traces and panic text are backend-specific; they are not claimed identical. Unicode source labels are decoded valid strings; invalid UTF-8 is not covered.

Selection refreshed 575 origin branches and read every one of 20 distinct helper claim blobs. This symbol tied the highest remaining unclaimed fan-out, four. The claim was pushed before code. Both own branches contain main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 and the sibling rule branch remains parked with shared blockers named. Shared finding/context/main/generator/oracle/comparison files are untouched. No push to main or area branches occurs.

The same four consumers are array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Together truthiness and labels remove two more dependencies from each, zero final blockers. The owned control_flow_readiness.json retains all remaining symbols. This is helper contract parity over every consumer fixture AST, not full native CFG or rule parity.

Commands after sourcing /workspace/adamic-tools/env.sh:

```sh
go test ./stage1/cohere/lint/helpers/wave12 -run '^TestLabels' -count=1 -v > /tmp/wave12-labels-oracle.log 2>&1
go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-labels-full.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-labels-vet.log 2>&1
```

Complete owned suite PASS 105.859s; vet clean. All seven delivered helper contracts and 20 existing/new negative controls pass.
