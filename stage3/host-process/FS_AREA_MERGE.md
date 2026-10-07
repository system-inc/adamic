Unknown-error source narrowing remains Refused at in; assigned language blockers are unchanged, while full 23_newLine now executes correctly.
Merged area/library 53f44e05 into c80c8bd on codex/host-process-land, preserving the FIFO barrier and all incoming macOS fixes; node:perf_hooks remains included.
Commands: Linux counts PASS 91.951s; affected packages and flow passed; 25 host records PASS 14.396s, with 11 and 23 agreeing on native and JavaScript; whole oracle PASS 364.432s.
Mutants: all 25 supplied host semantic mutants and the status-diagnostic mutant are caught; whole oracle also runs process, perf_hooks, cwd-barrier and fs mutants.
Not covered: macOS execution, remaining 23 blocked host sources, language fixes, native tsc, timers and nextTick scheduling.

Two conflicts were resolved: keep area's Darwin explanation together with the shared _DARWIN_C_SOURCE macro, and regenerate counts on Linux rather than select either parent's counts. The merge also keeps canonical filepath.EvalSymlinks temp directories together with this branch's relocated runtime-only witnesses. Terminal drains run while children are alive; detect_leaks=1 is selected only on Linux. The incoming process output/terminal/path-relative harness files are byte-for-byte equal to the area versions. The independent FIFO cwd barrier is retained.

The first host run failed only because the old compiler-stage records still named mkdtempSync. The supplied update gate verified 11_setModifiedTime and 23_newLine natively, and updated those successful records. The 17 changed blocked stages were refreshed from exact observed diagnostics; no Node observation, provenance field, or fixture source changed. The final 25-record run passed. Compiles means native execution agrees with Node under sanitizer/leak checks; both compiling fixtures were separately executed through the JavaScript backend and agree with the recorded Node bytes too. It does not mean the remaining sources are executable.

The supplied check.py collector was executed without editing its file. Its CLI diagnostics were normalized by removing the adamic prefix, the trailing go-run exit-status line and the repository path prefix, matching fixtures_test.go's stage-record format. The initial attempt missed the repository prefix and is preserved as an observation-format failure; the corrected collector compares all Node/native observations and catches all 25 supplied source mutants plus its diagnostic mutant.

## Fixture table

| Fixture | Stage | First blocker |
| --- | --- | --- |
| 01_readFile_utf8.a | Checker | stage3/fixtures/host/01_readFile_utf8.a:25:13: error TS2322: Type 'number \| undefined' is not assignable to type 'number'. |
| 02_readFile_utf16le.a | Checker | stage3/fixtures/host/02_readFile_utf16le.a:25:13: error TS2322: Type 'number \| undefined' is not assignable to type 'number'. |
| 03_readFile_utf16be.a | Checker | stage3/fixtures/host/03_readFile_utf16be.a:25:13: error TS2322: Type 'number \| undefined' is not assignable to type 'number'. |
| 04_readFile_missing.a | Checker | stage3/fixtures/host/04_readFile_missing.a:25:13: error TS2322: Type 'number \| undefined' is not assignable to type 'number'. |
| 05_writeFile.a | Refused | stage3/fixtures/host/05_writeFile.a:13:9: Adamic 0.1 refuses a boolean \| undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| 06_fileExists.a | NotYet | stage3/fixtures/host/06_fileExists.a:48:5: stage 0 can't lower node:fs.symlinkSync yet |
| 07_directoryExists.a | NotYet | stage3/fixtures/host/07_directoryExists.a:48:5: stage 0 can't lower node:fs.symlinkSync yet |
| 08_getDirectories.a | NotYet | stage3/fixtures/host/08_getDirectories.a:208:5: stage 0 can't lower node:fs.symlinkSync yet |
| 09_realpath.a | NotYet | stage3/fixtures/host/09_realpath.a:27:5: stage 0 can't lower node:fs.symlinkSync yet |
| 10_getModifiedTime.a | NotYet | stage3/fixtures/host/10_getModifiedTime.a:33:5: stage 0 can't lower node:fs.symlinkSync yet |
| 11_setModifiedTime.a | Compiles | Node agreement on native and JavaScript |
| 12_deleteFile.a | NotYet | stage3/fixtures/host/12_deleteFile.a:21:5: stage 0 can't lower node:fs.symlinkSync yet |
| 13_createDirectory.a | Checker | stage3/fixtures/host/13_createDirectory.a:54:17: error TS18046: 'e' is of type 'unknown'. |
| 14_getCurrentDirectory.a | Refused | stage3/fixtures/host/14_getCurrentDirectory.a:15:24: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 15_getExecutingFilePath.a | NotYet | stage3/fixtures/host/15_getExecutingFilePath.a:21:21: stage 0 can't lower node:path.basename yet |
| 16_getEnvironmentVariable.a | NotYet | stage3/fixtures/host/16_getEnvironmentVariable.a:11:12: stage 0 can't lower a BinaryExpression with a string and a string yet |
| 17_write.a | NotYet | stage3/fixtures/host/17_write.a:16:5: stage 0 can't lower a literal method call with an unrepresented result yet |
| 18_exit_0.a | NotYet | stage3/fixtures/host/18_exit_0.a:12:5: stage 0 can't lower a value of type undefined yet |
| 19_exit_1.a | NotYet | stage3/fixtures/host/19_exit_1.a:12:5: stage 0 can't lower a value of type undefined yet |
| 20_exit_2.a | NotYet | stage3/fixtures/host/20_exit_2.a:12:5: stage 0 can't lower a value of type undefined yet |
| 21_createHash.a | Refused | stage3/fixtures/host/21_createHash.a:13:18: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 22_createHash_fallback.a | NotYet | stage3/fixtures/host/22_createHash_fallback.a:17:7: stage 0 can't lower a value of type undefined yet |
| 23_newLine.a | Compiles | Node agreement on native and JavaScript |
| 24_useCaseSensitiveFileNames.a | NotYet | stage3/fixtures/host/24_useCaseSensitiveFileNames.a:20:16: stage 0 can't lower statSync: fs options other than a fixed literal or its plain const binding yet |
| 25_readDirectory.a | Checker | stage3/fixtures/host/25_readDirectory.a:1001:23: error TS2345: Argument of type 'T \| undefined' is not assignable to parameter of type 'T'. |

## Verification

All test output is retained under logs/fs-area-merge/. Linux nproc is 5. Counts differ from area's table in process_observations and node_fs_directory_system; the regenerated values are checked by the whole oracle. There was no rebase, force push, or push to main.

```sh
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go test ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/regexp -count=1 -timeout 30m
go vet ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/regexp ./internal/oracle ./stage3/fixtures
go test ./internal/oracle -count=1 -timeout 30m
go test ./stage3/fixtures -run '^TestFixtures/host$' -count=1 -timeout 30m -parallel 4 -v
```

Package outcomes: load 3.686s, lower 75.802s, native 212.053s, flow 171.573s and regexp 19.690s all PASS. JavaScript has no package tests; its behavior is exercised by the oracle and the separate host checks. Vet exited 0. Whole oracle PASS 364.432s.
