# Parser scout census — step 25 (#pw720wc)

Base: `ad7bd06632f1`, origin/area/stage1-lint; recovery ancestor `88f4a83d`. No parser changes. The comparison runs the actual port TypeScript on Node through the repository runtime hook and the unmodified pinned typescript-go parser through a Go overlay. This census does not claim native Adamic execution parity.

Canonical byte identity means preorder node kind, byte start/end, full node flags, literal flags, ordered children (depth), list metadata, operator, cooked/raw text, semantic fields, and parser diagnostics (code, byte position/length, category, text). Positions use UTF-8 bytes on both sides. Physical TypeScript test fixtures with @filename directives are parsed as physical inputs, not split into virtual harness files. JSDoc attachments are represented by flags; full lazy JSDoc trees are outside `ForEachChild` and need a separate docTypes comparison. Port storage represents OptionalChain and block-scoped declaration bits; other node flags have no stored counterpart and print as unset. Nothing is silently masked in the full comparison.

Classes overlap: a file may occur in several rows. Flag attribution requires matching preorder depth, kind, and span: when structure shifts, mismatched-span nodes are counted structurally rather than inventing a flag class from unrelated nodes. All 698 structurally divergent public files were replayed with this stricter rule. Structural comparison explicitly removes only node flags. “Shortest” means the shortest complete selected corpus file exhibiting that class, not a delta-minimized program. The full input and both complete trees are retained beside the first witness. Function names for flags identify missing production/context propagation. Structural attribution is provisional where several parser paths differ; use the retained tree to inspect the exact node.

Adapter: `census/oracle.go.txt` is a copy of the gate oracle with full flags, safe unterminated no-substitution template slicing, and per-file panic capture. Five incomplete-template probes parse and print on both sides: no adapter failures. The one-character probe differs only in ThisNodeHasError. The two-character input consisting of a backtick followed by the letter "a" also reveals a raw-text difference: Go retains `a`, while the port drops it. Responsible: Parser.literal/templateToken raw slicing. See census/templates.json and complete trees. Failure records are excluded from completed comparison denominators. The port worker marks exceptions during parser.file as parse-failure, exceptions during serialization as adapter-failure, and interrupts an individual parse after five seconds. Failure records preserve the side and reason, rather than fabricating an AST difference. Go panic capture is adapter-failure; inspect the stack/error before attributing it to parser semantics.

Reproduction: fetch cohere and its pinned TypeScript checkout, build with `python3 census/build.py /tmp/parser-census-oracle /workspace/adamic-tools/go/bin/go`, then `python3 census/run.py census/tsc.manifest census/tsc.json /tmp/parser-census-oracle`. `fetch.py` reads `public-pins.tsv`, shallow-fetches each exact SHA, executes no dependency installation or repository scripts, and writes `public.manifest`. Run the same driver on that manifest. Paths in manifests reflect this scout's `/workspace` layout.

## Entry points without an inherited list context

`expression()`, `assignment()`, `rootExpression()`, and `rootAssignment()` are public methods and do not establish an enclosing source/block list. The root wrappers only manage expressionDepth and roots. Called on a fresh Parser they can all recover differently from file parsing, for the same reason as tsprinter: recoverList consults the active outer listContexts to decide when to stop. Even legal `({m(){}})` is affected: speculative arrow-parameter recovery with no outer source context produces ArrowFunction and leaves CloseBraceToken; with source context it produces ParenthesizedExpression and reaches EndOfFile. This is measured for all six expression entry points in `census/entries.json`. The responsible path is Parser.arrowCandidate / arrowHeadKind and binding-parameter recovery in delimitedList/recoverList. Their ordinary calls from file/statements retain an inherited context.

Additional public paths `allowInAssignment()` and `allowInExpression()` wrap assignment/expression and only change disallowIn; both reproduce the exact same method-shorthand divergence on the probe. `type()`, `returnType()`, and `docTypes()` can also start with an empty outer list. `docTypes()` scans raw @type substrings, panics on a missing brace, and uses returnType rather than Go's complete JSDoc grammar; malformed types can diverge, but this is a type/JSDoc entry issue rather than an object-method-specific claim.

External production callers inspected: tsprinter `formatExpression` directly calls expression (the known pinned issue); tsprinter files, lint options/settings, lint reparsing, comments, and ESTree pipeline use file(), which establishes source context. Typeaware constructs Parser and delegates file parsing to its linter. JSX expression calls inherit source/list contexts during file parsing. No second external fresh-Parser expression caller was found in stage1. Malformed JSDoc probes in `census/docs.json` include `/** @type { */ let x;`: Go produces a missing TypeReference/Identifier while docTypes produces JSDocAllType and diagnostic 1005. For `/** @type */ let x;`, Go recovers a missing type while the port exits 70. The report refers to the pinned branch; formatter fix `06c6e7cbf` is not applied here.

## tsc

Files selected: **77**; both sides completed: **77**; byte-identical canonical trees and diagnostics with all flags: **0**; identical ignoring node flags: **77**; adapter/runner failures: **0**.

| Divergence class | Files affected | Shortest input bytes | Parser function / context |
|---|---:|---:|---|
| flags/HasJSDoc | 73 | 382 | Parser.make / Go finishNode and JSDoc attachment |
| flags/DisallowInContext | 51 | 5424 | Parser.rootAssignment / allowInAssignment / forStatement context propagation |
| flags/AwaitContext | 20 | 101 | Parser.file (top-level module context), methodBody / arrow / awaitContext propagation |
| flags/PossiblyContainsDeprecatedTag | 6 | 36796 | Parser.make / JSDoc comment scanning |
| flags/Ambient | 5 | 3255 | Statements.declaration / declare and declaration-file context |
| flags/PossiblyContainsDynamicImport | 3 | 3255 | Parser.primary / import expression and SourceFile aggregation |
| flags/YieldContext | 2 | 92419 | Parser.methodBody / arrow / yieldContext propagation |

### flags/HasJSDoc

Shortest corpus input: `/workspace/census-typescript-6/src/compiler/builderStatePublic.ts` (382 bytes). [Exact input](census/tsc/examples/flags-HasJSDoc.ts); complete preorder trees: [typescript-go](census/tsc/examples/flags-HasJSDoc.go.tree.gz), [port](census/tsc/examples/flags-HasJSDoc.port.tree.gz).

First witness (Go against port):
```text
Go: 2 PropertySignature 325 377 2097152 0 -1 0 0
Port: 2 PropertySignature 325 377 0 0 -1 0 0
```
```typescript
import {
    Diagnostic,
    WriteFileCallbackData,
} from "./_namespaces/ts.js";

export interface EmitOutput {
    outputFiles: OutputFile[];
    emitSkipped: boolean;
    diagnostics: readonly Diagnostic[];
}

export interface OutputFile {
    name: string;
    writeByteOrderMark: boolean;
    text: string;
    /** @internal */ data?: WriteFileCallbackData;
}

```

### flags/DisallowInContext

Shortest corpus input: `/workspace/census-typescript-6/src/compiler/transformers/taggedTemplate.ts` (5424 bytes). [Exact input](census/tsc/examples/flags-DisallowInContext.ts); complete preorder trees: [typescript-go](census/tsc/examples/flags-DisallowInContext.go.tree.gz), [port](census/tsc/examples/flags-DisallowInContext.port.tree.gz).

First witness (Go against port):
```text
Go: 7 VariableDeclaration 2159 2172 1024 0 -1 0 0
Port: 7 VariableDeclaration 2159 2172 0 0 -1 0 0
```

### flags/AwaitContext

Shortest corpus input: `/workspace/census-typescript-6/src/compiler/_namespaces/ts.performance.ts` (101 bytes). [Exact input](census/tsc/examples/flags-AwaitContext.ts); complete preorder trees: [typescript-go](census/tsc/examples/flags-AwaitContext.go.tree.gz), [port](census/tsc/examples/flags-AwaitContext.port.tree.gz).

First witness (Go against port):
```text
Go: 2 StringLiteral 78 98 8192 0 -1 0 0		../performance.js
Port: 2 StringLiteral 78 98 0 0 -1 0 0		../performance.js
```
```typescript
/* Generated file to emulate the ts.performance namespace. */

export * from "../performance.js";

```

### flags/PossiblyContainsDeprecatedTag

Shortest corpus input: `/workspace/census-typescript-6/src/compiler/factory/nodeTests.ts` (36796 bytes). [Exact input](census/tsc/examples/flags-PossiblyContainsDeprecatedTag.ts); complete preorder trees: [typescript-go](census/tsc/examples/flags-PossiblyContainsDeprecatedTag.go.tree.gz), [port](census/tsc/examples/flags-PossiblyContainsDeprecatedTag.port.tree.gz).

First witness (Go against port):
```text
Go: 1 FunctionDeclaration 25585 25729 69206016 0 -1 0 0
Port: 1 FunctionDeclaration 25585 25729 0 0 -1 0 0
```

### flags/Ambient

Shortest corpus input: `/workspace/census-typescript-6/src/compiler/performanceCore.ts` (3255 bytes). [Exact input](census/tsc/examples/flags-Ambient.ts); complete preorder trees: [typescript-go](census/tsc/examples/flags-Ambient.go.tree.gz), [port](census/tsc/examples/flags-Ambient.port.tree.gz).

First witness (Go against port):
```text
Go: 1 VariableStatement 736 852 8388608 0 -1 0 0
Port: 1 VariableStatement 736 852 0 0 -1 0 0
```

### flags/PossiblyContainsDynamicImport

Shortest corpus input: `/workspace/census-typescript-6/src/compiler/performanceCore.ts` (3255 bytes). [Exact input](census/tsc/examples/flags-PossiblyContainsDynamicImport.ts); complete preorder trees: [typescript-go](census/tsc/examples/flags-PossiblyContainsDynamicImport.go.tree.gz), [port](census/tsc/examples/flags-PossiblyContainsDynamicImport.port.tree.gz).

First witness (Go against port):
```text
Go: 0 SourceFile 0 3255 524288 0 -1 0 0
Port: 0 SourceFile 0 3255 0 0 -1 0 0
```

### flags/YieldContext

Shortest corpus input: `/workspace/census-typescript-6/src/compiler/core.ts` (92419 bytes). [Exact input](census/tsc/examples/flags-YieldContext.ts); complete preorder trees: [typescript-go](census/tsc/examples/flags-YieldContext.go.tree.gz), [port](census/tsc/examples/flags-YieldContext.port.tree.gz).

First witness (Go against port):
```text
Go: 2 Parameter 11283 11300 2048 0 -1 0 0
Port: 2 Parameter 11283 11300 0 0 -1 0 0
```

## gate

Files selected: **300**; both sides completed: **300**; byte-identical canonical trees and diagnostics with all flags: **202**; identical ignoring node flags: **298**; adapter/runner failures: **0**.

| Divergence class | Files affected | Shortest input bytes | Parser function / context |
|---|---:|---:|---|
| flags/Ambient | 65 | 60 | Statements.declaration / declare and declaration-file context |
| flags/AwaitContext | 21 | 115 | Parser.file (top-level module context), methodBody / arrow / awaitContext propagation |
| flags/DecoratorContext | 9 | 154 | Parser.decorator |
| flags/DisallowInContext | 9 | 87 | Parser.rootAssignment / allowInAssignment / forStatement context propagation |
| flags/YieldContext | 6 | 59 | Parser.methodBody / arrow / yieldContext propagation |
| flags/HasJSDoc | 3 | 101 | Parser.make / Go finishNode and JSDoc attachment |
| flags/DisallowConditionalTypesContext | 3 | 247 | Parser.type / conditional type context |
| flags/Reparsed | 2 | 97 | Statements.moduleDeclaration / synthesized dotted-module export token |
| kind/children | 2 | 413 | Parser.primary / Statements.statement / recovery dispatch (see first tree divergence) |
| diagnostics | 2 | 413 | Parser.error / expect / recoverList (see witness kind) |
| position | 1 | 413 | Parser.make / token / scanner.fullStart (see witness kind) |
| children/count | 1 | 742 | Parser.file / Statements.statement / delimitedList (see first tree divergence) |

### flags/Ambient

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/005_ambientEnumElementInitializer3/ambientEnumElementInitializer3.ts` (60 bytes). [Exact input](census/gate/examples/flags-Ambient.ts); complete preorder trees: [typescript-go](census/gate/examples/flags-Ambient.go.tree.gz), [port](census/gate/examples/flags-Ambient.port.tree.gz).

First witness (Go against port):
```text
Go: 1 EnumDeclaration 0 60 8388608 0 -1 0 0
Port: 1 EnumDeclaration 0 60 0 0 -1 0 0
```
```typescript
// @target: es2015
declare enum E {
 e = 3.3 // Decimal
}
```

### flags/AwaitContext

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/042_exportDefaultForNonInstantiatedModule/exportDefaultForNonInstantiatedModule.ts` (115 bytes). [Exact input](census/gate/examples/flags-AwaitContext.ts); complete preorder trees: [typescript-go](census/gate/examples/flags-AwaitContext.go.tree.gz), [port](census/gate/examples/flags-AwaitContext.port.tree.gz).

First witness (Go against port):
```text
Go: 2 Identifier 112 114 8192 0 -1 0 0		m
Port: 2 Identifier 112 114 0 0 -1 0 0		m
```
```typescript
// @target: ES6

namespace m {
    export interface foo {
    }
}
// Should not be emitted
export default m;
```

### flags/DecoratorContext

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/251_decoratorMetadataGenericTypeVariable/decoratorMetadataGenericTypeVariable.ts` (154 bytes). [Exact input](census/gate/examples/flags-DecoratorContext.ts); complete preorder trees: [typescript-go](census/gate/examples/flags-DecoratorContext.go.tree.gz), [port](census/gate/examples/flags-DecoratorContext.port.tree.gz).

First witness (Go against port):
```text
Go: 4 Identifier 119 127 12288 0 -1 0 0		Decorate
Port: 4 Identifier 119 127 0 0 -1 0 0		Decorate
```
```typescript
// @target: es2015
// @experimentalDecorators: true
// @emitDecoratorMetadata: true

export class C<TypeVariable> {
  @Decorate
  member: TypeVariable;
}

```

### flags/DisallowInContext

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/127_forInStatement4/forInStatement4.ts` (87 bytes). [Exact input](census/gate/examples/flags-DisallowInContext.ts); complete preorder trees: [typescript-go](census/gate/examples/flags-DisallowInContext.go.tree.gz), [port](census/gate/examples/flags-DisallowInContext.port.tree.gz).

First witness (Go against port):
```text
Go: 3 VariableDeclaration 63 73 1024 0 -1 0 0
Port: 3 VariableDeclaration 63 73 0 0 -1 0 0
```
```typescript
// @target: es2015
// @strict: false
var expr: any;
for (var a: number in expr) {
}
```

### flags/YieldContext

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/257_generatorES6_5/generatorES6_5.ts` (59 bytes). [Exact input](census/gate/examples/flags-YieldContext.ts); complete preorder trees: [typescript-go](census/gate/examples/flags-YieldContext.go.tree.gz), [port](census/gate/examples/flags-YieldContext.port.tree.gz).

First witness (Go against port):
```text
Go: 2 Block 32 59 2048 0 -1 0 0
Port: 2 Block 32 59 0 0 -1 0 0
```
```typescript
// @target: es6
function* foo() {
    yield a ? b : c;
}
```

### flags/HasJSDoc

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/003_commentOnClassAccessor1/commentOnClassAccessor1.ts` (101 bytes). [Exact input](census/gate/examples/flags-HasJSDoc.ts); complete preorder trees: [typescript-go](census/gate/examples/flags-HasJSDoc.go.tree.gz), [port](census/gate/examples/flags-HasJSDoc.port.tree.gz).

First witness (Go against port):
```text
Go: 2 GetAccessor 29 98 2097152 0 -1 0 0
Port: 2 GetAccessor 29 98 0 0 -1 0 0
```
```typescript
// @target: es2015
class C {
  /**
   * @type {number}
   */
  get bar(): number { return 1;}
}
```

### flags/DisallowConditionalTypesContext

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/206_recursiveResolveTypeMembers/recursiveResolveTypeMembers.ts` (247 bytes). [Exact input](census/gate/examples/flags-DisallowConditionalTypesContext.ts); complete preorder trees: [typescript-go](census/gate/examples/flags-DisallowConditionalTypesContext.go.tree.gz), [port](census/gate/examples/flags-DisallowConditionalTypesContext.port.tree.gz).

First witness (Go against port):
```text
Go: 3 FunctionType 122 190 16384 0 -1 0 0
Port: 3 FunctionType 122 190 0 0 -1 0 0
```
```typescript
// @target: es2015
// Repro from #25291

type PromisedTuple<L extends any[], U = (...args: L) => void> =
    U extends (h: infer H, ...args: infer R) => [Promise<H>, ...PromisedTuple<R>] ? [] : []

type Promised = PromisedTuple<[1, 2, 3]>

```

### flags/Reparsed

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/031_commentInNamespaceDeclarationWithIdentifierPathName/commentInNamespaceDeclarationWithIdentifierPathName.ts` (97 bytes). [Exact input](census/gate/examples/flags-Reparsed.ts); complete preorder trees: [typescript-go](census/gate/examples/flags-Reparsed.go.tree.gz), [port](census/gate/examples/flags-Reparsed.port.tree.gz).

First witness (Go against port):
```text
Go: 3 ExportKeyword 39 39 8 0 -1 0 0
Port: 3 ExportKeyword 39 39 0 0 -1 0 0
```
```typescript
﻿// @target: es2015
namespace hello.hi.world
{
    function foo() {}

    // TODO, blah
}
```

### kind/children

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/074_parseInvalidNullableTypes/parseInvalidNullableTypes.ts` (413 bytes). [Exact input](census/gate/examples/kind-children.ts); complete preorder trees: [typescript-go](census/gate/examples/kind-children.go.tree.gz), [port](census/gate/examples/kind-children.port.tree.gz).

First witness (Go against port):
```text
Go: 4 JSDocNullableType 289 297 0 0 -1 0 0
Port: 4 NumberKeyword 289 296 0 0 -1 0 0
```
```typescript
// @target: es2015
// @strict: true

function f1(a: string): a is ?string {
    return true;
}

function f2(a: string?) {}
function f3(a: number?) {}

function f4(a: ?string) {}
function f5(a: ?number) {}

function f6(a: string): ?string {
    return true;
}

const a = 1 as any?;
const b: number? = 1;

const c = 1 as ?any;
const d: ?number = 1;

let e: unknown?;
let f: never?;
let g: void?;
let h: undefined?;

```

### diagnostics

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/074_parseInvalidNullableTypes/parseInvalidNullableTypes.ts` (413 bytes). [Exact input](census/gate/examples/diagnostics.ts); complete preorder trees: [typescript-go](census/gate/examples/diagnostics.go.tree.gz), [port](census/gate/examples/diagnostics.port.tree.gz).

First witness (Go against port):
```text
Go: []
Port: ["diagnostic 1005 296 1 1\t',' expected.", 'diagnostic 1134 298 1 1\tVariable declaration expected.', 'diagnostic 1134 300 1 1\tVariable declaration expected.']
```
```typescript
// @target: es2015
// @strict: true

function f1(a: string): a is ?string {
    return true;
}

function f2(a: string?) {}
function f3(a: number?) {}

function f4(a: ?string) {}
function f5(a: ?number) {}

function f6(a: string): ?string {
    return true;
}

const a = 1 as any?;
const b: number? = 1;

const c = 1 as ?any;
const d: ?number = 1;

let e: unknown?;
let f: never?;
let g: void?;
let h: undefined?;

```

### position

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/074_parseInvalidNullableTypes/parseInvalidNullableTypes.ts` (413 bytes). [Exact input](census/gate/examples/position.ts); complete preorder trees: [typescript-go](census/gate/examples/position.go.tree.gz), [port](census/gate/examples/position.port.tree.gz).

First witness (Go against port):
```text
Go: 1 VariableStatement 280 302 0 0 -1 0 0
Port: 1 VariableStatement 280 299 0 0 -1 0 0
```
```typescript
// @target: es2015
// @strict: true

function f1(a: string): a is ?string {
    return true;
}

function f2(a: string?) {}
function f3(a: number?) {}

function f4(a: ?string) {}
function f5(a: ?number) {}

function f6(a: string): ?string {
    return true;
}

const a = 1 as any?;
const b: number? = 1;

const c = 1 as ?any;
const d: ?number = 1;

let e: unknown?;
let f: never?;
let g: void?;
let h: undefined?;

```

### children/count

Shortest corpus input: `/workspace/parser-scout/stage3/drivers/tsc/corpus/239_expressionWithJSDocTypeArguments/expressionWithJSDocTypeArguments.ts` (742 bytes). [Exact input](census/gate/examples/children-count.ts); complete preorder trees: [typescript-go](census/gate/examples/children-count.go.tree.gz), [port](census/gate/examples/children-count.port.tree.gz).

First witness (Go against port):
```text
Go: 146
Port: 216
```
```typescript
// @target: es2015
// @strict: true

// Repro from #51802

function foo<T>(x: T): T { return x }

class Bar<T> { constructor(public x: T) { } }

// Errors expected on all of the following

const WhatFoo = foo<?>;
const HuhFoo = foo<string?>;
const NopeFoo = foo<?string>;
const ComeOnFoo = foo<?string?>;

type TWhatFoo = typeof foo<?>;
type THuhFoo = typeof foo<string?>;
type TNopeFoo = typeof foo<?string>;
type TComeOnFoo = typeof foo<?string?>;

const WhatBar = Bar<?>;
const HuhBar = Bar<string?>;
const NopeBar = Bar<?string>;
const ComeOnBar = Bar<?string?>;

type TWhatBar = typeof Bar<?>;
type THuhBar = typeof Bar<string?>;
type TNopeBar = typeof Bar<?string>;
type TComeOnBar = typeof Bar<?string?>;

```

## public

Files selected: **152660**; both sides completed: **152609**; byte-identical canonical trees and diagnostics with all flags: **65282**; identical ignoring node flags: **151911**; adapter/runner failures: **51**.

| Divergence class | Files affected | Shortest input bytes | Parser function / context |
|---|---:|---:|---|
| flags/AwaitContext | 60159 | 9 | Parser.file (top-level module context), methodBody / arrow / awaitContext propagation |
| flags/HasJSDoc | 27146 | 51 | Parser.make / Go finishNode and JSDoc attachment |
| flags/DisallowInContext | 14356 | 35 | Parser.rootAssignment / allowInAssignment / forStatement context propagation |
| flags/DecoratorContext | 11896 | 30 | Parser.decorator |
| flags/Ambient | 9558 | 0 | Statements.declaration / declare and declaration-file context |
| flags/PossiblyContainsDynamicImport | 2549 | 21 | Parser.primary / import expression and SourceFile aggregation |
| flags/ThisNodeHasError | 2064 | 9 | Parser.make / Parser.error / scannerErrors error propagation |
| flags/YieldContext | 946 | 36 | Parser.methodBody / arrow / yieldContext propagation |
| flags/PossiblyContainsDeprecatedTag | 835 | 121 | Parser.make / JSDoc comment scanning |
| flags/DisallowConditionalTypesContext | 694 | 36 | Parser.type / conditional type context |
| flags/PossiblyContainsImportMeta | 616 | 84 | Parser.primary / meta-property and SourceFile aggregation |
| diagnostics | 541 | 13 | Parser.error / expect / recoverList (see witness kind) |
| kind/children | 531 | 13 | Parser.primary / Statements.statement / recovery dispatch (see first tree divergence) |
| children/count | 395 | 13 | Parser.file / Statements.statement / delimitedList (see first tree divergence) |
| position | 340 | 13 | Parser.make / token / scanner.fullStart (see witness kind) |
| payload/list/token-flags | 191 | 23 | Parser.literal / templateToken / delimitedList (see witness kind) |
| flags/Reparsed | 186 | 24 | Statements.moduleDeclaration / synthesized dotted-module export token |
| flags/InWithStatement | 60 | 35 | Statements.statement / with context |

### flags/AwaitContext

Shortest corpus input: `/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-parser/test/fixtures/typescript/async-call/with-optional-parameter/input.ts` (9 bytes). [Exact input](census/public/examples/flags-AwaitContext.ts); complete preorder trees: [typescript-go](census/public/examples/flags-AwaitContext.go.tree.gz), [port](census/public/examples/flags-AwaitContext.port.tree.gz).

First witness (Go against port):
```text
Go: 3 Parameter 6 8 8192 0 -1 0 0
Port: 3 Parameter 6 8 0 0 -1 0 0
```
```typescript
async(x?)
```

### flags/HasJSDoc

Shortest corpus input: `/workspace/parser-corpus/microsoft__TypeScript-50d70a3f5f45/tsc/testdata/tests/cases/compiler/emptyJSDocComment.ts` (51 bytes). [Exact input](census/public/examples/flags-HasJSDoc.ts); complete preorder trees: [typescript-go](census/public/examples/flags-HasJSDoc.go.tree.gz), [port](census/public/examples/flags-HasJSDoc.port.tree.gz).

First witness (Go against port):
```text
Go: 1 VariableStatement 0 50 2097152 0 -1 0 0
Port: 1 VariableStatement 0 50 0 0 -1 0 0
```
```typescript
// @declaration: true

/***/
export const foo = 1;

```

### flags/DisallowInContext

Shortest corpus input: `/workspace/parser-corpus/microsoft__TypeScript-050880ce59e3/tests/cases/conformance/parser/ecmascript6/Iterators/parserForOfStatement17.ts` (35 bytes). [Exact input](census/public/examples/flags-DisallowInContext.ts); complete preorder trees: [typescript-go](census/public/examples/flags-DisallowInContext.go.tree.gz), [port](census/public/examples/flags-DisallowInContext.port.tree.gz).

First witness (Go against port):
```text
Go: 3 VariableDeclaration 24 27 1024 0 -1 0 0
Port: 3 VariableDeclaration 24 27 0 0 -1 0 0
```
```typescript
//@target: ES6
for (var of; ;) { }
```

### flags/DecoratorContext

Shortest corpus input: `/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-parser/test/fixtures/experimental/decorators/typescript-declare-field/input.ts` (30 bytes). [Exact input](census/public/examples/flags-DecoratorContext.ts); complete preorder trees: [typescript-go](census/public/examples/flags-DecoratorContext.go.tree.gz), [port](census/public/examples/flags-DecoratorContext.port.tree.gz).

First witness (Go against port):
```text
Go: 4 Identifier 13 16 4096 0 -1 0 0		dec
Port: 4 Identifier 13 16 0 0 -1 0 0		dec
```
```typescript
class A {
  @dec declare a;
}

```

### flags/Ambient

Shortest corpus input: `/workspace/parser-corpus/excalidraw__excalidraw-ed10ac7dca7e/packages/fractional-indexing/global.d.ts` (0 bytes). [Exact input](census/public/examples/flags-Ambient.d.ts); complete preorder trees: [typescript-go](census/public/examples/flags-Ambient.go.tree.gz), [port](census/public/examples/flags-Ambient.port.tree.gz).

First witness (Go against port):
```text
Go: 0 SourceFile 0 0 8388608 0 -1 0 0
Port: 0 SourceFile 0 0 0 0 -1 0 0
```
```typescript

```

### flags/PossiblyContainsDynamicImport

Shortest corpus input: `/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-parser/test/fixtures/typescript/type-arguments/empty-type-import/input.ts` (21 bytes). [Exact input](census/public/examples/flags-PossiblyContainsDynamicImport.ts); complete preorder trees: [typescript-go](census/public/examples/flags-PossiblyContainsDynamicImport.go.tree.gz), [port](census/public/examples/flags-PossiblyContainsDynamicImport.port.tree.gz).

First witness (Go against port):
```text
Go: 0 SourceFile 0 21 524288 0 -1 0 0
Port: 0 SourceFile 0 21 0 0 -1 0 0
```
```typescript
let a: import("")<>;

```

### flags/ThisNodeHasError

Shortest corpus input: `/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-parser/test/fixtures/typescript/async-call/with-optional-parameter/input.ts` (9 bytes). [Exact input](census/public/examples/flags-ThisNodeHasError.ts); complete preorder trees: [typescript-go](census/public/examples/flags-ThisNodeHasError.go.tree.gz), [port](census/public/examples/flags-ThisNodeHasError.port.tree.gz).

First witness (Go against port):
```text
Go: 3 EqualsGreaterThanToken 9 9 32768 0 -1 0 0
Port: 3 EqualsGreaterThanToken 9 9 0 0 -1 0 0
```
```typescript
async(x?)
```

### flags/YieldContext

Shortest corpus input: `/workspace/parser-corpus/microsoft__TypeScript-50d70a3f5f45/tsc/testdata/tests/cases/conformance/es6/functionPropertyAssignments/FunctionPropertyAssignments2_es6.ts` (36 bytes). [Exact input](census/public/examples/flags-YieldContext.ts); complete preorder trees: [typescript-go](census/public/examples/flags-YieldContext.go.tree.gz), [port](census/public/examples/flags-YieldContext.port.tree.gz).

First witness (Go against port):
```text
Go: 6 Block 30 34 2048 0 -1 0 0
Port: 6 Block 30 34 0 0 -1 0 0
```
```typescript
// @target: es6
var v = { *() { } }
```

### flags/PossiblyContainsDeprecatedTag

Shortest corpus input: `/workspace/parser-corpus/calcom__cal.diy-54343aa685ae/apps/web/components/Loader.tsx` (121 bytes). [Exact input](census/public/examples/flags-PossiblyContainsDeprecatedTag.tsx); complete preorder trees: [typescript-go](census/public/examples/flags-PossiblyContainsDeprecatedTag.go.tree.gz), [port](census/public/examples/flags-PossiblyContainsDeprecatedTag.port.tree.gz).

First witness (Go against port):
```text
Go: 1 ExportDeclaration 0 120 69206016 0 -1 0 0				0
Port: 1 ExportDeclaration 0 120 0 0 -1 0 0				0
```
```typescript
/**
 * @deprecated Use custom Skeletons instead
 **/
export { Loader as default } from "@calcom/ui/components/skeleton";

```

### flags/DisallowConditionalTypesContext

Shortest corpus input: `/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-generator/test/fixtures/regression/ts-infer-keyword-compact-mode/input.ts` (36 bytes). [Exact input](census/public/examples/flags-DisallowConditionalTypesContext.ts); complete preorder trees: [typescript-go](census/public/examples/flags-DisallowConditionalTypesContext.go.tree.gz), [port](census/public/examples/flags-DisallowConditionalTypesContext.port.tree.gz).

First witness (Go against port):
```text
Go: 3 InferType 18 26 16384 0 -1 0 0
Port: 3 InferType 18 26 0 0 -1 0 0
```
```typescript
type T = A extends infer B ? C : D;

```

### flags/PossiblyContainsImportMeta

Shortest corpus input: `/workspace/parser-corpus/microsoft__TypeScript-050880ce59e3/tests/cases/conformance/importDefer/importMetaPropertyInvalidInCall.ts` (84 bytes). [Exact input](census/public/examples/flags-PossiblyContainsImportMeta.ts); complete preorder trees: [typescript-go](census/public/examples/flags-PossiblyContainsImportMeta.go.tree.gz), [port](census/public/examples/flags-PossiblyContainsImportMeta.port.tree.gz).

First witness (Go against port):
```text
Go: 0 SourceFile 0 84 1048576 0 -1 0 0
Port: 0 SourceFile 0 84 0 0 -1 0 0
```
```typescript
// @target: es2015
// @module: esnext

// @filename: b.ts
import.foo();
import.foo;

```

### diagnostics

Shortest corpus input: `/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-parser/test/fixtures/typescript/variable-declarator/invalid-definite-pattern-assignment-var/input.ts` (13 bytes). [Exact input](census/public/examples/diagnostics.ts); complete preorder trees: [typescript-go](census/public/examples/diagnostics.go.tree.gz), [port](census/public/examples/diagnostics.port.tree.gz).

First witness (Go against port):
```text
Go: ["diagnostic 1005 6 1 1\t',' expected.", 'diagnostic 1109 8 1 1\tExpression expected.']
Port: []
```
```typescript
var {}! = {};
```

### kind/children

Shortest corpus input: `/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-parser/test/fixtures/typescript/variable-declarator/invalid-definite-pattern-assignment-var/input.ts` (13 bytes). [Exact input](census/public/examples/kind-children.ts); complete preorder trees: [typescript-go](census/public/examples/kind-children.go.tree.gz), [port](census/public/examples/kind-children.port.tree.gz).

First witness (Go against port):
```text
Go: 1 ExpressionStatement 6 7 32768 0 -1 0 0
Port: 4 ExclamationToken 6 7 0 0 -1 0 0
```
```typescript
var {}! = {};
```

### children/count

Shortest corpus input: `/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-parser/test/fixtures/typescript/variable-declarator/invalid-definite-pattern-assignment-var/input.ts` (13 bytes). [Exact input](census/public/examples/children-count.ts); complete preorder trees: [typescript-go](census/public/examples/children-count.go.tree.gz), [port](census/public/examples/children-count.port.tree.gz).

First witness (Go against port):
```text
Go: 11
Port: 8
```
```typescript
var {}! = {};
```

### position

Shortest corpus input: `/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-parser/test/fixtures/typescript/variable-declarator/invalid-definite-pattern-assignment-var/input.ts` (13 bytes). [Exact input](census/public/examples/position.ts); complete preorder trees: [typescript-go](census/public/examples/position.go.tree.gz), [port](census/public/examples/position.port.tree.gz).

First witness (Go against port):
```text
Go: 1 VariableStatement 0 6 32768 0 -1 0 0
Port: 1 VariableStatement 0 13 0 0 -1 0 0
```
```typescript
var {}! = {};
```

### payload/list/token-flags

Shortest corpus input: `/workspace/parser-corpus/microsoft__TypeScript-050880ce59e3/tests/cases/conformance/es6/templates/templateStringUnterminated2_ES6.ts` (23 bytes). [Exact input](census/public/examples/payload-list-token-flags.ts); complete preorder trees: [typescript-go](census/public/examples/payload-list-token-flags.go.tree.gz), [port](census/public/examples/payload-list-token-flags.port.tree.gz).

First witness (Go against port):
```text
Go: 2 NoSubstitutionTemplateLiteral 0 23 32768 4 -1 0 0		`	\u005c`
Port: 2 NoSubstitutionTemplateLiteral 0 23 0 4 -1 0 0		`	\u005c
```
```typescript
﻿// @target: ES6
`\`
```

### flags/Reparsed

Shortest corpus input: `/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-parser/test/fixtures/typescript/module-namespace/head-export/input.ts` (24 bytes). [Exact input](census/public/examples/flags-Reparsed.ts); complete preorder trees: [typescript-go](census/public/examples/flags-Reparsed.go.tree.gz), [port](census/public/examples/flags-Reparsed.port.tree.gz).

First witness (Go against port):
```text
Go: 3 ExportKeyword 19 19 8 0 -1 0 0
Port: 3 ExportKeyword 19 19 0 0 -1 0 0
```
```typescript
export namespace X.Y {}

```

### flags/InWithStatement

Shortest corpus input: `/workspace/parser-corpus/microsoft__TypeScript-050880ce59e3/tests/cases/conformance/parser/ecmascript5/Statements/parserWithStatement1.d.ts` (35 bytes). [Exact input](census/public/examples/flags-InWithStatement.ts); complete preorder trees: [typescript-go](census/public/examples/flags-InWithStatement.go.tree.gz), [port](census/public/examples/flags-InWithStatement.port.tree.gz).

First witness (Go against port):
```text
Go: 2 Block 30 35 25165824 0 -1 0 0
Port: 2 Block 30 35 0 0 -1 0 0
```
```typescript
// @target: es2015
with (foo) {
}
```

Failures are listed separately in `census/public.json.gz`; they do not count as completed AST comparisons.

## Parser failures and adapter failures

All 152,660 Go inputs parsed and serialized successfully. The port completed 152,609 comparisons and failed on 51 inputs: 41 JSX exit-70 panics and 10 parses exceeding the five-second per-input budget. All 51 failed again in a separate replay. **Adapter failures: 0.** Failed parses count as unfinished parser work, outside AST/diagnostic comparison denominators. No port tree exists after a panic or timeout; the shortest input, Go tree, and port failure record are retained below.

| Failure class | Files | Shortest bytes | Responsible path |
|---|---:|---:|---|
| JSX expected name | 26 | 33 | JsxParser.identifier / name |
| Parse limit (5s) | 10 | 30 | Parser.file / Statements.statement recovery; propertyName/typeLiteral; suffix/binary speculative parsing (traces below) |
| JSX expected } | 5 | 296 | JsxParser.expression |
| JSX missing closing tag | 4 | 157 | JsxParser.element |
| JSX Unicode escape not allowed | 2 | 786 | JsxParser.identifier |
| JSX expected attribute value | 2 | 79 | JsxParser.attributes |
| JSX mismatched closing tag | 2 | 515 | JsxParser.element |

### parse-failure timeout (5s)

`/workspace/parser-corpus/microsoft__TypeScript-050880ce59e3/tests/cases/conformance/parser/ecmascript5/MissingTokens/parserMissingToken1.ts`: [input](census/failures/examples/failure-0.ts), [Go tree](census/failures/examples/failure-0.go.tree.gz), [port failure](census/failures/examples/failure-0.port.txt).
```typescript
// @target: es2015
a / finally
```

### parse-failure worker exit 70 adamic: panic: JSX missing closing tag

`/workspace/parser-corpus/microsoft__TypeScript-50d70a3f5f45/tsc/testdata/tests/cases/compiler/errorSpanForUnclosedJsxTag.tsx`: [input](census/failures/examples/failure-1.tsx), [Go tree](census/failures/examples/failure-1.go.tree.gz), [port failure](census/failures/examples/failure-1.port.txt).
```typescript
// @target: es2015
// @jsx: react
declare const React: any

let Foo = {
  Bar() {}
}

let Baz = () => {}

let x = <    Foo.Bar >Hello

let y = <   Baz >Hello
```

### parse-failure worker exit 70 adamic: panic: JSX expected name

`/workspace/parser-corpus/babel__babel-f67453d56391/packages/babel-parser/test/fixtures/typescript/type-arguments-bit-shift-left-like/jsx-opening-element/input.tsx`: [input](census/failures/examples/failure-2.tsx), [Go tree](census/failures/examples/failure-2.go.tree.gz), [port failure](census/failures/examples/failure-2.port.txt).
```typescript
<Component<<T>(v: T) => void> />

```

### parse-failure worker exit 70 adamic: panic: JSX expected }

`/workspace/parser-corpus/microsoft__TypeScript-50d70a3f5f45/tsc/testdata/tests/cases/compiler/jsxNestedWithinTernaryParsesCorrectly.tsx`: [input](census/failures/examples/failure-3.tsx), [Go tree](census/failures/examples/failure-3.go.tree.gz), [port failure](census/failures/examples/failure-3.port.txt).
```typescript
// @target: es2015
// @jsx: preserve
const emptyMessage = null as any;
const a = (
    <div>
      {0 ? (
        emptyMessage // must be identifier?
      ) : (
          // must be exactly two expression holes
        <span>
          {0}{0}
        </span>
      )}
    </div>
);
```

### parse-failure worker exit 70 adamic: panic: JSX Unicode escape not allowed

`/workspace/parser-corpus/microsoft__TypeScript-050880ce59e3/tests/cases/conformance/jsx/unicodeEscapesInJsxtags.tsx`: [input](census/failures/examples/failure-4.tsx), [Go tree](census/failures/examples/failure-4.go.tree.gz), [port failure](census/failures/examples/failure-4.port.txt).
```typescript
// @filename: file.tsx
// @jsx: react
// @skipLibCheck: true
// @target: es2015
// @moduleResolution: bundler
/// <reference path="/.lib/react.d.ts" />
import * as React from "react";
declare global {
    namespace JSX {
        interface IntrinsicElements {
            "a-b": any;
            "a-c": any;
        }
    }
}
const Compa = (x: {x: number}) => <div>{"" + x}</div>;
const x = { video: () => null }

// unicode escape sequence is not allowed in tag name or JSX attribute name.
// tag name:
; <\u0061></a>
; <\u0061-b></a-b>
; <a-\u0063></a-c>
; <Comp\u0061 x={12} />
; <x.\u0076ideo />
; <\u{0061}></a>
; <\u{0061}-b></a-b>
; <a-\u{0063}></a-c>
; <Comp\u{0061} x={12} />

// attribute name
;<video data-\u0076ideo />
;<video \u0073rc="" />

```

### parse-failure worker exit 70 adamic: panic: JSX expected attribute value

`/workspace/parser-corpus/microsoft__TypeScript-50d70a3f5f45/tsc/testdata/tests/cases/compiler/jsxAttributeMissingInitializer.tsx`: [input](census/failures/examples/failure-5.tsx), [Go tree](census/failures/examples/failure-5.go.tree.gz), [port failure](census/failures/examples/failure-5.port.txt).
```typescript
// @target: es2015
// @jsx: preserve
const x = <div foo= ></div>;
const y = 0;

```

### parse-failure worker exit 70 adamic: panic: JSX mismatched closing tag

`/workspace/parser-corpus/microsoft__TypeScript-50d70a3f5f45/tsc/testdata/tests/cases/conformance/jsx/jsxParsingError2.tsx`: [input](census/failures/examples/failure-6.tsx), [Go tree](census/failures/examples/failure-6.go.tree.gz), [port failure](census/failures/examples/failure-6.port.txt).
```typescript
// @target: es2015
//@jsx: preserve

//@filename: file.tsx
declare namespace JSX {
	interface Element { }
	interface IntrinsicElements {
		[s: string]: any;
	}
}

// @filename: Error1.tsx
// Issue error about missing span closing tag, not missing div closing tag
let x1 = <div><span></div>;

// @filename: Error2.tsx
let x2 = <div></span>;


// @filename: Error3.tsx
let x3 = <div>;


// @filename: Error4.tsx
let x4 = <div><div></span>;

// @filename: Error5.tsx
let x5 = <div><span>


```

Timeout attribution used runtime method wrappers without editing parser sources. Repeated operation traces are in `census/failure-traces.json`. Two snapshots each reproduce bigintPropertyName (propertyName/typeLiteral), incomplete constructor annotation (statementList/moduleDeclaration/semicolon recovery), invalid try statements (statementList/semicolon recovery), missing finally tokens (file/statement without progress), and deeply parenthesized expressions (suffix/binary speculation). These are measured time limits and repeated paths, not a claim that every case is an infinite loop.

## Structural witness attribution

The shortest public witness for kind/children, diagnostics, node count, and positions is `var {}! = {};` (13 bytes). `Statements.variableDeclaration` consumes an exclamation after any bindingName, including a destructuring pattern. Go ends the declaration before the exclamation and recovers subsequent statements. The divergence changes downstream children, spans, and diagnostics; it is one root cause appearing in four comparison categories.

The shortest payload witness is an unterminated template after a BOM and test directive. Go retains the last raw character; Parser.literal/templateToken drops it. The separate two-character incomplete-template probe isolates this raw-slicing difference.

Gate witnesses: `Parser.type` postfix-nullable lookahead omits the EqualsToken case in `const b: number? = 1;`; expression type-argument speculation (`Parser.typeArguments` / suffix) also differs on JSDoc-style unknown/nullable type arguments. These account for the two gate files differing beyond flags. The public structural categories include other causes; attribution beyond each shortest witness requires reviewing its stored tree and is not claimed as a complete root-cause partition.

## Corpus coverage

The 23 public snapshots are the supplied subset of Kirk’s quiet hundred, not the unavailable full quiet hundred. Repository pins and exact inventory are retained in `census/public-pins.tsv`, `census/fetch-results.json.gz`, and `census/public.manifest.gz`. The compiler corpus and stage3 gate have separate tables; they overlap the public TypeScript snapshot and are not added to its headline counts.

| Repository snapshot | Files | Completed | Full byte identity | Identity excluding flags | Parser failures |
|---|---:|---:|---:|---:|---:|
| actualbudget/actual `9732a4463aac` | 2023 | 2023 | 1185 | 2022 | 0 |
| angular/angular `c0dc8c4bbeea` | 6073 | 6073 | 269 | 6071 | 0 |
| babel/babel `f67453d56391` | 1664 | 1663 | 782 | 1578 | 1 |
| backstage/backstage `532931240eec` | 7409 | 7409 | 1397 | 7345 | 0 |
| calcom/cal.diy `54343aa685ae` | 5025 | 5025 | 1492 | 5020 | 0 |
| date-fns/date-fns `717ce0a807ea` | 1738 | 1738 | 1230 | 1738 | 0 |
| excalidraw/excalidraw `ed10ac7dca7e` | 674 | 674 | 214 | 674 | 0 |
| grafana/grafana `c7a7b797c198` | 9522 | 9522 | 3928 | 9521 | 0 |
| microsoft/playwright `2a8ba77a33a5` | 1390 | 1390 | 40 | 1389 | 0 |
| n8n-io/n8n `e77e30c7d92f` | 21895 | 21895 | 4520 | 21886 | 0 |
| nestjs/nest `35142c3eca8e` | 1966 | 1966 | 275 | 1966 | 0 |
| outline/outline `478e8121cbe5` | 2224 | 2224 | 334 | 2224 | 0 |
| prisma/prisma `c882b03377e7` | 5198 | 5198 | 1154 | 5196 | 0 |
| shadcn-ui/ui `a2e305c2e15b` | 3881 | 3881 | 3120 | 3881 | 0 |
| supabase/supabase `87681812a0b4` | 8257 | 8255 | 3709 | 8249 | 2 |
| tanstack/query `eaa75f4f8f81` | 995 | 995 | 293 | 981 | 0 |
| tldraw/tldraw `db1c86ea7857` | 2844 | 2844 | 844 | 2843 | 0 |
| trpc/trpc `d756e591a5e3` | 1022 | 1022 | 192 | 1017 | 0 |
| twentyhq/twenty `ddd166250b03` | 30263 | 30263 | 19730 | 30260 | 0 |
| typeorm/typeorm `c64a1f052fc3` | 3607 | 3607 | 222 | 3606 | 0 |
| microsoft/TypeScript `50d70a3f5f45` | 13260 | 13235 | 6682 | 12976 | 25 |
| microsoft/TypeScript `050880ce59e3` | 21241 | 21218 | 13484 | 20979 | 23 |
| vuejs/core `4ab865a848a1` | 489 | 489 | 186 | 489 | 0 |
