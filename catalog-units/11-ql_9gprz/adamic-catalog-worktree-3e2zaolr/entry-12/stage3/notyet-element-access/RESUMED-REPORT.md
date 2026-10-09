d1a091d2: built finite computed reads, fixed tuple indexing and indexed reference presence; 16/91 root shapes covered.
Commits: f7ffffc8, af1612c7, b255534a, d1a091d2; checked non-null merge c7095bc7; current-main merge 1c1fe5d1.
Final focused oracle, lowering, command, IR and SSA checks passed; recorded counts refreshed; vet and diff checks clean.
Nine mutants caught: eight runtime mismatches or sanitizer failures, plus the unrelated-getter refusal boundary mutant.
Uncovered: 72 array-derived roots and three open dictionaries; both binder examples still stop on ElementAccessExpression.

The resolved starting tip was 1024963a98d1f6e87cc586db256e93b94e4832a6, beyond
f4556ebf. The previous REPORT.md was read first. Census replay was already merged
at 9a1f14c5d994aa855625e7cfa295677060348fec; merging its branch again reported
Already up to date, and that step was pushed. The checked non-null branch appeared
later and was merged at the requested c41c0e062e99da37820f822968d4df1b48cdaee7.
Its .ts-only policy remains intact. Current main 749a69ad was merged into this
branch; no main or area branch was written.

Finite string-literal unions and exact integer-literal unions dispatch through
ordinary IR helpers. Receiver and key are evaluated once, in that order, and
held across a key that replaces the receiver's last external owner. Optional
own fields read their actual absence; differently represented result alternatives
are boxed through existing IR. Indexed reference narrowing consults declared
field types and inserts the existing Defined check with the named-read diagnostic.
An aliasing call that clears a narrowed field now matches Node's TypeError.

Fixed tuples and fixed tuple unions dispatch numeric keys against their existing
named slots. Shorter members read undefined past their own end. Negative,
fractional, NaN and infinite keys read undefined; negative zero selects slot zero.
Variable tuples, incompatible storage at a shared index, weak slots, and
unrepresented results retain explicit NotYet stops.

Required fields with a name also used by an unrelated getter use the existing
accessor dispatch pass. The fixture holds both ordinary data and an actual getter
viewed through an interface. Optional same-name fields remain conservative.
The previous session's intentionally unlowered write witness moved to
internal/oracle/testdata/notyet_element_access/fields.a, retaining its false
registration. This keeps it out of the flow sweep's glob of runnable .a programs.
The sweep originally failed on that witness; after the move it passes.

Only elementAccess changed in object.go. New lowering helpers are in
internal/lower/element_access_fields.go and element_access_tuples.go. The existing
finite-key boundary test was updated in census_small_test.go. No backend, runtime
C, or another worker's lowering function was edited. Ordinary IR means all
ownership and flow passes see the new work without new node cases.

## Table count and replay observations

Run `python3 stage3/notyet-element-access/coverage.py`. It selects exact-reason,
unblocked roots directly from codex/stage3-notyet-table at
57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7, roots/raw.csv, deduplicated by position.
It reconciles all 91 positions against the previous independent checker
classification. [resumed-coverage.csv](resumed-coverage.csv) records every root;
[resumed-coverage.json](resumed-coverage.json) records the counts:

| Group | Roots | New rule shape coverage |
| --- | ---: | ---: |
| Array-derived receivers | 72 | 0 |
| Finite data-field keys | 15 | 15 |
| Fixed tuple union | 1 | 1 |
| Open dictionaries | 3 | 0 |
| Total | 91 | 16 |

The previous seven optional ordinary-array shapes are additional roots and are
not added to the original 91-site count. There are no string receivers or
Map-entry bracket operations in the original reason's roots.

All 16 covered positions and both binder examples were replayed. None of the
16 selected covered signatures reproduces on the final implementation. This
is not a claim that 16 whole units compile: some selected units remain blocked
upstream, and several reach later operator or write stops. Exact commands,
exit codes and every non-boundary finding are in
[resumed-replays.json](resumed-replays.json). Replay exit 1 means the selected
signature did not reproduce; it alone does not establish successful lowering.

| Example | Before prerequisite merge | Final observation |
| --- | --- | --- |
| binder.ts:1746:21 | a NonNullExpression | an ElementAccessExpression at 1746:21 |
| binder.ts:1757:28 | a NonNullExpression | an ElementAccessExpression at 1757:28 |
| parser.ts:10787:26 | fixed tuple-union read unsupported | Read lowers; next stop reading args at 10788:14 |
| checker.ts:14176:16 | finite read unsupported | Read lowers; next stop a BinaryExpression with a value and a value at 14177:17 |
| checker.ts:45839:34 | unrelated getter-name guard | Read advances to a BinaryExpression with a value and a value at 45839:16 |
| checker.ts:46488:34 | unrelated getter-name guard | Read advances to a BinaryExpression with a value and a value at 46488:16 |

The requested original binder commands were run with the baseline elementAccess
hook removed and restored before implementation validation. Both original
ElementAccessExpression signatures failed to reproduce and findings recorded
NonNullExpression. After the approved prerequisite merge, both original commands
exit 0 and reproduce the remaining array-derived ElementAccessExpression stop.
Their lowering is not complete.

## Mutants and validation

The runner restores its source after every mutant, including failures. Final
source hashes and full nine failure logs are retained in
[resumed-mutants.json](resumed-mutants.json). Commands:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/notyet-element-access/read_mutants.py
python3 stage3/notyet-element-access/read_mutants.py tuple
python3 stage3/notyet-element-access/read_mutants.py presence
```

| Mutant | What caught it |
| --- | --- |
| finite wrong-dispatch: equality to inequality | Native and backend stdout differ from Node |
| unrelated-getter-name: restore blanket name guard | Fixture fails Lower with the computed own-field NotYet |
| finite missing-absence: Absent false | Native exit 70 instead of Node exit 0 |
| finite repeated-receiver: evaluate receiver as a third argument | Extra receiver output and count 2 instead of 1 in both backends |
| tuple fractional-index: equality to less-or-equal | Native and backend stdout differ from Node |
| tuple missing-slot: Absent false | ASan heap-buffer-overflow on shorter tuple member |
| tuple out-of-range: return slot zero instead of undefined | Native and backend stdout differ from Node |
| tuple repeated-receiver: evaluate receiver as a third argument | Extra receiver output and wrong call count |
| missing-presence-check: return unchecked value | UBSan null object access; release exit differs from Node's TypeError panic |

Each mutant's fixture exits 1 from go test. All three runner commands exit 0
only after detecting those failures and restoring their source. No Go or clang
build failure is counted as a killed mutant. The getter-name mutant is a lowering
boundary test, distinguished from the eight runtime mutants.

Final commands wrote directly to /tmp/element-access-final-*.log:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCheckedNonNull|TestNativeAgreesWithNode/internal/oracle/testdata/(element_access_(reads|tuple|optional|presence)|notyet_element_access/fields)' -count=1 -timeout 10m
go test ./internal/lower ./internal/flow ./internal/ir ./cmd/adamic -run 'TestCensusSmallFiniteKeyRead|TestInheritedLibraryReadsNeverLoadOwnFields|TestNonNull|TestExplainChecksDriver|TestEveryFunctionIsInSingleAssignment|TestCallTargetsIncludeEveryDescendant|TestClosureTargetsBoundOnlyProvenValues' -count=1 -timeout 10m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
go vet ./internal/lower ./internal/ir ./internal/flow ./cmd/adamic
git diff --check
```

Observed outputs: oracle ok 1.680s; lower ok 7.767s; flow ok 20.564s;
IR ok 0.214s; cmd/adamic ok 3.389s; counts ok 22.733s. Vet and whitespace
checks exited 0 without diagnostics. Runnable new fixtures are held to source
Node, JavaScript backend Node, native release, ASan and UBSan. Successful fixtures
also undergo the harness's leak check. The firing presence fixture intentionally
panics and retains what was alive at that stop, as Node does.

Final new fixture counts, allocations/frees/retains/releases/peak/regions:
reads 42/42/17/51/9/0; tuple 76/76/73/123/16/0; presence 4/3/5/7/4/0.
Every other existing counts row retains its measurements. No whole-package test
or full repository gate was run. The named SSA sweep spans the runnable corpus.

Setup command was `export GOPROXY='https://proxy.golang.org|direct'` followed by
`bash cloud/setup.sh`, logged to /tmp/element-access-setup.log. It exited 0.
Timing lines: Node ready 0.027s; Go ready 0.049s; clang ready 0.274s; markdown
installation 2.185s and ready 2.282s; submodules ready 23.418s; Go build ready
192.720s; test binaries deferred 192.831s; cache warm 192.832s; done 192.861s.
`nproc` is 5, cpu.max is 400000 100000. The printed
/workspace/adamic-tools/env.sh was sourced for every build and test command.

## Work still required

The 72 array-derived roots are 62 NodeArray, six Readonly<PathPathComponents>,
two TemplateStringsArray, one SortedReadonlyArray and one JSDocArray. Their
current representation is Object, while ArrayIndex expects native array storage.
NodeArray carries pos/end, hasTrailingComma and transformFlags;
TemplateStringsArray has raw, and JSDocArray has jsDocCache. Treating those objects
as plain arrays would discard observable fields or read the wrong layout.

This is implementation work, not a declaration that the 72 forms are forbidden.
It needs an agreed representation and construction/conversion path preserving
numeric slots, length, metadata and identity. representation is actively owned
by codex/notyet-representations (ownership check included 9813a901); it was not
edited. A separately fetched views-arrays-callables-parser branch has a much
larger view/brand implementation on a different compiler history. Its diff was
not imported into this small unit or substituted for the area representation.
The binder roots remain a concrete handoff to the representation owner, then
back to numeric ArrayIndex and its existing presence checks.

The three open dictionaries retain the index-signature/Record refusal in
CLAUDE.md's approved docs/0.1.md language table. Opening them needs a ruling
covering dynamic own keys, missing-key undefined, prototype names, numeric-key
conversion and ordering, and shape growth on writes. No dictionary refusal was
weakened. Required computed writes are already implemented on the separate
object-small branch at 9104bc07; setIndex stays that worker's function. Optional
computed writes still require shape growth. No shared runtime C helper was added.
