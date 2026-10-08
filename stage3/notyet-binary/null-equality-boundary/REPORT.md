# Null identity remains a representation boundary

Zero additional raw table sites claimed: no measured operand in the selected binary CSV contains null. The existing nullable RegExp match representation stores null as NULL; references permitting undefined also store their missing value as NULL. Differing reference normalization into Union and same-array-representation comparisons can therefore collapse distinct values.

Node independently reports `null === undefined` false and `null !== undefined` true. Two controls use RegExpExecArray|null against Sized|undefined and against RegExpExecArray|undefined. The explicit boundary now stops comparisons that would erase their nullish identity. Literal comparisons continue through their existing exact handling; comparisons with no null tag ambiguity keep the prior rule. A distinct boxed null (or explicit comparison operand metadata) would be a representation choice. No such choice is made here and no runtime C helper is added.

Observations: removing the boundary or omitting its same-representation undefined term compiles valid C and finishes with exit 0 in both backends. Source Node and the JavaScript backend print `false:true`; native prints `true:false`. Both mutants are caught by TestMixedEqualityBoxedNullBoundary and restored. Neither kill is a compilation failure. The initial ordinary number[]|null probe stopped earlier at the existing unsupported type; it is not counted as a reproduced miscompile. The admitted RegExp match controls prove the actual issue.

Commands and attached output:

- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestMixedEqualityBoxedNullBoundary -count=1 -timeout 15m`: pass 0.290s.
- `python /tmp/binary-null-equality-mutants-final.py`: both semantic mutants caught; C finishes normally and Node disagrees.
- `go test ./internal/lower -count=1 -timeout 15m`: pass 16.519s.
- `go vet ./internal/lower ./internal/oracle`: exit 0.
- Complete owned/imported oracle selection after the guard: pass 13.546s, including the nine checked assignment target repairs/controls and prior logical/equality/assignment fixtures.

The controls create temporary .a inputs; no new registered fixture count row is needed. The preceding normal counts refresh added the string accessor's row and changed no previous row. Files outside owned combine: internal/oracle/binary_null_boundary_test.go and this evidence directory. No other worker lower function changed.
