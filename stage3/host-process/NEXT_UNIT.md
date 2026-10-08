Language blockers remain in 14, 16, 17, 18-20 and 22; no language construct or acceptance fixture was rewritten.
Parent commit 776f9d6; checkpoint 2316c2c preserves Darwin malloc declarations and proves System.newLine namespace reads already work on this branch.
Commands: namespace oracle PASS 10.877s; complete process subset PASS 23.848s; native package PASS 132.735s; flow package PASS 86.527s; affected-package vet exit 0.
Mutants: native EOL LF changed to CRLF and generated JavaScript EOL LF changed to CRLF each run cleanly and are caught only by Node stdout comparison.
Not covered yet: full acceptance fixtures past fs.mkdtempSync pending the supplied area merge SHA, language changes, or a macOS compile/run.

## Exact language constructs

| Host fixture | One-line reproducer | Observed result |
| --- | --- | --- |
| 14_getCurrentDirectory | `let callback: () => void = () => {}; callback = undefined!;` | Refused non-null assertion at 1:49 |
| 16_getEnvironmentVariable | `const f = (s: string) => s || "";` | NotYet BinaryExpression with a string and a string at 1:26 |
| 17_write | `const sys = { write(s: string): void {} }; sys.write("x");` | NotYet literal method call with an unrepresented result at 1:44 |
| 18_exit_0, 19_exit_1, 20_exit_2 | `let activeSession: undefined;` | NotYet value of type undefined at 1:5 |
| 18-20, next independent construct | `const f = (exitCode?: number): void => {}; f(0);` | NotYet function value with an optional parameter at 1:12 |
| 22_createHash_fallback | `const crypto: undefined = undefined;` | NotYet value of type undefined at 1:7 |

The method-result refusal belongs to the plain object wrapper's void return: lower/iteration.go literalMethodCall rejects slotless results. It reproduces with an empty method and no Node imports, so it is not stdout.write's boolean result. The string || refusal also reproduces with no Node import. Both are compiler-owned language work. The current first undefined failure in 18-20 is the activeSession declaration; optional function values are an independent blocker rather than the observed first failure.

## Namespace and macOS checkpoint

23's exact library-bearing shape `import * as _os from "node:os"; const nodeSystem = { newLine: _os.EOL };` already compiles with the existing namespace lowering on this branch. The process namespace oracle now includes the exact JSON.stringify and diagnostic/version concatenation body from the acceptance fixture and compares source Node, sanitized native and generated JavaScript. This is a library proof only; the complete unmodified acceptance fixture still stops at mkdtempSync before the upcoming fs-file merge. The public host status.json is unchanged.

node_process.c now defines _DARWIN_C_SOURCE immediately after _POSIX_C_SOURCE. This keeps Darwin's malloc_zone_t declarations visible; Linux compilation is exercised by the namespace oracle. No new runtime translation unit was added, and no macOS outcome is claimed. Keep the identical macro from fs-file when the provided area commit is merged, without rebase or force push.

Logs are in logs/next-unit/. The native/flow package gate and complete process subset passed, including the existing host/performance mutants. Affected-package vet exited 0. The exact commands and outcomes are retained in process.log, packages.log and vet.log. All 25 host fixtures will be rerun after the user supplies the area SHA containing fs-file.

```sh
go test ./internal/native ./internal/flow -count=1 -timeout 30m > /tmp/process-next/packages.log 2>&1
go test ./internal/oracle -run '^TestNodeProcess|^TestProcessExitOutput$|TestInputAgreesWithNode/internal/oracle/testdata/node_process' -count=1 -timeout 30m -v > /tmp/process-next/process.log 2>&1
go vet ./internal/oracle ./internal/native ./internal/lower > /tmp/process-next/vet.log 2>&1
```
