Built runtime-kind dispatch for all seven predicates shared by StatsBase and Dirent.
Implementation commits: f096cedc and 0bc8dd3c; branch codex/host-stats-dirent-union, from f32f8f8.
Fs oracle with WASI PASS 137.308s; union oracle PASS 126.096s; Linux counts regenerated PASS 295.367s; final packages all PASS.
Static first-member dispatch and missing WASI refusal mutants caught by comparison with Node; existing method mutants retargeted and caught.
macOS and a positive block-device input were not run; WASI FIFO/socket ambiguity and the separate const-options refusal remain explicit.

The host lowerers recognize the declared Node methods and send both StatsBase
and Dirent receivers to the same runtime dispatcher. That dispatcher checks the
receiver's runtime shape identity before reading Stats' mode or Dirent's type.
The methods are isFile, isDirectory, isSymbolicLink, isBlockDevice,
isCharacterDevice, isFIFO and isSocket. Unsupported declarations and overloads
retain their existing named refusals. No declaration or compiler rule is changed.

The new node_fs_directory_union.a fixture assigns a Stats | Dirent | undefined
variable each kind in turn, then calls every predicate through union parameters.
It also declares the union in reverse order. Existing walking inputs include
files, directories, symlinks and a dangling link. A second harness run uses a
real FIFO, Unix socket and symlink to /dev/null, exercising positive special-kind
results on source Node, emitted JavaScript and sanitized native code.

Node 24.19.0's WASI host reports the FIFO as a socket. The runtime explicitly
refuses isFIFO/isSocket on that ambiguous WASI kind with:

```
adamic: panic: wasm32-wasi: fs.isFIFO/isSocket cannot distinguish FIFOs from sockets
```

Ordinary file/directory/symlink union cases agree on WASI. The special-file WASI
probe checks the exact refusal. Its mutant bypasses the new guards while retaining
runtime shape dispatch: it finishes successfully with no stderr and only Node's
stdout comparison catches the FIFO being misclassified as a socket. This target
limitation is observed on Linux; the runtime's Darwin feature macros are retained.

The requested dispatch mutant forces Dirent's layout for every union receiver.
It builds successfully, then exits 70 with the original missing-field compiler
panic when a Stats arrives. Source Node exits 0, so the behavioral comparison
catches it. Existing Stats isFile/isDirectory/isSymbolicLink and System fileExists
mutants now target the common dispatcher; their original semantic mutations and
requirement of exit 0, clean sanitizers and Node-only stdout disagreement remain.
All four retargeted mutants pass their catcher tests (3.360s).

Fixture 08 was run unchanged on a scratch of d18d2a413d8db84747ca85f9e35c105926170414.
The two unit commits were isolated onto that base and merged with a merge commit,
9beccfa68224d1876f3655e9ec278bea4ca9b879. The scratch retains the proof's compiler
and fixture intent. Its fixture blob remains c83a7d4fc45511802a8e53544755c35c67b4da99.
Source Node, the recorded status.json observation, emitted JavaScript through the
oracle loader and ASan/UBSan native with Linux leak detection agree exactly:

```
a,link,z
a,link,z

a,link,z
```

A merge of the entire task branch also imports the earlier optional-widening
compiler dependency, absent from d18d2a4. That full-branch scratch stops at
08_getDirectories.a:52:35, before runtime, on its const-asserted options. The
runtime acceptance above isolates this unit's delta; it does not remove the
refusal from the task branch or change the fixture. Verified one-line reproducer:

```a
import * as fs from 'node:fs'; const options={throwIfNoEntry:false} as const; fs.statSync('x',options);
```

The task compiler refuses it at 1:95:

```
Adamic 0.1 refuses optional property bigint in StatSyncOptions & { bigint?: false | undefined; throwIfNoEntry: false; } absent from structural source { readonly throwIfNoEntry: false; }, which can hide fields; declare bigint on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

Setup used GOPROXY='https://proxy.golang.org|direct' before cloud/setup.sh:
clang ready 0.575s; Go build ready 57.322s; test binaries deferred 57.646s;
build cache warm 57.648s; setup done 57.701s; nproc 5, CPU quota 4.
Node v24.19.0, Go 1.27.1, clang 20.1.8; Linux is the recorded count platform.

Commands used source /workspace/adamic-tools/env.sh and the configured
WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot:

```sh
go test ./internal/lower ./internal/ir ./internal/flow ./internal/native -count=1 -p=2 -parallel=2 -timeout=30m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=30m -args -update-counts
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run 'TestNodeFSDirectoryUnion|Test(InputAgreesWithNode|WASIInputAgreesWithNode)/internal/oracle/testdata/node_fs_directory_union' -count=1 -v -timeout=15m
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run 'TestNodeFS|Test(InputAgreesWithNode|WASIInputAgreesWithNode)/internal/oracle/testdata/node_fs_|TestWASIFileAgreesWithNode|TestCountsAreRecorded' -count=1 -parallel=2 -timeout=30m
```

Final package results: lower 172.998s; ir 36.150s; flow 450.878s;
native 426.677s. Counts regeneration 295.367s; union suite 126.096s.
The new Linux counts row is allocations/frees/retains/releases/peak/regions
119/119/35/116/28/0. Directory system's row advances because it enumerates the
new fixture. The complete fs oracle rerun passes (137.308s), including all 52 existing fs
mutants and both new mutants, Linux counts, and the fs WASI legs. There are no
failing tests in the final package or filtered oracle gates. The whole repository
and whole oracle were not rerun, using CLAUDE.md's touched-package/filtered-oracle
worker gate exception.

The first package run used a compiler test binary built before the fixture
expanded to all seven methods. Its three flow tests stopped on
node_fs_directory_union.a:6:91: stage 0 can't lower node:fs.Dirent.isBlockDevice yet:
TestEveryFunctionIsInSingleAssignment, TestEveryMutationIsInItsRange and
TestLivenessHoldsOnEveryPath. The frozen-source rerun passes all three.
The first complete fs oracle (287.513s) found four stale mutant substitutions:
TestNodeFSFileMutants/System_fileExists, Stats_isFile, Stats_isDirectory and
Stats_isSymbolicLink_follows, each reporting node_fs_file_test.go:208:
mutant changed nothing. All four were corrected; their focused rerun (3.360s) and final complete fs
oracle rerun (137.308s) pass.
No process fixture, language rule, main branch or remote history was modified.
