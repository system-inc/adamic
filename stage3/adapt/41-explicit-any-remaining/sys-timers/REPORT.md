# sys-timers: measured stop before construction

Task b5w3ycg. Base origin/main `e8f3723bd`; TypeScript 6.0.3 source `050880ce59e30b356b686bd3144efe24f875ebc8`. No adapter or lowering change is built.

Deleting the two adapted ambient declarations passes the type and emission gates, but fails the required no-new-sites census gate. The new site is **sys.ts:1499:13, NotYet: reading setTimeout**, phase lowering, in getNodeSystem's createSystemWatchFunctions argument. It is the bare global timer function read as a value. `internal/lower/expression.go:672`, in `(*lowering).enumNeverValue`, emits `l.notYet(node, "reading "+node.Text())` after the non-local builtin/function cases fail. The runtime timer provider alone does not satisfy this first-class value read. No conclusion is claimed about the later clearTimeout read behind this wall.

| Proof | Result |
| --- | --- |
| Required setup | source /workspace/adamic-tools/env.sh; cloud/setup.sh exit 0; API npm ci exit 0; source pin verified |
| Baseline full apply --build-product | Child exit 0, 971.663 seconds |
| Adamic types, all 79 sorted compiler roots | Baseline and candidate child exit 0; zero checker diagnostics |
| Four getNodeSystem shorthand sites | All contextual destination signatures unchanged; handle return remains NodeJS.Timeout; callbacks accepted |
| Ordinary compiler builds | Both child exit 0 |
| All nine emitted JavaScript files | Byte-identical |
| Public typescript.d.ts and tsserverlibrary.d.ts | Both byte-identical |
| typescript.internal.d.ts | Byte-identical; does not move |
| Full source scope | All 746 source files compared; only the two sys.ts declaration lines differ |
| sys.ts-only latent census | All three children exit 0; before 246 distinct observations / 123 walls, candidate 241 / 120, mutant 246 / 123 |
| Ambient refusals | Two before, zero after, two restored by mutant |
| New latent sites | FAIL: one NotYet, reading setTimeout, at candidate sys.ts:1499:13 |
| Restored-declaration mutant | Reproduces the entire baseline distinct-site set, not just the two ambient refusals |
| Amended full apply and paired replay | Not run: stopped before adapter construction after the census gate failed |

A wall is a distinct location/kind/reason/phase observation with kind Refused or NotYet. Six distinct observations disappear: the two ambient refusals, two no-body NotYets, and two associated continuation boundaries. One new NotYet appears. Source locations after the two deleted lines are adjusted by two lines for the comparison. Raw observations, including boundaries, are retained; none are suppressed.

The census uses the repository's guarded scratch overlay, `LATENT_FULL=1`, `LATENT_ONLY_FILE` set to each adapted sys.ts, and `LATENT_ASSERT_NO_OUTPUT=1`. The source-directory argument loads all 79 compiler roots. No speculative or replay flag is set. Ordinary Load and Lower cannot return a usable output from this measurement binary. Every run has zero checker diagnostics, despite the retained legacy label "measured on a checker-rejected program". Units are 140 / 138 / 140, with three deferred generic units on each side. This census is not a successful lowering claim.

## Ownership and exact type measurement

The initial required search covers every adapt.cjs and owners/sites JSON. Tracing the full apply's stage refs then identifies the exact statement pins in **41-explicit-any-remaining/rules.json, rows 13 and 14**. They currently rewrite the upstream any declarations to:

```ts
declare function setTimeout(handler: () => void, timeout: number): NodeJS.Timeout;
declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;
```

A future amendment belongs in 41. Adaptations 93-builtin-imports and 99-wasm-crypto-import additionally pin the whole getNodeSystem owner, including all four shorthand sites, at their own stages. The measured deletion changes none of that owner's bytes.

The loader's bundled @types/node 25.3.3 timer provider exposes a generic `setTimeout<TArgs extends any[]>(callback: (...args: TArgs) => void, delay?: number, ...args: MakeVoidParameterOptional<TArgs>): NodeJS.Timeout`. Its cancellation parameter is `string | number | NodeJS.Timeout | undefined`. The previous local callback was `() => void` and its handle was already NodeJS.Timeout. All four destination types, including their existing variadic any callback and optional System methods, remain exactly unchanged. The Go scratch probe uses the ordinary Adamic loader and records the actual signatures, return types, contextual destinations and enclosing owner; no replacement timer types are invented for this measurement.

## Retained evidence

result.json records counts, all twelve output hashes and compiler binary hashes. evidence.tar.gz retains every census JSONL, unit timing stream, accepted type/build/apply log, exact sys.ts bytes, the two complete source manifests, ownership pins, the deletion patch and the command/child-exit ledger. Its internal SHA256SUMS and adjacent evidence.sha256 cover the evidence. Scratch drivers and the loader probe are included alongside the report.

The preliminary tool build before submodule setup completed, and a preliminary deletion driver expecting upstream rather than adaptation 41's declarations, each exited 1. The latter stopped before changing the candidate. Corrected accepted measurements are explicitly named in the command ledger. The final census-comparison controller exits 1 because of the new site; all three actual census children exit 0. Child exits are recorded directly, with no pipeline or shell-status substitute. No process was terminated for silence.

The implementation and rule files of every adaptation remain unchanged. No new adaptation directory is introduced and no Loom result is submitted.
