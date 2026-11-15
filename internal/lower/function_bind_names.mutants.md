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
