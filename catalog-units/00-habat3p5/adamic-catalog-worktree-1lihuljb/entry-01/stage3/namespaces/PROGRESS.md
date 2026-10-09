# Namespace fixtures toward the October 8 parser deadline

## 08_parser_jsdoc_nested: 4 to 5 accepted

Commit ddb51eb runs the unchanged original-source JSDoc parser slice natively.
A directly returned scalar singleton assignment uses existing Assign and Return
IR: evaluate the right side once, store it, then return the stored value. Other
assignment-expression positions and reference/optional singleton storage stay
NotYet. This avoids a new expression form and keeps all ownership/flow passes
aware of the write.

Node, sanitized native and checked JavaScript print:

```text
true 1
false 1
3 3
```

`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/stage3/fixtures/namespaces/08|TestNamespaceSemanticMutants/drop_returned_assignment' -count=1 -v`
passed in 1.004s, with zero cache hits. Removing nextTokenJSDoc's returned store
finishes cleanly with sanitizers and is caught only by Node stdout comparison.
Removing the scalar-only boundary makes the reference-return refusal test fail;
the real compiler mutation was restored, and its regression passed afterward.
Focused namespace/shape lowering passed in 1.172s, counts regeneration in
19.145s, and `go vet ./...` and `git diff --check` passed.

The first row of [progress-matrix.json](progress-matrix.json) builds all twelve
unchanged slices against ddb51eb: five compile, five NotYet, two Refused. Every
accepted fixture matches fresh source Node in stdout, stderr and exit status,
under ASan/UBSan with leak detection, and in the checked JavaScript backend.
The two Refused fixtures are unchanged. The matrix is reproduced with:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/namespaces/progress-matrix.py --label '08 Parser.JSDocParser returned singleton assignment' --scratch /tmp/namespaces-progress-08 > /tmp/namespaces-progress-08.log 2>&1
```

This is a native parser slice with original token, nextTokenJSDoc and
parseOptionalJsdoc bodies, not unchanged parser.ts or its complete scanner.

## Parser factory var binding at parser.ts:1472

The original declaration uses 25 renamed factory bindings. The reproducible
fixture retains two of them and supplies small factory contracts; it does not
copy factory implementations. The exact executable cut is:

```typescript
// Cut from TypeScript 6.0.3 parser.ts:1472-1499. The factory contracts are driver support.
namespace Parser {
    const factory = {
        createNodeArray: (elements: number[]): number[] => elements,
        createNumericLiteral: (value: number): string => `number:${value}`,
        createLiteralLikeNode: (value: number): string => `other:${value}`,
    };
    var {
        createNodeArray: factoryCreateNodeArray,
        createNumericLiteral: factoryCreateNumericLiteral,
    } = factory;
    export function run(): void {
        console.log(factoryCreateNodeArray([1, 2, 3]).join(','));
        console.log(factoryCreateNumericLiteral(7));
    }
}
Parser.run();
```

Before the fix, calling namespace hoisting directly on the pattern reproduced
`the checker gave a declaration no symbol`. A textual location in an ordinary
error did not populate structured diagnostic Where. The public namespace guard
previously stopped the same input earlier as unsupported destructuring.
Now hoisting registers the checker symbol of each identifier leaf. Destructured
var initialization writes the hoisted storage instead of declaring it again.
A remaining symbol-less declaration returns located NotYet with the original
message. Flat object var only: arrays, defaults, rest and nesting remain NotYet.

Node, ASan/UBSan native, and checked JavaScript output `1,2,3` and `number:7`.
The wrong-factory-binding mutant selects createLiteralLikeNode with the same ABI;
it finishes cleanly and stdout comparison catches it. The location mutant
restores ordinary errors.New and the structured-Where assertion fails. Both
mutants were restored. The direct pre-fix hoisting repro fails and passes after
the fix. A separate attempted early-read driver was rejected by the checker
(TS2454); it provides no runtime evidence and was removed.

Focused uncached factory oracle plus semantic mutant passed in 0.984s (zero
cache hits), lower tests in 1.011s, counts regeneration in 20.094s, and the
post-location-mutant namespace regression run in 0.981s. Logs are
`/tmp/namespaces-factory-final-oracle.log`, `/tmp/namespaces-factory-final-lower.log`,
`/tmp/namespaces-factory-counts.log`, `/tmp/namespaces-factory-final-check.log`,
`/tmp/namespaces-factory-location-mutant.log` and
`/tmp/namespaces-factory-repro-before.log`. The twelve-fixture count remains five;
this cut is an additional parser obligation. The priority request arrived after
08 had already been pushed; this fix precedes all further fixture work.
