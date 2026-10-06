# Object prototype integration review

Base: codex/library-object at cdf632b47fefa315aefdeaeda468c13abed4d551.
Linux: Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5.
`bash cloud/setup.sh`: go, clang, node and submodules ready, each 0s.

## 1. Private mangled slots

`adamic_object_has` skips shape names beginning with #. The JavaScript
backend uses `adamicHasOwn`, evaluates both arguments once, and filters the
same private names before `Object.hasOwn`.
The supplied inherited-private fixture prints `false false true` on Node,
sanitized native, release native (-O2), and the JavaScript backend.

Uncached command: `go test -count=1 -timeout 10m ./internal/oracle -run
TestNativeAgreesWithNode/internal/oracle/testdata/library_object_private_mangled`.
Passed (9.321s). Mutant scanning # slots again: Node `false false true`,
native `true true true`, caught only by stdout comparison (9.213s).
No shared lowering or emission hooks changed for this item.
