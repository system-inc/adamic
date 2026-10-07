The following is the original W5 report. The corrected witness, named assertions, and expanded V8 coverage are recorded in [follow-up.md](follow-up.md). The runtime fix belongs to the separate `codex/wasi-empty-path` branch.

Built: pinned wasmtime installer and opt-in source differential tests over ordinary and input fixtures.
Commits: claim 753ff69; implementation ef23383; fixture mappings f58d76b; main merge 2a5c93f; final pushed SHA accompanies this report.
Commands: post-merge native oracle PASS 181.322s; V8 WASI PASS 277.968s; wasmtime 329 fixtures, 314 pass, 15 fail, 0 skip, 240.886s.
Mutants: stdout, exit 0 to 23, and stderr changes all compile and are caught under wasmtime; an altered archive fails the checksum check.
Not covered: reactor ABI, request lifecycle, performance, sanitizers in Wasm, and fixes outside this unit's files.

## Toolchain and setup

Node v24.19.0, Go 1.27.1, native clang 20.1.8, WASI SDK 27, wasmtime
38.0.3 (`d9dc16b28`, release build dated 2025-10-24). `nproc` is 5,
cgroup quota is four CPUs, reported memory is 17.6 GB.

Release: https://github.com/bytecodealliance/wasmtime/releases/tag/v38.0.3

| GitHub archive | SHA256 |
| --- | --- |
| wasmtime-v38.0.3-x86_64-linux.tar.xz | 101d79dff495b0392d583d11c3c78dd50941c3ae28e80cf1604ca43acdf05af7 |
| wasmtime-v38.0.3-aarch64-linux.tar.xz | 10f8dd0f4789075321a439a1fb4a3d1888e3a45c0620dc9c562e198095120120 |

Both archives were downloaded from GitHub and hashed. Installation checks the
pinned digest before extraction. The x86_64 installation and execution were
verified; the aarch64 archive was hashed but not executed on this x86_64 machine.

`bash cloud/setup.sh --wasi-sdk --wasmtime`: Go ready 0s, clang 1s, Node 1s,
WASI SDK 1s, wasmtime 4s, submodules 4s, cache warm 200s, done 200s.
`bash cloud/setup.sh --wasi-sdk` retry: Go 0s, clang 1s, Node 1s, SDK 1s,
submodules 1s, cache warm 79s, done 79s. Every subsequent command sourced
`/workspace/adamic-tools/env.sh`, which supplies WASI_SYSROOT.

The first SDK-only setup was still reading the script when this worker edited
it. It printed Go 0s, clang/Node 1s, SDK/submodules 5s, then failed with
`cloud/setup.sh: line 143: syntax error near unexpected token '('`. The
unchanged retry above passed. This was a worker race, not a claimed installer
failure in the final script. The complete logs preserve it.

## Harness and directory views

Run with `ADAMIC_ORACLE_WASMTIME=1`. Opting in with missing tools fails loudly.
Each lowering fixture is compiled once through the actual `adamic build
--target wasm32-wasi` command. Node executes its source through oracle/node.mjs;
no inserted-check backend substitution or successful-result cache is used.
stdout, stderr and exit status all use the existing byte-exact disagreement
check. On disagreement, V8 executes the same artifact for diagnostic evidence.

The ordinary runner uses `--dir /::/`, exactly the root preopen in
oracle/wasi.mjs. Input fixtures run from their source directory on Node. WASI
libc starts at `/`, regardless of the host process cwd. `preopen-cwd.c` proves
that both stock V8 and wasmtime runners miss `reading/hello.txt` with root only.
A dot preopen lets wasmtime find it. A dot preopen also wins the root-prefix
lookup for some absolute paths, so the input harness adds exact scratch-directory
and `/dev` preopens as well. These preserve all absolute paths the input fixtures
use, alongside their relative assets. They do not claim general filesystem
identity for arbitrary absolute paths outside those aliases.

Input diagnostics use an isolated V8 runner with the same aliases, without
editing oracle/wasi.mjs. Node and Wasm get identical argument bytes. Writing
fixtures reuse the exact same output directory path, restoring it between hosts,
and compare resulting file names, contents and modes. The host uid is retained;
when root, both executions use uid/gid 65534 as the existing input oracle does.

## Disagreements, read from Node's side first

The fixture table records every pass/fail/skip by name. No failures are converted
to skips or expected successes. The strict witness cannot re-green against raw
source while the deliberate inserted checks and host limitations below remain.

| Fixture basename | Classification and observed behavior |
| --- | --- |
| writes_past_end.a | Intentional Adamic inserted check: Node grows the array and continues; both Wasm hosts panic at index 3. |
| cast_fails.a | Intentional checked cast: Node prints undefined; both Wasm hosts panic. |
| map_shrinks.a | Intentional inserted check: Node continues mapping; both Wasm hosts panic after the same prefix. |
| find_shrinks.a | Intentional inserted check: Node continues with undefined; both Wasm hosts panic. |
| find_index_shrinks.a | Intentional inserted check: Node continues with undefined; both Wasm hosts panic. |
| narrowed_numbers.a | Intentional inserted check: Node prints NaN after invalidation; both Wasm hosts panic. |
| write_after_shrink.a | Intentional inserted bounds check: Node grows the array; both Wasm hosts panic. |
| e4eec87_f1_field_narrowed.a | Intentional inserted narrowing check: Node prints undefined/NaN; both Wasm hosts panic. |
| e4eec87_f1_class_narrowed.a | Intentional inserted narrowing check: Node prints undefined; both Wasm hosts panic. |
| e4eec87_f1_alias_narrowed.a | Intentional inserted narrowing check: Node prints undefined/NaN; both Wasm hosts panic. |
| write_stdout_order.a | Host capability difference: Node source and V8 write through /dev/stdout; wasmtime reports permission denied. |
| write_stderr_order.a | Host capability difference: Node source and V8 write second to stderr; wasmtime cannot open /dev/stderr. |
| prompt_then_read.a | Host capability difference: Node source and V8 read EOF through /dev/stdin; wasmtime reports permission denied. |
| arguments.a | Host CLI difference: Node source and V8 decode malformed UTF-8 arguments; wasmtime rejects them before module execution, exit 1. |
| walk.a | Two causes: host filename encoding and a shared Adamic empty-path bug, detailed below. |

The ten checked fixtures are marked `checked` by the existing oracle. Its normal
contract compares these to the JavaScript backend, which carries the same checks.
The new requested raw-source comparison intentionally exposes that documented
semantic distinction. Neither Wasm engine hides a miscompile in these cases.

`stdio-path.c` is a minimal independent host witness: the same module opens
/dev/stdout and prints `through path` under V8, exit 0; wasmtime returns errno 63
(`Operation not permitted`), exit 1. This is a path-capability difference, not
an Adamic buffering failure. The stdio hosts are granted root, but wasmtime's
capability host still does not support these special paths as Node does.

`directory-bytes.c` is a minimal independent host witness. Given a directory
containing `bad` + byte FF + ` name`, V8 opens and enumerates it with errno 0;
wasmtime's opendir returns errno 25, exit 1. In walk.a this prevents enumeration
of the scratch directory. Node and V8 list it decoded with U+FFFD. Inferring the
WASI host's UTF-8 filename boundary explains the observed difference; the C
probe establishes that it is outside Adamic's generated code.

The other walk.a difference is a real shared Adamic WASI runtime bug:
`readDirectory('')` lists the cwd and `fileStatus('')` returns directory. Node
source returns empty-path errors. Both V8 and wasmtime show the bug with matched
preopens. See handoff.md and the exact source in handoff.md. No Adamic bug uniquely
hidden by V8 was found. The shared bug is still useful coverage: the prior WASI
oracle does not run the separate input fixtures.

## Additional engine probe

`engine-stack.c` deliberately exhausts the engine stack without Adamic's guard.
V8 reports RangeError, exit 1; wasmtime reports an engine stack-overflow trap,
exit 134. Both print start with this libc witness. These diagnostic/exit
conventions are engine differences, reproduced independently of Adamic. They
are not counted as an oracle fixture failure. The real stack_tail_call.a and
stack_forever.a fixtures pass on this integration branch, which now has its
runtime guard; the older limitations in docs/wasm.md do not describe these
current observations.

## Mutants and verification

The dedication control passes. Its generated C is then changed independently:
adding `!` to its first string produces stdout differs; changing main's final
return 0 to 23 produces exit codes differ with output preserved; adding a write
to stderr produces stderr differs. All three compile and are caught only after
execution under wasmtime. The first stderr attempt lacked stdio declarations
and clang rejected it; it was discarded, corrected with `<stdio.h>`, and rerun.

Commands (all test output redirected to complete logs, never piped):

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/wasm-engines-postmerge-oracle.log 2>&1
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestWASI' -count=1 -v -timeout 30m > /tmp/wasm-engines-postmerge-wasi.log 2>&1
ADAMIC_ORACLE_WASMTIME=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestWasmtime' -count=1 -v -timeout 30m > /tmp/wasm-engines-postmerge-wasmtime.log 2>&1
go vet ./internal/oracle > /tmp/wasm-engines-vet-latest.log 2>&1
bash -n cloud/setup.sh > /tmp/wasm-engines-shell-latest.log 2>&1
gofmt -l internal/oracle/wasi_engines_test.go > /tmp/wasm-engines-format-latest.log
git diff --check > /tmp/wasm-engines-diff-latest.log
```

After merging origin/main f8013f0: full uncached native oracle PASS,
181.322s. Existing V8 WASI gate PASS, 277.968s. Complete wasmtime run: 329 fixtures, 314 pass, 15 fail, 0 skip;
exit 1, 240.886s. All three execution mutants pass their rejection assertions.
The added 15 main fixtures all pass. A release archive with one appended byte
fails the pinned SHA256 check, exit 1, before extraction.

Before this merge, the native oracle passed in 198.193s and the V8 WASI gate
passed in 157.060s; the 314-fixture wasmtime run had 299 pass and 15 fail in
110.450s. These are retained as earlier observations, not post-merge results. Vet, format, shell syntax and diff
checks have no diagnostics. The fixture table and preserved final logs give
the exact complete WASI and wasmtime outcomes and timings.

Not run: the full repository gate and ARM64 execution. No reactor/request or
performance measurements were made. No runtime or compiler production file
outside the setup option was edited. No pull request was opened.

The checkout initially fetched only main; the requested integration ref was
fetched explicitly before creating codex/wasm-engines. The initial main merge
was already up to date. The final fetch observed origin/main advancing from
e8ba3d5 to f8013f0, and merged it in commit 2a5c93f. All three oracle scopes
were then rerun. There was no rebase. Only codex/wasm-engines is pushed. The branch is
reviewable, but strict source failures mean this is not a claimed green landing.

Rebuild the independent engine-stack witness with the SDK compiler:

```sh
"$(dirname "$(dirname "$WASI_SYSROOT")")/bin/clang" -O2 -fno-optimize-sibling-calls cloud/reports/wasm-engines/engine-stack.c -o /tmp/engine-stack.wasm
node --disable-warning=ExperimentalWarning oracle/wasi.mjs /tmp/engine-stack.wasm
wasmtime run --dir /::/ /tmp/engine-stack.wasm
```

The stdio and directory-byte probes use the same compiler, without the recursion
flags. Pass the directory as the sole program argument to directory-bytes.wasm.

The complete logs are preserved as .log.gz files beside this report.
Uncompressed test logs remain in /tmp under the command paths above.
