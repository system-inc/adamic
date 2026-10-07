# Host errors and bounded byte reads

Adaptation 47 adds `src/compiler/hostErrors.ts` to the external TypeScript
6.0.3 tree, rewrites two caught-property reads, and adds seven BOM assertions
in sys.ts. Four caught-property sites are deliberately declined. No checker
option is relaxed. The full-series default oracle has one pre-existing API
failure, so this unit does **not** claim the requested empty baseline diff.

Input: microsoft/TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Adamic base:
`origin/area/stage3` at `a62e1f9de6c91080ddabfa375479fb437a46ed0c`, which
already contains current main `f8013f0`. The requested branch was created
from current main and the area branch merged by fast-forward. No rebase,
force-push, main push, or PR is used. Only this directory is delivered.
The generated shared scoreboard was saved here and restored in the repository.

## Helpers and source edits

Both functions are plain TypeScript with explicit narrowing, no cast, no any,
and no non-null assertion:

```ts
/** @internal */
export function errorCode(error: unknown): string | undefined {
    return typeof error === "object" && error !== null && "code" in error && typeof error.code === "string" ? error.code : undefined;
}

/** @internal */
export function errorMessage(error: unknown): string | undefined {
    return typeof error === "object" && error !== null && "message" in error && typeof error.message === "string" ? error.message : undefined;
}
```

The fs worker's required errorCode signature is unchanged. The new .ts file
exists only in the external tsc checkout; the Adamic repository contains .cjs
tooling and evidence, no committed tsc checkout or new .ts program.
The compiler's existing `**/*` includes the helper. Neither helper is added to
the public namespace barrel; both are marked internal.

adapt.cjs uses stock 6.0.3's parser and checker, identifies caught receivers by
symbol identity and unknown type, and imports directly from `./hostErrors.js`.
It edits current parsed spans, preserves existing CRLF source bytes, validates
all planned edits before writing, and rejects helper-content or reviewed-site
drift. It validates the fixed mkdir and builtin-require producers, byte-buffer
producer, guard expressions, pair-loop bounds, and counts. It never applies a
regular expression to source. Earlier assertions are recognized rather than
reinserted. There are no subdirectory catch-property reads to rewrite.

Incremental scoreboard: **3 files, 16 lines added, 5 removed**. The seven new
byte assertions occupy three existing condition lines. Two byte-swap reads
already have assertions from adaptation 32 and are retained.

## Catch census and behavior proof

The AST census finds **26 actual catch clauses: eight with bindings and 18
without bindings**. Comments and generated-helper template strings are not
compiler catch clauses. Six bound clauses read caught properties. The other
two already accept unknown without property access. All 18 bindingless catches
remain unchanged. evidence/stock-after.json records every clause.
Locations below use the pre-47 adapted tree, not edit addresses.

| Site | Action | Producer and behavior |
| --- | --- | --- |
| sys.ts:1553, createDirectory | `e.code` becomes `errorCode(e)` | The try contains only `_fs.mkdirSync(directoryName)`. Ordinary Node validates the path and calls its native mkdir binding; filesystem and validation errors are Error objects with string codes. EEXIST remains swallowed; other errors are rethrown by identity. Fixture 13 observes existing directory, existing file/EEXIST, and missing parent/ENOENT. |
| tracing.ts:66, startTracing | `e.message` becomes `errorMessage(e)` | The try contains only fixed builtin `require("fs")`. The normal Node builtin loader loads fs without running a user module. Its loader failures are Error objects with string messages. The existing `|| e` fallback remains, including empty messages. The fixed fs builtin is present on the measured Node. Six extracted catch-body cases compare name/message, including empty messages and ordinary primitives. |
| sys.ts:1280, watchPresentFileSystemEntry | decline | createSystemWatchFunctions takes fsWatchWorker and getModifiedTime from its caller; the try also calls the returned watcher's on method. The type contracts impose no thrown-value constraint. An injected worker throwing null makes the original code throw TypeError; unconditional errorCode would instead select polling. An extracted actual function demonstrates this. |
| program.ts:406, createGetSourceFile | decline | readFile is a supplied callback, and performance hooks also execute inside the try. A callback can throw null/undefined or an object with a numeric message. Original null/undefined reads throw TypeError; original numeric messages reach onError unchanged. A string-only helper would change each outcome. |
| program.ts:441, createWriteFileMeasuringIO | decline | actualWriteFile, createDirectory and directoryExists are supplied callbacks, reached through writeFileEnsuringDirectories. The same null/undefined and numeric-message counterexamples apply and are executed against the extracted original function. |
| commandLineParser.ts:2301, tryReadFile | decline | Public config-file readers accept a supplied readFile callback. Numeric messages are passed to diagnostic construction; null/undefined makes the original property read throw. All three cases are executed against the actual extracted function. |
| program.ts:2847, runWithCancellationToken | unchanged | instanceof OperationCanceledException narrows before use; the original value is rethrown. No TS18046. |
| sys.ts:1621, nodeSystem.require | unchanged | The caught unknown is returned as the error value, without reading properties. No TS18046. |

The two accepted proofs concern **unmodified Node builtins**, consistent with
the native host's stated Node-shaped error contract. They do not certify
monkeypatched fs functions, CommonJS loader hooks, proxies, or user getters.
Such alterations can throw arbitrary values even from the fixed builtin sites.
This limitation is explicit; the compiler's actual injected-host interfaces
above are declined rather than assigned that stronger contract.

For ordinary non-null primitives, a JavaScript property read usually produces
undefined and the helpers also return undefined. A primitive's prototype can
supply properties; a function can carry properties; and a non-string field can
be present on an object. The helpers filter all these cases. Null and undefined
are more significant: the original read throws while the helpers return
undefined. Stable inherited string fields on objects are retained by `in`.
Getter/proxy side effects need a separate proof because the helpers check and
then reread a field. Node's ordinary error fields are stable data properties.
No claim is made that a helper is a universal replacement for property access.

## Byte-order marks

All nine buffer reads are required values; two writes are left alone.
The Node readFileSync result is a dense Buffer, not a sparse caller array.
There is no callback or resizing between the guard and the reads.

| Reads | Presence proof | Change in 47 |
| --- | --- | --- |
| BE buffer[0], buffer[1] | `len >= 2`, len initialized from buffer.length | two assertions |
| Swap buffer[i], buffer[i + 1] | `len &= ~1`; i starts at zero, advances by two, and i < even len proves i + 1 < len <= buffer.length | existing two assertions retained |
| LE buffer[0], buffer[1] | `len >= 2` | two assertions |
| UTF-8 buffer[0], buffer[1], buffer[2] | `len >= 3` | three assertions |

This follows adaptation 30's checked-read pattern. Each read stays at its
original evaluation point. Stock TypeScript emits the readFile function
**byte-identically** before and after. Fixtures 01-04 receive the actual
adapted readFile function, covering UTF-8, UTF-8 BOM, UTF-16LE/BE, malformed
input, short and odd tails, missing files and directory failures.

## Measured checks

Stock census uses the compiler project and Node declarations, with Adamic's
strict options and ESNext/Bundler module settings. useUnknownInCatchVariables,
strictBindCallApply, noUncheckedIndexedAccess, exactOptionalPropertyTypes,
noImplicitReturns, noFallthroughCasesInSwitch, verbatimModuleSyntax and
erasableSyntaxOnly are enabled. No config file is edited or option weakened.
The upstream build retains upstream's own configuration; the strict census is
a separate proof, not a claim that upstream enables unknown catches itself.

| Meter | Before | After | Buffer mutant |
| --- | ---: | ---: | ---: |
| Stock TS18046 | 6 | 4 | 4 |
| Unasserted buffer reads | 7 | 0 | 1 |
| Stock diagnostics within readFile | 0 | 0 | 1, TS2322 |
| All compiler stock diagnostics | 733 | 731 | 732 |

The four stock TS18046 remainders are exactly the declined sites above.
BOM equality expressions themselves are accepted by stock TypeScript even
without assertions; the independent census presence check covers those too.
This is not a claim that all 731 general checker diagnostics were repaired.
The shared real Adamic census is recorded separately in evidence/adamic.json.
Its whole-program TS18046 count is **8 -> 6**: the same four declined catch
reads plus two unchanged non-catch callback reads at transformers/jsx.ts:187.
Those read s.propertyName and s.name in arrayFrom(importSpecifiersMap.values(),
s => ...); the loader infers that callback value as unknown. They are outside
this catch/BOM adaptation. Before checks 78 roots; after checks 79. Every other
whole-program diagnostic-code count is identical. The loader has no Node host
library declarations yet, so the stock census with Node declarations supplies
the independent Buffer type proof.

The helper alone passes the shared loader's checker but reaches **Refused: in**
at hostErrors.ts:3:66. The current lowerer still refuses this reflection
operator. This user-required plain-TypeScript narrowing helper is therefore
ready as a source adaptation, not a claim that it already compiles natively.
No production lowerer or native-host file was edited to bypass the refusal.

verify.cjs checks all **25 host fixtures** against their recorded exact Node
stdout, stderr, and exit codes. It inserts the actual adapted method into
fixture 13 and the actual adapted readFile into 01-04. The remaining 20 fixtures
serve as unchanged host controls. This is Node proof, not native host execution.
It also checks primitive/field helper contracts, six tracing catch cases,
all four declined-site counterexamples, absence of casts/any/! in the helper,
readFile JavaScript equality, and byte-identical idempotence. A real second
adapter CLI run reports zero edits. Fresh apply.sh discovery matches all three
changed-file SHA-256 values from the separately adapted tree.

Mutants actually run:

- The helper validates/returns message instead of code. The adapted fixture 13
  rethrows EEXIST instead of swallowing it; its exact Node observation differs.
  verify.cjs catches this at runtime, not by a compiler rejection.
- One byte-pair RHS `buffer[i + 1]!` loses its assertion in a separate scratch
  tree. census.cjs --check exits 1: one unasserted read and TS2322 at the buffer
  assignment, number | undefined not assignable to number. The original adapted
  tree remains intact. This assertion existed before 47 but belongs to the
  requested readFile buffer proof; the same census also covers the seven new
  BOM assertions. Earlier intermediate runs without the complete module/type
  setup only caught the presence violation; the final report includes TS2322.

## Public API and default oracle

api.cjs verifies pristine emission against upstream's pristine API reference,
reconstructs adaptation 20's **189 allowed changed lines** from its parsed,
validated optional-declaration owner ledger, and compares pre/post-47 public
API bytes. The SHA-256 values match; there is no errorCode/errorMessage function
or hostErrors module in the public snapshot. The helper's separate internal
module declaration is not a public namespace export.

The starting full-series pipeline also changes **28 public API lines beyond
that allowance**, from adaptation 40: 27 brand-field lines and ErrorCallback's
arg0 type. Its adapter changes the reference snapshot to match these, so a
naive full-series run would conceal them. This unit reconstructs the reference
with only the authorized 189 optional-property changes, exposing the other 28.
No such reference change is committed to Adamic. api-proof.json lists each
pre-existing difference. They exist before adaptation 47 and remain identical
after it. Fixing adaptation 40 is outside this unit's territory.

The **default all-runner oracle actually ran**, unfiltered, four workers:
**106,366 passing, 1 failing, 0 pending**. Install/build exited 0; tests exited 1.
Only `api/typescript.d.ts` differs, by those 28 pre-existing lines. There are no
other baseline differences. Install 2.294s, build 13.005s, tests 359.803s,
wall 375.172s. The baseline.diff artifact is deliberately nonempty. This is
an observed failure, not an empty-diff pass with an enlarged exception.

Setup passed: Go 0s, clang 1s, Node 1s, submodules 1s, cache warm 121s,
total 121s; nproc 5, cgroup quota 4, Node 24.19.0, Go 1.27.1, clang 20.1.8.
The printed environment is /workspace/adamic-tools/env.sh. All test output
went to logs. The full Adamic gate, native tsc, lint, browser integrations,
Windows execution and alternate Node versions were not run.

## Reproduction

Run from the Adamic repository with the setup environment sourced and stock
6.0.3 available on NODE_PATH (apply.sh installs it under its cache). Use fresh
output paths. The optional-owner ledger in evidence came from adaptation 20's
JSON stdout in the before apply log. The pristine checkout must also be built
with upstream's npm ci and npm run build before the API check.

```sh
bash cloud/setup.sh > /tmp/host-errors-setup.log 2>&1
source /workspace/adamic-tools/env.sh
unit=stage3/adapt/47-host-errors
stage3/apply.sh /tmp/tsc-47 > /tmp/tsc-47-apply.log 2>&1
node "$unit/census.cjs" /tmp/tsc-47 /tmp/tsc-47-census.json --check > /tmp/tsc-47-census.log 2>&1
node "$unit/verify.cjs" /tmp/pre-47 /tmp/tsc-47 /tmp/tsc-47-proof.json > /tmp/tsc-47-proof.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-47 > /tmp/tsc-47-idempotence.log 2>&1
node "$unit/api.cjs" /tmp/pristine /tmp/pre-47 /tmp/tsc-47 "$unit/evidence/optional-owners.json" /tmp/tsc-47-api.json --prepare-oracle > /tmp/tsc-47-api.log 2>&1
stage3/oracle/run.sh /tmp/tsc-47 /tmp/tsc-47-oracle > /tmp/tsc-47-oracle.log 2>&1
go build -o /tmp/tsc-47-census ./stage3/census/tool > /tmp/tsc-47-census-build.log 2>&1
/tmp/tsc-47-census /tmp/tsc-47/src/compiler /tmp/tsc-47-adamic.jsonl > /tmp/tsc-47-adamic.log 2>&1
node "$unit/mutate-buffer.cjs" /tmp/tsc-47-mutant-copy > /tmp/tsc-47-mutant-edit.log 2>&1
node "$unit/census.cjs" /tmp/tsc-47-mutant-copy /tmp/tsc-47-mutant.json --check > /tmp/tsc-47-mutant.log 2>&1
```

prepare-oracle must run after the upstream build has produced typescript.d.ts;
api.cjs reads the two existing build artifacts before preparing the reference.
Run another API check without that flag after the oracle rebuild to prove the
final artifact. A pre-47 tree is obtained by applying the series before copying
this directory into a scratch Adamic checkout. Do not delete this directory
from the working repository to obtain it.

Measured scratch paths and logs use /tmp/host-errors-*; the default oracle's
phase logs are /tmp/host-errors-oracle/{install,build,tests}.log. Source stays
outside Adamic. Evidence includes counts, fixture observations, actual mutants,
the API comparison, oracle report/diff, and the incremental scoreboard.
