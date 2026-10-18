| Item | Current main versus source Node | Historical fixing branch (`compiler/`) |
|---|---|---|
| 132 | Yes, 8/17 probes differ; literal-key reads | fx6-key-read |
| 133 | Yes, 1/1 probes differ; named callable producer | fx7-callable-producers |
| 134 | Yes, 5/5 probes differ; unrelated same-name fields | fx6-valid-programs |
| 135 | Yes, 5/9 probes differ; assignable callable producers | fx7-callable-producers |
| 136 | Yes, 3/4 probes differ; destructured contracts | fx6-valid-programs |
| 137 | Yes, 1/2 probes differ; tuple length object view | fx6-valid-programs (narrow refusal only) |
| 138 | Yes, 3/3 probes differ; tuple object recognition | fx7-wrong-aborts |
| 139 | Yes, 3/3 probes differ; read diagnostic locations (located refusal required, rather than acceptance) | fx6-key-read routing; fx6-cleanup tests |
| 140 | Yes, 3/3 probes differ; checked failure type text | fx7-depth-texts |
| 141 | N/A, cleanup with no executable probe | fx6-cleanup |
| 142 | Yes, 3/4 probes differ; unsupported descriptor loses fields | fx7-unsupported-view |
| 143 | Yes, 2/3 probes differ; snapshot reference payload | fx7-snapshot |
| 144 | Yes, 4/5 probes differ; maybe number/boolean union reads | fx7-scalar-union-view |
| 145 | Yes, 2/2 probes differ; non-closure producer registry | fx7-callable-producers |
| 146 | Yes, 9/9 probes differ; checked cast/callable metadata | fx7-wrong-aborts |
| 147 | Yes, 2/2 probes differ; undefined writes to same-named fields | fx6-valid-programs; fx7-name-family coverage |
| 148 | Yes, 2/2 probes differ; plain undefined reads | fx6-valid-programs; fx7-name-family coverage |
| 149 | Yes, 2/2 probes differ; nullable array/Map writes | fx6-valid-programs; fx7-name-family coverage |
| 150 | Yes, 2/2 probes differ; runtime-created entry tuples | none in named branches |
| 151 | Yes, 5/5 probes differ; recursive membership depth | fx7-depth-texts |
| 152 | Yes, 3/3 probes differ; failure runtime-kind text | fx7-depth-texts |
| 153 | N/A, cleanup with no executable probe | fx6-cleanup |
| 154 | Yes, 4/5 probes differ; callable union result representation | fx7-callable-union-result |
