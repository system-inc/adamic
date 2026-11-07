# elementAt measurement, task b5w3ycg

Measured before any existing adaptation, owner pin or elementAt source edit.
This is a stopped Step 1 checkpoint, not an elementAt repair.

Base is origin/main 251898645470552041f76a2cc1261372492b0690 plus 99r's
amendment 980870faa63a164ad82200d7688218dbc31d163a, chg_pyyr04zd. Main is
an ancestor of that amendment, so the branch fast-forwarded to it. Upstream is
050880ce59e30b356b686bd3144efe24f875ebc8, stock TypeScript 6.0.3.
The ordinary compiler has no compiler-source delta from origin/main.
Setup, the pinned API install and the complete baseline full apply with
--build-product all exited 0. The baseline includes 99r's readonly toSorted arm.

## Libs and the earlier record

internal/load/load.go:74 pins lib.es2024.d.ts; it does not read tsc's tsconfig.
Upstream src/compiler/tsconfig.json extends src/tsconfig-base.json, which declares
lib: ["es2020"]. The experiment checks ES2024 and ES2020 under Adamic's options,
then repeats the latter with the exact parsed upstream tsconfig options.
All complete-core semantic diagnostic codes agree between the two ES2020 rows.
These stock results concern the complete core.ts, not a whole-program stock
build claim. Actual Adamic types separately checks all 79 compiler roots.

Under ES2024, Array<T> and ReadonlyArray<T> declare the required method
`at(index: number): T | undefined` in lib.es2022.array.d.ts. Array.prototype is
any[], so Array.prototype.at is `(index: number) => any`. With both casts removed,
`!!Array.prototype.at` has literal type true, without TS2774 or any other core
semantic diagnostic. `array?.at` is
`((index: number) => T | undefined) | undefined`, and `array?.at(offset)` is
T | undefined. The actual Adamic checker also exits 0 for the isolated expressions.

Under ES2020, neither method is declared. Direct removal gives TS2550 at the
prototype read and TS2339 at the readonly-array call. Their reported any types
are error recovery, not typed support for at. The probe's type is boolean there.

I could not locate the exact earlier axochk ruling requested in the task.
The tracked core.ts:1637 record is in 99f, not 99q: its README says, "The ES2025
operations required by Adamic's prelude are absent at runtime; changing the
declared lib cannot repair that declaration." It introduces the truthful
HashedSet contract. The retained 99q README instead records the rejected H03
elementAt trial: cast removal exposed method-value refusals on its old compiler.
Those old lowering findings are not current results. A repository/history search,
a connected GitHub issue search for axochk, and a remote 99q branch-name search
did not recover a standalone lib ruling. No lib change or waiver is assumed here.
The relevant earlier READMEs and 99r's MEASURE.md are retained in the archive.

## Ownership

The complete stage3/adapt grep is retained as owner-grep.log. Adaptation 30's
sites.json entry pins the single array[offset] expression by occurrence, labels
its function <callback>, and declines an assertion because absence is intended.
Its adapter checks that expression and occurrence count, not the elementAt
initializer or its ancestry. Neither adapt.cjs nor whole-files.cjs pins the
initializer. A site inventory is not a complete initializer pin.
Adaptation 40 classifies and retains both casts; it does not pin or edit this
initializer. 99q explicitly dropped H03 and pins only its four accepted owners.
No other adaptation pins elementAt's initializer. The permitted new 99z is
therefore the checkpoint location. Its adapt.cjs is explicitly a no-op so this
evidence directory preserves numeric adaptation discovery without changing source.
No existing adaptation, owner pin or source transformation was edited.

## Complete-core candidates

Each candidate is substituted into the complete final adapted core.ts in an
in-memory stock CompilerHost. Ordinary probes use separate complete src copies.
The fallback body is unchanged throughout.

- baseline: both original as any casts.
- direct: `!!Array.prototype.at` and `array?.at(offset)`.
- probeOnly: remove only the feature-test cast.
- callOnly: remove only the call cast.
- satisfies: `!!(Array.prototype satisfies { at?: (index: number) => unknown }).at`
  and `array?.at(offset)`.
- indexed: `!!Array.prototype["at"]` and `array?.["at"](offset)`.

Every exported elementAt variable retains its declared
`<T>(array: readonly T[] | undefined, offset: number) => T | undefined` type.
Under ES2024 the initializer has that same signature for direct, callOnly,
satisfies and indexed; baseline and probeOnly infer an any result. Under ES2020
all initializers infer an any result, with the errors below except baseline.
Exact types, signatures, declarations and every core diagnostic are in stock.json.

| Candidate | Adamic types exit, all 79 roots | ES2020 core diagnostics | Whole-core JS identical, comments kept / removed | Ordinary binder first stop |
| --- | ---: | --- | --- | --- |
| baseline | 0 | none | yes / yes | core.ts:1076:100 |
| direct | 0 | TS2550, TS2339 | yes / yes | core.ts:1326:26 |
| probeOnly | 0 | TS2550 | yes / yes | core.ts:1077:27 |
| callOnly | 0 | TS2339 | yes / yes | core.ts:1076:100 |
| satisfies | 0 | TS2559, TS2550, TS2339 | yes / yes | core.ts:1326:26 |
| indexed | 0 | TS7015, TS7015 | no / no | core.ts:1326:26 |

Every stock ES2024 core check has zero semantic diagnostics. Every ordinary
binder row above is Refused, "a cast the runtime can't check". The standard
unmodified census tool emits a single-root and a whole-input row for a binder
file input; both rows agree for every candidate. Tool exit 0 means the census
was recorded, not that lowering succeeded. Each Adamic types exit is 0; its
normal declaration-type output is retained, without a diagnostic.

JavaScript comparison is complete-core TypeScript transpilation at ES2024/ESNext,
with comments kept and removed, exactly as 99r's measurement. Hashes and byte
sizes are retained. It is not a nine-file build or declaration comparison.
The indexed form changes emitted property syntax and does not qualify.
Satisfies checks assignability but preserves the source expression's type; it
cannot add a missing at member under ES2020. A new typed local changes JavaScript.
A structural as view would introduce a new cast, outside the requested rule.
No declaration augmentation or lib change is proposed.

## The newly exposed binder source

The direct candidate reaches core.ts:1326:26 in getOwnValues:

```ts
values.push((collection as MapLike<T>)[key]!);
```

The source is MapLike<T> | T[], target MapLike<T>, not an assignable upcast.
castProof constructs the refusal at internal/lower/cast_proof.go:188.
Its union-write gate is at :294, with the refusal return at :295;
castUnionWrites at :340 rejects a missing member property at :349. The stock
checker reports the union's mutable length property, while MapLike<T> has no
named length property and T[] does. This return-path explanation is source
inspection, not an instrumented compiler trace. next-stop.json retains the
complete function and the relation/property facts. The baseline any-cast branch
is castProof at :196-197. No compiler overlay was used.

This is one binder measurement, not a finding that all historical 73 roots
advance to the same new stop. No new per-file census groups were run.

## Other matching method feature-test stops in the historical latent census

- scanner.ts:4063:67, `(String as any).fromCodePoint`, utf16EncodeAsStringWorker;
  NotYet: String as a value outside equality or typeof (overloaded calls and
  static properties need their own representation).

The list uses the same historical checker-rejected unmerged4/main-unmerged
census as 99r's model, not a fresh ordinary census. Error.captureStackTrace has
the source shape but no retained finding at that test, so is not counted as a
matching stop. The raw census and exact source-span audit are retained.

## Decision and reproduction

No measured spelling both removes the casts without a new cast, passes both
lib configurations, and preserves complete-core JavaScript. Direct removal is
a useful ES2024-only candidate, but fails tsc's declared lib. Step 2 is not
entered. No owner amendment, mutant, paired replay, second full apply, nine-file
JavaScript/declaration comparison, or 79-root per-file census is claimed.
The exact earlier lib ruling remains a missing requested historical reference.

Commands, after the setup specified in the task:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
stage3/apply.sh --build-product /workspace/scratch/elementat/before
node --max-old-space-size=8192 stage3/adapt/99-correctness-axolotl-z/measure-elementat.cjs /workspace/scratch/elementat/before /workspace/scratch/elementat/raw
GOMAXPROCS=2 go build -p=2 -o /workspace/scratch/elementat/bin/ ./cmd/adamic ./stage3/census/tool
python3 stage3/adapt/99-correctness-axolotl-z/probe-elementat.py
```

All command logs are files. The first full-history upstream clone and a duplicate
compiler build were intentionally terminated; the exact pinned shallow clone,
completed setup and bounded ordinary build succeeded. The baseline src manifest
is unchanged after all probes. Raw results are packed in elementat-measure.tar.gz,
with an internal SHA256SUMS and an adjacent archive SHA-256 file. Every retained
Go source has a .go.txt filename. No task process remains running at closeout.
