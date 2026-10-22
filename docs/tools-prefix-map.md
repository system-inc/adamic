# Evidence: compiler/tools-prefix-map (#8wgk65j), stopped for disk

Shas under test: `1090d40694b541bf6d5f6d9cd88f93f322661d70` and its follow-up `92b7d6db5eae6b1411f826652d768266b6cdd170`,
both on base `34dfaca13e`. Result: **stopped before any test ran.** No test passed or failed on either sha. The mutant was
not run, and no two-root build was made. The rest of this note says why and what would get it done.

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

`docs/tools-prefix-map/1090d406/{gofmt,vet,maps}.log` and `docs/tools-prefix-map/92b7d6db/gofmt.log`, force-added.
