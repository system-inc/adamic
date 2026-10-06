# TypeScript compiler strictness survey

For Kirk and Ahra's design decision. The original survey was completed
2026-10-06 without changing the language, compiler, or adapter. The subsequent
[in-memory adapter implementation and re-measurement](tsc-strictness-adaptation.md)
are recorded separately. The next
[return adaptation, argument survey and syntax census](tsc-strictness-next-layer.md)
records the subsequent meter layer.

The evidence favors keeping both strictness rules. In the exact-optional sample,
96/100 findings describe values the implementation deliberately stores; honest
source declarations can represent them without changing JavaScript behavior.
In the possibly-undefined sample, indexed access dominates. Most reads rely on
construction or caller invariants that the array types do not express. Bounds
checks alone are insufficient. A future adapter should propose these separately
from mechanical rewrites, with an explicit invariant and a checked read.

## Scope and selection

Input: TypeScript **6.0.3**, tag `v6.0.3`, commit
[`050880ce59e30b356b686bd3144efe24f875ebc8`](https://github.com/microsoft/TypeScript/tree/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler).
Meter: Adamic main `d799ede40af1d947efe8c4a897efbb90c2ef2264`.
`cmd/adamic-meter --adapt --json` examined **77 root files** under `src/compiler`
and removed **3,718 TS1484 findings** using its existing in-memory type-import
rewrite. Imports outside that directory are followed by the checker. All 200
selected findings below are located inside `src/compiler`.

| Group | Code | Population | Sample |
| --- | --- | ---: | ---: |
| Exact optional | TS2412 | 617 | 85 |
| Exact optional | TS2375 | 76 | 11 |
| Exact optional | TS2379 | 31 | 4 |
| **Exact optional total** | | **724** | **100** |
| Possibly undefined | TS18048 | 354 | 61 |
| Possibly undefined | TS2532 | 224 | 39 |
| **Possibly undefined total** | | **578** | **100** |

The unit is a diagnostic, not a unique property, function, or required edit.
One diagnostic chain may discuss several properties; several diagnostics may
be fixed by one declaration change. These groups also overlap in root cause:
an unchecked array read can subsequently cause an optional-property write error.
Other diagnostic codes were not surveyed.

Selection is deterministic, proportional by diagnostic code, and spread across
source locations. Allocate 100 slots by largest remainder of `100 * N / total`.
Within each code, sort by relative filename, line, and column; select zero-based
rank `floor((k + 0.5) * N / n)` for `k = 0..n-1`. Sort the combined sample by
filename, line, column, and code, then assign E001..E100 or U001..U100. This is
systematic sampling, not independent random sampling. Do not extrapolate these
percentages to the entire compiler or attach confidence intervals.

[The complete 200-row ledger](tsc-strictness/sample.json) retains source lines,
nearby context, full diagnostic chains, ranks, pattern, decision, and rationale.
[The unabridged meter report](tsc-strictness/meter.json) records the population.
The source permalink for any ledger row is the pinned tree above plus its
`file` and `#L<line>`.

The meter uses Adamic's fixed checker options, not TypeScript's own tsconfig:
strict, exactOptionalPropertyTypes, noUncheckedIndexedAccess, noImplicitReturns,
noFallthroughCasesInSwitch, erasableSyntaxOnly, verbatimModuleSyntax, ES2024,
ESNext modules and Bundler resolution. Upstream's base config enables strict
but not these two extra rules. Missing dependencies, unsupported enum syntax,
and other diagnostics remain. In particular, a diagnostic code containing
"exactOptionalPropertyTypes" is not proof that the root cause is optionality.
The meter's checker comes from cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`
and its pinned typescript-go; independent reduced
probes below use stock npm TypeScript 6.0.3 and Node 24.

## Decision categories

**M: mechanical candidate.** A concrete meaning-preserving transformation exists
for the pattern. The 96 optional cases are candidates for *owned declaration*
corrections, not a claim that 96 automatic edits already pass a complete compiler
port. Change a declaration only after resolving its owner and rechecking all
consumers and specialized interfaces. Public API policy still needs review.

**C: checked invariant candidate.** No unconditional mechanical rewrite was
established in this survey. If the value must exist, use an explicit runtime
check, such as `const item = items[i] ?? panic("required compiler item")`, at the
existing read. Explain the site's invariant. If absence is legitimate, preserve
its existing handling or design that handling manually instead. A check changes
failure behavior on invalid states; it is not an equivalence proof over all
JavaScript inputs. Some sites can move to M after a stronger construction proof
or reuse of an existing check; the counts intentionally do not assume that proof.

**R: review another contract or typing issue.** Do not feed the finding to an
optional-property adapter. These need investigation of the actual diagnostic
chain or the generic API, rather than a blanket cast or assertion.

No sampled finding was demonstrated to be a reachable upstream compiler bug.
Two reduced source-local counterexamples do demonstrate why the existing types
and guards cannot be treated as sound proofs for arbitrary admitted inputs.
That distinction matters: a caller invariant can make an internally partial
operation safe; the diagnostic does not tell us whether that invariant holds.

## Exact optional properties: 100 findings

| Pattern | Count | Decision | Examples |
| --- | ---: | --- | --- |
| E-init: materialized undefined factory/constructor slots | 30 | M | E030, E068, E095, E096 |
| E-clear: reset caches or metadata to undefined | 8 | M | E005, E010, E087, E091 |
| E-forward: copy optional values or factory arguments | 44 | M | E002, E031, E044, E092 |
| E-literal: explicit maybe-undefined object properties/options | 13 | M | E004, E008, E080, E094 |
| E-view: structural optional view mismatch | 1 | M | E082 |
| E-index: optional write fed by an indexed read | 2 | C | E022, E023 |
| E-generic: arbitrary specialization of a factory type | 1 | R | E070 |
| E-other: discriminator/union issue under an optional error code | 1 | R | E027 |
| **Total** | **100** | **96 M, 2 C, 2 R** | |

For E030, `factory/nodeFactory.ts:1308` writes `node.flowNode = undefined`
with the comment "initialized by binder". E010, `checker.ts:13369`, clears
`type.members`. E031, `nodeFactory.ts:1626`, assigns the optional factory argument
`defaultType` to `node.default`. These are observable writes, not omitted keys.
The appropriate representation is:

```ts
interface Slot {
    p?: Value | undefined;
}
// Keep both the original write and the original optionality.
slot.p = maybeValue;
```

`p?: Value` distinguishes absence from a present undefined value;
`p?: Value | undefined` allows both. Changing to `p: Value | undefined` would
require the key on every object, a different structural contract. Deleting the
key, conditionally spreading it, or skipping the assignment changes `in`,
`Object.hasOwn`, key enumeration, object spread, or a setter's effects. Keeping
the write and changing only its truthful declaration preserves those effects.
An `in` guard will then correctly stop treating a present key as sufficient to
prove a non-undefined value; consumers must be rechecked.

E080, `moduleNameResolver.ts:1393`, returns
`{ moduleResolution: ModuleResolutionKind.Node10, traceResolution: options.traceResolution }`.
Do not "fix" it by dropping `traceResolution` when undefined. E082,
`program.ts:880`, passes `SourceFileImportsList` to a `Pick<SourceFile, ...>`;
the error identifies `packageJsonScope: PackageJsonInfo | undefined` versus the
optional destination property. This is a declaration/view alignment candidate,
not a runtime failure.

E022, `checker.ts:33923`, assigns `childrenTypes[0]` after `length === 1`.
E023, `checker.ts:43517`, caches `getTypeArguments(type as GenericType)[0]`.
Adding undefined to the destination could remove TS2412 while leaving an
unproved element or generic-arity invariant. Keep these separate from the
representation-only recommendations.

E070, `nodeFactory.ts:5202`, assigns a general comment into a `T` slot where
`T extends JSDocTag`. Its chain explicitly says `T["comment"]` might be a narrower
specialization. A truthful base optional declaration alone does not establish
that arbitrary specialization is constructible by this factory. Review the
actual instantiations and generic construction contract.

E027, `emitter.ts:1014`, reports TS2379 on
`printer.writeFile(sourceFile!, writer, sourceMapGenerator)`, but its chain says
`Bundle | SourceFile` is not `SourceFile` and lists missing required fields.
The preceding code derives `bundle` and `sourceFile` from `sourceFileOrBundle.kind`.
This is not an optional-field mismatch. Reproduce the discriminator typing in
isolation after resolving the other meter walls; do not attribute it to exact
optionality or automatically weaken Printer's parameter type.

## Possibly undefined values: 100 findings

| Pattern | Count | Decision | Examples |
| --- | ---: | --- | --- |
| U-loop: indexed traversal elements | 27 | C | U001, U014, U062, U093 |
| U-parallel: related arrays, tuple flags, template segments | 26 | C | U003, U031, U035, U087 |
| U-endpoint: first/last/fixed positions and nonempty assumptions | 26 | C | U017, U056, U059, U075 |
| U-position: computed search, caller, or stack index | 10 | C | U051, U071, U078 |
| U-table: dynamic entry lookup | 2 | C | U063, U070 |
| U-grid: allocated rows and cells | 4 | C | U065, U066, U067, U068 |
| U-regex: required capture in a successful match | 2 | C | U073, U079 |
| U-zero: deliberate numeric undefined-to-zero coercion | 2 | M | U083, U096 |
| U-typed: typed-array computed position | 1 | C | U099 |
| **Total** | **100** | **2 M, 98 C** | |

U062, `commandLineParser.ts:1982`, dereferences `s = args[i]` in the argument
traversal. U017, `checker.ts:16708`, dereferences the first parameter after a
`length === 1` guard. These are ordinary dense arrays in the intended program,
but TypeScript's `T[]` also admits sparse arrays. `new Array(1)` has length one
and its first element is undefined. A length guard is not the missing proof.
Replacing the traversal with `for...of` solely to satisfy tsc can still visit
holes; it also changes index-dependent behavior and mutation interactions.

U003, `checker.ts:7432`, indexes `elementFlags[i]` while mapping type arguments.
U087, `transformers/es2017.ts:764`, reads `outerParameters[i]` after asserting
`i < outerParameters.length`. Matching lengths, density, and consistent phase
construction remain obligations. Prefer one checked read per required value,
keeping the original evaluation point. U071, `factory/utilities.ts:1316`, checks
`machine.onOperator` but dereferences `nodeStack[stackIndex]` separately; checking
the callback does not check the stack element.

U063, `commandLineParser.ts:2693`, repeats `computedOptions[option]` in a `for...in`
loop. Enumeration does not in general prove a total own-entry record. Hoisting
or consolidating repeated reads also requires stable ordinary storage: getters
or proxies can make repeated lookups observable. Do not automatically reuse a
value across intervening calls or assignments. U070, `emitter.ts:6337`, indexes
`brackets[format & ListFormat.BracketsMask][0]`; the mask and table layout need
proof or a check of the selected entry. U099, `utilities.ts:10504`, updates
`segments[segment + 1]` in a `Uint16Array`. Density is guaranteed here, but the
index arithmetic still needs its own bounds proof.

U073, `parser.ts:10733`, uses `matchResult[1].length` for a matched pragma
attribute. `getNamedArgRegEx` at line 10707 makes group 1 mandatory; groups 2 and
3 are alternative captures and can be undefined. U079, `sourcemap.ts:394`, trims
group 1 of its specific source-map regexp. A reviewed literal-regexp proof could
remove a redundant check in future work; a generic successful-match rewrite
cannot treat every capture as present. This survey recommends checked required
captures unless that specific regexp proof is implemented and tested.

U083, `transformer.ts:403`, is different. Its feature table is intentionally
created as `new Array<SyntaxKindFeatureFlags>(SyntaxKind.Count)` at line 249.
Unset cells mean no flags. `(flags ?? 0) & mask` preserves the original
`undefined & mask` behavior without claiming the slot is populated.
U096, `utilities.ts:7686`, reads past the final base64 input block, uses bitwise
operators to produce zero contributions, and then pads the affected output
digits. `?? 0` at those **numeric bitwise operands** preserves that behavior.
Do not generalize this rule to strings, property dereferences, comparisons,
ordinary addition, or any place where undefined has different semantics.

Two source-local hazards deserve particular review:

* **U051, `checker.ts:38421`:** `if (pos < paramCount)` does not exclude a negative
  position before `signature.parameters[pos].escapedName`. A negative number
  satisfies the declared number type and the local guard, then throws. Valid
  upstream callers may ensure nonnegative positions; that reachability was not
  established here.
* **U059, `checker.ts:52403`:** a present empty `node.typeParameters` array passes
  the truthiness test, is not length greater than one, and can reach
  `node.typeParameters[0].constraint`. Empty arrays are admitted by the array
  type. The reduced condition throws on `[]`; parser/grammar suppression may
  prevent that path in the real checker. This is evidence of an incomplete
  *local* contract, not a demonstrated TypeScript compiler crash.

No selected possibly-undefined finding was established to be just a lost local
narrowing with a proven, unconditional syntax-only repair. The length-guard
cases are checker limitations for dense arrays, but also reject genuinely unsafe
sparse arrays; those two observations must not be conflated.

A loud check catches either state before using a value with a false type. It
does not decide whether the correct source behavior should be an early return,
a diagnostic, or a stricter input contract. Keep that domain decision manual.

## Recommendation

Keep exactOptionalPropertyTypes and noUncheckedIndexedAccess unchanged.

1. Extend the meter's report to resolve optional-property declaration owners and
   offer **type-only** `| undefined` corrections for deliberate writes. Preserve
   `?`, runtime assignments, and object keys. Recheck the whole affected program;
   report any newly exposed consumer errors. Do not batch-edit every optional
   declaration or promise that these 96 findings equal 96 independent fixes.
2. Report required indexed reads as invariant obligations. A reviewer may choose
   a loud unwrap, prove a dense/nonempty or parallel-array construction, or add
   legitimate missing-value handling. Automated insertion of a panic is an
   explicitly selected checked adaptation, not a silent meaning-preserving one.
3. Keep semantic coercions narrow. The two numeric bitwise sites support a
   local `?? 0` rule with number-or-undefined operands and unchanged reads.
   A general default, optional chain, skip, cast, or non-null assertion hides
   the very obligation the strict checker found.
4. Treat code labels as triage hints. Expand diagnostic chains before suggesting
   rewrites, particularly TS2379 and generic factory errors. Preserve an
   unresolved category rather than turning missing evidence into an assertion.

The current evidence supports a source-adaptation workflow, not a language
relaxation. It does not establish how many entire files would compile after
these changes, whether every inferred invariant holds, or that all remaining
undefined diagnostics are bugs.

## Reproduction and validation

The population above records the meter at the pinned main revision. On this
branch, `--adapt` also performs the subsequent optional adaptation; the included
exporter defaults to import-only adaptation to reproduce the original sample.
Run from the Adamic repository. Use a fresh scratch directory rather than
installing dependencies in the repository:

```sh
source /workspace/adamic-tools/env.sh
S=$(mktemp -d /tmp/adamic-strictness.XXXXXX)
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git "$S/typescript" > "$S/clone.log" 2>&1
npm install --prefix "$S/npm" typescript@6.0.3 --no-audit --no-fund > "$S/npm.log" 2>&1
go run ./cmd/adamic-meter --adapt --json "$S/typescript/src/compiler" > "$S/meter.json" 2> "$S/meter.stderr"
```

The public meter aggregates diagnostics. To export individual locations without
editing `cmd/` or `internal/`, use the included Go test source through an overlay:

```sh
export ADAMIC_SURVEY_ROOT="$S/typescript/src/compiler"
export ADAMIC_SURVEY_OUT="$S/diagnostics.json"
python3 - "$S/overlay.json" <<'PY'
import json, pathlib, sys
root = pathlib.Path.cwd()
json.dump({'Replace': {
    str(root / 'cmd/adamic-meter/survey_export_test.go'):
    str(root / 'docs/tsc-strictness/export_test.go.txt')
}}, open(sys.argv[1], 'w'))
PY
go test -overlay="$S/overlay.json" ./cmd/adamic-meter -run '^TestExportStrictnessSurvey$' -count=1 -v > "$S/export.log" 2>&1
python3 docs/tsc-strictness/audit.py "$S/diagnostics.json" "$S/typescript" > "$S/audit.log" 2>&1
TSC_SURVEY_TYPESCRIPT="$S/npm/node_modules/typescript/lib/typescript.js" node docs/tsc-strictness/probes.cjs > "$S/probes.log" 2>&1
go test ./cmd/adamic-meter > "$S/meter-tests.log" 2>&1
```

Observed output:

```text
Export: PASS, ok github.com/system-inc/adamic/cmd/adamic-meter 6.542s
Audit: E population 724 sample 100; M 96, C 2, R 2
Audit: U population 578 sample 100; M 2, C 98
PASS: all 200 IDs, source locations, source lines, diagnostic chains and quantile ranks verified
PASS: 5 stock TypeScript checks; identical emitted JS for honest optional type; Node presence, holes, bounds, flags, getters and captures; 3 mutants caught
ok github.com/system-inc/adamic/cmd/adamic-meter 0.292s
```

[The reduced probes](tsc-strictness/probes.cjs) establish stock tsc rejection of
present undefined without the explicit union, acceptance after the honest type
change, byte-identical emitted JavaScript, and continued rejection of a wrong
number value. They also show that the bounded first read still needs a check,
that its checked version typechecks, and that the sparse/negative/empty reduced
conditions actually fail on Node. These are pattern probes, not a compiled
Adamic port or an execution of all TypeScript compiler paths.

Three deliberately wrong candidate rewrites were run and caught by assertions,
not by compilation failure:

| Mutant | Catch |
| --- | --- |
| Delete an optional slot instead of storing undefined | Node own-key/presence/spread fingerprint differs |
| Remove a required-value check and return the input unchecked | Expected missing-value throw disappears |
| Default every missing required value to zero | Expected missing-value throw disappears |

All test output went to scratch log files and was read afterward. A full
native/oracle gate was not run for this documentation and scratch-tooling-only
change. No compiler fixtures, counts table, language rules, or adapter were
modified. Toolchain setup succeeded: Go 0s, clang 0s, Node 0s, submodules 0s,
build-cache warm 11s, total 11s; `nproc` reported 5 (cgroup CPU quota 4).
