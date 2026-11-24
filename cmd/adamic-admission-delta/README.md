# Admission delta

This developer gate compiles the **head corpus** with base and head compilers.
It runs every selected newly accepted program on source Node, the JavaScript
backend through `oracle/node.mjs`, and a native release build. Stdout bytes and
exit codes must agree. New refusals are informational. Compiler failures with
unrecognized diagnostics and crashes fail the gate. Timeouts and a missing native
toolchain make the run void (below). A run that hits a runtime timeout, a
program's or a witness repair's, is retried once alone before the timeout
counts, so load never reads as a red; every timed-out command is recorded, and
the summary line says how many runs timed out, were retried, and timed out
again (#pajbaba).

A crash the change fixes is judged by head alone (Compiler's rule on
#ppgk12h): a program the compiler crashed on at base, read from its own output
(a Go panic, or exit 4 with `adamic: compiler bug`), and doesn't at head is
newly accepted or newly refused by what head does, and a newly accepted one is
held to Node as every admission is. A signal or a runtime fatal error at base
may be the box's, so it still fails as a crash, and so does a crash at head.
Each fix is counted on its own line, `crashes fixed: N`, with one `crash fixed`
line per program and its class at head, and listed in `crash_fixed_paths`.

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
A program whose text holds `decodeJson` or `encodeJson` runs on Node with
`ADAMIC_JSON_TYPES` naming head's `oracle/json_types.go` built, as the oracle's
own tests run it: `oracle/node.mjs` otherwise runs `go run` on the tool in
head's checkout, which has no cohere, so the run failed before the program did.
The tool is built once, in a checkout laid as the compiler's is, cached by
head's SHA, and only when a program asks for it; `--json-types` names a built
one instead.
Classification calls only the revision's own loader and lowerer, never the C
emitter, in parallel (`--workers`, default available CPUs). Each classification
process uses one Go CPU; identical binaries reuse the same observation.
Only selected newly admitted programs build native executables. JSON
`phase_seconds` records checkout/provision, both builds, input verification,
classification, runtime checks and total wall times. Corpus programs and imports always come from a separate
head checkout, including when supplied binaries are used. `--corpus <dir>`
recursively selects `.a` and `.ts` files inside head for ad hoc runs.

The head checkout's `stage3/api` gets its packages by one of two roads
(#zvaxp1x). When the invoking repository already holds an install at
`stage3/api/node_modules` that its own `stage3/apply.py --packages-digest`
digests to the checkout's `stage3/api/node_modules.sha256`, the tool copies it
into the checkout: modes kept, symlinks kept as symlinks with their target text,
never a link to the invoking repository's copy, since the checkout's
`@types/node` is then corrected in place. The checkout's own `apply.py
--packages-digest` must then read the copy as that same pin, before any
correction, or the run fails. Otherwise the tool runs `npm ci --ignore-scripts`
in the checkout, with npm found as a build's tools are, under a 90 second limit.
A box with no npm fails, naming both roads: why npm wasn't found, and why the
laid install wasn't used (absent, a different `package-lock.json`, a checkout
that pins none, or `digest <actual> is not the pin <pinned>`).

No house box has npm, so the tool needs a worktree that already holds the pinned
`stage3/api/node_modules`, one `stage3/apply.py --packages-digest` reads as
`stage3/api/node_modules.sha256` before any correction. The old `ahra loom proof`
laid one in every proof (#th1nzka). A door job (`ahra loom job --kind proof`)
lays neither that install nor the submodules, so on a house box it fails at
provisioning until the door lays them (#08j5ptg for submodules; the install is
named to Jobs on #jobs, Nov 13). A copy of an install corrected in place reads as
a different digest and is refused, as it should be. Where the install is laid:

```sh
cd <proof job worktree>
python3 cloud/admission-corpus/manifest.py --sha HEAD > /tmp/manifest.json
go run ./cmd/adamic-admission-delta --base origin/main --head HEAD \
  --manifest /tmp/manifest.json \
  --manifest-generator cloud/admission-corpus/manifest.py --json
```

Base's and head's builds check out each revision's pinned cohere and its
TypeScript from the invoking repository's own submodule checkout when it is a
repository holding the pin, else from the box's mirrors (`<directory>/cohere.git`
and `TypeScript.git` under `~/scratch/system_loom/modules` and `~/loom-mirrors`,
the lay kit's own, or `ADAMIC_ADMISSION_MIRRORS`'s colon-separated list), and a
pin none of them holds is named with every place looked in (#t7b148j). The copied
install is made owner-writable before `@types/node` is corrected, since the door
lays it as read-only hard links; the digest reads only the executable bit.

Only head's checkout is provisioned, against head's own pin. When head pins a
different install from the worktree the tool runs from, head falls back to npm,
and on a box with none it fails loud.

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
`admitted` counts every corpus, not the diff alone, so the run names each program it counts: JSON lists them in
`admitted_paths`, and stderr prints `admitted: N`, then a line each, `admitted <path> (<corpus>): <judgment>`,
where the judgment is `agrees with Node`, `disagrees with Node`, a ruled negative witness, run-time stop or
backend divergence, `void, <why>`, or `not run on Node, omitted by the budget` (#1zscjp1).

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
Adamic, Oct 28). An element read through a checked view's array field is the
same hatch (#6wfvfhm, Adamic, Nov 6): its stop is one line of `adamic: panic:
element read failed: <expression> is not a <T>; expected <T>, found <kind>`, the
pattern grown by exactly that alternation, and its expected type is the
array's element, so the declaration's line must declare an array of exactly
that element, as `readonly T[]`, `T[]` or `ReadonlyArray<T>`, or a union as
`readonly (T)[]` or `ReadonlyArray<T>`. Its repair is arbitrary text rather than a pinned
literal, so a run that sees the stop also runs the repaired copy, written with the
witness's own extension so a `.ts` witness's repair compiles as `.ts`, and the entry
is a witness only when that copy agrees with Node (`repair_agree` in JSON).
A third, `"form": "cast_failed"` (#b5w3ycg, Adamic, Nov 16, applying rulings.md:9
and :385), is a `.ts` program that lies about a declared type and stops at the tag
test a cast inserts: stderr is one line of `adamic: panic: cast failed: this <S> is
not a <T>`, and `declared_type` is `<T>`, the type the cast claimed. That message
names no position, so the entry names the cast's line in `declaration`, a `.ts`
path, and in `column` where the cast's `as` keyword starts, 1-based. The tool holds
the source to it: the word `as` at that column, after whitespace or a closing
bracket, before whitespace, and outside any string. The stop could as well come from
another cast, and it names the concrete target, not the text written (an alias, a
generic `T`), so a witness holds one cast: a source holding the word `as` anywhere
else, strings and comments included, is refused, and only `as const`, which checks
nothing, is passed over (#s1gqkzm). The repair replaces exactly one
place and keeps the cast at the same line and column, so the copy that agrees with
Node shows the check doesn't fire on a true claim. Node erases the cast and goes
on, so both backends' stdout must equal each other and be what Node printed up to
the stop. A `.a` program holding a listed entry's bytes is refused, since in `.a`
these casts are refused, and only this form names a `column`.
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

## Darwin programs

The darwin class (#an8zcp1, Language and Adamic) is unjudged here. Apple's
framework programs belong to the darwin witnesses on a Mac, rather than Node.
The class needs two independent keys: the delta's own TypeScript parse finds a
module specifier beginning `apple/`, and head's lowering compile or native build
exits `platform.MissingExitCode` (5). The compiler reserves that exit
for the typed `platform.Missing` error; diagnostic words grant nothing.
Nothing on main raises that error yet; Apple's slice supplies its platform stops.

The parser reads Git blobs at the revision being judged, including every local
module reached through relative imports and re-exports. It recognizes ordinary,
type-only and side-effect imports, exports from a module, and literal dynamic
imports, including those inside functions. Comments, unrelated strings and a
computed module name give no key. Base's modules come from base, head's from
head, never the invoking working tree. Local module cycles terminate. Syntax diagnostics do not add a third key: the
parser still reads import nodes it recovers, and the darwin witnesses judge the
whole platform row.

A row with both keys is `darwin`, wholly void for that row: its observations
remain exactly as they happened, including Node, JavaScript and native results
when the stop is in the native build. `agree` is null; none of those outcomes is
judged, and the row is neither passed nor failed. A compile stop has no runtime
observations. Darwin rows do not make the run void or fail it, and are excluded
from `admitted` and Node-agreement counts. An old ruled stop at that unjudged path is not judged stale either. The other
rows compute the verdict.

There are two guards against hiding another program in this class. An admitted program with the parsed key always gets a native probe, even when
both compilers already admit it and even under a sampling budget. Head's exit
5 without its parsed import key fails as `platform-stop-without-apple`. A row
moves into darwin only from a judged base: its base file already reaches
`apple/`, and its base compile or matching base native build did not
platform-stop. When head stops natively, only a row whose base file already
reaches `apple/` gets one extra base native build, recorded as
`base_native_compile`; if that build also platform-stops, the row stays
`darwin` instead of failing as `moved-into-darwin`. A new file, or an existing file
whose import key changes from false at base to true at head, may enter. A
program leaving the platform stop is judged normally.

Every report includes `darwin_count` and `darwin_paths`, including zero and an
empty list. Every darwin row stays in `programs`, with class `darwin` and
`void_reason: "darwin: unjudged here; judged by the darwin witnesses"`. Raw
compile and runtime observations stay on that row. Each run prints:

```text
darwin: N (unjudged here; judged by the darwin witnesses)
```

This records the missing judgment without letting a missing SDK masquerade as
an unrecognized program diagnostic, or an unavailable native tool as agreement.
The darwin witnesses supply that judgment; this gate never invents it.

## Run-time NotYet stops

A run-time NotYet (#zhgxzjg, Adamic, Nov 2) is a third thing, neither an
agreement nor a negative witness: an honest stop where a key or a receiver
can't be known before the read. It is never a type lie, so it is never listed in
`negative-witnesses.json`, and the tool refuses a program listed in both. Its
own list is `cloud/admission-corpus/runtime-notyet-stops.json`, or an explicit
`--runtime-notyet-stops` path. Adamic rules each new form. A head may add a
program entry under an already ruled form (docs/rulings.md:412). The map key is
SHA-256 of exact program bytes, so an edited program falls out of the list and
demands Node agreement. Each entry names its `path`, its `ruling_task`, the
`lifting_task` that will lift the stop, and its exact `stderr`. The tool refuses
the list when an entry lacks either task or its message is not one line of
`adamic: panic: <form> (reading '<key>')` with `<form>` on the tool's closed
allowlist (`runtimeNotYetForms`), or, keyless, exactly
`adamic: panic: converting <what> to a string is not yet lowered` with `<what>`
one of the three on the second allowlist (`runtimeNotYetConversions`, #09dvwny:
String and a template of an unknown), matched whole, never as a pattern, or
exactly `adamic: panic: <message>` with `<message>` on the third
allowlist (`runtimeNotYetKeyless`, #ey7aysy under #0sps03r: Object.keys of a
runtime host value; #995spev: clearTimeout of an object that isn't a timer but
holds a truthy `_onTimeout`; #hh1cezv: `an Error's own property isn't represented yet`),
matched whole too. Adamic ruled the Error form on Nov 23 under docs/rulings.md:411.
A key no one sets on the Error, where Node answers `false`, stops too: an over-stop
that fails closed, counted by #nw1kfem. Each form names the ruling that made it, and the tool
refuses an entry whose `ruling_task` is not its form's ruling. The first
allowlist holds the ten `internal/ir` DynamicKeyedStops forms a ruled fixture
reaches; a `compiler bug:` stop never qualifies, listed or allowlisted.

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

## Ruled backend divergences

A ruled backend divergence is a fourth thing: a fixture whose two backends agree
with each other and are ruled to differ from the source on Node, such as
installed-layout and no-javascript-plugins (#redxes2), where Node, running the
source as an ES module, has no `require`. The tool keeps no list of its own. It
reads the oracle's rows, `internal/oracle/testdata/backend_divergences.json` in
the invoking repository, or an explicit `--backend-divergences` path: the same
file `internal/oracle`'s `ruledBackendDivergences` reads and holds each
fixture's runs to (#tqm95v6). Only Adamic adds a row: adding one is a ruling
routed to @system_adamic, and the file's `_comment` header says so. Each row
names its `fixture` (a repository path, once), its `ruling`, which must name its
task, its `reason`, and its `node`, `native` and `javascript` outcomes, each the
exact `stdout`, `stderr` and `exit_code`. The tool refuses the list when a row
lacks any of the first three, its ruling names no task, or a fixture is listed
twice.

A newly admitted program that disagrees with Node passes as a ruled divergence
only when the program's path is a row's fixture, on a row that sets `node`;
Node's run is exactly that row's `node` outcome, and native's exactly its
`native` outcome, stdout, stderr and exit code; and either native and
JavaScript agree with each other, as `outputsAgree` reads agreement, or
JavaScript's run is exactly the row's `javascript` outcome too. The second form
is for a row a test holds apart on purpose, a Weak read native frees where
Node's collector keeps the target (#d358hc1), so all three of its runs are
pinned whole. An outcome's `{root}` is the checkout the run judged, for a stop
that names its program's file. A program
on no row fails as any disagreement does, even when it prints a row's outputs,
and so does a row's program whose run differs from the row on any side. The
path, not the content hash, is the key, since the row pins every output an edit
could move. Counted, never hidden: JSON records `ruled_backend_divergences`,
`ruled_backend_divergence_paths`, `backend_divergence_count` (the list's size)
and each program's `ruled_backend_divergence`, with `agree` false, and every run
prints `ruled backend divergences: N` to stderr after the run-time stops line.

## Agreement mode

The delta runs only what head newly accepts, so a wrong answer already on main,
a native segfault where Node prints a value, is invisible to every head's gate.
`--agreement` needs no base and runs every program head accepts:

```sh
go run ./cmd/adamic-admission-delta --agreement --head HEAD \
  --manifest /tmp/manifest.json --open-tasks /tmp/open-tasks --shard 1/4 --json
```

It takes `--head`, `--manifest` and `--open-tasks`, all three required, refuses `--base`, and
refuses every flag it wouldn't read (`--base-binary`, `--corpus`,
`--corpus-filter`, `--budget`, `--seed` and the generator's two). Head's
compiler is built or supplied as the delta's is. Each program of the manifest is
classified by head's lowering alone: one head refuses is counted as refused and
never run, and one head's compiler crashes on, fails on with an unrecognized
diagnostic or times out on is named on a `compiler failure` line and in
`compiler_failures`. Every accepted program runs on Node, the JavaScript backend
and a native build under the address and undefined-behavior sanitizers
(`build <file> -o <out> --sanitize`), through the same runner as the delta's
newly accepted programs, with one retry after a runtime timeout, run beside the other workers, never alone as the delta's, and
the native build under `--compile-timeout`, since it's a compile. Before any of
them, one program of a single line builds natively with up to 15 minutes, so a
cold build cache makes the runtime's sanitized archives once; if it fails, the
run is void, naming what the build said. Then each accepted program runs and
is judged by the same rules: it agrees, or it is a listed negative witness, a
listed run-time stop or a ruled backend divergence, each exactly as ruled, or it
disagrees. Two more judgments then count apart what would otherwise disagree:

- `checked-stop`: a both-stop, where the JavaScript backend and native have the
  same exit, stdout and stderr, the exit is 70, and Node exits 0, at a path
  `cloud/admission-corpus/held-stops.json` lists, with the backends' stderr the
  row's byte for byte. Each row is a ruling and the pin of the stop's words:
  `{"path", "stderr", "hatch", "test"}`. `stderr` writes the checkout the run
  judges as `{root}`, since a stop naming its program's file names it there.
  `hatch` is the hatch the program uses, `<kind>[ (<detail>)], line <n>:
  <text>`, the kind one of cast, non-null, predicate, definite-assignment,
  relation-lie, index-write, index-read, computed-key, library-claim and
  language, the text that line of the program, trimmed. A `language` hatch is the
  checker's own unsoundness a ruling makes a checked stop, so its detail cites
  `docs/rulings.md:<line>, "<the row's opening words>"`. `test` is the test that
  holds the stop, `<repository path>_test.go: <TestName>`. A row without its
  hatch or its test is refused, and `TestHeldStopListHoldsItsRows` holds every
  row's hatch line, ruling and test to the tree. A both-stop with no row, or with
  a stderr that differs from its row's, disagrees. `--held-stops` names another
  list; a missing file is an empty one.
- `pending`: a program beside a pending sidecar, its name without its extension
  plus `.pending`, whose message is `awaits #<task>: <reason>`. The sidecar
  excuses the program only while that task is open, and the run is sealed and
  can't ask, so the entry carries the task in `task`, a `pending <path>: awaits
  #<task>` line names it on stderr. The run can't ask the task store, so
  `--open-tasks` names the tasks open when it was asked, one id a line without
  its `#`, an empty file holding none open and a line of any other form refused.
  A sidecar naming a task the list doesn't hold excuses nothing: its program is
  judged as if it had none, a `closed sidecar <path>: names #<task>, which isn't
  open, so it excuses nothing` line names it, and `closed_sidecars` lists it.
  That is how two live wrong answers hid for three weeks (#92cnfpf). A pending
  program that times out is pending, never void. A sidecar naming a
  branch, `awaits <branch>: <reason>`, names no task, so its program disagrees.
  An agreeing program with a sidecar is agreeing.

Accepted programs run `--workers` at once (default available CPUs),
and the lines and the report follow the manifest's order whatever order they
finish in. The native program runs with `ASAN_OPTIONS=detect_leaks=0`, as the
oracle's comparison run does: a program that stops with exit 70 still holds
what it held, and a leak report would break every ruled stop's exact stderr. A
disagreement is memory-unsafe when the native run's stderr names
AddressSanitizer or UndefinedBehaviorSanitizer, or the native process died on a
signal.

`--shard k/n` here keeps the programs whose position in the manifest's order,
each path counted once at its first listing, is `k-1` modulo `n`, so the shards
are disjoint and together cover the manifest. There is no diff corpus to keep
whole. A shard alone reads `partial` when it otherwise passes.

Each disagreement is one stderr line: `disagreeing <path>: exit node <n>,
javascript <n>, native <n>`, then, for stdout and for stderr where the three
differ, the first line at which they do, as each has it, and `; memory-unsafe`
when it is. The last stderr line is the summary:

```text
agreement: compared N, agreeing A, ruled R, checked stops C, pending P, disagreeing D, memory-unsafe U, refused F
```

`compared` is agreeing, ruled, checked stops, pending and disagreeing together;
a void program is in none of them. `--json` prints the report: those counts
(`checked_stops` and `pending` among them), `disagreements` (each with its
path, Node's, JavaScript's and native's observations, both compiles and
`memory_unsafe`), `compiler_failures`, `stale_runtime_stops`, `closed_sidecars`, every program
with its judgment, and `verdict`. A program judged by what it did carries
`outcome_hash`: hex SHA-256 over Node's, then JavaScript's, then native's exit,
stdout and stderr, each field written as its length, a colon and its bytes,
with the run's checkout written `{root}` in stdout and stderr, so a diff of two
days' reports sees any change in output or stop message even where the judgment
holds, and never the run's own scratch directory. A refused or void program has none, and a pending one
carries `task`. Exit codes are the delta's: a disagreement, a
compiler failure or a listed run-time stop that now agrees exits 1, a run that
proved nothing (a compile timeout, a runtime timeout that held after its retry,
or no clang) exits 3 as `void`, and anything else exits 0. A usage error exits 1,
as every usage error of the command does.
