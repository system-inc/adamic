# Admission delta

This developer gate compiles the **head corpus** with base and head compilers.
It runs every selected newly accepted program on source Node, the JavaScript
backend through `oracle/node.mjs`, and a native release build. Stdout bytes and
exit codes must agree. New refusals are informational. Compiler failures with
unrecognized diagnostics, crashes, and timeouts fail the gate.

```sh
python3 cloud/admission-corpus/manifest.py --sha HEAD > /tmp/manifest.json
go run ./cmd/adamic-admission-delta --base origin/main --head HEAD \
  --manifest /tmp/manifest.json \
  --manifest-generator cloud/admission-corpus/manifest.py --json
```

Use `--base-binary /path/to/base-adamic --head-binary /path/to/head-adamic`
with the corresponding `--base` and `--head` source revisions to skip compiler
builds. Otherwise each compiler builds in its own detached checkout with its
pinned submodules. Corpus programs and imports always come from a separate
head checkout, including when supplied binaries are used. `--corpus <dir>`
recursively selects `.a` and `.ts` files inside head for ad hoc runs.

The manifest's `sha` must equal resolved head. Its named `corpora` each have a
`programs` list of `{ "path": "...", "blob": "git blob hash" }`. Empty corpora
are retained in JSON. Every entry is verified against head. The result records
the generator blob, the manifest's Git blob hash even for an external manifest,
and every program blob. The tool also computes a `diff` corpus from every `.a`
added or changed between base and head, in any directory, including rename
destinations. Deleted programs are excluded. JSON always lists `diff` paths and
blobs and `diff_count`, including an empty list and zero for equal revisions.
The diff corpus is first; duplicate paths are classified once, while every
manifest blob is still verified. The name `diff` is reserved for this coverage.

`--compile-timeout 45s` bounds each base/head classification compile.
`--timeout 10s` bounds each backend build and execution. Both kill the process
group on timeout. The JSON records both limits and each command wall time.
`--budget <seconds>` limits selection, reserving five command timeouts per
program. Classification and compiler preparation are outside this runtime
budget. All newly admitted diff programs run first, followed by all witness programs,
even if their reservation exceeds the budget.
Remaining paths are sorted then shuffled using a SHA-256-derived seed from head.
Zero budget selects all programs. `sampling_seed`, `sampling_size`, `omitted`,
`budget_seconds`, and actual `budget_used_seconds` describe the selection.

`verdict` is `pass`, `sampled`, or `fail`. `sampled` exits zero but explicitly
means the entire admission delta has not been proven. A mismatch or compiler
error exits nonzero. Unrun observations and `agree` are null; `sampled` is false
on an unrun program. `admitted: 0` is explicit for equal admission sets.

Tests use tiny fake compiler executables to test command contracts independently
of the compiler. The real compiler proof and its lowering mutant are recorded
under `review/compiler/admission-delta/`.
