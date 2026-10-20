# Draft 90: internal undefined returns

The separate compiler-owned source scan finds 116 sites. This draft removes 111,
leaving four open-hook findings and the unclassified noop timer fallback. It
changes 69 return annotations in seven compiler files. The requested complete
full latent comparison is **blocked**, so this is not a completed landing proof.
See [the census report](../../census/void-widening/README.md) for the limits and
[the per-site table](../../census/void-widening/void-widening-2026-10-20.json).

Run `node adapt.cjs <tree>`. Stock TypeScript 6.0.3 resolves every target signature.
Exact file hashes, declaration positions, lines and return types reject moved
targets. All guards and full-file transpile comparisons run before any write.
A second run writes zero files. targets.json records every original declaration
and exact replacement. Explicit void becomes undefined; inferred void gains an
undefined return annotation. Existing local parser callback assertions are
refined to the same checked type. Single-parameter forwarding arrows keep their
runtime parameter spelling and infer undefined from the refined callees.

Every edited implementation and required returned-call dependency is checked by
stock TypeScript. The compiler-project comparison stays 140 to 140 diagnostics,
with none introduced. A new number/string assignment mutant adds exactly TS2322,
140 to 141. The complete upstream compiler, harness and test build exits 0.

Every edited file emits byte-identical transpileModule JavaScript with comments
retained and removed. A runtime-statement mutant in each of the seven files is
caught by that byte comparison. All ten built JavaScript artifacts and both
public declaration artifacts also match the unchanged control byte for byte.
Public typescript.d.ts remains 11,449 lines and the existing 223 sanctioned API
declarations are unchanged. No public signature, harness source, sanctioned API
manifest, internal/ file or other adaptation is edited.

The full numbered apply exits 0. Its generated patch-set row is
`90-void-widening | 7 | 66 | 66`; 69 annotations share 66 source lines. Its 746
source files are byte-identical to the isolated checked/built tree, and a rerun
writes zero files. apply discovers the directory automatically, so the checked-in
patch-set table does not need a dispatch change.

The instrumented full Node corpus reports 106,366 passing, one existing API
baseline failure and zero pending, exactly like the uninstrumented control.
Both actual baseline.diff files are identical, SHA-256
`5333568a661aea24fc8af3149e9f4d2dfca70731842461c677f385b092adc4f0`.
Counters only wrap the two excluded open-hook calls, preserve their values, and
have a caught increment-removal mutant. The uninstrumented runtime artifacts
are separately byte-identical to the control.

The source-scan comparison is 116 to 5 Refused sites; its 26 structural-depth
NotYet findings are unchanged. There is no new finding from that check. Restoring
bindFunctionOrConstructorType to void returns exactly binder.ts:2995:24, five to
six in the whole-source scan. A separate full-mode binder-root control also
changes zero to one. The planted new-site mutant changes the whole-source scan
116 to 117 at core.ts:2601:50, with the full-mode core-root control zero to one.
These do not replace the missing complete full latent before/after comparison
or prove no new refusal from every other compiler check.

The moved-declaration mutant is caught before writing. Seven erasure mutants,
the moved-declaration mutant and the TS2322 mutant pass their intended controls.
All mutations are confined to scratch and restored in finally blocks.

The excluded noop declaration has an empty body, but globally retyping it breaks
four harness/test type checks. No permitted checked repair preserving the full
tree has been established. The two patcher registries accept external callbacks;
Node witnesses return 42 and 17 and stop forEach early. Those remain findings.
The full latent measurement blocker, not a public API ruling, prevents an exact
full-census clearance claim for this draft.

Commands used /workspace/adamic-tools/env.sh and stock TypeScript from
stage3/api/node_modules. All test output went to logs:

```sh
NODE_PATH=/workspace/adamic/stage3/api/node_modules VOID_WIDENING_COMPILER_HELPER=/workspace/adamic/stage3/adapt/47-host-errors/adapt.cjs timeout 180 node /workspace/wks9251-scratch/adapt.cjs /workspace/wks9251-scratch/adapter-proof > /tmp/wks9251-adapter-first.log 2>&1
NODE_PATH=/workspace/adamic/stage3/api/node_modules VOID_WIDENING_COMPILER_HELPER=/workspace/adamic/stage3/adapt/47-host-errors/adapt.cjs timeout 180 node /workspace/wks9251-scratch/adapt.cjs /tmp/wks9251-after > /tmp/wks9251-adapter-idempotence.log 2>&1
# In /tmp/wks9251-after:
timeout 300 npm run build > /tmp/wks9251-after-build-final.log 2>&1
# In /workspace/adamic, with TMPDIR=/workspace/wks9251-scratch/tmp:
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-gate/stage3-meter.0XrKU8/upstream timeout 840 bash stage3/apply.sh /workspace/wks9251-scratch/applied90 > /tmp/wks9251-apply90.log 2>&1
NODE_PATH=/workspace/adamic/stage3/api/node_modules timeout 180 node stage3/adapt/90-void-widening/adapt.cjs /workspace/wks9251-scratch/applied90 > /tmp/wks9251-apply90-rerun.log 2>&1
```

The implementation was exercised at its scratch path before being copied into
this directory. The copied implementation is byte-identical. Compressed logs and
structured proof results are retained under proofs/. No complete Adamic Go gate
or native tsc execution is claimed for these external-source type annotations.
