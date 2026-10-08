# Step 41 (#8kytk6j): flat parser snapshots, with native adoption blocked

The first piece is an isolated **Go reference encoder and mapped reader**, plus
an independent Node typed-array reader. It is real executable format tooling,
not an Adamic port or a replacement for the parser. Native typed-array support
is absent; no ordinary-array, text-file, C bridge, or GC workaround was added.
Shared parser, lint harness, bridge, runtime and compiler files are untouched.

## Five headline findings

1. All 143 main-corpus files round-trip every current node-table field and expression root byte-identically: 917,981 table nodes, 922,041 child indices, 50,128,226 transport bytes.
2. Go whole-tree canonical bytes agree on 142/143 files; Babel exposes an existing clean-source parser defect, reduced to `let a=({b:c=>c});`. The mismatch is pinned, tested, and never counted as agreement.
3. A native walk of TypeScript `parser.ts` adds 101,529 retains and releases for 50,764 visits, with zero additional allocations. The required zero-retain critical path is not achieved.
4. The reference format uses 48 column bytes per node plus edges and strings, averaging 55.214 total bytes per table node; Node column walks were about 2.7 times faster on the measured file.
5. `new Uint32Array(1)` prints `0` on Node but is `lower.NotYet` in stage 0. Native encoding, mapping and printer adoption stop at the shared-file boundary.

## Base, reading and ownership

Branch base: `origin/area/stage1-lint`, `9156bf5c579a44d687c9955d13e44f9ad8bbb6f8`.
Oracle cohere: `7945d102a6c18dd36adf9114a758ce646e8b2359`.
Its TypeScript submodule: `d92d9bfee114c80be2c375d72edae966176e3a4f`.
Compiler corpus: TypeScript 6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.

Read README.md, CLAUDE.md, docs/0.1.md, docs/memory.md, docs/utf16-views.md,
docs/stage1-progress.md, parser GAPS.md, lint README.md/GAPS.md/PERFORMANCE.md,
lint helpers README.md, tsprinter README.md/GAPS.md and regex corpus notes.
There is no stage1/README.md or parser/README.md on this base. Historical
performance and coverage statements are distinguished from fresh observations.

## Where it bites, with source evidence

| Area | Evidence at this base | Consequence |
| --- | --- | --- |
| Current node representation | stage1/typescript/parser/nodes.ts:4, :8, :18 | Thirteen fields, including a separately owned mutable child array. `readonly children` freezes the field binding, not array contents. |
| Construction and access | stage1/typescript/parser/parser.ts:46, :244, :247 | Indexed table, object-returning accessor, object per make. Speculative table entries need not be reachable from the file root. |
| Canonical printer and count walk | stage1/typescript/parser/nodes.ts:40, :63 | Object lookup followed by numeric child traversal. Strings are materialized only for canonical output. |
| Port lint ancestry and dispatch | stage1/cohere/lint/lint.ts:115, :121, :124 | Separate ancestry and listener walks; parent indices avoid owning cycles. |
| Go lint dispatch | cohere/internal/lint/linter/linter.go:43, :79 | Combines listeners by numeric ast.Kind; walks each node with ForEachChild, following pointers. |
| Port printer reads | stage1/cohere/tsprinter/expressions.ts:48, :57 | Indexed access returns ParseNode; child helpers unwrap parentheses. |
| Port printer writes | stage1/cohere/tsprinter/expressions.ts:79, :93, :106 | Normalization rewrites child indices and appends a node during logical-tree rotations. A mapped immutable base alone cannot implement this. |
| Go AST base and lists | cohere/TypeScript/tsc/internal/ast/ast.go:124, :180 | Node has numeric Kind/Flags, TextRange, atomic id, Parent pointer, and interface-valued concrete data; lists contain pointers. |
| Go allocation | cohere/TypeScript/tsc/internal/ast/ast_generated.go:20, :627; internal/core/arena.go:13 | Per-concrete-kind arenas, not a malloc for every node. Identifier embeds the node base. |
| Go parent links | cohere/TypeScript/tsc/internal/parser/parser.go:5994, :6006 | finishNode sets context flags and installs strong parent links in immediate children. |
| Go formatting | cohere/internal/format/estree/convert.go:97, :129; internal/format/javascript/print.go:92 | Converts the pointer AST to estree and prints documents. Flat storage does not itself replace either conversion or document layout. |
| False arrow classification | stage1/typescript/parser/parser.ts:1886, :1907; statements.ts:258 | Variable initializer containing a parenthesized object with an arrow property produces false diagnostics. Speculative parameter parsing can accept an interior arrow; definitive repair belongs to the parser owner. |
| Existing retain evidence | stage1/cohere/lint/PERFORMANCE.md:191 | Historical generated-C evidence of owned node returns; fresh count measurements below confirm remaining walk traffic. |

Textual `.node(` call-site census over production `.ts` files, excluding gaps,
testdata and generated registries: lint **283 sites / 42 files**, tsprinter
**171 sites / 1 file**, parser **84 sites / 3 files**. This is a lexical census,
not a static count of calls executed. Direct `nodes[index]` accesses are additional.

## What a parsed file costs today

The measured file is TypeScript 6.0.3 `src/compiler/parser.ts`, 539,685 source
bytes, 50,808 table nodes and 50,764 reachable visits. Forty-four entries are
not reached by the counted root walk. No compaction renumbers them.

| Observation | Current representation | Flat reference |
| --- | ---: | ---: |
| Native structural sizes | ParseNode object 144 bytes; child-array header 56 bytes; table/child slots 8 bytes each | 48 bytes of columns/node, 4 bytes/child, 8 bytes/string range, 2 bytes/UTF-16 unit |
| V8 retained heap delta after GC | 12,887,648 bytes, 253.654 bytes/table node | Binary 2,766,288 bytes, 54.446 bytes/table node |
| Go AST sizes, amd64 | Node base 48 bytes; Identifier 72 bytes; NodeList 32 bytes | Go reference Reader stores a byte slice and scalar offsets |
| Go parse allocations / allocated bytes | 2,518 / 5,463,896 | Encoder allocation costs are not a parse replacement benchmark |
| Go heap delta after GC | 4,332,264 bytes | Mapping does not deserialize nodes or strings |
| Go pointer walk allocations | 0 | Go flat walk: 0 B/op, 0 allocs/op in auxiliary benchmark |
| Native parse-only process, counted driver | 191,072 allocations/frees; 2,456,851 retains; 2,237,027 releases | Native flat path blocked |
| One additional native tree walk | 0 allocations; 101,529 retains and releases | Zero retains is the target, not a measured native result |

Native structural sizes come from a compiled sizeof probe, not guessed JS object
sizes. A node plus its empty child-array header is already 200 bytes; add table
slots, child capacity, strings, allocator rounding and parser temporaries. The
parse-process counts include scanner, source, arguments and output; they are not
exclusive node-construction counts or total allocated-byte measurements. V8 GC
heap deltas are approximate retained-memory measurements, not object-size facts.

For the shortest `x;`, four visits add exactly nine retains and releases. On the
large file, ten walks add 1,015,290 of each. These match `2 * visits + 1` per
invocation in both probes. Counted ASan/UBSan builds have equal allocations and
frees, with no sanitizer/leak diagnostics. Go has GC but no reference counts;
Node has GC and no comparable retain counter.

## Proposed layout and the executable reference

`flat.go` writes version-1 little-endian bytes. The 40-byte header has eight magic
bytes, file size, node count, edge count, expression-root count, string count,
UTF-16-unit count, file root, and a reserved zero word. Twelve separate uint32
columns follow, in this exact order:

| Column | Meaning |
| --- | --- |
| 0 | kind string-table ID |
| 1 / 2 | UTF-16 pos / end |
| 3 | optional-chain bit 5; trailing-list bit 30; multiline bit 29 |
| 4 | literal flags |
| 5 | signed int32 list count, -1 means absent |
| 6 / 7 | first child / child count |
| 8 / 9 | cooked text / raw text string-table IDs |
| 10 / 11 | operator / semantic string-table IDs |

Following the columns: uint32 child indices in existing visitor order, uint32
expression roots, string ranges as (first UTF-16 unit, unit count), then UTF-16LE
units. Empty string is ID zero. Interning order is deterministic by node-table
order and field order. Every node, including speculative leftovers, is preserved.
The format formula is `40 + 48*N + 4*E + 4*R + 8*S + 2*U` bytes. Pooling includes
kind/operator/semantic names; IDs are snapshot-local, not assumed to equal Go
ast.Kind. A stable kind dictionary and numeric dispatch would be a later format
revision. The reference freezes no external AST ABI.

Reader.Open validates magic/version, size arithmetic, roots, positions, flags,
list sentinel, string IDs/ranges, edge ranges and cycles. It rebuilds no tree or
string table. Cycle validation uses one temporary byte per node; **opening is
not zero-allocation**, and validation is O(nodes + edges), not constant time.
Scalar and child access return values, not ParseNode objects. StringUnits is a
view into the backing bytes. Unix Map owns a read-only MAP_PRIVATE mapping; its
reader and views must finish before Close. Linux execution was tested; Darwin
uses the same syscall interface but has not been executed here. Files must stay
immutable and untruncated for the owner's lifetime. Concurrent truncation can
SIGBUS a mapping; this reference is not an untrusted-file service.

Node reference.mjs constructs typed-array views over one fs.readFileSync buffer;
that read allocates/copies the file. It does **zero node deserialization**, but
is not a Node mmap implementation. Uint32Array views require a little-endian
host and aligned base. Go accesses little-endian words explicitly. Restoration
and printing intentionally allocate strings and a snapshot in the test layer;
they are excluded from walk measurements.

Adamic-side design proposal: unique encoder construction, then a sealed immutable
owner whose read-only column/string views are borrowed for a complete walk.
Typed arrays, ArrayBuffer and DataView are ordinary TypeScript syntax. There are
no invented ownership annotations, unchecked casts, syntax extensions or GC.
Construction buffers must not remain mutable aliases after sealing. The printer
would need a separately owned normalization overlay keyed by base index, with
appended indices in a disjoint range; mapped columns stay immutable. **Neither
the typed-array ownership rules nor the overlay representation was ruled in the
first scout commit; the received rulings below now govern both.**

The format stores exactly the present ParseNode fields. It does not invent Go
context/binder flags, SourceFile metadata, JSDoc attachments, symbol links, source
bytes or diagnostics omitted by the tree transport. Production caches also need
source identity, parser/format version, script mode, diagnostics and policy
metadata. No zero-deserialization claim is made for a complete lint/checker cache.

## Original questions for @system_adamic, now ruled below

1. Will stage 0 support ordinary Uint32Array/Uint16Array/ArrayBuffer/DataView, including typed-array conversion, alignment, bounds and signed list sentinels? The exact smallest probe is gaps/typed_arrays.ts; Node prints 0, stage 0 returns lower.NotYet, What `new an Identifier`, at 1:15.
2. What approved immutable TypeScript API can seal a backing buffer and borrow typed-array/string views without retaining on each node read? Readonly<Uint32Array> still exposes mutating methods; is a checked read-only view/library type needed? No invented syntax or erased interface dispatch is proposed.
3. What binary-file and read-only mapping API belongs in 'adamic', and who owns unmapping when borrowed views escape, throw, or run concurrently? How should Node preserve observable behavior while native maps? Should mutable/truncated input files be refused or copied, and under which explicit contract?
4. How should pooled UTF-16LE strings become borrowed native strings without decoding, copying, per-read retains or losing lone surrogates? Current native strings use WTF-8/UTF-8 and cached unit views. Decoding once would violate this step's stated zero-deserialization goal.
5. Is restricting mapped word views to little-endian/aligned hosts acceptable, or is a standard scalar DataView access path required? No silent byte-swapping copy is proposed.
6. How is a persistent printer overlay owned and shared under immutable-by-default rules, including appended indices and aliasing during normalization? Existing mutation must be represented explicitly, not performed against mapped bytes.
7. Can the existing type/lifetime proof establish a borrowed tree owner across recursive listeners and printer callbacks, including early exit and exceptions, so measured walk retains reach zero? This is a compiler proof question, not permission to remove retains manually.

These were open questions at the first scout commit. The follow-up rulings below
supersede their open status. Serialization field order and version remain
reference-tool choices, not an approved product ABI.

## Shared files needed, and where this scout stops

- internal/lower/object.go:1336: typed-array construction currently reaches the generic new-expression NotYet path. Related IR/native value and array emission would need approved typed-view support.
- internal/load/prelude.d.ts:23, internal/native/runtime/input.c, internal/native/runtime/adamic.h, oracle/adamic.mjs: an approved binary mapping/owner/view API and matching Node behavior are absent. docs/0.1.md would need the ruling recorded by its owner.
- stage1/typescript/parser/parser.ts:1886: fix the false arrow-head/speculation behavior in the reduced clean-source witness. No parser workaround is added to the encoder.
- stage1/cohere/lint/lint.ts:115 and stage1/cohere/tsprinter/expressions.ts:79: consumers need an index/view API and a designed normalization overlay after the rulings and parser repair. Their owners must make those changes.

No lint-listener migration, finding/fix comparison through a flat-backed linter,
formatter comparison through a flat-backed printer, native mmap reader, or native
flat-walk speedup is claimed. This package's oracle is exact current-tree storage
plus Go canonical-tree comparison with explicit parser boundaries. This is the
boundary-limited first piece, not completion of the native port requested in the
brief.

## Fixture list and exact coverage

Kirk clarified that the quiet hundred lives on his Mac. GitHub was reachable.
All 23 supplied commits were shallow-fetched into scratch with `--depth=1
--filter=blob:none`, without checkout, install, package scripts or source execution.
One tracked .ts source was extracted from each. The unversioned TypeScript pin
has the native repository layout; its sample comes from packages/, not src/compiler.
Fixture provenance, commit, source path and SHA256 live in testdata/public/sources.json.
This is a public sample from the supplied pins, **not the full quiet hundred**.
The selection is lexical first eligible source, excluding declaration/test/fixture
paths; TypeScript 6.0.3 selects src/compiler. It is not a claim about repository
coverage or diagnostic-free checking of whole projects.

| Family | Count | Fixture list / coverage |
| --- | ---: | --- |
| Pinned compiler | 77 | Every .ts under TypeScript 6.0.3 src/compiler, checked commit, sorted paths |
| Supplied public pins | 23 | Table below, complete source snapshots and SHA256 manifest |
| Go cohere tests | 23 | testdata/upstream/00..22.ts.txt, sources.json names original Go test file:line; no-var and no-debugger tests |
| Clean TypeScript test corpus | 3 | testdata/typescript/sources.json, optional-chain arrow comments, constant template evaluation, Unicode literal |
| Shortest clean controls | 13 | testdata/edges/00..12.ts.txt |
| Adamic sources | 4 | parser/nodes.ts, parser/main.ts, cohere/lint/lint.ts, cohere/tsprinter/expressions.ts |
| Additional corpus/recovery probes | 4 | testdata/recovery/sources.json and answers.json; two disagree with Go, two agree |
| Reduced parser boundaries | 3 | testdata/reduced-recovery/00..02.ts.txt, full .go.txt/.node.txt answers |
| Isolated clean parser reduction | 1 | testdata/blocked-parser/arrow.ts.txt |
| Language capability probe | 1 | gaps/typed_arrays.ts |

Main corpus totals 143 files, 917,981 table nodes, 922,041 edges, 61,580
per-file interned strings, 50,685,082 encoded bytes, 46,129,063 Go canonical
bytes, 50,128,226 exact transport bytes. Node typed-array reading, Go reading,
re-encoding and Linux mapping all agree with the unchanged current node table
for all 143. Go canonical comparison agrees on 142; Babel's first mismatch is
FunctionDeclaration End 25,854 versus port 25,772. testdata/parser-boundary.json
pins both full canonical SHA256 values and fails if that boundary changes.
The four additional recovery fixtures round-trip the current tree; Go/Node
recovery output agrees on two and differs on two. Those differences are held to
recorded hashes, not counted as parity.

| Shortest input | Why it matters / observed result |
| --- | --- |
| `x;` | Four reachable nodes; nine native retain/release pairs per walk |
| `a?.b;` | Optional-chain flag and punctuation child |
| `f(); f;` | List count 0 versus absent -1 |
| `[,,];` | Omitted-element indices and traversal order |
| `[x,];` / multiline `[x]` | Trailing comma and multiline flags |
| `'\ud800';` | Cooked lone surrogate preserved as UTF-16 units, not U+FFFD |
| `` `a${x}b`; `` | Cooked/raw template strings stay distinct |
| `/*😀*/x;` | UTF-16 positions differ from UTF-8 canonical byte offsets |
| `import type {x} from "m";` | Semantic fields cannot be dropped |
| `(x)=>x; (x);` | Arrow/parenthesis tree shape and speculative entries |
| `a&&(b&&c);` | Logical-tree shape the printer normalizes |
| `const x=1; let y=2;` | BlockScoped semantic payload |
| `let a=({b:c=>c});` | Go clean, port false diagnostics and wrong recovered tree |
| `a?.b<T>.c;` | Go diagnostic 1477, port omits it |
| `` `\u`; `` | Go diagnostic 1125, port omits it |

The exact literal controls are checked-in raw fixtures, not code executed by the
source projects. Existing Go flags other than retained port fields are outside
this format's oracle. No findings/fixes/formatting counts were fabricated.

### Public sample inventory

| Fixture | Repository | Commit | Source path |
| --- | --- | --- | --- |
| 00.ts.txt | actualbudget/actual | `9732a4463aac2909ac2aad1627085e0eb7a207b5` | `packages/api/app/query.ts` |
| 01.ts.txt | angular/angular | `c0dc8c4bbeea70879aef54e9fcc7888359dfd1a5` | `adev/shared-docs/components/algolia-icon/algolia-icon.component.ts` |
| 02.ts.txt | babel/babel | `f67453d563918a1ea4a1dcb1d7af01d4a528d99b` | `Gulpfile.ts` |
| 03.ts.txt | backstage/backstage | `532931240eec03b318efaadb39e24d9ee6d7076c` | `.storybook/main.ts` |
| 04.ts.txt | calcom/cal.diy | `54343aa685ae8f33159d2f485ec4a57bad5c574a` | `.snaplet/transform.ts` |
| 05.ts.txt | date-fns/date-fns | `717ce0a807ea4c6b540d015b5c408723175b2838` | `pkgs/core/examples/vite/vite.config.ts` |
| 06.ts.txt | excalidraw/excalidraw | `ed10ac7dca7e40f3f4a31269b4bfba980d0db41e` | `examples/with-script-in-browser/utils.ts` |
| 07.ts.txt | grafana/grafana | `c7a7b797c1980a886efdd01b433a8d92b7ae7166` | `apps/alerting/rules/plugin/src/generated/alertrule/v0alpha1/alertrule_object_gen.ts` |
| 08.ts.txt | microsoft/playwright | `2a8ba77a33a5a39a52372c42f12d254506296c76` | `.github/workflows/merge.config.ts` |
| 09.ts.txt | n8n-io/n8n | `e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073` | `knip.ts` |
| 10.ts.txt | nestjs/nest | `35142c3eca8edaaf6abc5984d915da2fbd458aa2` | `integration/_support/register-local-packages.ts` |
| 11.ts.txt | outline/outline | `478e8121cbe517b9d7773e9d68d20bfe26056c59` | `app/actions/definitions/webmcp.ts` |
| 12.ts.txt | prisma/prisma | `c882b03377e70c090d7bbe05cf85bf04fe9e33b2` | `apps/lsp-playground/src/bridge.ts` |
| 13.ts.txt | shadcn-ui/ui | `a2e305c2e15b6affdb760e65058be69e8212713f` | `apps/v4/app/(app)/(create)/hooks/use-action-menu.ts` |
| 14.ts.txt | supabase/supabase | `87681812a0b4aef538d18ea78a78e5ab6257952e` | `apps/design-system/config/docs.ts` |
| 15.ts.txt | TanStack/query | `eaa75f4f8f819237febca9f7e887455367b7ab97` | `examples/angular/auto-refetching/src/app/app.component.ts` |
| 16.ts.txt | tldraw/tldraw | `db1c86ea7857483c47aa333cf4e92abf455a67cf` | `apps/analytics-worker/src/worker.ts` |
| 17.ts.txt | trpc/trpc | `d756e591a5e37ef20b8d75ecd4d736c195497289` | `examples/.experimental/next-app-dir/next.config.ts` |
| 18.ts.txt | twentyhq/twenty | `ddd166250b03e4cb262766034e8ef3e444b516c0` | `packages/create-twenty-app/src/cli.ts` |
| 19.ts.txt | typeorm/typeorm | `c64a1f052fc39f6688b6b73b83d065d7147ba8bb` | `docs/docusaurus.config.ts` |
| 20.ts.txt | microsoft/TypeScript | `50d70a3f5f453a79a4323b263165da51f656a4e3` | `packages/typescript/scripts/generate.ts` |
| 21.ts.txt | microsoft/TypeScript | `050880ce59e30b356b686bd3144efe24f875ebc8` | `src/compiler/_namespaces/ts.moduleSpecifiers.ts` |
| 22.ts.txt | vuejs/core | `4ab865a848a1da3d10fb674f857e5fff13094644` | `packages-private/dts-built-test/src/index.ts` |

## Mutants and checks that can fail

- Ten valid column mutants change kind, pos, end, optional flags, literal flags, list, cooked/raw text, operator and semantic payload. Go and Node execute successfully; exact snapshot comparison catches every change.
- Five additional valid mutants alter optional/trailing/multiline bits, swap child order, or shorten a child range. Both readers execute; equality checks catch each. A count-only walk cannot detect reordered children, which is why ordered full-tree comparison is required.
- Ten invalid storage mutants cover magic, version, file length, root, edge, child range, string ID, pool range, cycles and the reserved word. Open refuses every one. Every shorter prefix of the sample file is rejected.
- Recorded parser/recovery answer hashes and clean-source refusal checks deliberately stay red if the boundary changes; a repair needs the tests and evidence updated, not a silent exemption.
- Typed-array test holds Node's exact output and lower.NotYet category; closing the gap fails its current expected-boundary assertion.

No compiler/runtime mutation was made. Performance numbers are observations;
the Go tiny-tree benchmark is explicitly auxiliary, not a native-port claim.

## Reproduction and flags

Set up the existing oracle submodules and a scratch TypeScript checkout at the
pin above. Every package test runs by default. Missing compiler input is a hard
failure; there are no t.Skip paths. Tests use t.Parallel at every new top level.

```sh
source /workspace/adamic-tools/env.sh
mkdir -p /tmp/adamic-gate
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/flat41/typescript
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/flat41/typescript ADAMIC_FLAT_ARTIFACTS=/workspace/scratch/flat41/artifacts go test ./stage1/typescript/flattree -count=1 -v -timeout=30m > /tmp/flat-tests.log 2>&1
go vet ./stage1/typescript/flattree > /tmp/flat-vet.log 2>&1
gofmt -l stage1/typescript/flattree
node --expose-gc --disable-warning=ExperimentalWarning oracle/node.mjs stage1/typescript/flattree/benchmark.mjs /workspace/scratch/flat41/typescript/src/compiler/parser.ts /workspace/scratch/flat41/artifacts/parser.flat 100 > /tmp/flat-node-bench.json
GOMAXPROCS=1 go test ./stage1/typescript/flattree -run '^$' -bench BenchmarkWalk -benchmem -count=5 -cpu=1 > /tmp/flat-go-bench.log 2>&1
```

Measured platform: Linux amd64, AMD EPYC 9V74, five visible processors, cgroup
cpu.max `400000 100000`, four available cores. Node v24.19.0, Go go1.27.1.
Node benchmark uses one thread, 100 walks per row, five alternating rounds;
parser construction, fs.readFileSync, column views and GC precede timing. Each
walk visits 50,764 nodes. Current times: 96.545..117.037 ms; flat:
35.356..40.334 ms per 100 walks. Best ratio is 0.366, median ratio 0.370.
Native counts use -O1 -g -DADAMIC_COUNT -fsanitize=address,undefined
-fno-sanitize-recover=all and the repository's C11 warning flags. Logical counts
are not native release timing. Go auxiliary benchmark uses three synthetic
nodes: current snapshot 11.20..12.02 ns, flat columns 43.59..46.61 ns, both
0 B/op and 0 allocs/op. Flat does not beat Go's pointer representation in that
benchmark. Benchmarks do not include lint listeners, printer normalization,
checker binding, page-fault distributions or parallel walks.

Evidence files contain the package log, clean vet output, Node/Go benchmarks,
Go AST metrics and C sizeof results. testdata/go_metrics.go builds via an overlay
inside the pinned typescript-go module; it changes no oracle production file.
Go's retained heap and V8's retained heap are different measurements, not a
cross-language exact allocation-byte comparison.

## Next scout on this step

Take this format, fixture manifests, exact reader mutants, native retain deltas,
Node walk benchmark and shortest parser reductions. Use the received typed-view,
mapping-owner, string-view and borrow rulings below, wait for runtime #4gkdjsz to
land, and hand the clean-source reduction to the parser owner. Then implement a
native scalar column reader in the approved APIs and prove zero retains across
a read-only walk. The
printer owner needs the overlay design before adoption. Expand the sampled
public pins to Kirk's full quiet-hundred manifest, and run actual lint findings,
fixes and formatting byte comparisons after the consumers can read flat trees.
Do not claim the current public sample or reference tool is that completion.


## Follow-up: the parser parity miss, step 25 handoff

The mismatch is **the port's parser versus typescript-go**, not Babel's reading.
Babel is only the repository from which the first real witness was extracted.
Both runs take raw source directly through their parser drivers. No Babel parser,
transform, dependency installation or project script participates.

Shortest measured clean whole-file input: **`let a=([b=>c])` (14 bytes)**.
The original object witness reduces to `let a=({b:c=>c})` (16 bytes); removing
its property colon/name yields the smaller array form. Both hit the same
speculation path. All fourteen single-character deletions were checked: ten have
Go parse diagnostics; four are valid and agree with Node. This establishes
single-deletion minimality, not an exhaustive proof over every possible program.

The expected typescript-go tree, in testdata/parser-parity/shortest.go.txt, is:
SourceFile -> VariableStatement -> VariableDeclarationList -> VariableDeclaration,
with Identifier a and ParenthesizedExpression -> ArrayLiteralExpression ->
ArrowFunction(Parameter b, EqualsGreaterThanToken, Identifier c), then EOF.
Go emits no diagnostics. The port reports a missing comma at the interior arrow,
then declaration/statement diagnostics, and returns a recovered wrong tree.
The original object expected tree remains in testdata/reduced-recovery/02.go.txt.
The reduced object, array and expression/assignment controls have complete Go
and Node answers in testdata/parser-parity/.

### Exact parser function and the divergent condition

**Parser.arrowCandidate**, stage1/typescript/parser/parser.ts:1886, is the
responsible acceptance function. It invokes **Parser.parameters** at :1069,
which ignores `expect('CloseParenToken')`'s false result at :1074.
arrowCandidate then accepts `EqualsGreaterThanToken` at :1907 without requiring
a successfully closed parameter list. It rewinds the speculation, returns true,
and the ordinary arrow parser reparses the parenthesized array/object as an
arrow signature with recovery.

The trace of the 14-byte input proves the sequence, without changing parser
source or fixing grammar behavior:

1. At OpenParenToken, pos 6, active contexts are `source, variables`.
2. Speculative binding-array parsing reaches the **interior** `=>` at pos 9.
3. recoverList('bindingArray') and recoverList('parameters') stop there. The
   outer variables list deliberately treats `=>` as a recovery terminator;
   see parser.ts:196 and recovery.ts:53.
4. expect('CloseParenToken') returns false at that same token.
5. arrowCandidate nevertheless returns true, after rewinding diagnostics.

The corresponding Go guard is cohere/TypeScript/tsc/internal/parser/parser.go:4445:
`if !p.parseExpected(ast.KindCloseParenToken) && !allowAmbiguity { return nil }`.
tryParseParenthesizedArrowFunctionExpression at :4394 selects the strict,
allowAmbiguity=false path for an uncertain head, then rewinds a nil result.
The variables-list recovery terminator is also present in Go at :927; the
recovery terminator itself is not the divergence. The missing successful-close
requirement in the port is what lets an interior arrow certify an outer head.
No repair is applied: this belongs to **step 25**.

### Comparison with the parser census

Attempted fetch of `parser-scout/census` failed because that remote ref was not
pushed. A successful remote-head listing also found no `parser-scout/*` or
parser census branch. Consequently no census SCOUT.md, SHA or actual census
results can be quoted from that branch. The comparison here is explicitly to
**the two census findings named in the brief**, not a claimed review of an
unavailable artifact.

- **Missing node flags:** this defect changes node kinds, child structure and
  diagnostics before canonical flag selection. Adding missing node flags cannot
  turn the incorrectly accepted ArrowFunction into the required parenthesized
  array/object. This is an additional grammar/speculation defect, not a flag-only
  mismatch. The existing canonical protocol does not inventory all Go flags.
- **Entry points parsed outside a list context:** the failing input uses
  Parser.file(), and the trace proves both source and variables contexts are
  present. It is not a missing-list-context entry-point fixture. Context still
  matters: `({b:c=>c})` parsed through Parser.expression() with an empty context
  list returns ParenthesizedExpression with no diagnostics; as a complete file,
  inside an array, or on an assignment's right side, the controls also agree
  with Go. Active variables-list recovery exposes the missing speculative
  close-paren guard. An outside-list census harness could therefore miss this
  production whole-file defect. Keep both entry modes in step 25's corpus.

trace.mjs wraps Node calls to the original parser methods only for observation;
recorded trace JSON includes the failed close-paren and accepted-arrow events.
Package tests replay the Go expected trees, Node boundaries, controls, traces and
single-character deletions. Parser source, the shared harness and the other
shared lanes remain untouched.

## Rulings received from @system_adamic, steps 41 and 43

The user supplied these rulings for the next piece. Step 41 questions are now
answered; native work waits on the runtime area landing, **#4gkdjsz**.

| Step 41 question | Ruling |
| --- | --- |
| Typed arrays | Runtime owns them; Uint8, Int32, Float64 and Uint16 are already built there. Step 41 waits for that area to land. The old probe still accurately records this branch's older stage 0. |
| Sealed immutable views | Readonly typed-array views are admitted, with compile-time readonly and no writable alias into the same backing. |
| Mapping lifetime | A mapped file is an owned resource. Every view and borrowed string retains the mapping; the last release unmaps it. |
| Borrowed UTF-16 strings | Immutable strings retain their backing, with no copy and no early free. |
| Disk layout | Little-endian and naturally aligned; a versioned header carries magic, version and endianness. Foreign/mismatched layouts are refused loudly. |
| Printer overlay | The printer owns a store keyed by node index; the shared tree stays immutable. |
| Zero-retain walks | Borrow inference proves reads that never store or return node references need no retain/release. The measured 101,529 pairs per 50,764 visits are the pass's acceptance target. |

The prototype's manual Go mapping lifetime is a host-tool precondition, not the
ruled product lifetime API. Its v1 header implies little-endian through the format
version; the product revision must carry explicit endianness and reject foreign
layouts. Those requirements are not retroactively claimed as implemented here.
The next native piece follows the rulings once the runtime area lands; it must
not emulate missing typed arrays or bypass backing-resource ownership.

Related step 43 rulings, recorded for context: the scoreboard may remain Go host
tooling, measuring Adamic binaries through stdout, exit code and timing without
linking into the product. A snapshot without installed dependencies is a distinct
labeled input and may only be compared with snapshots of that same kind. No
scoreboard code or dependency snapshot is changed by this parser follow-up.
