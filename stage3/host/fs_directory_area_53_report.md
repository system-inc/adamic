Merged area/library 53f44e05 and retained symlinkSync, POSIX basename, const stat options, shared fs-file support and macOS fixes.
Commits: pieces e5dba453, c2878c78 and ed957219; merge receipt SHA is supplied in the handoff.
Checks: complete Linux counts 220.176s; lower 206.004s, native 510.436s, fresh 217.582s, load 13.400s, flow 367.874s; JavaScript has no package tests; whole uncached oracle PASS 1281.011s; vet and edited-source whitespace checks passed.
Mutants: all 25 audited host source mutants caught by Node comparison; all five owned source mutants caught only by stdout; six new runtime mutants and existing directory/path mutants are included in the whole oracle.
Uncovered: 22 unchanged host fixtures remain blocked; only 11, 12 and 15 agree on both backends; macOS and native tsc execution were not tested.

No rebase, force-push or main push. This is a merge of the requested area into codex/host-fs-directory-land. Audited NOTICE and status.json match the area exactly.

| Fixture | Native | JavaScript | First blocker |
|---|---|---|---|
| 01_readFile_utf8.a | Checker | Checker | stage3/fixtures/host/01_readFile_utf8.a:25:13: error TS2322: Type 'number \| undefined' is not assignable to type 'number'. |
| 02_readFile_utf16le.a | Checker | Checker | stage3/fixtures/host/02_readFile_utf16le.a:25:13: error TS2322: Type 'number \| undefined' is not assignable to type 'number'. |
| 03_readFile_utf16be.a | Checker | Checker | stage3/fixtures/host/03_readFile_utf16be.a:25:13: error TS2322: Type 'number \| undefined' is not assignable to type 'number'. |
| 04_readFile_missing.a | Checker | Checker | stage3/fixtures/host/04_readFile_missing.a:25:13: error TS2322: Type 'number \| undefined' is not assignable to type 'number'. |
| 05_writeFile.a | Refused | Refused | adamic: stage3/fixtures/host/05_writeFile.a:13:9: Adamic 0.1 refuses a boolean \| undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| 06_fileExists.a | NotYet | NotYet | adamic: stage3/fixtures/host/06_fileExists.a:29:9: stage 0 can't lower a PrefixUnaryExpression on a value yet |
| 07_directoryExists.a | NotYet | NotYet | adamic: stage3/fixtures/host/07_directoryExists.a:29:9: stage 0 can't lower a PrefixUnaryExpression on a value yet |
| 08_getDirectories.a | NotYet | NotYet | adamic: stage3/fixtures/host/08_getDirectories.a:143:44: stage 0 can't lower a parameter that isn't a plain name yet |
| 09_realpath.a | NotYet | NotYet | adamic: stage3/fixtures/host/09_realpath.a:12:21: stage 0 can't lower a PrefixUnaryExpression on a value yet |
| 10_getModifiedTime.a | NotYet | NotYet | adamic: stage3/fixtures/host/10_getModifiedTime.a:31:24: stage 0 can't lower a call through ?. (an optional call) yet |
| 11_setModifiedTime.a | Agrees | Agrees | Exact stdout, stderr and exit agree with recorded Node |
| 12_deleteFile.a | Agrees | Agrees | Exact stdout, stderr and exit agree with recorded Node |
| 13_createDirectory.a | Checker | Checker | stage3/fixtures/host/13_createDirectory.a:54:17: error TS18046: 'e' is of type 'unknown'. |
| 14_getCurrentDirectory.a | Refused | Refused | adamic: stage3/fixtures/host/14_getCurrentDirectory.a:15:24: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 15_getExecutingFilePath.a | Agrees | Agrees | Exact stdout, stderr and exit agree with recorded Node |
| 16_getEnvironmentVariable.a | NotYet | NotYet | adamic: stage3/fixtures/host/16_getEnvironmentVariable.a:11:12: stage 0 can't lower a BinaryExpression with a string and a string yet |
| 17_write.a | NotYet | NotYet | adamic: stage3/fixtures/host/17_write.a:16:5: stage 0 can't lower a literal method call with an unrepresented result yet |
| 18_exit_0.a | NotYet | NotYet | adamic: stage3/fixtures/host/18_exit_0.a:12:5: stage 0 can't lower a value of type undefined yet |
| 19_exit_1.a | NotYet | NotYet | adamic: stage3/fixtures/host/19_exit_1.a:12:5: stage 0 can't lower a value of type undefined yet |
| 20_exit_2.a | NotYet | NotYet | adamic: stage3/fixtures/host/20_exit_2.a:12:5: stage 0 can't lower a value of type undefined yet |
| 21_createHash.a | Refused | Refused | adamic: stage3/fixtures/host/21_createHash.a:13:18: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 22_createHash_fallback.a | NotYet | NotYet | adamic: stage3/fixtures/host/22_createHash_fallback.a:17:7: stage 0 can't lower a value of type undefined yet |
| 23_newLine.a | NotYet | NotYet | adamic: stage3/fixtures/host/23_newLine.a:10:31: stage 0 can't lower reading _os yet |
| 24_useCaseSensitiveFileNames.a | NotYet | NotYet | adamic: stage3/fixtures/host/24_useCaseSensitiveFileNames.a:32:9: stage 0 can't lower a PrefixUnaryExpression on a value yet |
| 25_readDirectory.a | Checker | Checker | stage3/fixtures/host/25_readDirectory.a:1001:23: error TS2345: Argument of type 'T \| undefined' is not assignable to parameter of type 'T'. |

Language reproducers confirmed with the merged compiler:

- `import { statSync } from 'node:fs'; console.log(String(!statSync('missing', {throwIfNoEntry:false})));` fails with a PrefixUnaryExpression refusal.
- `const f=(path:string):string=>path; console.log(String(!!f));` fails with the same refusal without Node imports.
- `function combine(path:string, ...paths:(string|undefined)[]):string {return path;} console.log(combine('x'));` fails on a parameter that is not a plain name.
- `function find<T>(a: readonly T[], p: (x:T)=>boolean): void { for(let i=0;i<a.length;i++) p(a[i]); }` reproduces fixture 25's TS2345: T | undefined is not assignable to T.

Commands (all output goes to log files):

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower ./internal/native ./internal/javascript ./internal/fresh ./internal/load ./internal/flow -count=1 -timeout=30m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=30m -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout=30m -parallel=1
python3 internal/oracle/node_fs_file_host_check.py --compiler /tmp/fs-directory-area-53-compiler --all --mutants --logs /tmp/fs-directory-area-53-host --report /tmp/fs-directory-area-53-host.json
python3 stage3/host/check_fs_directory.py --mutants --logs /tmp/fs-directory-area-53-owned
go vet ./internal/lower ./internal/native ./internal/fresh ./internal/load ./internal/flow ./internal/oracle
```

The all-host and owned-host runners exit 1 because acceptance compilation is incomplete; they do not claim blocked fixtures pass. All compilable fixtures agree with Node. No stock acceptance source was modified. The stock check.py mutant list was extracted with Python ast.literal_eval and each mutation run against its recorded Node observation. Fixture 13's intended omission of directory creation changes stderr and exit as well as stdout; it is counted under the stock observation-comparison rule, not the stronger stdout-only rule used for our runtime mutants. The earlier stricter source-mutant attempt is retained as a failed development check.

The Darwin feature macro remains in node_fs_directory.c, node_fs_file.c, node_process.c and node_path.c. Existing area harnesses retain terminal draining during child execution and canonical temp directories. New mutant harnesses and the owned host runner enable detect_leaks=1 only on Linux.

The final post-oracle rerun recorded all 25 fixtures and exactly reproduced the initial merge-run outcomes. All 25 Node observations match; 11, 12 and 15 agree on both compiled backends. The six new runtime mutants are symlink target/code/message, basename suffix/dirname, and stat throwIfNoEntry.
