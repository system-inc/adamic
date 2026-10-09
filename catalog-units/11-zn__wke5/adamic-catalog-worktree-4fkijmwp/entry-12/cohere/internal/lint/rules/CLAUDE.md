# Adding a rule to cohere

This file is auto-loaded into every agent working in this directory, so it stays short.
It carries only what you must know before reading anything else. **The full standard is
`internal/lint/rules/PortingARule.md`, and porting a rule without reading it is how this
repository ships its defects.**

## Read this before you touch anything

**Three families of the word `verify`, and only one of them changed.** This program used
to be called `verify` and is now `cohere`. Ordinary English ("verify that the fixture
fails") is not the program. And **Base's validation decorators are not the program
either**: `VerifyIsEmail`, `VerifyIsArray`, `VerifyBy` and about forty siblings, plus the
rules `base/correctness-require-verify-array-parity` and `base/correctness-require-verify-optional-parity`, are Kam's public API
in the Base framework and keep that spelling forever. `base/doc.go` explains why. Those
rules key on those literal strings, so a renaming sweep makes them match nothing, report
nothing, and leave the tree green having checked less than it appears to. A sweep already
nearly took them once.

**`--no-fix` is not optional, and `--fix` is not what causes writing.** The fix phase runs
by default, so a bare `cohere --timing` mutates the tree it is pointed at. Every dry run
against real code is `cohere --no-fix --lint`.

**`-r` is blocked at the harness. Delete file by file, named.**

**Never pipe a build or a test through `head` when you care about the exit code.** A pipe
reports the exit of the last stage, not the first, and `go build -o` leaves the previous
binary in place when compilation fails, so the stale one keeps answering `--rules` with a
plausible number.

**Do not edit another author's file, and do not open a package's shared `register.go`.**
Register with your own `<rule>_register.go`; there are over two hundred of them and that
is the convention. Other agents work in this tree concurrently: commit by pathspec
(`git add <your files> && git commit -F <msg> -- <your files>`), never `-A`, never `-a`.

**Put scratch outside `internal/lint/rules/`.** A probe package or a failing debug test
inside the rules tree turns the package red for every other agent, naming a file they have
never opened. `internal/<your_rule>_probe/` costs nobody anything.

**"Complete" is not a status you can report before the fixtures have run.** Say what
compiles and say what has executed, and keep those two claims separate.

## The one failure this whole standard exists to prevent

**Absence of a signal reported as a negative result.** A measurement that did not happen
looks exactly like a measurement that found nothing wrong — a `go test -run` pattern that
matched no test and printed `ok`, a grep that matched nothing and read as a clean tree, a
mutant that would not compile and read as a passing control. So: never accept a zero
without a control, and make your instruments refuse to print a result when they produced
no evidence.

## Then read the full document

    internal/lint/rules/PortingARule.md

It is the standard and the order, and every line in it was earned by a defect that
actually shipped. It covers, in the order you will need them:

- where the upstream trees and corpora live, and which artifact is the authority
- how to size a rule before starting, and when the deliverable is a measurement rather
  than a port
- reading the reference implementation, and what fidelity means here (the decision, not
  the workaround)
- taking fixtures verbatim from upstream's corpus, and the four ways a fixture lies
- the shelf at `internal/lint/ecmascript/` and `internal/lint/checking/`, and why to probe
  a helper against upstream rather than against its doc comment
- the mutation sweep (`tools/score_one_mutation.sh`), spans, message text, and fixers
- registering, enabling in `CohereSettings.json`, and the several ways a rule ends up
  registered and inert
- running dry against the real tree, and the six distinct kinds of zero
- the guards in `internal/lint/registry/` that catch what your own tests cannot

## A `.md` with no `.go` beside it is a rule nobody built

Every rule that was audited carries a document named for it, sitting where its code would
sit: `core/no_alert.md` beside `no_alert.go`, and `core/no_bitwise.md` beside nothing.
That second shape is the whole signal. The rule was read, measured against this tree, and
either declined or never reached, and the document says which.

So a namespace directory answers two questions at once. What is enforced is what has
`.go` files. What was considered and left is what does not, and the reason is in the file.

Building one of those means adding code beside a document already in the right place, and
re-measuring its violation count first: those counts were taken when the audit ran and the
tree has moved. A recommendation of Yes is a judgment from that day, not a commitment.
Current status lives in the rules spreadsheet, not here.

## Layout, so you can find things

    internal/lint/rules/<namespace>/    base core next nexus react structure tailwind typescript
    internal/lint/ecmascript/           the shared shelf, subdivided by question
    internal/lint/checking/             type-checker-facing helpers
    internal/lint/rule/                 rule.Rule, rule.Register
    internal/lint/testing/              Run, RunTyped, ExpectFindings, ExpectFixedSource
    internal/lint/registry/             the guards
    internal/lint/configuration/        config resolution
    internal/types/program/             the type graph
    internal/edit/                      the write layer
    TypeScript/ and TypeScript-shim/    the vendored compiler and its linkname shims

The gate is `go build ./... && go test -count=1 ./...`, and the config that decides whether
a rule ever runs is `/Users/kirkouimet/Projects/ahra/CohereSettings.json`.
