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
with the corresponding `--base` and `--head` source revisions to skip runtime
compiler builds. A supplied compiler gets a cached lowering companion
from its revision, or use `--base-lower-binary` / `--head-lower-binary` for
companions that implement `admission-lower <path>`. Otherwise each compiler
builds in its own detached checkout with its pinned submodules.
Compiler products use `internal/buildcache`, keyed by SHA, lowering adapter,
build settings and toolchain. `ADAMIC_BUILD_LOG` records hits and misses.
Classification calls only the revision's own loader and lowerer, never the C
emitter, in parallel (`--workers`, default available CPUs). Each classification
process uses one Go CPU; identical binaries reuse the same observation.
Only selected newly admitted programs build native executables. JSON
`phase_seconds` records checkout/provision, both builds, input verification,
classification, runtime checks and total wall times. Corpus programs and imports always come from a separate
head checkout, including when supplied binaries are used. `--corpus <dir>`
recursively selects `.a` and `.ts` files inside head for ad hoc runs.

The manifest's `sha` must equal resolved head. Its named `corpora` each have a
`programs` list of `{ "path": "...", "blob": "git blob hash" }`. Empty corpora
are retained in JSON. Every entry is verified against head. The result records
the generator blob, the manifest's Git blob hash even for an external manifest,
and every program blob. `--manifest-generator-revision <sha>` records an
explicit generator revision when the generator has not landed at head yet.
The default still verifies the generator at head. The tool also computes a `diff` corpus from every `.a`
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

`--corpus-filter diff` runs the entire revision diff alone. A comma-separated
list selects other named corpora, always retaining the whole diff.
`--shard i/N` partitions non-diff inputs deterministically by SHA-256 of their
path, with 1-based shard indices. All shards retain the entire diff, so their
union covers the full generated corpora while every unit checks changed inputs.
JSON records `corpus_filter`, `shard`, and `complete_corpus`. Filtered and
multi-shard runs report `partial` when otherwise passing, preventing a
measurement unit from masquerading as a complete gate.

`verdict` is `pass`, `partial`, `sampled`, or `fail`. `sampled` exits zero but explicitly
means the entire admission delta has not been proven. A mismatch or compiler
error exits nonzero. Unrun observations and `agree` are null; `sampled` is false
on an unrun program. `admitted: 0` is explicit for equal admission sets.

Tests use tiny fake compiler executables to test command contracts independently
of the compiler. The real compiler proof and its lowering mutant are recorded
under `review/compiler/admission-delta/`.

Ruled negative witnesses are read from
`cloud/admission-corpus/negative-witnesses.json` in the invoking repository,
or an explicit `--negative-witnesses` path. This allows a separately delivered
compiler branch to use this tool without merging its unlanded tool dependency.
The valid JSON header is the `_comment` field; adding an entry is a ruling
routed to @system_adamic. The map key is SHA-256 of exact program bytes,
never a path. The declaration and ruling task are documentary evidence. Imports and exports
are refused for listed sources so a dependency edit cannot inherit a file hash.
The tool validates the declaration's literal type contradiction and a pinned
literal repair. Supported forms are declared string initialized with
`undefined!` or `null!`, and the bounded speculative callback returning
`'wrong'` asserted as number. Other forms require an independent evidence
validator before a ruling can use them. The repair's content hash cannot be
listed. Every edit, including a type-correct repair, demands ordinary Node
agreement unless independently re-ruled. Valid programs must never be listed.

Ordinary comparison is exactly `outputsAgree`: stdout bytes and exit code,
with execution errors rejected; stderr is retained in the report but not
compared. An agreeing program remains an agreement even if its hash is listed.
A negative witness instead requires source Node exit 0 and both backends'
exact ruled exit 70 and entire stderr, plus equal backend stdout. `{root}`
in the ruled diagnostic expands only to the head checkout's absolute path.
Compilation failures never qualify. The list size prints to stderr on every
run that loads it, including empty lists; JSON records
`negative_witness_count`, `node_agreements`, and `accepted_witnesses`.
`agree` remains false on an accepted listed witness; `negative_witness` records
the separate exception. A wrong exit, diagnostic, or source repair fails.

Runtime coverage also includes listed bytes, pinned repair bytes, and every
admitted edit at a ruled declaration's path, even if accepted by both compilers.
These comparisons are mandatory under a budget. A path schedules coverage only;
it never grants an exemption. This prevents an already-admitted repair from
escaping Node agreement merely because it is no longer a new admission.
