# Views fix-forward D on main

Branch: `compiler/loom-r-views-fx-d-ac-m`.

Base: `561aeb4b851ca0820f39440a09dc5fb3fd27ba40`, also confirmed as remote main after the test runs started.

Applied D and its work-in-progress walk fix using `git merge-tree --write-tree --merge-base=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 origin/main b4b56e2de851`, producing tree `721f0d688d02edd6617c2b1bdb721c87f1d7dfab`. The merge had no conflicts. A second merge-tree using D alone gives tree 6b53c2e10b11374386373dad693be8e65498cfe9. [runtime-beyond-D.patch](runtime-beyond-D.patch) lists every removed and added runtime line relative to that tree, with exact text and locations. Only view_unions_untagged.c changes beyond D.

Assumptions: D means commit 065366f362ea04ce2dd14c06ffe2fcf7baae0d82, including its original audit files. The requested original message means that commit's complete message. A private runtime restoring its exact recursive matcher is the revert mutant; it does not change the fixture, receiver tag, generated program, or wasm stack size.

The iterative matcher stores active calls on the heap. It preserves branch order, active-path cycle checks, finite tags, open kind selection, optional fields and callable producer matching. It frees its work allocation after an answer. The entry stack check preserves the forced-limit exit. The deep-chain fixture is registered as an ordinary Node comparison, with its receiver tag and checked-view cast intact. Its focused native check also checks chain release with the leak checker.

Counts were regenerated on Linux by `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=60m -args -update-counts` using clang 20.1.8. The command passed in 197.857 seconds. No existing main row changed. The exact added rows follow (columns: fixture, allocations, frees, retains, releases, peak live, in regions). [counts-changed.patch](counts-changed.patch) preserves the complete diff.

| Fixture | Allocations | Frees | Retains | Releases | Peak live | In regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/views_deep_chain.a | 2001 | 2001 | 2004 | 2007 | 2001 | 0 |
| internal/oracle/testdata/review/agree/fxspptb_oct9_native_p18_cast_kind_no_field_types.a | 4 | 4 | 5 | 9 | 4 | 0 |
| internal/oracle/testdata/review/agree/fxspptb_oct9_native_p20_callable_member_call_good.a | 7 | 7 | 6 | 14 | 5 | 0 |
| internal/lower/testdata/fx7_wrong_aborts/wide.a | 5 | 5 | 6 | 11 | 5 | 0 |
| stage3/interface-downcasts/fx7/depth-65.a | 66 | 66 | 70 | 74 | 66 | 0 |
| stage3/interface-downcasts/fx7/depth-200.a | 201 | 201 | 205 | 209 | 201 | 0 |
| stage3/interface-downcasts/fx7/depth-2000.a | 2001 | 2001 | 2005 | 2009 | 2001 | 0 |
| stage3/interface-downcasts/fx7/p09.a | 2 | 0 | 6 | 5 | 2 | 0 |
| stage3/interface-downcasts/fx7/p32.a | 4 | 2 | 9 | 9 | 3 | 0 |
| stage3/interface-downcasts/fx7/p33.a | 5 | 2 | 10 | 9 | 4 | 0 |
| stage3/interface-downcasts/fx7/p67.a | 2 | 0 | 8 | 6 | 2 | 0 |
| stage3/interface-downcasts/fx7/p72.a | 3 | 0 | 6 | 5 | 3 | 0 |
| stage3/interface-downcasts/fx7/scalar-tuple.a | 2 | 0 | 5 | 5 | 2 | 0 |
| stage3/interface-downcasts/fx7/scalar-static.a | 2 | 0 | 6 | 6 | 2 | 0 |
| internal/oracle/testdata/review/agree/fxspptb_oct9_views_p68_object_view_over_tuple.a | 4 | 4 | 8 | 14 | 4 | 0 |
| internal/oracle/testdata/review/agree/fxspptb_oct9_views_p69_tuple_view_over_tuple_union.a | 3 | 3 | 8 | 13 | 3 | 0 |
| internal/lower/testdata/fx7_tuple_recognition/p69.a | 4 | 4 | 9 | 15 | 4 | 0 |
| internal/lower/testdata/fx7_wrong_aborts/bad_callable.a | 2 | 0 | 6 | 4 | 2 | 0 |
| internal/oracle/testdata/loom_views_d/checked-cast.a | 4 | 4 | 5 | 9 | 4 | 0 |

Focused checks passed:

- `TestFX7ViewCycleAndStackGuard`: cycle match, 2,000-link match with a 512 KiB native stack, chain release, forced-limit exit 70; unsanitized and sanitized builds.
- Both `TestFX7TupleSelectionBeforeDiagnosticKind` controls.
- `TestViewsDeepChain`: Node, native, sanitized native, JavaScript and a separate leak check.
- `TestWASIAgreesWithNode_007`: all seven fixtures, including the original depth-2000 witness.
- `TestViewsDeepChainRecursiveWalkMutant`: iterative WASI control agrees with Node. Restoring D's recursive matcher gives exit 70, empty stdout and `adamic: panic: RangeError: Maximum call stack size exceeded`; Node gives exit 0 and `true\n`. The final test also pins this exact failure reason.
- `go vet ./...`, including a second run after the mutant assertion was tightened.
- `gofmt` for every changed Go file; compiler-source whitespace check (see final check scope below).

The complete WASI family passed with `ADAMIC_ORACLE_WASI=1 ADAMIC_TEST_WASI=1 go test ./internal/oracle -run '^TestWASIAgreesWithNode_' -count=1 -timeout=60m -json`: 130 top-level tests (128 numbered shards and two solo fixtures) and 1,067 fixture subtests passed, with no failures or skips, in 282.770 seconds. The registered deep-chain fixture passed in shard 015. Shard 007 was also run separately.

`internal/oracle` passed with both WASI switches enabled: `go test ./internal/oracle -count=1 -timeout=90m -json`, 6279 passed tests/subtests, 138 skips, no failures, 1216.309 seconds. `oracle-results.json` separates top-level totals; `oracle-skips.txt` records every skip and its exact test message. The strict recursive mutant passed in this complete run as well.

`internal/lower` passed in 114.715 seconds; `internal/javascript` passed in 0.486 seconds. They ran with both WASI switches enabled in `go test -p 1 ./internal/native ./internal/lower ./internal/javascript -count=1 -timeout=60m`. That command reported the native environment failure described below; the other two packages passed.

Environment correction during the complete native run: `TestLongUnicodePortabilityGCC` was the only reported failure in that run (1166.340 seconds). Debian GCC 14.2.0 and libasan were installed, but the system did not include `sanitizer/allocator_interface.h`, now needed by main's host_runtime.c. GCC's libasan exports both functions the runtime calls. The header is present in the installed clang 20.1.8 SDK. Passing `C_INCLUDE_PATH` did not work because the repository intentionally strips it from compiler commands. No runtime source was changed for this.

The local setup now links `/workspace/adamic-tools/gcc-sanitizer-headers/sanitizer` to `/workspace/adamic-tools/llvm/lib/clang/20/include/sanitizer` and provides `/workspace/adamic-tools/bin/gcc` as:

```sh
#!/bin/sh
exec /usr/bin/gcc -idirafter /workspace/adamic-tools/gcc-sanitizer-headers "$@"
```

The focused GCC check then passed in 20.187 seconds: both backends match Node, with ASan, UBSan and leak detection clean. The complete native package then passed with both WASI switches enabled: `go test ./internal/native -count=1 -timeout=60m -json`, 931 passed tests/subtests, three skips, no failures, 449.072 seconds. `native-results.json` separates top-level totals and `native-skips.txt` records the exact skip messages. Assumption: supplying the installed sanitizer interface header to the existing GCC toolchain is an environment repair within the requested setup and validation work.


Final checks: the runtime changes beyond D still match [runtime-beyond-D.patch](runtime-beyond-D.patch) exactly, all changed Go files pass gofmt, and the compiler-source `git diff --check` passes. The whole diff reports whitespace in retained D audit files and blank context lines in the exact runtime patch; `whitespace-warnings.txt` preserves those warnings. These evidence files are kept intact. The counts diff still matches the saved measurement. The branch contains one commit on the stated main base, with D's original message followed by the requested re-cut line and co-author trailer. The final commit SHA and push result are supplied in the delivery message, since a commit cannot contain its own SHA.
