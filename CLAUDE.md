# Adamic

> In dedication, with gratitude, to Kenneth Lane Thompson, whose work we stand on. - Kirk and Ahra

You're working on Adamic: TypeScript whose types are true, compiled to native code through C, on every core, with no garbage collector. Read `README.md`, `docs/0.1.md` (the language, approved by Kirk) and `docs/memory.md` (the memory model) before you change anything. The goals, in order: cohere compiled by Adamic, faster and leaner than Go cohere; TypeScript's own compiler compiled by Adamic, faster than typescript-go; Claude Code compiled by Adamic; and the code that runs Ahra's body.

## The layout

- `cmd/adamic`: `adamic types | c | build | js <file>`.
- `internal/load`: typescript-go in process. Adamic's compiler options are set in code, `.a` files are shown to the checker as TypeScript, the 0.1 library is embedded.
- `internal/ir`: the typed tree IR.
- `internal/lower`: the checker's types to IR. `NotYet` means stage 0 hasn't learned it; `Refused` means 0.1 forbids it, with the fix. `refusals.go` is the up-front refusal pass.
- `internal/native`: IR to C11 (`emit.go`) plus the runtime (`runtime/*.c`, embedded), compiled by clang with `-Wall -Wextra -Werror -pedantic`.
- `internal/javascript`: the second backend, IR to JavaScript, carrying the same inserted checks as native.
- `internal/flow`: a control-flow graph built from the IR, with single assignment lifted from cohere's high-level IR, for the analyses reuse in place and arenas need (#5jck546).
- `internal/oracle`: the differential test. Every fixture runs three ways: its source on Node (the truth), native under ASan and UBSan, and the JavaScript backend on Node. stdout, stderr and the exit code must match byte for byte, and every program that finishes must leak nothing.
- `cohere/`: a submodule (cohere, which carries typescript-go). Over HTTPS: `git config submodule.cohere.url https://github.com/system-inc/cohere.git && git submodule update --init --recursive --depth 1`.

## Compiler file ownership

Whole functions are kept together; trace extracted files with `git log --follow -C1% -- <file>`.

- Native emission: `emit.go` owns program assembly and emitter state; `emit_functions.go` signatures, calls and returns; `emit_statements.go` statements and loops; `emit_expressions.go` expression dispatch and operators; `emit_branches.go` short-circuit branches; `emit_ownership.go` retain/release and temporaries; `emit_values.go` C value representations; `emit_locals.go` globals, locals and capture cells; `emit_objects.go` shapes, fields and method thunks; `emit_arrays.go`, `emit_maps.go`, `emit_numbers.go` and `emit_strings.go` own their named areas. Existing `library_*.go`, `regexp.go`, `class_inheritance.go`, `reuse.go` and `region.go` retain their specialized work.
- Lowering: `lower.go` owns entry orchestration and state; `diagnostics.go` diagnostic types and descriptions; `modules.go` module order and registration; `functions.go` signatures, parameters and bodies; `statements.go` statement dispatch; `locals.go` bindings, captures and constants; `assignments.go` writes and compound updates; `control.go` conditions and loops; `prelude.go` console and panic recognition. Existing expression, class, collection and library files retain their areas.
- String runtime: `string.c` remains the sole translation unit, including private `string_{build,decode,trim,walk,builder,slice,repeat,search,replace,split}_impl.h` implementations in their original order. These headers own allocation/concatenation, decoding, trimming, unit walking, buffers, slicing, repetition/padding, searching, replacement and splitting respectively. Relative-index and affix entry points stay in `string.c`; existing string cache, sharing, append, case and normalization `.c` files retain their areas. Include implementation headers only from `string.c`, preserving static linkage.

## The doctrine

- **Never a silent miscompile.** Wrong output that looks right is the worst thing a compiler can do. Every change is held by the oracle against Node.
- **Prove every check can fail.** A green run proves nothing until a mutant fails it. The mutant must be catchable only by the check you're proving (one killed by `-Werror` doesn't count), and it must change real input (string literals are immortal, so a use-after-free fixture needs strings the program builds). Say which mutant you ran and that it was caught.
- **Read Node's side first** when a fixture disagrees: the fixture can be wrong too.
- **Not yet and never are different promises.** Something stage 0 can't compile yet says `NotYet` with where and what; it never reaches clang as bad C, and it never compiles wrong.
- **No garbage collector**, not as a fallback, not for now.
- **Trust the artifact, not the story.** Report what you ran and what it printed, what you covered and what you didn't, and keep what you observed apart from what you infer.

## The gate

The oracle pins Node v24.19.0 in `internal/nodepin`. `bash cloud/setup.sh` supplies it
from nodejs.org on Linux and macOS, checking published SHA-256 sums against
`cloud/node-pin.json`. Source the printed `env.sh` to put it first on PATH. A different
version fails the oracle before comparisons; test262 and the fuzzer check it too.

```
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > "$TMPDIR/test.log" 2>&1; echo "exit=$?"
```

The oracle caches fixture observations under the user cache directory. It still regenerates C and
JavaScript and rechecks ordinary comparisons and recorded counts. Timed stream and permission probes
reuse successful checks only with the same full harness and inputs. Changing a fixture, compiler output,
runtime library, Node version or test harness invalidates the affected observations.
`ADAMIC_GATE_UNCACHED=1` bypasses all oracle result caches. **Integration always runs uncached before
main moves.** Ordinary worker gates may use the cache. Keep `-timeout 30m` for the complete gate.

The input tests drop to uid 65534 when run as root, so TMPDIR must be world-traversable (for example /tmp/adamic-gate, mode 1777) and Node must not live under /root. A new fixture needs its row in internal/oracle/counts.md: `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`.

Send test output to a log and read the log. Never pipe a test run into `head` or `tail`: it kills the run mid-way and can orphan the fixtures' processes. Tests are parallel by default; a test that can't be says why in a "Not parallel:" comment.

On Linux there's no `leaks` tool: LeakSanitizer (part of ASan there) does that job. On macOS the leak check is the counted build, whose allocations must be its frees and its values in regions, then `leaks --atExit` on the same binary for malloc memory outside the counts; `leaks` alone can't see a leaked value, since the size-class allocator's chunks stay reachable.

## Adamic's own code passes Adamic's own gate

cohere checks and formats every Adamic program in this repository, `.ts` and `.a` alike. `tsconfig.json` carries the same compiler options stage 0 sets in `internal/load/load.go` (keep the two identical), the prelude as the one global declaration file, and `"sourceExtensions": [".a"]`, which cohere's TypeScript reads and stock tools ignore. `CohereSettings.json` turns on `cohere:typescript` (soundness, style and correctness) and ignores the files that are wrong on purpose, for these reasons:

- `internal/load/testdata/0.1/refuse/**`: docs/0.1.md's five refused programs, whose job is to be refused.
- `stage1/**/gaps/**`: each stage-1 slice's smallest programs for what stage 0 can't hold yet, kept exactly as written so their gaps tests notice when a gap closes.
- `review/**`: reviewers' probes, written to break things.

## Working here

- Write code that reads like the code around it: plain Go, comments that say why, names spelled out in full.
- Never overwrite a file wholesale; edit it. Never force-push, never rewrite history, never delete a branch you didn't make.
- Shell state doesn't persist between Bash calls, so a variable set in one call is empty in the next. Set any variable in the same command that uses it, and write removals as `rm -f "${S:?}/name"`: an unset variable then stops the command instead of collapsing the path to `/`. Delete only named files under the scratch directory or the repository, never by a path that begins with a bare variable. `.claude/hooks/guard-removals.mjs` refuses a command that breaks this and says why, so no one has to press Deny.
- Commit messages say what changed and why, in plain sentences. The first commit was `stuff`, after Ken; only the first one gets to be.
- No em-dashes in docs, comments or commit messages.
