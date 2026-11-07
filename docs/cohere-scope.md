# What cohere judges in adamic

A bare `cohere --no-fix` at adamic's root judges adamic's hand-written source: its Go, its TypeScript and its `.a`. It
leaves alone what adamic's tests and records read byte for byte. Those trees are inputs, not code under judgment, and
reformatting one moves an expected output, a count or a key (Adamic's ruling on #h2c34ad, Nov 7, with cohere's
conditions).

Each glob below sits in `CohereSettings.json`'s `ignorePatterns`, and each one has its row here: what kind it is, why
it stays out, and what reads it. `cloud/cohere_scope_test.go` fails when a glob is in one place and not the other, so
this table can't fall behind the settings. When cohere's ignore entries can carry their own reason and kind (#cxeq6jw),
the table moves into the settings in one change, and that test proves the move lost nothing.

Never add a glob here to clear a finding in hand-written source. An ignore names a tree of inputs or a generated file,
scoped to its directory or its file, never a bare extension.

| Glob | Kind | Why it stays out | Read by |
| --- | --- | --- | --- |
| `**/testdata/**` | input | Test inputs, read by bytes: oracle fixtures, counts, refused programs, analyzer inputs and golden files. Reformatting one moves the output it pins. | Each package's own tests. internal/oracle/testdata is type-checked by internal/oracle's load-and-check test against internal/load/prelude.d.ts and internal/load/node_types, which is its gate until cohere can judge a tree by types alone (#cxeq6jw). |
| `**/fixtures/**` | input | Fixture programs and their expected outputs. | stage3/fixtures' tests, stage3/interface-downcasts' lanes, and the oracle. |
| `**/corpus/**` | input | Third-party corpora, kept byte for byte: test262 cases and tsc's own test corpus, some of which fail to parse on purpose. | cmd/adamic-test262 and stage3/drivers. |
| `**/evidence/**` | record | Committed evidence: logs, measurements and the inputs a ruling was proven on. | The tasks and reviews that cite them. |
| `**/proof/**` | record | Committed proofs, the same as evidence. | The tasks and reviews that cite them. |
| `stage1/**/gaps/**` | input | stage1's gap cases, written to be wrong in exactly one way (CRLF, a line separator, empty JSX type arguments). | stage1's parity tests. |
| `review/**` | record | Review records: logs, JSON and the witnesses a review read. | The reviews that cite them. |
| `stage3/meter/runs/**` | generated | The daily stage 3 meter's records. | Written by cloud/meter-daily/run.sh and stage3/meter/twice-daily.sh, and read by the meter's reports. |
| `verify/catalog/**/*.{json,jsonl}` | record | The regression catalog's recorded detections. | verify/catalog/check.sh. |
| `cloud/reports/**/*.{json,jsonl}` | record | Committed measurement reports. | The tasks that cite them. |
| `docs/verification/**/*.{json,jsonl}` | record | Committed verification records. | The documents beside them. |
| `stage3/lane/shards.json` | generated | The lane's unit manifest. | Written by stage3/lane/generate_shards.py, and checked by its `--check`. |
| `stage3/lane/shard-plan.json` | generated | The lane's shard plan. | Written by stage3/lane/shard_plan.py, and read by generate_shards.py. |
| `stage3/patch-set.md` | generated | The adaptations' patch table. | Written by `stage3/apply.sh <out> --write-table`, and read by stage3/adapt/75-optional-widening/coverage/run.py. |

## Out of tsconfig.json's plain type check

Three trees are left out of the root tsconfig, because the plain checker can't judge them honestly. Each is still
checked by the test that reads it.

| Tree | Why | Checked by |
| --- | --- | --- |
| `internal/oracle/testdata` | Adamic programs against the host adamic's loader serves: the prelude plus `internal/load/node_types` (@types/node 25.3.3, with `host_axioms.json`'s corrections). The plain checker sees neither, so its 458 diagnostics were configuration, not code. | internal/oracle's load-and-check test, until cohere can judge a tree by types alone (#cxeq6jw) |
| `internal/load/testdata/node_activation` | Fixtures that prove the loader switches on the bundled Node types for a builtin `require` or an import type. The plain checker doesn't. | internal/load's node activation tests |
| `stage1/**/gaps` | stage1's gap cases, each written to be wrong in exactly one way: CRLF, a line separator, empty JSX type arguments. | stage1's parity tests |
