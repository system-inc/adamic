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
