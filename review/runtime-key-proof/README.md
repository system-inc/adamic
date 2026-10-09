Proved TestRuntimeKeyKeepsBoundaries as written; no test or runtime correction needed.
Feature base: runtime/guard-runtime-key 6cdec81378d243013870e97dfcbf1f67aa920ccb.
Control: both TestRuntimeKey tests pass on the branch and main fc31f9f15f95d841d4cd230a37ba55852866f2c8.
M04: flags and compiler/version assertions fail; the existing test and file name/contents case remain green.
Scope: native key tests and native vet, followed by the test-only lane checks; no full oracle or repository test suite.

The audit patch is exactly
origin/test-audit/internal-native-library c15f21ef81ffe692bf00b1728e98bd55cc80b255,
review/test-audit/internal-native-library/M04.diff. It removes the string-part
length prefix in runtimeKey while leaving the explicit file-contents length prefix.
The file case is retained to guard the full implementation's boundary.

The original binary build stopped before compilation because the sourced TMPDIR,
/tmp/adamic-gate, did not exist in the fresh environment. bash cloud/setup.sh fixed
that and completed in 43.940s: Go 0.026s, submodules 0.070s, Node 0.073s,
markdown dependencies 0.206s, clang 0.434s, cache warm 43.917s. Go 1.27.1,
clang 20.1.8, Node 24.19.0, nproc 5, cgroup quota 4 CPUs.

Each binary build ran in the background under timeout 900 and was checked at least
every minute, outside the test execution budget. The test command was exactly:

```sh
timeout 90 /tmp/native.test -test.run '^TestRuntimeKey' -test.count=1 -test.timeout 90s -test.v
```

The unmodified branch exited 0: TestRuntimeKeyKeepsBoundaries 0.00s,
TestRuntimeKeyIncludesEveryInput 0.07s. `timeout 90 go vet ./internal/native`
exited 0 with no output.

Applied M04, rebuilt successfully, and ran the same command. It exited 1 solely
on these two assertions:

```text
library_test.go:56: flags ["a" "bc"] and ["ab" "c"] share a key
library_test.go:59: compiler "clang" with version "version 1" and "clangv" with "ersion 1" share a key
--- FAIL: TestRuntimeKeyKeepsBoundaries (0.00s)
--- PASS: TestRuntimeKeyIncludesEveryInput (0.07s)
```

The file case executes after both nonfatal t.Error calls and produces no failure.
This is the permitted outcome: its independent contents-length prefix still
separates the two inputs under M04. Nothing was skipped or weakened.

Reversed M04 with git apply -R, verified zero diff in internal/native/library.go,
rebuilt, and reran the exact command: exit 0, both tests green (0.00s and 0.07s).
Native vet again exited 0 with no output.

A detached checkout of fetched main fc31f9f1 received only the original test diff
from 6cdec813. Its pinned cohere and TypeScript submodules were initialized from
the existing local checkouts. Built /tmp/native-main.test from that checkout under
the same background deadline, then ran the same test command against that binary:
exit 0, boundary test 0.00s and existing input test 0.08s. This is an actual main
control, without merging or rewriting the feature branch or main.

The evidence files are gzip-compressed raw logs, with timestamps disabled in their
headers. Builds produce empty logs on success. Production code is unchanged;
the new commit records the observed proof instead of the author's unrun forecast.
