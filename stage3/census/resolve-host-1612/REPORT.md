# Resolve host 1612 proof checkpoint

| Exception | Path counter | Citations |
| --- | --- | --- |
| getDirectories | 0 | moduleNameResolver.ts:820,825; getAutomaticTypeDirectiveNames is outside the resolver closure |
| trace | 0 | moduleNameResolver.ts:1755,124; traceResolution is absent |
| getCurrentDirectory | 0 | moduleNameResolver.ts:1750-1764,1897,3017,3023; successful Node10 JS resolutions and the zero-extension retry cannot reach it |

All three have separate direct forwarder checks for exact arguments and nodeSystem receiver. No other exceptions. The lane never exercises this path, so its path counts are vacuous.

Implementation: `635baacb5dfa6c4d61aa6fb9f26d720552b1520e`, one commit on ad8ca5a3. No Loom submission.

Final independent checks pass: full sys.ts census 184 -> 183, exactly 1612:77 gone, new 0, changed 0; restoring mutant exact original bytes and all 184 findings restored; stock TypeScript 0 diagnostics on 79 roots; adamic types exit 0; 128 presence cases; built Node output byte-identical in both optional cases; readFile receiver mutant red; JavaScript scope only at this site across 604 files/1208 comparisons and the actual bundle; hereby dts public declarations byte-identical.

The fresh whole unfiltered lane is RUNNING at this checkpoint. Its full apply/test receipts will be appended without force.
