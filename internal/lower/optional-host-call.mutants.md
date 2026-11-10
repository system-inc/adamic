# Mutants for a call through ?. to a Node host method (#3625x53)

The pins are `TestOptionalHostCalls` in internal/oracle/optional_host_call_test.go,
walking internal/oracle/testdata/optional_host_call/. n1 (Dirent), n2 (Stats) and
n3 (Stats | Dirent | undefined), in both .a and .ts, now lower through the shared
receiver guard and agree with Node on undefined and present receivers. n3 presents
both host types. n4 keeps its namespace-value stop, n5 its optional-function-call
stop, and p1 the program-declared method's agreement.

The unguarded witness is n6_unguarded_process_cwd.a and its .ts twin. A parameter
is typed NodeJS.Process | undefined and called only with undefined; unlike n4,
it never hands the process namespace over as a value. Its host method is
`node:process.Process.cwd`. nodeHost lowers that call to host_cwd, which does not
read or guard the receiver. Node prints `undefined\n`; an admitted unguarded call
prints `present\n` on release native, sanitized native and JavaScript, exit 0,
stderr empty. The difference is stdout line 1, not a compiler or tool failure.

## (a) The rule removed

In internal/lower/expression.go, optionalCall, replace exactly:

```go
return l.optionalHostCall(call, method)
```

with:

```go
return nil
```

The direct pins `TestOptionalHostCalls/n6_unguarded_process_cwd.a` and
`TestOptionalHostCalls/n6_unguarded_process_cwd.ts` must fail at
`lowered, want "...node:process.Process.cwd; #3625x53) yet"`.
The mutated lowering must admit each program, then disagree with Node on all
three backends with `present\n` versus `undefined\n`. n1 through n3 now guard
their own receivers and stay green: they are no longer unguarded witnesses.

## (b) An unguarded member put in the table

In internal/lower/optional_host_call.go, add this one entry to the fourteen
StatsBase/Dirent members in optionalHostCallGuarded:

```go
"node:process.Process.cwd": true,
```

The same n6 pair must fail at the same `lowered, want` catch line and the admitted
programs must show the same stdout-line-1 difference on all three backends.
nodeFSEntryMethod reads the table's StatsBase/Dirent keys only: admitting some
other host must not silently install an optionalIntrinsic guard for it. Thus
this single-entry mutant bypasses the rule but retains cwd's unguarded host
lowering. No mark or nullish guard is removed in either mutant.

Run each single-edit mutant through the n6 refusal pair, requiring a real red,
and through a run of the admitted program against Node, requiring the wrong
line above. Restore the edit; the unmutated fourteen-program walk must pass.
Keep these oracle checks in the oracle unit, with its declared runtime tools;
the lower mark catchers remain lower-only.
