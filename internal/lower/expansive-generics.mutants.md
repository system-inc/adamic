# Expansive generics (#ykd89v9)

One edit, run on a house box or a sealed one, then put back.

## Every edge plain

In `occurrences` (expansive_generics.go), keep the call and make the expanding test answer false, so
no edge is ever expanding and no cycle is ever refused:

| Old text | New text |
|---|---|
| `			each(parameter, exposure == expansiveExpanding)` | `			each(parameter, false)` |

Run:

```
go test ./internal/lower -run 'TestExpansiveGenerics' -count=1 -v -timeout 60s
```

The red is Go's test timeout, `panic: test timed out after 1m0s`, with rows of
`TestExpansiveGenericsStopsWhereTheGenericIsDeclared` still running, as on main before the rule.
The rows expected to time out are "in an array", "in an object type" and "through a second
generic", in .a and .ts: holdsVoidWalk and the other walks with no bound meet them first.

Rows "in itself" and "beside a map" may not hang. The cycle finder's use walk keeps typeNesting's
bound (cycles.go, #ykd89v9 -1) as defense, so they can stop there in bounded time, red with the cycle
finder's words, "finding cycles through a type that instantiates itself deeper at each level", where
this rule's are wanted. "in a tuple" may stop there too: cycles_nesting_test.go's tuple row is the
same program and is held by that bound. "a type alias", "through a base" and
`TestExpansiveGenericsStopsAcrossTwoModules` are red either way, by a timeout or by the bound's words.

`TestExpansiveGenericsControlsLower` and `TestExpansiveGenericsLeavesDeepReadonly` make no expanding
edge, so the mutant changes nothing for them, and they pass before the timeout.
