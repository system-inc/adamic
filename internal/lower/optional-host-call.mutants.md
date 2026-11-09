# Mutants for a call through ?. to a Node host method (#3625x53)

The pins are `TestOptionalHostCalls` (internal/oracle/optional_host_call_test.go) over
internal/oracle/testdata/optional_host_call/: n1 a Dirent's `entry?.isFile()` on an undefined entry, n2 a
Stats', n3 the `Stats | Dirent | undefined` shape, each stopping by the rule in its .a and .ts twin, n4 `host?.cwd()` on a local
`NodeJS.Process | undefined`, which stops first at handing process over as a namespace value, and p1 a method the program declares, which lowers and
agrees with Node. Each mutant is one edit, run on a house box,
then put back. Both measured on Codex 01a1beee at 197aece1.

## (a) The rule removed

internal/lower/expression.go, `optionalCall`. Old:

```go
		if method := l.checker.GetSymbolAtLocation(callee); method != nil && len(method.Declarations) > 0 && !load.IsLibrary(ast.GetSourceFileOfNode(method.Declarations[0])) {
			return l.optionalHostCall(call, method)
		}
```

New:

```go
		if method := l.checker.GetSymbolAtLocation(callee); method != nil && len(method.Declarations) > 0 && !load.IsLibrary(ast.GetSourceFileOfNode(method.Declarations[0])) {
			return nil
		}
```

Red: `TestOptionalHostCalls/n1_dirent.a` and `/n1_dirent.ts`, `lowered, want "...(node:fs.Dirent.isFile; #3625x53) yet"`.
The call reaches nodeFSDirectoryCall's Dirent arm, which never reads the optional flag, and lowers as a plain
host call. Run, `kind(undefined)` is a null access natively and a TypeError on the JavaScript backend, where
Node prints `undefined`. n2 stops at nodeFSFile's own `node:fs.StatsBase.isFile through an optional call`, red
on its stop text. n3 goes whichever way the union's first declaration sends it: nodeHostMember stops reading at
StatsBase's, so it's nodeFSFile's stop, and with Dirent's first it lowers as n1 does. Measured: n1 and n3 lower; at the parent both end in a null-member access natively and a
TypeError on the JavaScript backend where Node prints undefined. n4 keeps its namespace-value stop. p1 lowers as before.

## (b) An unguarded member put in the table

The table is empty on f2dcbfd9, so emptying it changes nothing and isn't a mutant here. The edit that shows the
table is read is the opposite one, an entry for a host lowering that doesn't guard its receiver.

internal/lower/optional_host_call.go. Old:

```go
var optionalHostCallGuarded = map[string]bool{}
```

New:

```go
var optionalHostCallGuarded = map[string]bool{"node:fs.Dirent.isFile": true}
```

Red: `TestOptionalHostCalls/n1_dirent.a` and `/n1_dirent.ts` lower, with the run-time failure of (a), and
`/n3_union.a` and `/n3_union.ts` still stop but name only `node:fs.StatsBase.isFile`, red on their stop text.
n2 and p1 are unchanged. When a host lowering comes to guard its receiver and its member goes in the table, its
pin row runs it on an undefined and a present receiver against Node, and emptying the table turns that row into
this NotYet.
