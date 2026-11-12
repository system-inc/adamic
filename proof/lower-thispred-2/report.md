# Method `this is T` agreement probe

Repository: `b8d5f493df0227efcb33c6040ba40c906eec73c9` (requested main HEAD), checked out detached after `git fetch origin main`. Recursive submodules initialized: cohere `5360753e19e4417675a90e92a1823bea66b13918`, its TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.

## Setup and method

Sourced `/workspace/adamic-tools/env.sh` first. Created its missing `TMPDIR=/tmp/adamic-gate` outside the repository. Cloned upstream TypeScript outside the repository to `/tmp/lower-thispred-2/TypeScript`, checked out `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE` to that directory. Copied stage3/api's package.json and lockfile to `/tmp/lower-thispred-2/packages` and ran `npm ci` there: TypeScript 6.0.3, @types/node 25.3.3, undici-types 7.18.2. Set `NODE_EXTRA_CA_CERTS` and `npm_config_cafile` to `/etc/ssl/certs/ca-certificates.crt`; put the external node_modules/.bin on PATH. Node v24.19.0; Go go1.27.1 linux/amd64.

Read `internal/oracle/static_self_construction_test.go` and `loomViewsDAgrees` / `loomViewsDAgreesBackend` in `internal/oracle/loom_views_d_test.go`. Used the same `onNode`, `lowered`, `releasedUncached`, `nativelyUncached`, `onJavaScriptBackend`, and `disagreement` helpers from a throwaway same-package test at `/tmp/lower-thispred-2/probe_test.go`, injected as a virtual test file with `go test -overlay`. The harness logs refusals rather than failing immediately, and permits Node exit 70 so a refusal can still be examined for the executed bad reads.

```
go test -overlay /tmp/lower-thispred-2/overlay.json ./internal/oracle -run '^TestThisPredicateProbe$' -count=1 -v
```

Each source below exists under `/tmp/lower-thispred-2` as byte-identical `.ts` and `.a` files. Ran `tsc 6.0.3 --strict --noEmit` on each `.ts` (all exit 0, no diagnostics); this also checks the identical TypeScript text used in its `.a` twin. Ran both extensions with `node --disable-warning=ExperimentalWarning oracle/node.mjs <file>` and again using the oracle helper.

## Polarity and source programs

The supplied t1 returns true for Full, so its predicate lies on its **true answer**; the false-branch read is not executed on Node. It is preserved verbatim. The requested t2 body returns false for Empty, which is a **false-answer lie**, despite the request calling it a true-side lie. t2 keeps that body and puts the read in the predicate's true branch, which Node does not enter. Two supplemental controls exercise the intended lying false and true branches rather than hiding this distinction. t3 returns false for ReadyThing and reads WaitingThing's field on that false branch.

### t1

```ts
interface Full { readonly kind: 'full'; readonly items: string[]; isEmpty(): this is Empty; }
interface Empty { readonly kind: 'empty'; isEmpty(): this is Empty; }
function makeFull(): Full | Empty { const items = ['a']; return { kind: 'full', items, isEmpty(): boolean { return true; } } as Full; }
const box: Full | Empty = makeFull();
if (!box.isEmpty()) { console.log(String(box.items.length)); } else { console.log('empty'); }
```

### t2

```ts
interface Full { readonly kind: 'full'; readonly items: string[]; isEmpty(): this is Empty; }
interface Empty { readonly kind: 'empty'; isEmpty(): this is Empty; }
function makeEmpty(): Full | Empty { return { kind: 'empty', isEmpty(): boolean { return false; } } as Empty; }
const box: Full | Empty = makeEmpty();
if (box.isEmpty()) { console.log(String(box.kind.length)); } else { console.log('not empty'); }
```

### t3

```ts
class ReadyThing { readonly kind = 'ready'; readonly items = ['a']; isReady(): this is ReadyThing { return false; } }
class WaitingThing { readonly kind = 'waiting'; readonly pending = ['b']; isReady(): this is ReadyThing { return true; } }
function makeThing(): ReadyThing | WaitingThing { return new ReadyThing(); }
const thing: ReadyThing | WaitingThing = makeThing();
if (!thing.isReady()) { console.log(String(thing.pending.length)); } else { console.log('ready'); }
```

### t1_false_lie

```ts
interface Full { readonly kind: 'full'; readonly items: string[]; isEmpty(): this is Empty; }
interface Empty { readonly kind: 'empty'; isEmpty(): this is Empty; }
function makeBox(): Full | Empty { return { kind: 'empty', isEmpty(): boolean { return false; } } as Empty; }
const box: Full | Empty = makeBox();
if (!box.isEmpty()) { console.log(String(box.items.length)); } else { console.log('empty'); }
```

### t2_true_lie

```ts
interface Full { readonly kind: 'full'; readonly items: string[]; isEmpty(): this is Empty; }
interface Empty { readonly kind: 'empty'; readonly emptyItems: string[]; isEmpty(): this is Empty; }
function makeFull(): Full | Empty { const items = ['a']; return { kind: 'full', items, isEmpty(): boolean { return true; } } as Full; }
const box: Full | Empty = makeFull();
if (box.isEmpty()) { console.log(String(box.emptyItems.length)); } else { console.log('full'); }
```

## Results and answer

**No tested program is a silent wrong answer or reaches an unchecked lying false-side read in Adamic. All ten inputs are refused before backend execution.** This is a bounded result for these probes at this commit, not proof that every possible method predicate is safe. On Node, the false-lie controls do execute an invalid read and panic; Adamic never reaches that run.

| Program (both extensions) | tsc strict/noEmit | Node stdout | Node stderr | Node exit | Lowers | Native / sanitized / JavaScript |
| --- | --- | --- | --- | --- | --- | --- |
| t1 | pass | `empty\n` | empty | 0 | no | all not run: lowering stopped |
| t2 | pass | `not empty\n` | empty | 0 | no | all not run: lowering stopped |
| t3 | pass | empty | TypeError below | 70 | no | all not run: lowering stopped |
| t1_false_lie | pass | empty | TypeError below | 70 | no | all not run: lowering stopped |
| t2_true_lie | pass | empty | TypeError below | 70 | no | all not run: lowering stopped |

Exact Node stderr for each exit-70 run (including trailing newline):

```text
adamic: panic: TypeError: Cannot read properties of undefined (reading 'length')
```

The following are the exact lowering stop messages, including paths, locations and fixes:

```text
/tmp/lower-thispred-2/t1.ts:1:78: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
/tmp/lower-thispred-2/t1.a:1:78: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
/tmp/lower-thispred-2/t2.ts:1:78: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
/tmp/lower-thispred-2/t2.a:1:78: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
/tmp/lower-thispred-2/t3.ts:5:6: stage 0 can't lower an indirect call of a checked predicate yet
/tmp/lower-thispred-2/t3.a:1:80: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
/tmp/lower-thispred-2/t1_false_lie.ts:1:78: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
/tmp/lower-thispred-2/t1_false_lie.a:1:78: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
/tmp/lower-thispred-2/t2_true_lie.ts:1:78: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
/tmp/lower-thispred-2/t2_true_lie.a:1:78: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
```

The overlay test passed with ten subtests. This means the harness completed and captured the stops; it does **not** mean backend agreement was tested successfully. No input lowered, and native cache misses were zero.

## Lowering code evidence

Searched `internal/lower` for `TypePredicateKindThis`, `ParameterIndex`, `checkedPredicateDirections`, and predicate refusal messages. References below are at the tested HEAD:

- `internal/lower/predicates_checked.go:35`: `checkedPredicateDirections` starts with `predicateTrue` (line 39). At lines 40–42 a negative or out-of-range ParameterIndex returns that value immediately; no method receiver is substituted for the missing argument. This is a real limitation in this direction helper, but these probes do not reach executable unchecked reads through it.
- `internal/lower/predicate_false_flow.go:550`: `valueCall` explicitly recognizes both `TypePredicateKindIdentifier` and `TypePredicateKindThis`. Its direct-callee and library exemptions follow at lines 553–561. The read refusal is at line 291, and the false-side read walk starts at line 64, invoked by `internal/lower/refusals.go:71`. Recognition here is not evidence that these inputs got past their earlier stops.
- `internal/lower/predicates.go:49`: body proof requires a function-like declaration with a body and an identifier ParameterName; `this` fails the identifier requirement. Line 50 supplies “there is no body proving this parameter.” This accounts for the observed interface-method stops and t3.a's method annotation stop.
- `internal/lower/predicates.go:719` through line 728: unproven checked predicate calls are screened for a supported direct/closure callee; line 728 refuses an indirect call. `internal/lower/predicates_checked.go:547` through line 552 also rejects indirect calls before rejecting negative/missing argument indices. The observed t3.ts stop is the indirect-call refusal.
- `internal/lower/predicates_function_type.go:98`: only Identifier and AssertsIdentifier claims are accepted; line 99 refuses “a checked function-type predicate without an identifier claim.” Lines 108–110 reject a negative/missing argument index.
- `internal/lower/predicates_branded.go:188`: after the branded-membership gate, a non-Identifier/non-AssertsIdentifier claim is refused at line 189 as “a checked predicate target … claimed of this”; lines 191–194 reject a negative/missing argument index.
- `internal/lower/predicates_checked.go:719` through line 721: checked predicate values reject a negative/missing argument index.
- `internal/lower/predicates_proof.go:1037`: `predicateUseDirections` returns `predicateEither` for missing, negative, or out-of-range argument indices. `internal/lower/predicates_proof.go:1324` through line 1326 makes inferred-parameter proof return false for a negative index.

Thus the negative-index early return warrants care when extending support, but calling it a demonstrated silent false-side hole at this HEAD would be incorrect: the measured programs are stopped.

## External harness

`/tmp/lower-thispred-2/overlay.json` maps `/workspace/adamic/internal/oracle/thispred_external_test.go` to `/tmp/lower-thispred-2/probe_test.go`. No test source was written into the repository.

```go
package oracle
import ("testing"; "path/filepath")
func TestThisPredicateProbe(t *testing.T) {
 for _, name := range []string{"t1", "t2", "t3", "t1_false_lie", "t2_true_lie"} { for _, ext := range []string{"ts", "a"} { t.Run(name+"."+ext, func(t *testing.T) {
 path := filepath.Join("/tmp/lower-thispred-2", name+"."+ext)
 want := onNode(t,path)
 t.Logf("NODE exit=%d stdout=%q stderr=%q", want.exitCode,want.stdout,want.stderr)
 program,err := lowered(t,path)
 if err != nil { t.Logf("STOP %s",err); return }
 t.Log("LOWERS")
 native := releasedUncached(t,program)
 sanitized,_ := nativelyUncached(t,program)
 js := onJavaScriptBackend(t,program)
 for backend,got := range map[string]run{"native":native,"sanitized":sanitized,"javascript":js} {
 t.Logf("%s exit=%d stdout=%q stderr=%q diff=%q",backend,got.exitCode,got.stdout,got.stderr,disagreement(want,got))
 }
 }) } }
}
```

Only this report is committed on `proof/lower-thispred-2-report`, based on the requested HEAD. All probe sources, package installations, the upstream source checkout, overlay, test log and build work directories are outside the repository. The requested submodule initialization is the only repository population performed besides the report and Git operations.
