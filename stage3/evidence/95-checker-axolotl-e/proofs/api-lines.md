# Declared public API difference

Pinned TypeScript commit: `050880ce59e30b356b686bd3144efe24f875ebc8`.
Baseline: `tests/baselines/reference/api/typescript.d.ts`, SHA-256
`c967e03c8dbb4770f7e2f0b71b5d13593de679a41cc2b60266c4b69f1342a263`.

Exactly upstream line **6218** moves, in the public MethodSignature beginning
at line **6216**, under `ts.TypeChecker.signatureToSignatureDeclaration`:

```diff
-                typeArguments?: NodeArray<TypeNode>;
+                typeArguments?: NodeArray<TypeNode> | undefined;
```

The sanction's before return-type text is 117 bytes and matches the parsed
upstream baseline text byte for byte. `verify-api.cjs UPSTREAM_BASELINE` proves
this without a printer, trimming, or whitespace normalization. Its optional
BEFORE_BUILT AFTER_BUILT arguments also require the entire built public API
artifact to change by exactly that one return-type edit and nothing else.

The source has two overloads carrying the intersection. Both change in
`src/compiler/types.ts`; the second is @internal and contributes no additional
public baseline line. Correctness's wave regenerates the lane sanctions and
runs the baseline undo mutant, as requested. No lane source changes here.

The composed built declaration moves the corresponding line **6219**
(MethodSignature starts at **6217**), because earlier adaptations shift
its location by one. The full built API grows from 590,996 to 591,008 bytes;
verify-api.cjs requires every other byte to remain identical.
