# Step 32 local witness registry

All witnesses run on source Node; native builds intentionally stop at checking. No native allocation counts are available.

| Stop | Program | Node exit | Stdout mutant killed | Native build exit |
|---|---|---:|---|---:|
| 1 | probes/01-indexed-path.a | 0 | Yes | 1 |
| 2 | probes/02-tuple-parameter.a | 0 | Yes | 1 |
| 3 | probes/03-present-undefined-field.a | 0 | Yes | 1 |
| 4 | probes/04-optional-path-argument.a | 0 | Yes | 1 |
| 5 | probes/05-optional-path-argument.a | 0 | Yes | 1 |
| 6 | probes/06-optional-symbol-argument.a | 0 | Yes | 1 |
| 7 | probes/07-optional-symbol-argument.a | 0 | Yes | 1 |
| 8 | probes/08-optional-member-read.a | 0 | Yes | 1 |
| 9 | probes/09-never-argument.a | 0 | Yes | 1 |
| 10 | probes/10-optional-symbol-argument.a | 0 | Yes | 1 |
| 11 | probes/11-optional-symbol-array.a | 0 | Yes | 1 |
| 12 | probes/12-present-undefined-argument.a | 0 | Yes | 1 |
| 13 | probes/13-optional-file-read.a | 0 | Yes | 1 |
| 14 | probes/14-void-result.a | 0 | Yes | 1 |
| 15 | probes/15-optional-declaration-argument.a | 0 | Yes | 1 |
| 16 | probes/step32-symbol-array.a | 0 | Yes | 1 |
