# System timer slot: C-noinfer-b

Adamic, task b5w3ycg, Nov 7 21:47:55Z. This head is held for Runtime
#995spev and stacks on the adaptation 41 ambient-head amendment. It uses
TypeScript 6.0.3, source `050880ce59e30b356b686bd3144efe24f875ebc8`, on the
pinned main `0cb816a0ec439f63542b3f83bb13941fd79f8f41`.

```
setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;
clearTimeout?(timeoutId: unknown): void;
```

The two private `pollScheduled` annotations become `unknown`. The measured
inventory has 22 calls and no member read on a returned handle. `NoInfer`
lets the actual extra arguments choose A while callbacks may ignore a suffix,
including the two label-only calls. The adapted typescript.d.ts needs
TypeScript 5.4 or later to read because of `NoInfer`. This is a declared public
API difference, recorded in api-sanctions.json and the upstream ledger.

## Owner contract

owners.json pins the whole System interface and both complete polling
functions before and after, with exact text and SHA-256. All three owners
must be uniformly before or after. The adapter plans and checks the complete
file before writing; a half-applied tree, a partial member change, changed,
missing, duplicate or moved owner, or parse error refuses without writes.
Replay is idempotent and preserves mtimes. LF and CRLF are preserved. Unrelated
source outside the owners is accepted. Full-file standalone JavaScript is
checked equal with comments kept and removed. No runtime statement changes.

## Census gate and the newly accepted initializer

All coordinates below use original fully adapted main. The two deleted
ambient lines shift actual head coordinates down two after line 57.
Head 1 already exposes `reading setTimeout` at main 1501:13, actual 1499:13,
owned by Runtime #995spev. Against Head 1, C-noinfer-b exposes eight sites;
against original main, it exposes nine including that existing read. All are
expected compiler walls, not a claim that native lowering succeeds.

The eighth additional site, main 1519:36 (actual 1517:36), says exactly:

`a member typed any (require()module) related to {} at the initializer site`

It names `const nodeSystem: System = { ... }`. Its `require` member returns
`{ module: require(modulePath), modulePath, error: undefined }`, where
`modulePath = resolveJSModule(moduleName, baseDir, nodeSystem)`. The `any` is
Node's `require(modulePath)` result in that object, not the separate
`require("source-map-support")` or `require("inspector")` calls. System's
`require?(baseDir: string, moduleName: string): ModuleImportResult` uses
`ModuleImportResult<T = {}>`; its success member is `module: T`. Thus the
nested `module` result is related to `{}`. The stock 6.0.3 checker confirms
`module: require(modulePath)` at actual 1613:38 has value type `any` and
contextual type `{}`; head2-stock.json retains that query.

The original slot, C-noinfer-b, and the any-slot-restoration mutant report the
same recording unit. Its deferred relation changes from main 384:79's callback
`any` relation to this initializer relation and back. No initializer finding
is recorded in Head 1 or the mutant. `internal/lower/any_relations.go:24-48`
keeps only the first deferred any relation (`if l.refusal.deferred == nil`);
line 43 constructs the quoted message. Inference, supported by the measured
replacement and reversal: the slot's earlier `any` relation masked the later
initializer relation by occupying that deferred slot. The census does not
trace every intermediate relation candidate, so this is not presented as a
complete scan trace.

The gate requires exactly the approved location, kind, reason and phase set,
including that initializer. Any other new finding stops the proof. The
any-slot-restoration mutant keeps the two private annotations as unknown and
must restore main 384:79.

## Reproduction and receipts

Source `/workspace/adamic-tools/env.sh`, install the pinned stage3/api parser,
and export ADAMIC_TYPESCRIPT_SOURCE to the pinned upstream checkout. Run
`bash stage3/apply.sh --build-product NEW_TREE` from a pristine checkout.
`node contract.cjs HEAD1_TREE` exercises owner refusal and replay. The retained
proof drivers log each child's actual exit and bound each piece at 30 minutes.
The full lane executes the entire unchanged manifest in separate units,
validating test identities, multiplicities, counts and exact sanctioned API
changes. No expected failure, timeout or test inventory is relaxed.

The accepted storage survey found no direct executable rule pinning either
private annotation; storage-pin-survey.json retains its scope and limitations.
See [PROOF.md](PROOF.md) for the complete build, lane and census receipts.
