# isPropertyNameLiteral claims what its body accepts (#ad161ms)

Adamic ruled on task #ad161ms. This is a fork fix. It changes JavaScript only
where a bigint literal names a class member or property that a transformer
names at run time, and every such input crashes stock tsc 6.0.3. Four forms
crash it, at es2022 and es2015 unless named:

```ts
declare function dec(value: unknown, context: unknown): any;
class C { @dec 1n = 0; }
class D { @dec 1n() {} }
class E { @dec static 1n = class {}; }
const o = { 1n: class { static x = 1; } }; // es2015, no decorator
```

`isPropertyNameLiteral` in `src/compiler/utilities.ts` returns true for four
kinds: `Identifier`, `StringLiteral`, `NoSubstitutionTemplateLiteral` and
`NumericLiteral`. Its predicate claimed `node is PropertyNameLiteral`, which
also holds `BigIntLiteral` and `JsxNamespacedName`. Three callers took the
false side for a `ComputedPropertyName` and read `name.expression`, which a
bigint name doesn't have:

- `partialTransformClassElement` in `esDecorators.ts`, the name a decorator's
  context carries;
- `visitReferencedPropertyName` in `esDecorators.ts`;
- `getAssignedNameOfPropertyName` in `namedEvaluation.ts`, the name an
  anonymous class gets from the key it's assigned to.

The predicate now names the three kinds its body accepts:

```ts
export function isPropertyNameLiteral(node: Node): node is Identifier | StringLiteralLike | NumericLiteral {
```

With that claim, stock tsc reports 11 errors at exactly those three callers
(J-ssn5bf6). Each caller now gives a bigint name its property key as a string
literal, `factory.createStringLiteral(parsePseudoBigInt(name.text))`, which is
the decimal digits a `[1n]` key gets from `__propKey` at run time. The literal
branch's `createStringLiteralFromNode` doesn't fit a bigint: its emit falls
through to the bigint's own text, so the name would be the bigint `1n`, not the
key `"1"`. Both transformer files import `isBigIntLiteral` and
`parsePseudoBigInt` in their lists' own order. `utilities.ts` already imports
the claim's three names, and still names `PropertyNameLiteral` elsewhere.

## The adapter

`adapt.cjs` finds each owner through the stock TypeScript 6.0.3 AST and pins
its whole text, upstream or adapted: `isPropertyNameLiteral`, the if statement
that names a decorated member in `partialTransformClassElement`,
`visitReferencedPropertyName`, and `getAssignedNameOfPropertyName`. It refuses
any other text, a parse error, mixed line endings, and a partly adapted tree,
meaning any owner or import name adapted while another isn't. A second run
writes nothing. `rewrite(tree, { callers: false })` is the mutant: it narrows
the claim and leaves the callers and imports as they were.

The directory sorts after 99-correctness-receiver-transformers, which pins
`transformESDecorators` and esDecorators.ts's namespace import whole by hash,
so it would refuse either one edited before it. No directory after this one
touches `isPropertyNameLiteral`, esDecorators.ts or namedEvaluation.ts.

## What is proven

**The whole-oracle count, J-pbbkng3.** An instrumented copy of the adapted
tree recorded the kind of every name reaching each caller, on each side of its
`isPropertyNameLiteral` test, and ran the lane oracle whole: upstream's suite,
all runners, eight workers, no filter. The suite read 106,366 passing, 1
failing and 0 pending. That's the lane's expectation: the one failure is the
sanctioned Public APIs test, and the one baseline diff is `api/typescript.d.ts`.
The callers were reached 327 times:

| Caller | Calls | Literal side | False side |
| --- | --- | --- | --- |
| partialTransformClassElement | 164 | 0 | 164 ComputedPropertyName |
| visitReferencedPropertyName | 87 | 0 | 87 ComputedPropertyName |
| getAssignedNameOfPropertyName | 76 | 46 (Identifier 22, StringLiteral 12, NumericLiteral 8, PrivateIdentifier 4) | 30 ComputedPropertyName |

No `BigIntLiteral` and no `JsxNamespacedName` reached any of them. The new
branches are never taken, so every baseline computes the same. The same
recorder caught a `BigIntLiteral` at partialTransformClassElement and at
getAssignedNameOfPropertyName on the crash programs, before each one threw.

**The fix, J-d41h4ew, with the adapter after the whole chain.** Stock tsc
6.0.3 reads 0 over the fixed tree. The mutant narrows the claim and leaves the
callers, and stock tsc reports exactly 11 errors, all in esDecorators.ts and
namedEvaluation.ts. The public `typescript.d.ts` and `tsserverlibrary.d.ts`
are byte-identical. `typescript.internal.d.ts` moves by the one
`isPropertyNameLiteral` line.

**The commit's proof job** applies the adaptations twice into fresh trees,
compares them byte for byte and runs the adapter a second time for no writes.
It runs stock tsc for 0, and holds the declarations to that rule. It pins
the four crash forms: stock tsc throws and writes nothing, and the adapted
tree's tsc writes output. At run time, that output gives the decorator
context names and the assigned class name that stock tsc's output gives for
the same program with `[1n]` in place of `1n`. That's `"1"`, and the empty
name stock gives a decorated static anonymous class. It runs the mutant again.

## Draft upstream issue

Not filed. Kirk decides whether and when.

Title: Decorated class member named by a bigint literal crashes emit in TypeScript 6.0.3

Version: TypeScript 6.0.3 (commit 050880ce59e30b356b686bd3144efe24f875ebc8).

Reproduction, `repro.ts`, run as `tsc --target es2022 repro.ts` (es2015 too):

```ts
declare function dec(value: unknown, context: unknown): any;
class C { @dec 1n = 0; }
```

The method form `class C { @dec 1n() {} }` and the static form
`class C { @dec static 1n = class {}; }` crash the same way. So does an object
literal with no decorator, `const o = { 1n: class { static x = 1; } };`, at
`--target es2015`, in `getAssignedNameOfPropertyName`.

Expected: emitted JavaScript whose decorator context names the member `"1"`,
and whose `o[1].name` is `"1"`, as the computed forms `@dec [1n] = 0` and
`{ [1n]: class { static x = 1; } }` give. The field forms report TS1539, "A
'bigint' literal cannot be used as a property name", and the undecorated
`class C { 1n = 0; }` reports it and emits.

Actual: no output file, and

```text
TypeError: Cannot read properties of undefined (reading 'kind')
    at isPropertyNameLiteral
    at partialTransformClassElement
    at visitPropertyDeclaration
```

Cause: `isPropertyNameLiteral` (src/compiler/utilities.ts) declares
`node is PropertyNameLiteral`, but its body returns false for `BigIntLiteral`
and `JsxNamespacedName`. `partialTransformClassElement` and
`visitReferencedPropertyName` in esDecorators.ts and
`getAssignedNameOfPropertyName` in namedEvaluation.ts treat that false side as
a `ComputedPropertyName` and read `name.expression`, which is undefined for a
bigint name.

Suggested change: declare the predicate as
`node is Identifier | StringLiteralLike | NumericLiteral`. The compiler then
reports each of these three callers. At each, give a bigint name its property
key, `factory.createStringLiteral(parsePseudoBigInt(name.text))`.
