# Parser scout census — step 25 (#pw720wc)

Base: `ad7bd06632f1`, origin/area/stage1-lint; recovery ancestor `88f4a83d`. No parser changes. The comparison runs the actual port TypeScript on Node through the repository runtime hook and the unmodified pinned typescript-go parser through a Go overlay. This census does not claim native Adamic execution parity.

Canonical byte identity means preorder node kind, byte start/end, full node flags, literal flags, ordered children (depth), list metadata, operator, cooked/raw text, semantic fields, and parser diagnostics (code, byte position/length, category, text). Positions use UTF-8 bytes on both sides. JSDoc attachments are represented by flags; full lazy JSDoc trees are outside `ForEachChild` and need a separate docTypes comparison. Port storage represents OptionalChain and block-scoped declaration bits; other node flags have no stored counterpart and print as unset. Nothing is silently masked in the full comparison.

Classes overlap: a file may occur in several rows. Structural comparison explicitly removes only node flags. “Shortest” means the shortest complete selected corpus file exhibiting that class, not a delta-minimized program. The full input and both complete trees are retained beside the first witness. Function names for flags identify missing production/context propagation. Structural attribution is provisional where several parser paths differ; use the retained tree to inspect the exact node.

Adapter: `census/oracle.go.txt` is a copy of the gate oracle with full flags, safe unterminated no-substitution template slicing, and per-file panic capture. The one-character incomplete-template probe parses and prints on both sides: no adapter failure; only ThisNodeHasError differs. Failure records are excluded from completed comparison denominators. A port panic during parsing is currently reported as a runner failure, rather than fabricated as an AST difference.

Reproduction: fetch cohere and its pinned TypeScript checkout, build the Go overlay at the path in `census/overlay.json` (regenerate absolute paths if relocated), then `python3 census/run.py census/tsc.manifest census/tsc.json /tmp/parser-census-oracle`. `fetch.py` reads `public-pins.tsv`, shallow-fetches each exact SHA, executes no dependency installation or repository scripts, and writes `public.manifest`. Run the same driver on that manifest. Paths in manifests reflect this scout's `/workspace` layout.

## Entry points without an inherited list context

`expression()`, `assignment()`, `rootExpression()`, and `rootAssignment()` are public methods and do not establish an enclosing source/block list. The root wrappers only manage expressionDepth and roots. Called on a fresh Parser they can all recover differently from file parsing, for the same reason as tsprinter: recoverList consults the active outer listContexts to decide when to stop. A legal object method shorthand remains supported; malformed surroundings determine whether recovery consumes the method or returns to the enclosing list. Their ordinary calls from file/statements retain an inherited context.

Additional public paths `allowInAssignment()` and `allowInExpression()` wrap assignment/expression and only change disallowIn; they likewise inherit rather than establish a list context. `type()`, `returnType()`, and `docTypes()` can also start with an empty outer list. `docTypes()` scans raw @type substrings, panics on a missing brace, and uses returnType rather than Go's complete JSDoc grammar; malformed types can diverge, but this is a type/JSDoc entry issue rather than an object-method-specific claim.

External production callers inspected: tsprinter `formatExpression` directly calls expression (the known pinned issue); tsprinter files, lint options/settings, lint reparsing, comments, and ESTree pipeline use file(), which establishes source context. Typeaware constructs Parser and delegates file parsing to its linter. JSX expression calls inherit source/list contexts during file parsing. No second external fresh-Parser expression caller was found in stage1. The report refers to the pinned branch; formatter fix `06c6e7cbf` is not applied here.

## tsc
Files selected: **77**; both sides completed: **77**; byte-identical canonical trees and diagnostics with all flags: **0**; identical ignoring node flags: **77**; adapter/runner failures: **0**.

| Divergence class | Files affected | Shortest input bytes | Parser function / context |
|---|---:|---:|---|
| flags/HasJSDoc | 73 | 382 | Parser.make / Go finishNode and JSDoc attachment |
| flags/DisallowInContext | 51 | 5424 | Parser.rootAssignment / allowInAssignment / forStatement context propagation |
| flags/AwaitContext | 20 | 101 | Parser.methodBody / arrow / awaitContext propagation |
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
