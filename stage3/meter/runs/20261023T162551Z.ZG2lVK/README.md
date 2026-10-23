# Stage 3 meter 20261023T162551Z

The area label repeats the pinned main tree for reporter compatibility.

| Run | Main SHA | Whole program | Own file | Lowering | tsc entry diagnostics |
| --- | --- | ---: | ---: | ---: | ---: |
| Current | e50e502c21f12f9df64f7e716e980ed84695aa6f | 2/79 | 58/79 | 2/79 | 144 |
| 20261020T014828Z.POASDN | 8c318e4528bffb67019116408a819c06e83df9fa | 2/79 | 57/79 | 2/79 | 175 |

Lowering counts ordinary successful IR lowering. No native output is claimed.

Elapsed: 22m 33s.

## Oct 22 comparison

| Main record | Whole program | Own file | tsc entry diagnostics |
| --- | ---: | ---: | ---: |
| Oct 22, 1293af7c (the supplied record) | 2/79 | 58/79 | 144 |
| Oct 23, e50e502c21f12f9df64f7e716e980ed84695aa6f | 2/79 | 58/79 | 144 |

The requested counts are unchanged. Both independent copies agree.
From main/report.md's latent table, `an object refinement using an open numeric enum as a literal tag`
has **NotYet 0, Refused 502**. Main's latent totals are NotYet 929, Refused 3403,
SkippedDependency 2, error 0, panic 2. Both retained panic reasons also occur in
the Oct 20 committed stream: a nil-pointer dereference in moduleSpecifiers.ts and
Unhandled case in Node.Text for ast.ComputedPropertyName in declarations/diagnostics.ts.
These are retained measurement findings, not an overlay build failure or native output.
