# Evidence: compiler/tools-prefix-map (#8wgk65j)

Shas: head `92b7d6db5eae6b1411f826652d768266b6cdd170` (two commits on `34dfaca13e`: 1090d406, then 92b7d6db), base
`34dfaca13e`. The second run, below, covers 92b7d6db and the base. 1090d406 was skipped there, since its test is a subset
of 92b7d6db's. The first run, further down, stopped for disk before any test ran.

## Second run: results

- The three maps tests PASS on 92b7d6db, and the control is real: TestAnObjectNamesNotWhereTheCompilerIsInstalled logs
  `the unmapped object names /opt/llvm-a; the mapped one doesn't` (PASS, not SKIP).
- Both mutants go red, each through its own test.
- The two-root sanitized build: on the base the binaries differ and each names its own root twice. On 92b7d6db they
  are identical (same sha256) and name neither root.
- **Not run:** `go test ./internal/native/runtimeproducts` and the full `go test ./internal/native -timeout 40m`. The
  first failed on disk twice (below). The second needs more than that, so I didn't start it. 34dfaca13e's tests were not
  run either: the brief reruns only what fails on the head, and nothing in the maps tests did.

### How the second run fit

The first run's two full LLVM trees (`/opt/llvm-a`, `/opt/llvm-b-longer-name`, 11 GB each) and `/opt/dl` (2.1 GB) stay
where they were. Nothing under /opt was removed or moved. The second root is a hardlink copy,
`cp -al /opt/llvm-a /opt/llvm-c-longer-name`. It costs no file space (`/opt/llvm-a/bin/clang-20` and
`/opt/llvm-c-longer-name/bin/clang-20` are inode 1024484 with link count 2), and clang sees it as its own installation:
`/opt/llvm-c-longer-name/bin/clang -print-resource-dir` prints `/opt/llvm-c-longer-name/lib/clang/20`. llvm-b-longer-name
already existed, so the new root is llvm-c-longer-name, as the follow-up said. The two roots' names are 10 and 23
characters long.

df showed 428 MB free at the start. Freed in this order: `go clean -cache` (1.1 GB). Then `git worktree remove --force
/home/user/wt-base`, the base worktree I made in the first run (1 GB). From then on both shas used the one worktree
`/home/user/wt-head`, switched with `git checkout --detach <sha>` and `git submodule update --init --recursive --depth 1`.
Then `go clean -cache` once more before the runtimeproducts retry. The worktree was clean (`git status --short` empty)
after each mutant and at the end.

### Environment for every command

```
export PATH=/opt/llvm-a/bin:/opt/node24/bin:$PATH TMPDIR=/tmp/adamic-gate \
  WASI_SYSROOT=/opt/wasi-sdk/share/wasi-sysroot GOFLAGS="-buildvcs=false -trimpath"
```

ADAMIC_TOOLS is unset, so `toolchain.Path` takes clang from PATH (`which clang`: `/opt/llvm-a/bin/clang`). The two-root
build swaps only the first PATH entry.

### Step 1 on 92b7d6db

| Command | Exit | Log |
| --- | --- | --- |
| `gofmt -l internal/native` | 0, no files listed | `92b7d6db/gofmt.log` (empty) |
| `go vet ./internal/native` | 0 | `92b7d6db/vet.log` (empty) |
| `go test ./internal/native -count=1 -v -run '^(TestAnObjectNamesNotWhereTheCompilerIsInstalled\|TestToolMapsTakeTheSysrootOut\|TestASanitizedObjectNamesNoHeadersDirectory)$'` | 0 | `92b7d6db/maps.log` |
| `go test ./internal/native/runtimeproducts -count=1` | 1, out of disk | `92b7d6db/runtimeproducts.log` |
| the same, after `go clean -cache` (2.5 GB free) | 1, out of disk | `92b7d6db/runtimeproducts-retry.log` |
| `go test ./internal/native -count=1 -timeout 40m` | not run | |

maps.log:

```
--- PASS: TestToolMapsTakeTheSysrootOut (0.00s)
--- PASS: TestASanitizedObjectNamesNoHeadersDirectory (0.03s)
    tool_maps_test.go:56: the unmapped object names /opt/llvm-a; the mapped one doesn't
--- PASS: TestAnObjectNamesNotWhereTheCompilerIsInstalled (0.06s)
PASS
ok  	github.com/system-inc/adamic/internal/native	0.067s
```

Before that, the same command exited 1 twice with a log of `FAIL` alone, once with vet exiting 1 and an empty log too.
Both runs left under 450 MB free, so I take them to be the disk running out while Go built or linked. I didn't prove
that, but no test ran: the same test binary built with `go test -c` and run directly passed all three
(`92b7d6db/build-test.log`, empty, and `92b7d6db/maps-direct.log`). The 6.7 s against 0.03 s for the sanitized test is
that first direct run against the later ones; I didn't look into what it reuses between runs.

Both runtimeproducts runs failed in the same three tests, TestProduct_LeanCensusExports,
TestTheRuntimeBuildReachesOnlyItsDeclaredFiles and TestTheLeanCensusFollowsTheBuildAndFailsClosed, at
`build go list -export ./internal/native listing.json: write /root/.cache/adamic-build/local/.building-.../exports/NNN.a: no space left on device`
(the retry's first one at `compile: writing output: write $WORK/b281/_pkg_.a: no space left on device`). These are disk
failures, not test verdicts, so they weren't rerun on the base.

### Step 2: the mutants on 92b7d6db

Each was made in place in the worktree (no room for a scratch copy), recorded with `git diff`, run, then undone with
`git checkout -- <file>`.

**toolMaps returns nil** (`92b7d6db/mutant-toolmaps.diff`: `return nil` as toolMaps' first line in
internal/native/library.go). The maps command exited 1 (`92b7d6db/mutant-toolmaps.log`):

```
    tool_maps_test.go:71: maps [], want "-ffile-prefix-map=/tmp/adamic-gate/TestToolMapsTakeTheSysrootOut3170298484/001=/adamic-sysroot" among them
--- FAIL: TestToolMapsTakeTheSysrootOut (0.00s)
--- PASS: TestASanitizedObjectNamesNoHeadersDirectory (0.03s)
    tool_maps_test.go:54: the object names where clang is installed (/opt/llvm-a), so its bytes depend on where the tools sit
--- FAIL: TestAnObjectNamesNotWhereTheCompilerIsInstalled (0.05s)
```

TestAnObjectNamesNotWhereTheCompilerIsInstalled's failure line, verbatim:
`tool_maps_test.go:54: the object names where clang is installed (/opt/llvm-a), so its bytes depend on where the tools sit`.

**The strip flag dropped from Flags** (`92b7d6db/mutant-strip.diff`: `"-fsanitize-undefined-strip-path-components=-1"`
removed from the sanitized return in internal/native/native.go:152). The maps command exited 1 (`92b7d6db/mutant-strip.log`):

```
--- PASS: TestToolMapsTakeTheSysrootOut (0.00s)
    tool_maps_test.go:108: the sanitized object names the headers' directory /tmp/adamic-gate/TestASanitizedObjectNamesNoHeadersDirectory1512955177/002, so its bytes depend on where they sit
--- FAIL: TestASanitizedObjectNamesNoHeadersDirectory (0.03s)
    tool_maps_test.go:56: the unmapped object names /opt/llvm-a; the mapped one doesn't
--- PASS: TestAnObjectNamesNotWhereTheCompilerIsInstalled (0.05s)
```

Both mutants compile, so `-Werror` doesn't catch either of them; a test does. Each is caught only by the tests that hold
what it breaks.

### Step 3: two roots, base against head

For each sha: `go build -o /home/user/wt-head/adamic ./cmd/adamic` (exit 0), then the script below with
`<label>` 34dfaca1 and 92b7d6db. The fixture is `internal/oracle/testdata/arguments.a`. `--sanitize` gives
`-O1 -g -fsanitize=address,undefined`, and the 92b7d6db binary carries 10 `.debug*` sections (`readelf -S | grep -c debug`).
`adamic build` takes the file first: `adamic build --sanitize <file> -o <out>` printed the usage and exited 2, so the
first try was rerun as below.

```
fixture=$worktree/internal/oracle/testdata/arguments.a
for root in llvm-a llvm-c-longer-name; do
	cache=$out/cache-$root   # made fresh and empty for each build
	PATH=/opt/$root/bin:/opt/node24/bin:$PATH TMPDIR=/tmp/adamic-gate ADAMIC_BUILD_CACHE_DIR=$cache \
		"$worktree/adamic" build "$fixture" -o "$out/arguments-$root" --sanitize
done
cmp "$out/arguments-llvm-a" "$out/arguments-llvm-c-longer-name"
strings -a "$out/arguments-$root" | grep -c /opt/llvm   # for each root
```

| sha | builds | cmp | `/opt/llvm` in the llvm-a binary | `/opt/llvm` in the llvm-c-longer-name binary |
| --- | --- | --- | --- | --- |
| 34dfaca1 (base) | 0, 0 | exit 1: `differ: char 41, line 1` | 2 | 2 |
| 92b7d6db (head) | 0, 0 | exit 0, identical | 0 | 0 |

On the base, the strings are each root's `lib/clang/20/include` and `lib/clang/20/include/sanitizer`. The sha256s are
`427c8c03...` against `af5db8f0...`. On the head, both are
`9593d4dea3346ccb9e3ffe7c48d85e5ccec24fe5be8cefb8de17b69edf232546`, and `/adamic-compiler` appears twice in place of the
root. Each binary ran with exit 0. Logs: `34dfaca1/two-root.log`, `92b7d6db/two-root.log`.

What this covers: one small fixture, clang 20.1.8 on x86_64 Linux, native and sanitized, two install paths of different
lengths. Not covered: a wasm32 build (so the sysroot map was held only by its unit test), a non-sanitized build, and any
other build path than the one `adamic build` takes for this fixture by default.

## First run: stopped for disk

What follows is the first run's report as it stood, about 1090d406.

## Box facts

- Linux 6.18.44 x86_64, 4 cores, 15 GiB memory, running as root (a Claude Code cloud container).
- Go 1.27.1 (go.mod's toolchain, downloaded by GOTOOLCHAIN). GOFLAGS=`-buildvcs=false -trimpath`, as cloud/write-env.sh sets it.
- LLVM: the official `LLVM-20.1.8-Linux-X64.tar.xz` from the llvm-project llvmorg-20.1.8 release, sha256
  `1ead36b3dfcb774b57be530df42bec70ab2d239fbce9889447c7a29a4ddc1ae6`, extracted twice with `--strip-components 1` to
  `/opt/llvm-a` and `/opt/llvm-b-longer-name`. Both report `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`.
- Node v24.22.1 (the newest v24 in nodejs.org/dist/index.json) at `/opt/node24`, its tarball checked against SHASUMS256.txt.
- wasi-sdk 27 (`wasi-sdk-27.0-x86_64-linux.tar.gz`) at `/opt/wasi-sdk`, `WASI_SYSROOT=/opt/wasi-sdk/share/wasi-sysroot`. As
  cloud/setup.sh does, its `lib/clang/20/lib/wasm32-*` builtins were linked into each LLVM tree's resource directory
  (`lib/clang/20/lib`), the same links in both trees.
- `TMPDIR=/tmp/adamic-gate`, mode 1777.
- Worktrees: `git worktree add --detach` of 1090d406 and of 34dfaca13e, each with `git submodule update --init --recursive --depth 1`
  (both exit 0; cohere/TypeScript at d92d9bfee114).
- How adamic finds clang: `internal/toolchain.Path` takes `<ADAMIC_TOOLS>/bin/clang`, then `<ADAMIC_TOOLS>/go/bin/clang`,
  then (outside ADAMIC_SANDBOX=1) clang on PATH. ADAMIC_TOOLS was unset, so PATH decides; `which clang` printed
  `/opt/llvm-a/bin/clang` with `/opt/llvm-a/bin` first.

## What ran on 1090d406

With `PATH=/opt/llvm-a/bin:/opt/node24/bin:$PATH`:

| Command | Exit | Log |
| --- | --- | --- |
| `gofmt -l internal/native` | 0, no files listed | `1090d406/gofmt.log` (empty) |
| `go vet ./internal/native` | 1 | `1090d406/vet.log` |
| `go test ./internal/native -count=1 -v -run '^(TestAnObjectNamesNotWhereTheCompilerIsInstalled\|TestToolMapsTakeTheSysrootOut)$'` | 1 | `1090d406/maps.log` |

Neither failure is the change's. vet.log ends:

```
# github.com/microsoft/TypeScript/tsc/internal/contentmapper
compile: writing output: write $WORK/b256/_pkg_.a: no space left on device
```

and maps.log holds only `FAIL`: the test binary never built, for the same reason. So the base was not rerun: there is
no test failure to compare.

## What ran on 92b7d6db

Only `gofmt -l` over `git archive 92b7d6db internal/native`: exit 0, no files listed (`92b7d6db/gofmt.log`, empty). No
worktree, vet, test, mutant (including dropping `-fsanitize-undefined-strip-path-components=-1` from Flags for
TestASanitizedObjectNamesNoHeadersDirectory) or build ran on it.

## Why it stopped

The container's writable allowance is about 39.5 GB (df showed 10G used, 29G available at the start). The release
tarball is 2.0 GB and each extracted tree is 11 GB (8.1 GB of it bin/: flang, mlir, lldb, bolt, clangd and the like;
1.7 GB of static libraries in lib/*.a). Two trees, the tarballs, two worktrees and the Go caches left 433 MB, and the
first compile of typescript-go's packages ran out.

The way out was to delete what adamic never runs (it calls only `clang`, `clang++` and `llvm-ar`, by a grep of
cmd/ and internal/) from both trees identically, and the downloaded tarballs. The repository's
`.claude/hooks/guard-removals.mjs` refused both `rm` and `mv` of anything under /opt, since it allows removals only
inside the repository and /tmp. I didn't work around the guard, so the trees stayed as they were and the run stopped.

## To finish it

On a fresh box, or with the guard's permission to trim /opt, stream the tarball into tar once per root without keeping
it, and leave out what clang never needs, so each tree is well under the 11 GB a full one takes (not measured here):

```
for root in /opt/llvm-a /opt/llvm-b-longer-name; do
  mkdir -p "$root"
  curl -fsSL https://github.com/llvm/llvm-project/releases/download/llvmorg-20.1.8/LLVM-20.1.8-Linux-X64.tar.xz |
    tar --no-same-owner -xJ -C "$root" --strip-components 1 --no-wildcards-match-slash \
      --exclude='LLVM-20.1.8-Linux-X64/bin/flang*' --exclude='LLVM-20.1.8-Linux-X64/bin/mlir-*' \
      --exclude='LLVM-20.1.8-Linux-X64/bin/lldb*' --exclude='LLVM-20.1.8-Linux-X64/bin/llvm-bolt*' \
      --exclude='LLVM-20.1.8-Linux-X64/bin/clangd' --exclude='LLVM-20.1.8-Linux-X64/bin/bbc' \
      --exclude='LLVM-20.1.8-Linux-X64/lib/lib*.a'
done
```

The patterns are matched against the archive's own names, before --strip-components, and with
`--no-wildcards-match-slash` the lib/ one takes only the top-level static libraries, never the sanitizer runtimes under
lib/clang/20/lib. Check `lib/clang/20/lib/x86_64-unknown-linux-gnu/libclang_rt.asan.a` is there before building.
Then the brief's steps 1 to 3 as written, for 34dfaca13e, 1090d406 and 92b7d6db.

## Logs

All force-added under `docs/tools-prefix-map/`:

- First run, 1090d406: `1090d406/{gofmt,vet,maps}.log`. 92b7d6db's first-run gofmt is `92b7d6db/gofmt.log` (empty);
  the second run's gofmt on the worktree was empty too and wrote the same file.
- Second run, 92b7d6db: `vet.log`, `maps.log`, `build-test.log`, `maps-direct.log`, `runtimeproducts.log`,
  `runtimeproducts-retry.log`, `mutant-toolmaps.{diff,log}`, `mutant-strip.{diff,log}`, `two-root.log`.
- Second run, base: `34dfaca1/two-root.log`.
- `two-root.sh`, the script step 3 ran.
