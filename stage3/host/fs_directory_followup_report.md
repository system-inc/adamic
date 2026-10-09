Verified the already implemented symlinkSync and exact const statSync options on the current integrated landing branch.
Implementation commits: e5dba453abd9fd6ffb81a81f133c4f0e383f18cc (symlink), ed9572199cbedd76b2ece670810455256024f268 (stat options); integrated in 1edda4ef463e6fe188ca40d9ba235c7d99e4221f.
Fresh commands: uncached symlink oracle PASS 3.463s; uncached stat-options oracle PASS 1.365s; all 25 host fixtures rerun on native and JavaScript, exactly reproducing the previous outcomes.
Mutants: wrong symlink target, error code and message plus forced throwing stat option caught only by Node stdout; all ran with exit zero and clean ASan, UBSan and Linux leak checks.
Remaining: 3/25 acceptance fixtures green; other callers remain Checker, NotYet or Refused. Fixture 25 retains its recorded strict diagnostic as an adaptation issue owned elsewhere. macOS was not run.

The symlink implementation was pushed before stat options. The fresh symlink check preceded the fresh stat-options check; an ordinary branch push confirmed 1edda4e was already up to date. This follow-up changes only the verification receipt and logs. No rebase, force-push or main push.

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
| 11_setModifiedTime.a | Agrees | Agrees | Exact stdout/stderr/exit agree with recorded Node |
| 12_deleteFile.a | Agrees | Agrees | Exact stdout/stderr/exit agree with recorded Node |
| 13_createDirectory.a | Checker | Checker | stage3/fixtures/host/13_createDirectory.a:54:17: error TS18046: 'e' is of type 'unknown'. |
| 14_getCurrentDirectory.a | Refused | Refused | adamic: stage3/fixtures/host/14_getCurrentDirectory.a:15:24: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 15_getExecutingFilePath.a | Agrees | Agrees | Exact stdout/stderr/exit agree with recorded Node |
| 16_getEnvironmentVariable.a | NotYet | NotYet | adamic: stage3/fixtures/host/16_getEnvironmentVariable.a:11:12: stage 0 can't lower a BinaryExpression with a string and a string yet |
| 17_write.a | NotYet | NotYet | adamic: stage3/fixtures/host/17_write.a:16:5: stage 0 can't lower a literal method call with an unrepresented result yet |
| 18_exit_0.a | NotYet | NotYet | adamic: stage3/fixtures/host/18_exit_0.a:12:5: stage 0 can't lower a value of type undefined yet |
| 19_exit_1.a | NotYet | NotYet | adamic: stage3/fixtures/host/19_exit_1.a:12:5: stage 0 can't lower a value of type undefined yet |
| 20_exit_2.a | NotYet | NotYet | adamic: stage3/fixtures/host/20_exit_2.a:12:5: stage 0 can't lower a value of type undefined yet |
| 21_createHash.a | Refused | Refused | adamic: stage3/fixtures/host/21_createHash.a:13:18: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 22_createHash_fallback.a | NotYet | NotYet | adamic: stage3/fixtures/host/22_createHash_fallback.a:17:7: stage 0 can't lower a value of type undefined yet |
| 23_newLine.a | NotYet | NotYet | adamic: stage3/fixtures/host/23_newLine.a:10:31: stage 0 can't lower reading _os yet |
| 24_useCaseSensitiveFileNames.a | NotYet | NotYet | adamic: stage3/fixtures/host/24_useCaseSensitiveFileNames.a:32:9: stage 0 can't lower a PrefixUnaryExpression on a value yet |
| 25_readDirectory.a | Checker | Checker | stage3/fixtures/host/25_readDirectory.a:1001:23: error TS2345: Argument of type 'T \| undefined' is not assignable to parameter of type 'T'. (recorded adaptation diagnostic) |

Both member fixtures run on both backends. The symlink fixture exercises omitted type and file/dir/junction/null/undefined constants, relative targets, broken links, catchable EEXIST/ENOENT/EACCES and remaining filesystem effects. Broader paths and dynamic type arguments remain named NotYet. The stat fixture uses the same const-asserted binding as fixture 24, reused fs-file stat semantics and existing error shape.

Commands (each stdout/stderr redirected to the retained logs):

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNodeFSDirectorySymlinkMutants|TestInputAgreesWithNode/internal/oracle/testdata/node_fs_directory_symlink.a' -count=1 -timeout=30m -parallel=3 -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNodeFSDirectoryStatConstOptionsMutant|TestInputAgreesWithNode/internal/oracle/testdata/node_fs_directory_stat_options.a' -count=1 -timeout=30m -parallel=3 -v
go build -o /tmp/fs-directory-followup/adamic ./cmd/adamic
python3 internal/oracle/node_fs_file_host_check.py --compiler /tmp/fs-directory-followup/adamic --all --mutants --logs /tmp/fs-directory-followup/host --report /tmp/fs-directory-followup/host.json
```

The host runner exits 1 for the 22 blocked fixtures; that is not reported as a green acceptance gate. Its ten fs source mutants are caught. All 25 Node runs match the recorded observations. No acceptance source, runtime, lowering, declaration or compiler option changed.
