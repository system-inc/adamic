# Admission delta

This developer gate compiles the **head corpus** with base and head compilers.
It runs every selected newly accepted program on source Node, the JavaScript
backend through `oracle/node.mjs`, and a native release build. Stdout bytes and
exit codes must agree. New refusals are informational. Compiler failures with
unrecognized diagnostics and crashes fail the gate. Timeouts and a missing native
toolchain make the run void (below).

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
destinations, and every changed `.ts` the manifest lists (most `.ts` files are
modules a program imports, so a `.ts` is a program only where the manifest says
so). A changed `.ts` it doesn't list, or every one in a `--corpus` run, goes to
`diff_ts_unmatched` instead, by path, so a new program in an unglobbed directory
is seen rather than silently skipped. Deleted programs are excluded. JSON always lists `diff` paths and
blobs and `diff_count`, including an empty list and zero for equal revisions,
and stderr says it as `changed programs: N`: programs added or changed, not programs that differ from Node.
The diff corpus is first; duplicate paths are classified once, while every
manifest blob is still verified. The name `diff` is reserved for this coverage.

`--compile-timeout 45s` bounds each base/head classification compile.
`--timeout 10s` bounds each backend build and execution. Both kill the process
group on timeout. The JSON records both limits and each command wall time.
`--budget <seconds>` limits selection, reserving five command timeouts per
program. Classification and compiler preparation are outside this runtime
budget. All newly admitted diff programs run first, followed by all witness programs,
even if their reservation exceeds the budget.
Remaining paths are sorted then shuffled using head's SHA-256 as the seed, or
`--seed <hex>` (at least 8 bytes) in its place, so one unit key always samples
the same programs whatever head is. Zero budget selects all programs.
`sampling_seed` (the seed used, as hex), `seed_given`, `sampling_size`, `omitted`,
`budget_seconds`, and actual `budget_used_seconds` describe the selection.

`--corpus-filter diff` runs the entire revision diff alone. A comma-separated
list selects other named corpora, always retaining the whole diff.
`--shard i/N` partitions non-diff inputs deterministically by SHA-256 of their
path, with 1-based shard indices. All shards retain the entire diff, so their
union covers the full generated corpora while every unit checks changed inputs.
JSON records `corpus_filter`, `shard`, and `complete_corpus`. Filtered and
multi-shard runs report `partial` when otherwise passing, preventing a
measurement unit from masquerading as a complete gate.

`verdict` is `pass`, `partial`, `sampled`, `void`, or `fail`. `sampled` exits zero but explicitly
means the entire admission delta has not been proven. A mismatch or an
unrecognized compiler error or crash exits 1. A run that proved nothing about the
change is `void`: a compile timeout, a runtime command's timeout (Node, either
backend's compile, or either backend's run) on a program that didn't agree, or no
clang where the native build resolves it (`internal/toolchain`'s `Path`), checked
before anything runs. It exits 3 with `"void": true` and `void_reason`, and each
program it names carries its own `void_reason`, so the judge reruns it instead of
charging the change. A red beside a void is still a red: it exits 1, records the
void's reason, and keeps `"void": false`.
Unrun observations and `agree` are null; `sampled` is false
on an unrun program. `admitted: 0` is explicit for equal admission sets.

`inputs.txt` names the repository paths whose change can move a candidate's
admission set, one per line, sorted: the lowering companion's Go import closure.
`admission-lower` calls cmd/adamic's `compile`, which loads and lowers and stops,
so the list is cmd/adamic, every in-repository package internal/load and
internal/lower link on Linux and macOS, each module file pinning one, and go.work.
The C emitter and the JavaScript backend move runtime outcomes, never the
admission set, and are not on it. `TestInputsAreWhatTheLoweringCompanionLinks`
fails when a linked package is under no entry or an entry is not needed.

Tests use tiny fake compiler executables to test command contracts independently
of the compiler. `TestProduct_AdmissionCompiler` builds the revision fixture's
lowering companion ahead under the product trace. `TestProduct_AdmissionCompilerReady`
waits for that owner in the tree's product order and only finds it. Ordinary
`TestBuildRevisions` finds the same product before entering its synthetic Git
repository and passes the binary to the command. Its base and head are the same
pinned revision, so there is one compiler product. A missing product fails the
test without building. The real compiler proof and its lowering mutant are recorded
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
literal repair. Supported literal forms are declared string initialized with
`undefined!` or `null!`, and the bounded speculative callback returning
`'wrong'` asserted as number. One more form, `"form": "checked_boundary"`, is a
type lie a checked boundary catches (#995z2fa, Adamic, Oct 25): a value a checked
cast let in, stopped at a checked view's own read or write. Its entry names the
ruling in `ruling` beside `ruling_task`, and its stderr must be the view check's
own located message, one line of `adamic: panic: field read failed: ...
expected <type>, found <type>`, whose expected type (a read) or found type (a
write) is `declared_type`. A `compiler bug:` panic or any other guard never
qualifies. The declaration's line must declare that type, and the repair must
replace exactly one place. An optional field's read expects its declared type
plus undefined, so `name?: Identifier` declares a stop that expected
`Identifier | undefined`; a required field declares exactly its type (#1yr7rdh,
Adamic, Oct 28). Its repair is arbitrary text rather than a pinned
literal, so a run that sees the stop also runs the repaired copy, written with the
witness's own extension so a `.ts` witness's repair compiles as `.ts`, and the entry
is a witness only when that copy agrees with Node (`repair_agree` in JSON).
Other forms require an independent evidence validator before a ruling can use
them. The repair's content hash cannot be listed. Every edit, including a type-correct repair, demands ordinary Node
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

## Run-time NotYet stops

A run-time NotYet (#zhgxzjg, Adamic, Nov 2) is a third thing, neither an
agreement nor a negative witness: an honest stop where a key or a receiver
can't be known before the read. It is never a type lie, so it is never listed in
`negative-witnesses.json`, and the tool refuses a program listed in both. Its
own list is `cloud/admission-corpus/runtime-notyet-stops.json`, or an explicit
`--runtime-notyet-stops` path. Only Adamic adds an entry: adding one is a ruling
routed to @system_adamic, and the `_comment` header says so. The map key is
SHA-256 of exact program bytes, so an edited program falls out of the list and
demands Node agreement. Each entry names its `path`, its `ruling_task`, the
`lifting_task` that will lift the stop, and its exact `stderr`. The tool refuses
the list when an entry lacks either task or its message is not one line of
`adamic: panic: <form> (reading '<key>')` with `<form>` on the tool's closed
allowlist (`runtimeNotYetForms`), or, keyless, exactly
`adamic: panic: converting <what> to a string is not yet lowered` with `<what>`
one of the three on the second allowlist (`runtimeNotYetConversions`, #09dvwny:
String and a template of an unknown), matched whole, never as a pattern. Each
form names the ruling that made it, and the tool refuses an entry whose
`ruling_task` is not its form's ruling. The first allowlist holds the nine
`internal/ir` DynamicKeyedStops forms a ruled fixture reaches; a `compiler bug:`
stop never qualifies, listed or allowlisted.

A listed program passes only when the tool sees all of it: both backends exit 70
with exactly the entry's stderr, their stdout is the same bytes and is byte for
byte the start of Node's, and Node runs the source to the end, exit 0, printing
more than that. Anything else is a lie or a crash, and fails as any disagreement
does. A listed program that now runs to the end on both backends and agrees with
Node is stale: it fails the gate, named in `stale_runtime_stops`, so its entry
leaves the list. Like a witness, a listed program, or an edit at a listed path,
runs even when both compilers already admit it, and under any budget, so the
list can't rot between changes.

An entry whose bytes no program in the run holds is judged only in a run that
holds every corpus (no `--corpus-filter`, no `--shard`), since a partial run
can't know what the rest held. If its `path` holds other bytes at head, the
listed program was edited in place: the entry is stale, named in
`stale_runtime_stops`, and fails the gate as an agreeing one does, even when the
edit agrees with Node. If nothing is at its `path`, the listed program never
arrived or has since been deleted: the run passes and names it in
`pending_runtime_stops`, so the list can land before its programs. That case
fails toward showing, listed in the JSON, never toward hiding. Counted, never
hidden: JSON records `ruled_runtime_stops`, `ruled_runtime_stop_paths`,
`runtime_notyet_count` (the list's size) and each program's `ruled_runtime_stop`,
with `agree` false, and every run prints `ruled run-time stops: N` to stderr.
