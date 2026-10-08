Built closure coverage and pinned the branded-name refusal; no production lowering changed.
Branch codex/hidden-12-name-binding; base dcdbb9098f77f30ad41790c56df1bd63ad462b63.
Exact head replay, focused oracle and counts refresh pass; region remains 7,705 hidden bytes.
Empty-string capture mutant fails the nested-call oracle on stdout in both compiled backends.
Step 30 gains verified coverage; the historical boundary remains, with zero bytes revealed.

## What the replay establishes

All 82 adapted compiler file hashes match census pin `388096e6`. The binder hash is
`8ff0292668cdbe20584f06f5ca54c0a9ae419b16bc099c8d6261b1bd595a2d39`.
The inherited replay ran on the assigned area-next-fixtures base, with the full
compiler project and ancestor binding context retained:

```sh
source /workspace/adamic-tools/env.sh
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/binder.ts:760:13 \
  -kind NotYet -reason 'reading name' > /tmp/hidden-boundary-12-replay.log 2>&1
```

Exit 0, exact position/kind/reason reproduced. Worker load/register/lower took
2.38139714s; total including overlay/build was 4.880707364s. Raw replay is preserved
in `replay.txt.gz`.

Observation: the selected unit is `declareSymbol`, binder.ts:749:5. Its local
`const name` declaration at 755:9 fails at 755:15 on `__String | undefined`.
The measurement restores its pre-statement state and continues. The next finding
is `reading name` at 760:13, blocking [29754,37459). The declaration and read are
in the same attempted function, not separate ancestor scopes. The replay's
ancestor seeding therefore does not repair this local. The full binder record
preserves the enclosing blocker too: `__String` at 630:47, [17474,188866).

Inference: this is a cascade from a rolled-back unsupported declaration, not
proof of a production closure binding bug. The short closure witness already
works on the assigned base. I conservatively retain the refusal rather than
invent a local for a declaration whose initializer never lowered. Removing the
historical stop needs a separately sound representation for TypeScript's
`__String`: its definition in the pinned source includes string/void brand
intersections and `InternalSymbolName`. It cannot be fixed by substituting the
witness's ordinary string capture. This unit does not implement that representation
or bypass its checks.

## Correctness and mutant

The proposed fixture is unchanged from the brief. A separate negative fixture
uses `string & { readonly __escapedIdentifier: void } | undefined`; its exact
`a value of type Name | undefined` refusal is pinned by
`TestHiddenBoundaryNameBindingRefusal`.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run 'TestHiddenBoundaryNameBindingRefusal|TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_name_binding' \
  -count=1 -v > /tmp/hidden-12-oracle-final.log 2>&1
```

Pass, 0.514s. Source Node, JavaScript backend, release native and sanitized native
agree: stdout `node\n`, empty stderr, exit 0; LeakSanitizer passes. Node 24.19.0;
clang 20.1.8 release `-O2`; sanitized `-O1 -g -fsanitize=address,undefined
-fno-sanitize-recover=all`; leak run `ASAN_OPTIONS=detect_leaks=1`.
Both builds also use `-std=c11 -Wall -Wextra -Werror
-Wcast-function-type-strict -pedantic -Wno-unused-variable
-Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter
-Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls`. Release
adds `-DADAMIC_SLABS`; the counted build uses `-DADAMIC_COUNT -O2`. Independent source
Node observation used `NODE_NO_WARNINGS=1 node --input-type=module`,
`node:module.stripTypeScriptTypes` and a data-module import, exit 0.

The independent compiler mutant uses a scratch Go overlay of expression.go.
Immediately after `local, isLocal := l.local(node)`, it returns
`ir.StringConstant{Index: l.constant("")}` when the local is captured, named `name`
and represented as `ir.String`. It changes no source fixture or production file.

```sh
ADAMIC_GATE_UNCACHED=1 go test -overlay=/tmp/hidden-12-mutant-overlay.json \
  ./internal/oracle \
  -run 'TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_name_binding.a$' \
  -count=1 -v > /tmp/hidden-12-mutant.log 2>&1
```

Exit 1, intended catcher: stdout disagreement, not a build failure. Node prints
`node\n`; native and backend each print `\n`, empty stderr and exit 0.
`mutant.txt` preserves both disagreements. No check was elided or added to production.

## Counts, every changed row

```sh
go test ./internal/oracle -run TestCountsAreRecorded -count=1 \
  -args -update-counts > /tmp/hidden-12-counts-retry.log 2>&1
```

Pass, 59.806s. Allocations/frees/retains/releases/peak/regions:

- New `hidden_boundary_name_binding.a`: 2/2/3/3/2/0, the captured cell and returned
  closure. The refused fixture has no counted row.
- `logical_and_reference_maybe.a`: moved to the fixture registration order;
  its 8/8/11/22/4/0 numbers are unchanged.
- Removed stale `stage3/fixtures/taste/17_binder_flow.a`: the assigned base's
  taste_stage3_test.go already registers it with `lowers=false` because optional
  presence is unsupported. Its former 51/51/0/51/7/0 row was not reproducible as
  an accepted fixture. No registration or production behavior changed here.

All other rows and numbers are unchanged. The first refresh failed because
stage3/api's pinned `@types/node` was absent. `npm ci --ignore-scripts --no-audit
--no-fund` in stage3/api installed the locked dependencies; the retry above passed.

## Assigned-region measurement

The same adaptation and full-project checker were used. The inherited full latent
census completed every binder attempt before it entered unrelated checker.ts
work; it was then interrupted. This is a region measurement, not a complete new
whole-corpus result. `binder-census.jsonl.gz` contains the complete binder record
and full-mode header; `binder-stock.json.gz` contains the independently generated
stock TypeScript 6.0.3 spans. All 82 stock hashes also match the adapted files.

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/hidden-12-overlay \
  > /tmp/hidden-12-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/hidden-12-overlay/overlay.json \
  -o /tmp/hidden-12-census ./stage3/census/latent/tool \
  > /tmp/hidden-12-census-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/hidden-12-census \
  /tmp/hidden-adapted/src/compiler /tmp/hidden-12-full.jsonl \
  > /tmp/hidden-12-census.log 2>&1
NODE_PATH=/workspace/adamic/stage3/api/node_modules node \
  /tmp/hidden-census-source/stage3/census/hidden/units.cjs \
  /tmp/hidden-adapted/src/compiler /tmp/hidden-12-stock.json \
  > /tmp/hidden-12-stock.log 2>&1
```

Using census pin's hidden.py `calculate` on the complete binder record and stock
catalogue, then intersecting hidden_ranges with [29754,37459), gives:

| Measurement | Hidden intersection | Bytes |
|---|---|---:|
| Historical pin | [29754,37459) | 7,705 |
| Assigned base and final production compiler | [29754,37459) | 7,705 |
| Revealed difference | empty | 0 |

An independent byte-set audit of binder boundaries minus independently exposed
binder attempts also gives 7,705. This full-interval lower bound is exact: other
files' records cannot expose binder bytes (the census exposure loop uses the
record's own file); cross-file dependency events can only add blocked spans or
remove exposure. No interval can contain more than its own 7,705 bytes.
There is no production compiler diff between this measurement and the delivery.
`measurement.json` and `region.txt` preserve unit counts and arithmetic.

The next boundary is unchanged `reading name` at binder.ts:760:13. The preceding
cause to resolve is `__String | undefined` at binder.ts:755:15. The enclosing
`__String` blocker remains too. No group total is claimed as revealed bytes.

## Setup and scope

Ran GOPROXY='https://proxy.golang.org|direct' and cloud/setup.sh, then sourced
/workspace/adamic-tools/env.sh. The initial setup failed with missing cohere
rule_runner and shim API mismatches from the previous checkout's submodule.
`git submodule update --init --recursive` aligned cohere to the assigned base's
7945d102 revision; no cohere code was copied or edited. Retry timing lines:
Node 0.020s, Go 0.022s, submodules 0.057s, clang 0.149s, markdown dependency
step 0.818s, markdown ready 0.997s, Go build 184.111s, test binaries deferred
184.269s, cache warm 184.270s, done 184.316s. nproc 5; cpu.max 400000 100000;
Go 1.27.1, Node 24.19.0, clang 20.1.8.

Only the focused oracle and required counts selector ran, plus the required setup
build and measurement tools. No whole-package semantic tests or full gate ran.
No production lowering, runtime, native emitter, forbidden central file or
cohere source changed. No historical bytes were burned down. Toward roadmap
step 30 this lands the requested semantic witness, its independently caught
capture mutant, conservative refusal coverage and the measured dependency on
branded-string representation. Broader representation support and full-corpus
coverage remain unimplemented and unclaimed.

The preserved region evidence can be audited without another compiler run:

```sh
python3 stage3/hidden-12/measure.py \
  /tmp/hidden-census-source/stage3/census/hidden/hidden.py \
  > /tmp/hidden-12-audit.log 2>&1
```

Pass: hidden intersection [29754,37459) = 7705 bytes; revealed 0.
