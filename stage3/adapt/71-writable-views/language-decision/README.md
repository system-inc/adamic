# Retained writable views: language-decision evidence

This audit classifies all 615 retained observations at adaptation commit
`12fef29628bfda6cf8bec4ba2a326fa4ec4a66d5`. It changes no adaptation,
compiler source, runtime computation, public declaration, or API sanction.

| Retained family | (a) reads only | (b) compatible writes | (c) counterexample | (d) unresolved | Total |
| --- | ---: | ---: | ---: | ---: | ---: |
| Shared `never[]` | 0 | 0 | 0 | 82 | 82 |
| Other views | 91 | 5 | 2 | 272 | 370 |
| Diagnostics | 0 | 2 | 0 | 161 | 163 |
| All | 91 | 7 | 2 | 515 | 615 |

[results.json](results.json) contains every site, original census reason, receiving
expression, class, reads, writes, local aliases, resolved calls, and unresolved
reasons. IDs follow the immutable
[retained census list](../evidence/followup/remaining-selected.json).
[counts.csv](counts.csv) gives counts for every source-type family as well as the
three requested groups. [c-sites.json](c-sites.json) enumerates every (c) site.

## Method and limits

`audit.cjs` uses stock TypeScript 6.0.3's compiler API over the entire implementation
program: 603 implementation files, 90 declaration files, and 115,001 indexed
symbols. References use resolved symbols, including shorthand properties. For
each receiving expression it follows local aliases, nested property aliases,
statically unique function bodies, actual argument bindings, and returned aliases
back into their concrete caller. It checks field/index write types against the
original holder's read domains. Erased assertions on RHS values are removed:
`undefined!` still writes `undefined`.

(a) means the analyzed receiving graph closes without a write or unresolved
escape; (b) closes with writes all compatible with the original declared read
domains; (c) has a disjoint write or an executable legal generic instantiation
that violates its holder's promise; (d) records missing information. Generic
constraints alone cannot establish (b): a holder may narrow a constraint field.

The analysis conservatively stops at stored aliases, callbacks, virtual/external
consumers, accessors, reassigned bindings, destructuring, unsupported syntax,
ambiguous value/branch/union facts, recursion, 16 call levels, or 600 traversal
steps. Array mutators have explicit stock-library models. It does not prove
arbitrary heap aliasing or runtime reachability of the census observations.
Plain declared properties and the modeled standard-library operations are its
semantic assumptions. A latent refusal's wider abstract type is not itself proof
of an incompatible runtime write. Unresolved cases stay (d), even when a likely
reader or writer is visible.

Many (a) observations are truthiness expressions such as `object && otherValue`:
the object is tested but never transferred into the result. Those are evidence
about the observed receiving expression, not an endorsement of writable variance.

## Examples

All locations below refer to the pinned adapted `src/compiler/` tree.

Three (a) examples:

- `emitter.ts:3499:112` (ID 415): the `childrenTextRange` argument reaches
  `getClosingLineTerminatorCount`; its receiver only reads range end/presence.
- `emitter.ts:4711:127` (418): the range goes through `emitListWorker` and the same
  closing-line helper; the receiver graph only reads it.
- `emitter.ts:5461:90` (420): generated identifier passed to `isPrivateIdentifier`;
  that receiver only reads `kind`.

Three (b) examples:

- `checker.ts:37379:53` (262): diagnostic recovery assigns/pushes
  `relatedInformation` values accepted by the original diagnostic field.
- `checker.ts:51498:32` (352): the mutable index-signature view initializes
  `modifiers` with a compatible `NodeArray<ModifierLike>`.
- `parser.ts:4286:54` (514): the property-signature view assigns an initializer
  in its original `Expression | undefined` domain.

Every (c) site (only two belong to the 615-site cohort):

- `utilities.ts:10689:10` (603), `setNodeFlags`: `(node as Mutable<T>).flags =
  newFlags`. A legal `T` with `flags: NodeFlags.Synthesized` (16) is returned as
  the same `T` after writing `NodeFlags.None` (0). Its guard only tests that the
  node exists. The census mentions readonly `pos`; the actual violating field
  here is **flags**.
- `utilities.ts:12365:6` (606), deep clone with replacements: `(visited as
  Mutable<T>).parent = undefined!`. A binary expression with a present SourceFile
  parent is cloned and returned as `T`, but its returned **parent** is undefined.
  The original object's parent remains intact. The census again names `pos`,
  while the actual incompatible write is to parent.

[controls/tsc-writers.a](controls/tsc-writers.a) checks both with zero stock-TS
diagnostics and executes the actual unchanged built tsc exports under Node.
Observed output: `16 0 true undefined`. This proves legal-call counterexamples;
it does not assert that the baseline compiler tests make those calls.

For the requested third (c) illustration, **outside the retained cohort**,
`utilities.ts:10645:5` writes `pos` through a `TextRange` view of a readonly literal
position holder. The existing [text-range witness](../language-questions/text-range.a)
promises `0` and observes `1`. It is not counted as a third retained site.

Three (d) examples, one per retained family:

- `builder.ts:2044:49` (37, shared never[]): the sentinel is stored in a diagnostics
  object. Its subsequent container aliases require heap analysis.
- `binder.ts:1034:95` (1, other): a flow node is stored in `endFlowNode`; the wider
  field's later readers/writers and their aliases are not resolved by this graph.
- `binder.ts:853:49` (34, diagnostics): the diagnostic is pushed through a method
  consumer; element/container aliases remain unresolved.

## Validation and reproduction

Use the already adapted and built tree from unit 71's lane, or reproduce that tree
with `stage3/apply.sh` and its existing build workflow. Set `NODE_PATH` to the pinned
stock TypeScript 6.0.3 installation. From the repository root:

```sh
node stage3/adapt/71-writable-views/language-decision/audit.cjs "$TREE" stage3/adapt/71-writable-views/evidence/followup/remaining-selected.json stage3/adapt/71-writable-views/language-decision/results.json > audit.log 2>&1
node stage3/adapt/71-writable-views/language-decision/validate.cjs > validation.log 2>&1
node stage3/adapt/71-writable-views/language-decision/writers.cjs "$BUILT_TREE/built/local/typescript.js" > writers.log 2>&1
python3 stage3/adapt/71-writable-views/language-decision/verify.py "$TREE" "$BUILT_TREE/built/local/typescript.js" > verification.log 2>&1
node stage3/adapt/71-writable-views/witness.cjs > position.log 2>&1
node stage3/adapt/71-writable-views/witness.cjs --mutant > position-mutant.log 2>&1
```

The last command must exit 1; all others exit 0. `.a` controls are compiler-API
inputs mapped to virtual TypeScript paths, not newly registered Adamic fixtures.
No fixture registry or `counts.md` changes are required.

[validation.json](validation.json) records 16 receiving controls, a generic-bound
control, Node observations, and four caught input mutants: reader starts writing,
compatible writer writes outside its domain, incompatible writer becomes
compatible, and shorthand container escape becomes a scalar read. Each mutant
exits 1 at its classification assertion while stock-TS diagnostics remain zero.
[writers.json](writers.json) records two actual-library repair mutants: preserving
flags and preserving parent each makes its counterexample assertion fail (exit 1).
The position witness's write-to-zero mutant also fails its Node observation
(exit 1, observed `0` instead of `1`).

[verification.json](verification.json) verifies exact coverage, all family/count
summaries, the 603 physical source hashes against the previous source pins, and
unchanged built library/public declaration hashes. Removing a retained site and
changing a source-file comment both fail their respective verification checks.
These are nine caught mutants altogether. No adaptation/source changed, so the
previous lane/oracle results remain historical evidence; neither was rerun for
this evidence-only commit. No full gate or whole Go package was run.
