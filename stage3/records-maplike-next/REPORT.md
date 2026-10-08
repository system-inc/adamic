Merged the records topic branch with integration's area-next fixture candidate.
Parents: fcddeb299460708f76e8436b4d887a260a6b84f1 and 506441dd; final commit is reported separately.
Validation: records fixtures, scanner MapLike replay, lowering, flow, stage3 fixtures and counts are recorded below.
Mutants: absent entry, named member and alias checks, record operations and runtime reads are recorded below.
Not covered: full scanner token output, complete repository gate, other platforms; no main or area branch changed.

## Conflict resolutions and retained refusals

The ten conflicts match integration's list. CohereSettings.json keeps both ignore lists.
IR keeps Record and all three typed-array kinds, reference counting and area-next's
argument, field-readiness, predicate and nested-frame metadata. JavaScript emits
both records and area-next's closure, count, field and Buffer paths.

The compiler overlaps that affect lowering or refusals are:

- expression.go keeps typed arrays, unknown/object union tags and Buffer storage
  before record selection. It preserves typed-array and Buffer view checks before
  recursively checking record values. Namespace, library method, arguments and
  enum dispatch stay before record expressions. Boxed library String remains on
  its existing storage path. Nullable record containers remain NotYet.
- library_method_values.go retains area-next's complete call/apply/bind and callback
  adapters, including their receiver, arity, presence, contextual-result and escape
  refusals. Alias readiness reads use the actual local type, including the boolean
  marker used for detached own-property aliases.
- library_object.go dispatches record inspection using the written arguments;
  fixed-object descriptor, prototype, shape and result-type refusals remain.
- locals.go keeps evolving-object inference and assertion readiness and adds the
  detached own-property marker and proven opaque helper parameter storage.
- object.go keeps Array.isArray, optional join and Buffer dispatch, then the
  detached own-property call, then the established user and library adapters.
  Record literals use dictionary storage before the fixed-object literal path.
- refusals.go keeps predicates, Node-library, typed-array, namespace, cast, method
  escape, optional view, class view, invariance, generator and suppression checks.
  It also retains record storage, named-write, literal prototype-name and detached
  method refusals. The blanket index-signature refusal becomes the records gate;
  unsupported signatures remain NotYet. Delete is permitted only for records;
  fixed-object delete remains Refused. Area-next's namespace, void and checked
  assertion paths replace the topic parent's obsolete blanket syntax refusals;
  their targeted restrictions remain in force. The ordinary unbound-method gate
  permits only aliases validated by these adapters.
- detached_own.go restricts its opaque-object argument check to the own-property
  helper whose body proves that storage. Other object parameters retain area-next's
  tagged representation checks. The helper still rejects a record through that
  fixed-object ABI.
- prototype.go keeps the records hasOwnProperty dense-apply path beside area-next's
  descriptor adapters. Other unsupported record prototype calls remain NotYet.
- A newly reachable record && string logical value would otherwise panic in native
  emission. expression.go refuses that unproven conversion as NotYet; its lowering
  refusal is pinned in records_test.go. No native emitter workaround is used.

The counts conflict keeps both fixture sets and area-next's regexp ownership row,
then uses the complete generator. Exact allocation and predicate deltas are in
this directory's evidence. Auto-merged native, flow and fresh visitors were checked
for both their new node kinds and Record nodes.

## Stage 3 status changes

Six records fixtures now compile: 01_has_property, 02_get_property, 03_own_keys,
05_integer_order, 06_delete_readd and 16_built_strings. The runner updated them only
after recorded Node, current Node, native, sanitizer and leak observations agree.
Five earlier index-signature refusals now name the actual remaining NotYet stop:
objects/10_build_options (unsafe write to named tscBuild), objects/27_delete_substitution
(numeric signatures), records/07_optional_view (fixed-object/record storage view),
records/13_strict_option (unsupported boolean | undefined record value) and
taste/06_localized_message (logical record result conversion). Exact changes are
in evidence/status-delta.json. Node output and provenance are unchanged.

The first broad gates lacked the candidate's required @types/node 25.3.3.
npm ci --prefix stage3/api --ignore-scripts installed its locked dependencies;
the final gates ran after installation. The initial stage3 run also exposed the
logical record emitter panic. The added lowering refusal stops before emission.

## Commands and observations

All test output went directly to logs, compressed under evidence. These completed
commands passed; the final scanner replay uses the compiler rebuilt after mutant
restoration.

```sh
ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api --ignore-scripts
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run 'TestPartialRecord|TestNamedRecord|TestNativeAgreesWithNode/internal/oracle/testdata/(records_|detached_own_|library_method_values|namespace|typed_arrays|non_null|taste_void|object_prototype)'
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(optional_|object_|enum|records_|detached_own_)'
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run '^TestRecord|^TestDetachedOwn'
go test ./internal/lower ./internal/flow -count=1 -timeout 30m
go test ./internal/fresh -count=1 -timeout 30m
go test ./stage3/fixtures -count=1 -timeout 30m
go test ./internal/oracle -count=1 -timeout 30m -run '^TestCountsAreRecorded$' -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 10m -v -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(library_string_raw.a|method_coverage_object_descriptors.a|library_method_values.a)$'
NAMED_INDEX_MUTANT_LOGS=/tmp/maplike-next-static-mutants python3 stage3/named-index-records/run-mutants.py
python3 stage3/records-maplike-next/run-mutants.py
go test ./internal/native -count=1 -timeout 10m -v -run '^TestRecordsAgainstNode$|^TestRecordReadMutants$'
go test ./internal/lower -count=1 -timeout 10m -run '^TestRecordRefusals$|^TestNamedRecord'
go vet ./...
gofmt -l cmd internal
git diff --check
```

The original fixture selection has grown on area-next: 89 native fixtures now
pass, covering the requested original 73 and six named fixtures; separate tests
also compare the partial and named fixtures with emitted JavaScript. The broader
merge-overlap selection passes 97 native fixtures. Both native builds (sanitized
and release), emitted JavaScript, stdout/stderr, checked-stop pins and ownership
checks pass. The original records operation/JavaScript and detached-helper tests
pass separately. The three changed-count witnesses pass uncached too.

Final lower passes in 169.722s, flow 254.705s, fresh 160.760s, stage3 52.881s and
counts 177.392s. IR passed in the initial package run, 26.536s. The broad overlap
oracle passes in 154.616s (native hits 0/misses 242; Node hits 0/misses 217), the
original selection in 60.914s (native hits 0/misses 239; Node hits 0/misses 176).
Record/helper mutants pass in 15.520s. Runtime comparison and read mutants pass
in 110.156s. The restored refusal tests pass in 1.680s. These are observed timings,
not estimates.

Setup reports Go 0.020s, Node 0.023s, submodules 0.064s, markdown 0.075s, clang
0.191s, build 96.074s, cache warm 98.049s and completion 98.186s. nproc is 5,
CPU quota is 4, Go is 1.27.1, Node v24.19.0 and clang 20.1.8.

## Scanner MapLike replay

The unchanged source is stage3/records-maplike-main/corePublic-MapLike.a, copied
by the topic branch from scanner 503d2e0d's exact corePublic MapLike witness.
Source Node, native and emitted JavaScript each print `ok` followed by a newline,
exit 0 and have empty stderr. Both byte comparisons are empty. This closes the
reported type-only MapLike stop, without claiming a full scanner build or token
comparison.

## Mutants observed

- Finite partial absent entry and named absent entry: synthesize an own undefined
  property; both backends are caught by Node stdout.
- Named member-kind guard removal: JavaScript runs on like Node, violating the
  required checked-stop pin. Named alias readiness removal: both backends exit 0
  instead of Node's initialization stop 70.
- Real named-index source changes bypass read type, named writes or alias contracts;
  TestNamedRecordReadTypes or TestNamedRecordRefusals catches each. Source restored.
- Logical record conversion and nullable record container admission: real source
  guard removals are caught by TestRecordRefusals, before native compilation.
- Eleven operation mutants alter read, write, delete, in, hasOwn, keys, values,
  entries, spread, for-in or stringify. Every mutation is caught by Node stdout.
- A prototype-read mutation violates the loud-stop pin; missing releases are
  caught by LeakSanitizer; eager coalesce and unchecked narrowed reads are caught
  by stdout or the checked-stop pin.
- Three observation mutations remove own-key guards where a read is discarded
  or compared; the missing inherited-key stop is caught where Node completes.
- Detached inherited-as-own mutations are caught by Node stdout for records and
  fixed objects; dropping helper binding readiness is caught by the stop pin.
- Runtime prototype membership restoration, silent missing reads and treating
  an own hit as missing are caught by exact diagnostic, exit and own-hit fixtures.

Each test's raw log and the five real source-mutation failure logs are retained.
No mutant was counted merely because C compilation failed.


The final full stage3 run also pins two stale host outcomes already present in
the candidate: host/09_realpath reaches unsupported process.platform, and
host/14_getCurrentDirectory reaches the existing .a non-null-assertion refusal.
Their source and current Node observations are unchanged. The installed locked
Node types also let host/24_useCaseSensitiveFileNames compile; the update runner
verified native against Node before updating that pin. The runner's six
compiling records updates and all seven remaining diagnostic changes are listed
in evidence/status-delta.json.

## Counts changes against 506441dd

The regenerated table contains 850 keyed rows (allocation and predicate tables),
compared with 818 in the candidate. It adds 33 compiler records/partial/named/helper
fixtures, with the same measured values as the topic parent's rows, and removes
one stale row: taste/17_binder_flow is already registered as unsupported because
writing an absent optional own field is NotYet. No supported fixture was removed.
All other candidate rows remain unchanged except these three:

| Fixture | Before | After | Observed reason |
| --- | --- | --- | --- |
| library_string_raw.a | 45/45/55/89/11/0 | 45/45/51/85/11/0 | Four raw-array temporaries transfer directly into literal object slots; generated C removes four retain/release pairs. |
| method_coverage_object_descriptors.a | 47/47/88/127/7/0 | 46/46/61/98/6/0 | Detached own-property aliases use a boolean readiness marker instead of an allocated closure, with their specialized descriptor calls and actual marker reads. |
| library_method_values.a | 137/137/154/295/42/0 | 136/136/152/291/41/0 | The own-property token closure becomes the readiness marker; other library adapters and their checks remain. |

Columns in each tuple are allocations/frees/retains/releases/peak/regions.
The untouched candidate was built in a detached scratch worktree at 506441dd;
its counted executions reproduce all three Before rows. The comparison links
cohere through the existing submodule, with no copied code. Go's VCS stamping
was disabled for that scratch build because of the linked submodule. The merged
fixtures were separately held to uncached Node in both backends. Counts added,
removed and changed are all listed in evidence/counts-delta.json; predicate rows
are unchanged. The String.raw generated-C diff is retained as evidence/raw.diff.gz.

Only codex/records-maplike-next is pushed, once, without force. Neither parent is
rewritten, and no main or area branch is merged into or pushed.

Merge commit: 7a2a4132ac763f18445da17c76ebdd1f6488aa63, with the two requested
parents. A following evidence-only commit compresses the generated-C diff and
records this validation note. Go formatting and vet are clean. The full merge
diff's whitespace check reports inherited raw sanitizer/patch/fixture evidence
from 506441dd; those bytes were retained. The records delta against 506441dd
passes its whitespace check. The generated-C context diff is compressed so its
intentional diff prefixes do not appear as source indentation warnings.
