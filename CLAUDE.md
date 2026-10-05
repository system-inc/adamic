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
- `internal/oracle`: the differential test. Every fixture runs three ways: its source on Node (the truth), native under ASan and UBSan, and the JavaScript backend on Node. stdout, stderr and the exit code must match byte for byte, and every program that finishes must leak nothing.
- `cohere/`: a submodule (cohere, which carries typescript-go). Over HTTPS: `git config submodule.cohere.url https://github.com/system-inc/cohere.git && git submodule update --init --recursive --depth 1`.

## The doctrine

- **Never a silent miscompile.** Wrong output that looks right is the worst thing a compiler can do. Every change is held by the oracle against Node.
- **Prove every check can fail.** A green run proves nothing until a mutant fails it. The mutant must be catchable only by the check you're proving (one killed by `-Werror` doesn't count), and it must change real input (string literals are immortal, so a use-after-free fixture needs strings the program builds). Say which mutant you ran and that it was caught.
- **Read Node's side first** when a fixture disagrees: the fixture can be wrong too.
- **Not yet and never are different promises.** Something stage 0 can't compile yet says `NotYet` with where and what; it never reaches clang as bad C, and it never compiles wrong.
- **No garbage collector**, not as a fallback, not for now.
- **Trust the artifact, not the story.** Report what you ran and what it printed, what you covered and what you didn't, and keep what you observed apart from what you infer.

## The gate

```
gofmt -l cmd internal
go vet ./...
go test -count=1 ./... > "$TMPDIR/test.log" 2>&1; echo "exit=$?"
```

Send test output to a log and read the log. Never pipe a test run into `head` or `tail`: it kills the run mid-way and can orphan the fixtures' processes. Tests are parallel by default; a test that can't be says why in a "Not parallel:" comment.

On Linux there's no `leaks` tool: LeakSanitizer (part of ASan there) does that job.

## Working here

- Write code that reads like the code around it: plain Go, comments that say why, names spelled out in full.
- Never overwrite a file wholesale; edit it. Never force-push, never rewrite history, never delete a branch you didn't make.
- Commit messages say what changed and why, in plain sentences. The first commit was `stuff`, after Ken; only the first one gets to be.
- No em-dashes in docs, comments or commit messages.
