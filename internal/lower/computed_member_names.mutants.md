# Computed member names: mutants (#ppgk12h)

Each mutant puts one walk back to asking a member's name node for its text, which panics on a computed
name ("Unhandled case in Node.Text: *ast.ComputedPropertyName"). TestComputedMemberNamesStopWithoutPanic
(computed_member_names_test.go) recovers that panic and fails by name. Run as J-enhkzt9 (sealed), each
mutant applied alone and the file restored after.

| Mutant | Where | Red |
|---|---|---|
| helper: memberKeyText returns `node.Name().Text(), true` | class_accessors.go | every pin |
| noteAccessorNames asks `node.Name().Text()` | class_accessors.go | the seven getter pins (the harness's SortedMap, string, numeric and non-constant getters in a class, non-constant and symbol getters in a literal, bind beside a symbol getter) |
| bindGetters asks `current.Name().Text()` | function_bind.go | a bind beside a symbol getter |
| dynamicReadHazard asks `member.Name().Text()` | unknown.go | both unknown-read pins beside an iterable class |
| presenceHazard asks `member.Name().Text()` for a method | unknown.go | both unknown-read pins beside an iterable class |

interface_cast.go's two getter walks and view_sites.go's getter filter take the same helper; no pin
reaches them on their own, so they're held by the helper mutant and by reading.

internal/oracle/testdata/computed_member_names_iterable.ts holds the unknown-read shape to Node on the
native backend and on the three WebAssembly families (J-jvx5kkx); before this change it panicked.
