# Exact published parser scratch refusal

The disposable worktree starts at unchanged 33bf53aca3ce199aa05fe7bd2e7898e7c5ef2d5b,
not a reconstructed integration. Publication be7bdb8b adds only its quarantine
note. No scratch commit, merge or worktree is pushed. The cohere submodule pin
7945d102 matches this checkout; the worktree references the existing submodule
through a local symlink, without copying SDK code or changing its pin.

The original compiler builds and the exact native-arguments-length-value.a
witness compiles successfully. It prints 9, exits 0 and has empty stderr.
Source Node prints 1. Its generated C is saved unchanged alongside this report.
The wrong binding reads arguments[0].number; the actual count is a separate
parameter which the body ignores. The stored-slot bug is visible in the exact
published compiler, independently of the earlier argument-order inference.

Applied enforcement is saved verbatim as exact-parser-enforcement.patch.
It introduces the branch's one typedef per plain/counted shape, underlying
function typedef declarations, constructor _Generic source checks and
-Wcast-function-type-strict under -Werror; it also removes both erased nominal
virtual pointer casts using the existing numeric-identity/typed-direct-call
implementation. Because this old IR lacks PackedCountNeeded, prototype selection
uses its ReadsArguments/rest observations. This witness is an arguments reader,
so the counted prototype has exactly the branch's shape. No reconciliation of
scratch function signatures, constructor selection, call sites, or count-slot
binding is performed. Those remain deliberately incompatible.

The enforced Go compiler builds. The native build now exits 1 in clang:

```
array.c:202:72: error: too many arguments to function call, expected 2, have 3
  double result = compare->code(compare, (adamic_value[]){left, right}, 2).number;
```

Native Build compiles its runtime first, so that error stops it before the
witness's translation unit. Compiled the unchanged enforced generated C
separately against that same runtime header and the same strict release flags,
with -c, to prove the witness itself also fails for its typed constructor:

```
exact-parser-enforced.c:39:40: error: controlling expression type
'adamic_counted_code_function *' (...) not compatible with any generic association type
  adamic_closure * adamic_temporary_4 = adamic_closure_new(adamic_function_1_reader_value, 0);
```

The old caller stores a counted function through the plain constructor.
The new header refuses that conversion before it reaches storage. Both complete
compiler diagnostics are saved. No unrelated missing host declaration occurs
in the generated-C diagnostic. No enforced executable exists or is run.
The branch's reconciled source-on-Node release/sanitized acceptance tests remain
necessary: a correctly typed function that reads the wrong buffer slot is a
semantic bug, which C's type system cannot itself detect.

Commands in /tmp/closure-exact-parser-scratch, after sourcing tool env:

```
go build -buildvcs=false -o /tmp/closure-exact-parser-scratch-adamic ./cmd/adamic
/tmp/closure-exact-parser-scratch-adamic build /workspace/adamic/internal/oracle/testdata/native-arguments-length-value.a -o /tmp/closure-exact-parser-scratch-original
/tmp/closure-exact-parser-scratch-original
git apply /path/to/exact-parser-enforcement.patch
go build -buildvcs=false -o /tmp/closure-exact-parser-enforced-adamic ./cmd/adamic
/tmp/closure-exact-parser-enforced-adamic c /workspace/adamic/internal/oracle/testdata/native-arguments-length-value.a
/tmp/closure-exact-parser-enforced-adamic build /workspace/adamic/internal/oracle/testdata/native-arguments-length-value.a -o /tmp/closure-exact-parser-enforced
clang -std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -O2 -ffp-contract=off -fno-optimize-sibling-calls -I internal/native/runtime -c /tmp/closure-exact-parser-enforced.c -o /tmp/closure-exact-parser-enforced.o
```

Each command's output was redirected to its adjacent log; the C command's
stdout was saved as exact-parser-enforced.c.txt. Hashes, exit statuses and exact
input commits are in exact-parser-replay.json. All mutation is confined to the
detached scratch; only report, patch and raw evidence are committed here.
