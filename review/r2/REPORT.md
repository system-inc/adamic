# Stream R2: the afternoon's merges, reviewed

Reviewed `git log --first-parent 80c3098..2385966`: the constant-time string index (stream C, 4023b5b), writeTextFile and buffered stdout (C2, 380dc77), the benchmark harness (B, 6faf9b3), the Array.from fix (B2, 618298f), the gitignore port (P, 6fdd7c7), and Set with number | undefined in collections (C2, e869a74). Nothing on main was changed. Everything here is on this branch only, under `review/r2/`.

Environment: Linux x86_64 (4 cores, Docker on a Firecracker VM), Go 1.27.0, Ubuntu clang 18.1.3 (with `libclang-rt-18-dev` installed for ASan, which the container lacked), Node v24.21.0 at /opt/node24.

How to run a probe: `./review/r2/run3.sh review/r2/probes/<probe>.a`, from the repository root. It builds the probe under ASan and UBSan exactly as the oracle does (`review/r2/sanitized`, which calls `native.Build` with `Sanitize`) and runs it three ways: the source on Node (`oracle/node.mjs`), native with LeakSanitizer on, and the JavaScript backend on Node. With `PIPE=1`, stdout and stderr share one pipe. Proof that run3.sh's leak check can fail: a mutant retaining each element twice in `set.c`'s `adamic_set_add_all` made `set_duplicate_strings.a` exit 1 with "LeakSanitizer: detected memory leaks, 150 bytes in 3 allocations", while Node and the backend still agreed.

## Confirmed findings

### 1. Native reorders output around writeTextFile to /dev/stdout or /dev/stderr (C2, buffered stdout)

`adamic_write_text_file` (input.c) doesn't flush stdout's buffer before it opens and writes its file. So when the file is the program's own stdout or stderr, its text lands before lines printed earlier. Exit codes match, so only the byte order is wrong. The oracle can't see this because each run writes into a directory of its own.

`review/r2/probes/write_stdout_order.a`, `PIPE=1 ./review/r2/run3.sh ...`:

```
== node      first / second / third   exit=0
== native    second / first / third   exit=0
== backend   first / second / third   exit=0
```

`write_stderr_order.a` (writeTextFile('/dev/stderr') between two console.logs, both streams on one pipe) does the same: Node `first second third`, native `second first third`. If stdout is redirected to a file and the program writes that same file, both orders and contents differ, because the writer's O_TRUNC and the buffer's later write race.

Proposed fix (not applied, since this stream changes nothing on main): expose adamic.c's `flush` (for example as `adamic_output_flush`) and call it at the top of `adamic_write_text_file` and `adamic_read_text_file` (finding 2). docs/0.1.md's list of flush points ("before every write to stderr, before a panic's message, at exit, after every line on a terminal") then gains "before a file is read or written".

### 2. A prompt isn't shown before a read of stdin (C2, buffered stdout)

`readTextFile('/dev/stdin')` blocks without flushing stdout. With stdout a pipe, a program's prompt stays in the buffer while it waits for the answer to that prompt. A driver that waits for the prompt before answering deadlocks. Here it waits out its timeout instead.

`review/r2/probes/prompt_then_read.a` under `review/r2/interact.py`, which waits five seconds for the first line before answering:

```
== node     prompt: ready                            rest: got yes exit 0
== native   no prompt within 5s; answering anyway    rest: ready / got yes exit 0
== backend  prompt: ready                            rest: got yes exit 0
```

The fix is the same as for finding 1.

### 3. A signal throws away up to 64 KiB of printed output (C2, buffered stdout)

When the program is stopped from outside (timeout, Ctrl-C, a supervisor's SIGTERM), Node has already written every line, but native loses whatever is still in its buffer. `review/r2/probes/killed_after_output.a` prints one line, then spins. Each side was run under `timeout -s TERM 3 ... > file`:

```
== node     exit=124 stdout: 39 bytes: started, about to work for a long time
== native   exit=124 stdout: 0 bytes
== backend  exit=124 stdout: 39 bytes
```

docs/0.1.md says stdout is flushed "everywhere the difference could be seen", and this is one such place it isn't. A conservative fix: handlers for SIGTERM, SIGINT and SIGHUP that write out `output[0, output_used)` with `write` (async-signal-safe; `output_used` is advanced only after the memcpy, so only whole lines go out), then restore the default action and re-raise, so the exit status stays the signal's. A SIGKILL can't be helped, and the docs could say so.

### 4. sort with a comparator that returns NaN orders differently from Node (pre-existing on main; the C2 undefined-last sort inherits it)

Main's `adamic_array_sort` is a bottom-up merge sort, not V8's TimSort. docs/0.1.md says the runtime "uses TimSort the way V8 does", and bench/README.md (landed this afternoon in 6faf9b3) says "TimSort over unboxed doubles". For a consistent comparator, any stable sort gives the same answer. For an inconsistent one, and `(a, b) => a - b` over data holding a NaN is the commonest kind, the answer is wrong, exit 0, with nothing printed to say so.

`review/r2/probes/sort_nan_plain.a`:

```
== node     plain [5,NaN,0,1,3]   maybe, one undefined [5,NaN,0,1,3,]
== native   plain [0,1,3,5,NaN]   maybe, one undefined [0,1,3,5,NaN,]
== backend  plain [5,NaN,0,1,3]   maybe, one undefined [5,NaN,0,1,3,]
```

(`packed_crossings.a` found it first: its sort line differs the same way.) A fix exists and isn't merged: `origin/cloud/s2n8gc6` carries 8ae599e, "runtime: sort is V8's TimSort, step for step". I built that branch (c60cdad) and ran the plain half of the probe (it predates sort on number | undefined): Node, native and the backend all print `plain [5,NaN,0,1,3]`. Merging it, with this probe as a fixture, should close the finding. `sort_undefined.c` sorts its copy with `adamic_array_sort`, so it picks up the fix with no change.

### 5. The gitignore test can't see five wrong ports (P)

The test's mutant check is honest: its own five mutants were each caught, natively and on Node, in both runs below. But the questions it asks leave whole features unasked. Without `COHERE_GIT_SOURCE` (the default, and how the gate runs), the only character classes anything asks about are `[[:digit:]]`, the generator's `[[:alpha:]]` and an unknown class. Every query is already a clean path. No pattern line ends in a tab, no escaped range end appears, and no ignore file is near 100 MiB.

I added six mutants to the test's own `mutants` list (`review/r2/gitignore/mutants.go.txt`; `python3 review/r2/gitignore/apply.py` puts them in place, and `git checkout stage1/cohere/gitignore/gitignore_test.go` takes them out). Then I ran `go test -count=1 -v -run TestThePortAnswersAsGoCohereAndGitDo ./stage1/cohere/gitignore`, once without and once with `COHERE_GIT_SOURCE` pointing at a shallow clone of github.com/git/git (6,325 paths and 421 globs compared). The logs are beside the mutants.

| Mutant (each a port that no longer reads as the Go does) | Without git's corpora | With them |
|---|---|---|
| `[[:upper:]]` stops at Y (`<= 0x59`) | survives | survives |
| `[[:xdigit:]]` takes G (`<= 0x47`) | survives | survives |
| a trailing tab trimmed as a space would be | survives | survives |
| a path never cleaned (`cleanRelative` returns what it was given) | survives | survives |
| an escaped range end (`[a-\z]`) read as the backslash | survives | caught |
| the size limit `>=` where the Go has `>` | survives | survives |

"Survives" means the test reported "natively the mutant agrees with Go cohere" and "on Node the mutant agrees with Go cohere". The last row is marginal, since it differs only at exactly 100 MiB. The first four are real wrong ports that pass the gate. Each needs one more question: glob cases for every class name with a member at each end of its range (`[[:upper:]]` against `Z`, `[[:xdigit:]]` against `G` and `F`); a pattern line ending in a tab; a few unclean queries (`./a`, `a//b`, `a/../b`) put to `ignoredPath` and `Patterns.ignored`; and `[a-\z]` in the documented globs, so the default run catches it too.

## What turned out fine

Every probe below printed the same bytes and exit code all three ways, and native leaked nothing (LeakSanitizer on).

- **String index (C).** Reading string_index.c and every caller: strings are never written after `allocate`, `new_string` (input.c) and every stack piece start with `units = 0, index = NULL`, the index is freed only in `free_one`, and nothing frees a string any other way. `string_index_stale.a` builds a 480-unit string mixing ASCII, two- and three-byte characters, pairs, lone halves that join, and combining marks. It reads it forward, backward, scattered and in slices, and does the same to its upper-, lower-, NFC-, NFD- and NFKC-mapped forms, its trims, pads, repeats and splits, to slices rejoined at 80 split points, and to slices doubled so halves meet. It also runs two cursors alternating on one string. All three sides match. **Mutant:** checkpoints without their low-half bit (`(uint32_t)(offset << 1)`) made native's checksums differ from the first line on, while Node and the backend still agreed. Source restored.
- **Buffered stdout in its other paths (C2).** `output_then_overflow.a`: stdout, stderr, more stdout, then a stack overflow inside a sort comparator, on one pipe. All three print the same lines in the same order, with exit 70. `output_then_check.a`: 3,000 lines (more than the buffer), then an inserted check firing in a `map` callback. Native and the backend match byte for byte. Node differs only because JavaScript grows the array where Adamic panics, which docs/0.1.md expects. `stdout_full.a` with stdout on /dev/full: all three keep running, write stderr and the file, and exit 70.
- **number | undefined crossings (C2, B2).** `packed_crossings.a` covers includes and indexOf of undefined, NaN and -0 (including `[NaN].includes(undefined)` and `[undefined].includes(NaN)`, which a packed NaN could confuse), join, fill, Map values with undefined and NaN through for-of, entries and values(), reduce with a number | undefined accumulator, Array.from returning undefined, and a captured cell. Everything matched except the sort line (finding 4). `from_first_parameter.a`: a first parameter typed `string | undefined` or `number | undefined`, the latter pushed into an array afterwards; matched. A destructured first parameter is NotYet ("a parameter that isn't a plain name"), so the identifier-only refusal leaves no hole.
- **Set (C2).** `set_edges.a`: chained `add` on a new Set, `add`'s result aliasing (`===`), -0, 0 and NaN, strings built at run time against literals, delete-and-re-add during a live loop, a Set inside a Map, a `ReadonlySet` parameter iterated by `values()`. `set_duplicate_strings.a`: duplicate built strings through `new Set(array)` and `add`, with no leak. All matched.

## What I didn't cover

- macOS: everything here ran on Linux. The terminal path of buffered stdout (`isatty`, a flush after every line) wasn't run under a pty.
- Strings near the index's limits (an offset past `UINT32_MAX >> 1`, which falls back to walking from the start) and the 536,870,888-byte readTextFile edge: read, not run.
- The benchmark harness (bench/run.go): read only for the TimSort claim (finding 4).
- writeTextFile's own failure paths (permissions, umask, a path through a file), which `write_files.a` and `input_test.go` already hold to Node. I didn't re-probe them.
- Sorting a `number | undefined` array whose comparator shrinks or grows the array: `sort_undefined.c`'s write-back reads correctly, but I didn't run it.
- The gitignore port's performance and the GAPS.md workarounds.

## The gate

Run on this branch (main plus `review/` only) with `TMPDIR=/tmp/adamic-gate` (mode 1777) and Node 24 first on PATH. `gofmt -l cmd internal` printed nothing, `go vet ./...` printed nothing, and `go test -count=1 ./...` exited 0: load 1.9s, lower 3.2s, native 134.5s, oracle 268.3s, stage1/cohere/gitignore 87.7s. Note: the first baseline gate on unchanged main hit Go's default 10-minute test timeout in `internal/oracle` (still inside `TestCountsAreRecorded/fixtures` at 10m0s). A leftover probe of mine was spinning a core at the time, so that run says nothing against main. On a 4-core machine the oracle comes close to the default, and a `-timeout` in the gate line would make it robust.
