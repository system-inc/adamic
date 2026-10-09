# Honest optional values at nonliteral nullish unwraps

This adaptation repairs 11 of the 22 nonliteral assertions that received
`undefined` while stock TypeScript 6.0.3 ran the 301 acceptance projects. Eleven
remain blocked by the simultaneous unchanged-JavaScript and unchanged-public-API
requirements. [BUGS.md](BUGS.md) lists every stock file:line:column, expression,
input set, first triggering input, Node value and repair or blocker.
[disposition.json](disposition.json) is the complete machine-readable ledger.

The October 8, 12:35 ruling reserves literal `undefined!` and `null!` for compiler
placeholder support (#9wc5q5j). This adaptation leaves those 50 observed sites
alone. Every other assertion remains a checked unwrap; moving a false assertion
to a cast would retain its failure. The original split was 50 placeholders,
seven property/call results, and 15 other nullish expressions. The ruling's type
lie category includes all seven plus all 15, totaling 22.

Run `node stage3/adapt/49-honest-optionals/adapt.cjs <tree>`. The script uses stock
`typescript@6.0.3` ASTs and binding symbols, operates on the current text and
matches owners and expressions without line-number assumptions or source regexes.
It accepts earlier adaptations, preserves source outside its explicit edits, and
checks JavaScript bytes with comments both retained and removed before writing.
All proposed files are checked before any write. Missing owning expressions stop
it rather than silently losing a repair. apply.sh invokes it unchanged between
48 and 50; no pipeline, compiler production source, baseline reference or public
declaration file is edited in Adamic.

The repairs

- Remove the false list assertion on `source.typeParameters` and the false `then`
  property assertion. Their existing consumers already handle an absent value.
- Widen four private argument declarations: `getSignatureOfTypeTag.node`,
  `getTypeWithDefault.defaultExpression`, `isAccessible.symbolFromSymbolTable`,
  and `areTypeParametersIdentical.targetParameters`. Existing presence, equality
  or length tests handle absence. Four explicit checked reads remain inside those
  existing guards instead of checking the absent value at the call boundary.
- Allow `instantiateTypes` to forward an optional mapper. Parameterize the three
  private `instantiateList` signatures by mapper type `M`, retaining a required
  default for the other instantiators. `instantiateType` already handles absence.
- Make the `reduceLeft` implementation accumulator and callback accumulator
  optional; its generic overload already represents the caller's actual initial
  type. Make `or`'s accumulator and general return optional, with a nonempty tuple
  overload retaining precision for fixed nonempty calls.
- Own scanner `text` and printer `writer` as optional storage. The scanner passes
  its optional initial value to the existing `setText`, which normalizes it before
  returning methods. Printer entrypoints install a writer before emitting, and
  preserve optional saved writers on reset. Check active reads while keeping
  optional setter inputs, saved writers and presence tests optional.

The source patch has four files, 147 lines added and 144 removed in apply.sh's
incremental table. There are 11 repaired observed assertions and 116 explicit
read assertions, including the four guarded reads and scanner/printer lifecycle
reads. This is type syntax only; it introduces no runtime branch in stock tsc.

The blockers

Three observed `symbol!` arguments flow through `createTypeWithSymbol` into
public `Type.symbol: Symbol`. The object really stores undefined. Widening that
owner to an honest optional type changes the public API snapshot. Leaving the
required property type and adding another assertion would preserve the lie.

Eight sites involve numeric optional values: two `contextFlags!`, two `flags!`,
one `meaning!`, two `pos!`, and `result.isReferenced!`. Node coerces undefined to
zero for bit operations or to NaN for position arithmetic. Honest optional
handling needs runtime syntax at that read; `Number(value)` or a default changes
emitted JavaScript bytes. The `meaning` chain runs through the private node
builder, symbol name/expression helpers, lookupSymbolChain, getSymbolChain, and
`needsQualification`'s `flags & meaning`. Public required parameter types can
stay narrower than an accepting private implementation; they alone do not prove
that widening `meaning` changes the API. Its numeric read is the blocker.

[all-removals-diagnostics.json](all-removals-diagnostics.json) records the trial:
zero upstream semantic diagnostics before removing all 22 false assertions and
81 afterward. `blocker-proof.cjs` rejects explicit coercion edits by JavaScript
identity and shows that widening public Type.symbol changes the API bytes. It
records the separate meaning propagation limitation. These sites are retained
and reported, not claimed repaired. Completing all 22 requires a revised byte
constraint or compiler support for truthful optional numeric operands.

Proof

The reference is origin/main 45487a809f89885a3fc651cd590e7dabf31362dc, measured
before adding 49. Source input is TypeScript v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8. Node is v24.19.0, nproc is 5 and the
CPU quota is four cores. The previously completed setup is reused: node 0.035s,
Go 0.048s, submodules 0.107s, clang 0.268s, markdown dependencies 1.317s,
go build 63.548s, complete 63.760s; its full log remains in the scout evidence.

Commands and observed outcomes are preserved in evidence and verification.json.
The fresh landing lane invokes apply.sh followed by oracle/run.sh with default
all suites and eight workers, serially. The reference default oracle has build
exit 0, 106,366 passing, one sanctioned API-snapshot failure, zero pending, and
only `api/typescript.d.ts` differing from pristine. The final candidate matches that exact observation and the lane's sanctioned API normalization.

`artifact-proof.cjs` compares all ten emitted built/local JavaScript files,
`built/local/typescript.d.ts`, every local baseline file and the complete baseline
diff bytes against the reference. It also mutates one JavaScript artifact, one
public API byte and one baseline byte; each must be rejected by its own byte
check. All ten JavaScript artifacts and the 590,847-byte public API are byte-identical; all three mutants were caught. Only the sanctioned API local baseline exists, and its diff bytes also match.

The moved-read probe inserts real nullish stops at all 116 assertions via the
compiler API. All 301 project output triples remain identical; 38 sites execute
910,898,035 checked reads without a nullish value, and 78 sites are unvisited.
Those 78 remain outside the runtime evidence. The lifecycle/guard argument and
unchanged emitted stock JavaScript are separate evidence, not a claim that the
probe executed them. [read-summary.json](read-summary.json) and
[read-observations.json](read-observations.json) retain the exact counts and inputs.

`idempotence.cjs` applies the script twice to one adapted tree and requires source
byte identity. `mutants.cjs` restores the bad unwrap at each of the 11 repaired
sites; the owner-expression census catches all 11. `runtime-mutants.cjs` checks
all 22 original sites on a real triggering acceptance input: replacing the
identity observer by a checked unwrap for that site yields exit 70 and names the
site and undefined. Stock exits 0 or 2. This includes the blocked sites, so their
runtime incompatibility is demonstrated rather than inferred.

Reproduce the focused checks with NODE_PATH pointing to stage3/api/node_modules
and esbuild@0.27.3 where the bundle probes use it:

```
node stage3/adapt/49-honest-optionals/check.cjs <adapted-tree>
node stage3/adapt/49-honest-optionals/idempotence.cjs <adapted-tree>
node stage3/adapt/49-honest-optionals/mutants.cjs <adapted-tree>
bash stage3/lane/run.sh <fresh-results>
node stage3/adapt/49-honest-optionals/artifact-proof.cjs <main-tree> <adapted-tree> <main-oracle> <adapted-oracle>
```

No whole Adamic gate or package confirmation was run. No new .a fixture was
added; the original scout fixtures and counts remain its recorded historical
feature-area observations. The main reference apply/oracle, complete slot probe,
complete checked-read probe, focused mutants and fresh default landing lane are
the required measurements for this follow-up.

Latest-main gate follow-up

Merged origin/main 73352e874ddbda5a78c95c5c670860d43275e4d0. Its production compiler and stage3 adaptation/pipeline sources are identical to the measured reference main; the intervening changes are tests and reports. The default lane passed in 777.147 seconds, and an independent `python3 stage3/lane/check.py` passed: 106,366 passing, exactly one sanctioned Public APIs failure, zero pending. Build exited zero. Every other upstream baseline remains identical. See [verification.json](verification.json) and the final evidence logs.

The gate's reported failing check was `test counts`: expected one failing upstream test, observed four, with three extra `unusedTypeParamet...` compiler tests. Those extra failures did not reproduce after the merge. A source/emitted-JavaScript audit scanned 756 files and found zero observation hooks; its injected-hook mutant was rejected and restored. Probes generate separate scratch bundles and never write instrumentation into apply.sh's adapted tree. No cause for the historical three failures is claimed.
