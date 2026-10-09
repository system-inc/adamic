# Proofs through assignment expressions

October 8, 11:16 ruling, compiler task #cvhj5fk: a simple assignment expression
returns its RHS value and carries its membership, narrowing and definedness
proofs. The target write is still checked independently. Compound assignment
returns the newly computed target value, which needs its own proof; its result
is not the RHS operand.

This branch starts at origin/main `45487a809f89885a3fc651cd590e7dabf31362dc`.
No feature branch or compiler change is merged. The scanner witness is based on
scout `91c49dcd` and preserves the entire upstream getIdentifierToken function,
including `return token = keyword;` at scanner.ts:1823. expectations.json pins
TypeScript 6.0.3 commit `050880ce59e30b356b686bd3144efe24f875ebc8`, its scanner
source hash, and the exact function text. Only the enum, keyword table and driver
are reduced; CharacterCodes preserves the original function's identifiers.
The reduced typed Map does not prove the upstream Object.entries construction.

The other five programs are authored ruling probes, explicitly labeled as such.
All begin with an a-check pass header. Current NotYet observations are not
refusal headers; a-check treats NotYet as checked. The ruled target is Compiles,
recorded separately in [expectations.json](expectations.json), including Node
bytes, proof obligations and the independent target-write requirement.
[status.json](status.json) records current-main observations in the shared fixture
schema with portable exact diagnostics.

| Fixture | Proof and observable effects | Mutant caught by Node stdout bytes |
| --- | --- | --- |
| 01_scanner_keyword | Defined keyword membership flows into the result while token has the full enum type; tests two keywords and three misses. | Store/return Identifier in the keyword branch. |
| 02_if_narrowing | `if ((x = f()) !== undefined)` narrows x; both present and absent branches execute; f runs once per call. | The absent input becomes `missing`, taking the present branch. |
| 03_return_defined | `return (cache = compute())` returns a defined string through an optional cache; two calls prove single evaluation and stores. | Return compute directly, omitting the cache write. |
| 04_chained | `a = b = value` carries definedness through both results and writes both targets once. | Omit b's assignment. |
| 05_compound | `x += y` returns updated x: 12 then 14, with one RHS computation per call. | Replace += with =, returning/storing y. |
| 06_wider_target | A Keyword member proof survives storage into full Kind; the member-typed return is preserved. | Return value directly, omitting the wider target write. |

These mutants change actual source/input and remain valid strict TypeScript.
They fail only at the source Node byte comparison, not a checker/build failure.
They prove the fixture oracle can fail; they do not claim that the missing
compiler implementation has been mutation-tested. Mutants are generated into
scratch .a files, so intentionally wrong permanent programs need no refusal
headers or gate exemptions.

## Reproduction and observations

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules"
python3 stage3/fixtures/assignment-proofs/check.py > /tmp/step12-assign-check.log 2>&1
go test ./stage3/fixtures -run 'TestFixtures/assignment-proofs' -count=1 -timeout 5m > /tmp/step12-assign-fixtures.log 2>&1
```

check.py runs stock TypeScript 6.0.3, source Node via oracle/node.mjs, and current
`go run ./cmd/adamic build` on every fixture; any compiled binary must equal
Node stdout, stderr and exit. Six isolated child mutant runs must exit 1 at
`Node stdout byte comparison`, after zero TypeScript diagnostics. It writes
raw build/type/Node/mutant logs to the scratch path printed at completion.
[counts.md](counts.md) and [mutants.json](mutants.json) record the bucket counts
and individual catchers. No central native oracle rows were added.

Observed on current main: six Node goldens and six mutants pass their checks;
all six Adamic builds stop NotYet at assignment-expression lowering. No native
binary or native no-check proof is claimed. Use `check.py --require-adamic` after
#cvhj5fk lands to require all six Compiles outcomes and byte equality. Future
emission proof/check inspection remains the compiler worker's obligation; the
metadata's no-recheck expectation alone is not emitted-code evidence.

Setup timings: Node 0.020s, Go 0.023s, submodules 0.056s, markdown 0.067s,
clang 0.142s, validated Go-build stamp 0.631s, deferred test binaries 0.633s,
warm cache 0.634s, done 0.659s. nproc 5, CPU quota 4; Go 1.27.1, clang 20.1.8,
Node 24.19.0. Only this bucket's checks and the filtered shared fixture test ran;
no full package or full gate ran. Property setters, throwing RHSs, aliasing
references and compound enum membership are not covered by these six probes.
