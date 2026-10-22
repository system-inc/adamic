# internal/native whole on 92b7d6db, Linux (#8wgk65j)

Sha under test: `92b7d6db5eae6b1411f826652d768266b6cdd170` (branch `compiler/tools-prefix-map`, two commits on
`34dfaca13e`: 1090d406 and 92b7d6db). No code was changed. The logs are beside this note under
`docs/tools-prefix-map-full/`, unedited.

## Verdict

Red. The full package fails one test, `TestNonShippingFlagsStayIdentical`, and it passes on `34dfaca13e` on the same
box. The cause is 92b7d6db: `Flags` now appends `-fsanitize-undefined-strip-path-components=-1` to every sanitized
recipe (`internal/native/native.go:152`), and the test's pinned literals in `internal/native/release_build_test.go` still
end at `-fno-sanitize-recover=all`. Every failing row differs from its want by that one trailing flag and nothing else.
gofmt, vet, runtimeproducts and the three targeted tests are clean.

## The box

- Linux `vm 6.18.44-fc-v140`, x86_64, `nproc` 4, run as root.
- clang: the official `LLVM-20.1.8-Linux-X64.tar.xz` from llvm-project's releases (sha256
  `1ead36b3dfcb774b57be530df42bec70ab2d239fbce9889447c7a29a4ddc1ae6`), extracted once into `/opt/adamic-tools/llvm`,
  first on PATH through `/opt/adamic-tools/bin/clang`. `clang --version`:
  ```
  clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
  Target: x86_64-unknown-linux-gnu
  Thread model: posix
  InstalledDir: /opt/adamic-tools/llvm/bin
  ```
- Node: `v24.22.1` (the newest v24 from nodejs.org, fetched by `cloud/setup.sh`), `/opt/adamic-tools/bin/node`, next on
  PATH and outside /root.
- wasi-sdk 27 installed by `ADAMIC_TOOLS=/opt/adamic-tools bash cloud/setup.sh --wasi-sdk` (its clang reports
  `20.1.8-wasi-sdk`), `WASI_SYSROOT=/opt/adamic-tools/wasi-sdk/share/wasi-sysroot`. setup.sh's own wasm32 link probe
  passed.
- Go: `go1.27.1 linux/amd64`, as go.mod's toolchain line asks.
- `TMPDIR=/tmp/adamic-gate`, mode 1777.
- Checkout: `git worktree add --detach /tmp/adamic-gate/wt 92b7d6db...`, then
  `git submodule update --init --recursive --depth 1` (cohere at `06af2e39`).
- `/proc/loadavg` before each run: runtimeproducts `3.99 3.74 1.83`, full `2.62 3.95 2.32`, after full `3.95 4.03 3.19`.
  The box ran nothing else of mine during the runs.
- Disk: 13 GB free after setup, short of the 15 GB asked for. `go clean -cache` (2.2 GB of Go build cache that setup had
  warmed for all of `./...`) brought it to 15 GB before runtimeproducts. 12 GB were free at the start of the full run
  and the lowest sample during it (every 30 s) was 9.1 GB. Nothing hit the disk limit.

## What ran and what it printed

| Step | Exit | Time |
| --- | --- | --- |
| `gofmt -l internal/native` | 0, no output | |
| `go vet ./internal/native` | 0, no output | |
| `go test ./internal/native/runtimeproducts -count=1 -v -timeout 20m` | 0 | 172 s wall, `ok ... 153.052s` |
| `go test ./internal/native -count=1 -timeout 40m` | 1 | 679 s wall, `FAIL ... 674.139s` |
| targeted run of the three maps tests | 0 | 2 s wall, `ok ... 0.048s` |

runtimeproducts, its outcome line for the test asked about, verbatim:

```
--- PASS: TestTheRuntimeBuildReachesOnlyItsDeclaredFiles (153.02s)
```

All five tests in that package passed (`TestProduct_LeanCensusExports`, `TestTheLeanCensusFollowsTheBuildAndFailsClosed`,
`TestTheJobserverNeverReachesAnArchivesBytes`, `TestAnArchivesBuildIsJournaledAsRuntimeproducts` and the one above).

The full package's one failure, its first message line verbatim (the log has twelve such lines: sanitized, oracle
counts and slab probe, release false and true, each for `Flags` and `LinkFlags`; the counted and thread sanitized rows
and the ordinary-flags check pass):

```
--- FAIL: TestNonShippingFlagsStayIdentical (0.00s)
    release_build_test.go:35: sanitized release=false: got ["-std=c11" "-Wall" "-Wextra" "-Werror" "-Wcast-function-type-strict" "-pedantic" "-Wno-unused-variable" "-Wno-unused-but-set-variable" "-Wno-unused-function" "-Wno-unused-parameter" "-Wno-self-assign" "-ffp-contract=off" "-fno-optimize-sibling-calls" "-pthread" "-O1" "-g" "-fsanitize=address,undefined" "-fno-sanitize-recover=all" "-fsanitize-undefined-strip-path-components=-1"], want "-std=c11\x00-Wall\x00-Wextra\x00-Werror\x00-Wcast-function-type-strict\x00-pedantic\x00-Wno-unused-variable\x00-Wno-unused-but-set-variable\x00-Wno-unused-function\x00-Wno-unused-parameter\x00-Wno-self-assign\x00-ffp-contract=off\x00-fno-optimize-sibling-calls\x00-pthread\x00-O1\x00-g\x00-fsanitize=address,undefined\x00-fno-sanitize-recover=all"
```

No other test in the package failed. The run was not verbose, so `full.log` holds only the failure, the unit-cache lines
other tests print on purpose, and the package verdict.

The targeted run: `TestToolMapsTakeTheSysrootOut`, `TestASanitizedObjectNamesNoHeadersDirectory` and
`TestAnObjectNamesNotWhereTheCompilerIsInstalled` pass, the last printing
`the unmapped object names /opt/adamic-tools/llvm; the mapped one doesn't`.

## Against main

`go test ./internal/native -count=1 -v -run '^TestNonShippingFlagsStayIdentical$'` on `34dfaca13e2078244f80a45bd4bcfdf0fca4b922`
(a second detached worktree, same box, same tools): exit 0, `--- PASS: TestNonShippingFlagsStayIdentical (0.00s)`. There,
`native.go:151` ends the sanitized recipe at `-fno-sanitize-recover=all`, which is what the test pins. So the failure
is the branch's, not the base's. The log is `34dfaca1/failing-test.log`.

Inferred, not run: the test's three sanitized suffixes need ` -fsanitize-undefined-strip-path-components=-1` appended
for the branch to go green here. Whether that literal should move with the flag is the branch owner's call; the
test's comment says it pins order and bytes on purpose.

## A false start, for the record

My first attempt at the full run backgrounded `source env.sh && cd worktree && sampler` as one list, so `go test` ran
in the original checkout, which has no cohere, and died in under a second with
`go: cannot load module cohere/TypeScript/tsc listed in go.work file`. No test ran. It was rerun from the worktree at
the sha (the `full.log` here is that rerun, and its wrapper printed the worktree path and sha before starting).

## Files

- `docs/tools-prefix-map-full/92b7d6db/runtimeproducts.log`
- `docs/tools-prefix-map-full/92b7d6db/full.log`
- `docs/tools-prefix-map-full/92b7d6db/targeted.log`
- `docs/tools-prefix-map-full/34dfaca1/failing-test.log`
