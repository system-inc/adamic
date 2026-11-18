# Bind name reach mutant (#x909e6q)

Run on this Linux box on 2026-11-15. The mutant uses Go's file overlay, so the
real compiler source and concurrent unmutated builds stay unchanged.

In `internal/lower/function_bind_names.go`, in the function returned by
`boundNameReach`, change exactly:

```go
if entry == nil || entry.unreadable || l.boundMemberUnreadable(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression), target.Name().Text()) {
```

to:

```go
if entry == nil || l.boundMemberUnreadable(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression), target.Name().Text()) {
```

This treats an unreadable name as its readable candidates alone, which can be
empty, instead of falling back to every function of the bound type.

Command (after sourcing `/workspace/adamic-tools/env.sh`):

```sh
go test -overlay=/tmp/bind-name-mutant-overlay.json -json ./internal/lower -run '^TestBoundNameReach$/^computed_union$' -count=1
```

Exit 1. The intended red assertion, without a build or sanitizer failure:

```text
--- FAIL: TestBoundNameReach/computed_union
function_bind_names_test.go:58: want NotYet "bind of a function that declares 2 parameters where the type it's bound as names 1", got <nil>
```

The union key is `f | g`. Its logical write reaches IR as a `SetProperty` under
`LogicalAssignment.KeyNames`; fresh records both names as `WriteField`. The
stored alias cannot be read by the plain-assignment reader, so f is unreadable.
The readable literal initializations under f take one parameter, while `held`
may store the two-parameter function. Dropping the fallback loses that function.

The restored compiler's full bind selection and whole-package checks are
reported with this head. The raw mutant log is `/tmp/bind-name-mutant-final.log`.

## Roads that stop before this scan

Ordinary shaped-object index assignment stops in `setIndex`; a write through
any stops on its representation. A finite logical assignment is the admitted
computed-field path above. An unresolved key is conservatively charged to its
receiver's shaped members, with any/unknown affecting every name, a union
including any member, and a type parameter following its constraint. An index
signature writes record storage, so it never charges a shaped field.

Computed `Object.defineProperty` fails `recordDefineShape`'s literal-key test
and keeps the property-descriptor refusal. `Reflect.set` stops as a namespace
member without a lowered binding. Neither produces a `WriteUnknown`: neither
reaches IR. `TestBoundNameEarlierWriteStops` pins these stops.

Function-valued `Object.assign` sources are already refused. The admitted
`Object.assign(target, {})` still falls back when its result includes f,
as required; `assign_result` pins this independently of the earlier refusal.
A subclass adding a parameter already stops at the override-count check. The
narrower-interface subclass pin also reaches the bind count and keeps every
method from `MethodsNamed`.

A final soundness probe found that `const host: Record<string, Unary> = { f: two };
host.f?.bind(otherReceiver)` also has fixed-name syntax but reads record storage.
The first implementation admitted it with an empty name set. Record entries are
outside this table, so indexed-storage bind targets now keep the type-wide
fallback. `TestBoundNameRecordTargetKeepsTypeWideFallback` pins that boundary.
The original probe is preserved at `/tmp/bind-record-probe.ts`; it printed
`Lower: <nil>` before this guard. The final pin requires the two-versus-one stop.

## Accessor probes on eb086fcd

`TestBoundNameAccessorsStayStopped` runs each source with Node, requiring exit 0,
empty stderr, and stdout `1\n9\n`. Each source calls a closure before the bind
and passes the bound unary view to a callback consumer that supplies two arguments.

- `bind_name_accessors/literal_getter.ts:8:15` stops with
  `stage 0 can't lower bind of a getter-backed method value yet`.
- `bind_name_accessors/class_getter.ts:8:28` stops because its view's
  `readonly field f becomes writable` (`adamic/invariant-mutable`).
- `bind_name_accessors/class_getter_readonly.ts:9:15` preserves readonly to reach
  the bind, which stops with `stage 0 can't lower bind of a getter-backed method value yet`.
- `bind_name_accessors/setter.ts:9:15` stops with
  `stage 0 can't lower bind of a function that declares 2 parameters where the type it's bound as names 1 yet`.
  This source isolates a setter named f, storing its argument into the data field f
  later bound. `TestBoundValueReach` additionally pins its allocation reach as unknown.

The getter guard in `bindReceiverRule` scans source getters before the name reach
can admit these binds. The setter write is unreadable and retains the type-wide
count. None of these probes lowers, so none reaches either backend with a wrong
answer. No lowering change or accessor-check mutant was made: the requested new
check and its mutant were conditional on a probe lowering. The unreadable-write
mutant above remains the regression for dropping that fallback.

Accessor follow-up validation (after sourcing `/workspace/adamic-tools/env.sh`):

```sh
go test -json ./internal/lower -run '^TestBound(NameAccessorsStayStopped|ValueReach)$' -count=1 -timeout 10m
GOMEMLIMIT=6GiB go test -json ./internal/lower -run 'Bind|Bound' -count=1 -parallel 8 -timeout 30m
gofmt -l -e internal/lower internal/oracle
go vet ./internal/lower
go test -json ./cmd/adamic-gate/parallelcensus -count=1
ADAMIC_GATE_UNCACHED=1 go test -json ./internal/oracle -run '^Test(Function|Method)Bind' -count=1 -timeout 30m
```

The focused accessor/reach run passes 6 tests and subtests; the lower bind
selection passes 313; parallelcensus passes 6. Formatting emits no paths and vet
passes. Lowering code is unchanged, so the conditional whole-lower rerun is not
needed for this follow-up.
The function/method bind oracle selection passes 58 tests and subtests with 2
WASI skips. Raw follow-up logs are `/tmp/bind-accessor-pins.log`,
`/tmp/bind-accessors-bind-tests.log`, `/tmp/bind-accessors-parallelcensus.log`,
and `/tmp/bind-accessors-oracle.log`.
