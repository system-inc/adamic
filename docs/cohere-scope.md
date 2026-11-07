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
| `stage1/cohere/tsprinter/results/**` | record | tsprinter's committed results. | The tasks that cite them. |
| `stage1/cohere/lint/validation/**` | record | lint's committed validation records. | The tasks that cite them. |
| `stage1/cohere/lint/performance/**` | record | lint's committed performance records. | The tasks that cite them. |
| `stage1/cohere/css/performance/**` | record | css's committed performance records. | The tasks that cite them. |
| `stage1/cohere/typeaware/validation-volume*/**` | record | typeaware's committed validation-volume records. | The tasks that cite them. |
| `stage1/cohere/markdownblocks/mutant_evidence/**` | record | markdownblocks' mutant evidence. | The tasks that cite it. |
| `stage1/cohere/markdownblocks/fixture_speed_evidence/**` | record | markdownblocks' fixture speed evidence. | The tasks that cite it. |
| `stage1/cohere/lint/inventory/inventory.json` | generated | lint's rule inventory, and also the json port's largest corpus case, pinned at a cohere commit, so its bytes are an input. | Written by stage1/cohere/lint/inventory/main.go, and read by stage1/cohere/json's corpus tests. |
| `stage1/cohere/lint/.generated/**` | generated | lint's rule registry, registry.ts and registry.go. lint's TestMain refuses any byte registry.Stale doesn't share with the generator's output, so they are never formatted by hand. Revisited when the generator emits cohere's format. | Written by stage1/cohere/lint/registry's Generate (go run ./cmd/lint-registry), and checked byte for byte by registry.Stale. |
| `stage1/cohere/lint/rules/*/rule.json` | input | Rule data whose bytes registry's committed `.generated` hashes and some product keys read. Revisited when those readers stop depending on bytes. | stage1/cohere/lint's registry generator and its product keys. |
| `stage1/cohere/lint/rules/*/mutant.json` | input | Rule mutants whose bytes the same hashes and keys read. | stage1/cohere/lint's mutant tests. |
| `stage1/**/*.md` | input | For this pass: markdownblocks' preflight parses its own GAPS.md as a witness, and the markdown packages' corpora read .md files, so each stage1 doc needs a per-file read before it's formatted. | stage1/cohere/markdownblocks and markdowninline's tests. |

## Out of tsconfig.json's plain type check

Three trees are left out of the root tsconfig, because the plain checker can't judge them honestly. Each is still
checked by the test that reads it.

The first road for the first two was a tsconfig of their own that includes `internal/load/prelude.d.ts` and the
bundled Node types. It isn't open yet: cohere has no types-only scope (`ignorePatterns` cuts checking and formatting
together, and overrides have no all-off), so a nested project there would also be linted and formatted, and these
are inputs that never are. When cohere's #cxeq6jw lands, an ignore entry with `check: types` type-checks such a tree
through its own tsconfig without linting or formatting it, and both come back under cohere's types that way.

| Tree | Why | Checked by |
| --- | --- | --- |
| `internal/oracle/testdata` | Adamic programs against the host adamic's loader serves: the prelude plus `internal/load/node_types` (@types/node 25.3.3, with `host_axioms.json`'s corrections). The plain checker sees neither, so its 458 diagnostics were configuration, not code. | internal/oracle's load-and-check test, until cohere can judge a tree by types alone (#cxeq6jw) |
| `internal/load/testdata/node_activation` | Fixtures that prove the loader switches on the bundled Node types for a builtin `require` or an import type. The plain checker doesn't. | internal/load's node activation tests, until #cxeq6jw's types-only scope |
| `stage1/**/gaps` | stage1's gap cases, each written to be wrong in exactly one way: CRLF, a line separator, empty JSX type arguments. | stage1's parity tests |
