# 95e: the ruled signature declaration return contract

lane: run as part of a Correctness wave

Base: `typescript/axolotl-checker-a`,
`ef4d26a4a38b7890317e1ba71f8ad928fd8588f6`, including `d5c15cbf`.
This head is independent of 95c and contains only this adaptation directory.

@system_adamic's November 1 ruling authorizes one public API difference:
`TypeChecker.signatureToSignatureDeclaration` returns a
`SignatureDeclaration & { typeArguments?: NodeArray<TypeNode> | undefined; }`.
The implementation can return an ArrowFunction whose typeArguments property
is present and undefined. Both source TypeChecker overloads receive the same
truthful intersection; one overload is public and the other is `@internal`.
Nothing else in the public API changes. The upstream API baseline is not edited.

`adapt.cjs` uses stock TypeScript 6.0.3 through 47-host-errors' `compiler()`.
It addresses both parsed MethodSignatures in TypeChecker, binds their shared
symbol and both call signatures, checks public/internal order, binds the
NodeArray type and typeArguments property symbols, and checks the exact
before/after slot types. Each overload's normalized text has an
`ownerBeforeSha256` pin; types.ts and checker.ts have no whole-file input pin.
Both edits are planned, parsed, and transpiled with and without comments before
one write. Source slicing preserves CRLF and uses no printer. No runtime code,
assertion, cast, optional chain, or default is introduced. A second run and
`--check` are read-only when the repairs are present.

`verify.cjs` independently reconstructs both complete overload edits, compares
every source file and CRLF count, and compares the complete types.ts JavaScript.
`prove.py BEFORE AFTER CENSUS OUTPUT` replays source identity, idempotence,
whole-program census, and a one-public-overload undo mutant. The census uses
95a's exact entry-root normalization: root and aggregate observations must
agree, and count once. CENSUS is the unchanged production tool built with
`go build -o CENSUS ./stage3/census/tool`.

## Declared API line and sanction

[api-sanctions.json](api-sanctions.json) is an array, not a lane-file update.
It records the return type text of the unique public MethodSignature.
`verify-api.cjs` proves that its `before` is the exact **117-byte** type text
in upstream's pinned `tests/baselines/reference/api/typescript.d.ts`, without
formatting or whitespace normalization. Baseline line **6218**, under the
MethodSignature starting at **6216**, changes only:

```diff
-                typeArguments?: NodeArray<TypeNode>;
+                typeArguments?: NodeArray<TypeNode> | undefined;
```

The source overloads are inventoried separately; the internal overload is
stripped from the public declaration artifact. The path is
`["TypeChecker", "signatureToSignatureDeclaration"]`: this is the actual
`owned()`/`ownerPath()` convention in this branch's
`stage3/lane/generate-api.cjs` (lines 51 to 58 explicitly omit the `ts` module).
The normalized API-difference tool uses its separate namespace-bearing key
`member:["ts","TypeChecker","signatureToSignatureDeclaration"]#0`.

[proofs/api-lines.md](https://github.com/system-inc/adamic/blob/typescript/axolotl-checker-e-evidence/stage3/evidence/95-checker-axolotl-e/proofs/api-lines.md) and
[proofs/api-lines.json](https://github.com/system-inc/adamic/blob/typescript/axolotl-checker-e-evidence/stage3/evidence/95-checker-axolotl-e/proofs/api-lines.json) record the line and bytes.
Correctness's wave owns regeneration of sanctioned-api.json and its baseline
undo mutant. This head neither writes lane files nor claims lane acceptance.
The local source undo mutant is additional checker/verifier evidence.

## Measured checker and JavaScript proof

The exact entry-root census measures **131 to 130 whole-program** and
**17 to 16 checker own-file**, removing only TS2322 at checker.ts:1683:9.
No diagnostic appears anywhere. Undoing only the public overload's edit restores
the exact original census: **131 whole-program, 17 checker own-file**. Both the
independent verifier and adapter `--check` exit 1 on that mutant. The internal
overload stays widened in the mutant. No mutant enters the oracle tree.

The independent source verifier compares **746 files under src**. Exactly
one changes. CRLF counts stay **10,681**. Complete stock types.ts JavaScript
is byte-identical, with comments retained and removed:

| Comments | Bytes | SHA-256 before and after |
| --- | ---: | --- |
| Retained | 139,659 | `4839bedc439cad06969bad2a3d805a7c0b480f06df73fcff62d2cd0957009fc3` |
| Removed | 121,398 | `f041d720cfda4ccb137ffc32198ec3d2fce46dbf13404df737dc637e21290639` |

A second run reports `files: 0, declarations: 0`; `--check` passes without
writing. Setup was already completed for 95c: Node 0.023s, Go 0.031s, clang
0.258s, tsc 0.547s, Markdown 1.431s, submodules 163.991s, Go build 436.508s,
cache warm 436.625s, done 436.662s. `nproc=5`, CPU quota 4.
All build/test shells source `/workspace/adamic-tools/env.sh`.

## Evidence and reproduction

Only adapt.cjs, verify.cjs, sites.json, README.md and api-sanctions.json live
in this shipping directory. Logs, censuses, oracle results and helper scripts
live on [`typescript/axolotl-checker-e-evidence`](https://github.com/system-inc/adamic/tree/typescript/axolotl-checker-e-evidence/stage3/evidence/95-checker-axolotl-e).
Every shipping-directory file feeds apply's input key. Run apply from this
branch and proof helper commands from the evidence branch.

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH="$PWD/stage3/api/node_modules"
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/95-upstream
bash stage3/apply.sh /workspace/scratch/95e-clean > /tmp/95e-clean-apply.log 2>&1
python3 stage3/evidence/95-checker-axolotl-e/prove.py /workspace/scratch/95a-tree /workspace/scratch/95e-tree /workspace/scratch/95-census /workspace/scratch/95e-proofs > /tmp/95e-prove.log 2>&1
node stage3/adapt/95-checker-axolotl-e/verify.cjs /workspace/scratch/95a-tree /workspace/scratch/95e-clean > /tmp/95e-clean-source.json 2>&1
node stage3/evidence/95-checker-axolotl-e/verify-api.cjs /workspace/scratch/95-upstream/tests/baselines/reference/api/typescript.d.ts /workspace/scratch/95c-final/built/local/typescript.d.ts /workspace/scratch/95e-tree/built/local/typescript.d.ts > /tmp/95e-api.json 2>&1
timeout --kill-after=20s 880 bash stage3/oracle/run.sh /workspace/scratch/95e-tree /workspace/scratch/95e-oracle --limit-seconds 720 > /tmp/95e-oracle.log 2>&1
```

The evidence branch retains the complete end-to-end apply, fresh source/census
checks, and bounded upstream oracle observation. Main's comparison is 106,366
passing, one standing Public APIs acknowledgement failure, and zero pending;
this head declares its additional API line rather than hiding that failure.
No lane acceptance predicate is run, as Correctness's wave owns it.

Not covered: native tsc compilation/lowering, browser/ESLint integration, and
the full repository Go gate. The single public API widening is explicitly
sanctioned above; no other public API byte is authorized to move.
