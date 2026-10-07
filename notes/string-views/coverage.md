# String view coverage

Base: origin/codex/string-views at ba9c9ef. Coverage branch is cut directly from that commit.
Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the requested three-dot diff and commit messages. The three-dot diff includes inherited runtime-area work (WASI, borrowing, normalization and their evidence). The feature implementation itself is in 19be8f5 and d3ccac2; 66534dc and ba9c9ef record measurements and validation. Reviewed string runtime files, the native string-view test, oracle registration/counting, existing string programs and the feature report and mutant runner.

Nine new programs, zero output disagreements. All filenames below are in internal/oracle/testdata. Existing means the file was already present at ba9c9ef; new filenames have the prefix string_views_. A condition can be reached in a program without its allocation policy being observable in stdout. Recorded counts hold allocation changes; ASan and LeakSanitizer hold lifetime. The native TestRuntimeStringViews additionally inspects ownership and the exact policy boundary.

| Code case | Existing program evidence | Added coverage |
|---|---|---|
| slice relative positive, negative, fractional, NaN, infinities, clamp below zero and above length | string_build_boundaries.a, string_positions.a, shared_slices.a | methods.a on short nested views |
| slice with omitted end and omitted both bounds | shared_slices.a, string_build_boundaries.a | lifetime.a, methods.a, policy.a |
| slice from >= to, equal bounds, empty source | ascii_scan.a, string_build_boundaries.a | methods.a, policy.a |
| flagged ASCII slice uses bytes | shared_slices.a, shared_slice_append.a | lifetime.a, holders.a, loops.a, policy.a |
| Unicode whole-code-point bounds use located bytes and carry unit count | shared_slices.a, string_build_boundaries.a | lifetime.a, methods.a, policy.a |
| Unicode slice begins on low half only | shared_slices.a, string_build_boundaries.a | surrogates.a |
| Unicode slice ends on high half only | shared_slices.a, string_build_boundaries.a | surrogates.a |
| Unicode slice begins low and ends high, with or without middle bytes | shared_slices.a, string_build_boundaries.a | surrogates.a, methods.a |
| existing lone high/low surrogates retained by slicing | string_build_boundaries.a | surrogates.a, methods.a |
| substring clamps negative to zero, truncates/normalizes, swaps start and end, omitted end | library_string_indices.a, string_build_boundaries.a | methods.a, calls.a, holders.a |
| share size == 0 returns immortal empty (trim calls share directly) | maps_and_text.a trims empty and all-space strings | policy.a, methods.a cover the separate empty-slice entry path |
| share owner is original string | shared_slices.a | lifetime.a, holders.a |
| share owner is ultimate owner of a view | shared_slices.a, shared_slice_append.a | lifetime.a, holders.a, policy.a |
| share whole string retains itself | shared_slices.a | methods.a, policy.a |
| small heap view of small owner now shares (no 64-byte minimum) | string_build_boundaries.a reaches small views but does not escape them | lifetime.a, holders.a, throw.a |
| constant owner pins nothing; view has no retained owner | shared_slice_append.a, string_build_calls.a | methods.a |
| heap owner capacity equals length | string_build_boundaries.a via repeat | policy.a |
| heap owner has spare append capacity; storage uses capacity | string_append.a | loops.a, policy.a |
| storage/header threshold below, equal, above eightfold bound; share and copy | string_build_boundaries.a samples lengths, no exact new threshold | policy.a: parents 511/512/513 and views 7/8/9 (64-byte header here) |
| tiny view of large owner copies | string_build_boundaries.a | policy.a, including tiny view of large view |
| large view of large owner shares | shared_slices.a | policy.a |
| copied and shared ASCII results preserve flag | string_build_boundaries.a | policy.a and reads/searches afterward |
| shared views have independent Unicode caches, relative to their first byte | shared_slices.a, string_positions.a | policy.a |
| parent reassigned while child lives | shared_slices.a (intermediate parent) | lifetime.a (original and intermediate), loops.a |
| parent goes out of scope while child lives | shared_slices.a (large array-held results) | lifetime.a (small returned result), holders.a |
| views stored in arrays then read later | shared_slices.a | holders.a, loops.a, surrogates.a |
| views stored in Map keys | shared_slices.a | holders.a, throw.a |
| views stored in Map values then read later | no dedicated escaping small-view program found | holders.a, throw.a |
| views stored in object fields then read later | no dedicated escaping small-view program found | holders.a, throw.a |
| views escape a throwing frame, error message is also a view; catch/finally read holders | no dedicated small-view program found | throw.a |
| views of strings built in loops, then parent extended | shared_slices.a builds parents; shared_slice_append.a extends views | loops.a extends parent with earlier small views still held |
| indexing invalid bracket index: negative, fraction, NaN, end, infinity | string_index.a | characters.a valid bracket indexes; methods.a method normalization |
| ASCII character pool first initialization and reuse for every value 0..127 | no complete pool-value/lifetime program found | characters.a, including NUL and DEL, ASCII and long Unicode sources |
| cached ASCII unit from short Unicode string | string_index.a | lifetime.a, methods.a |
| cached ASCII unit from long Unicode string | string_positions.a | characters.a |
| non-ASCII one-unit result uses slice, including surrogate halves | string_index.a, library_string_indices.a | methods.a, surrogates.a |
| charAt truncates, NaN -> 0, invalid -> empty | library_string_indices.a | methods.a on nested views |
| charCodeAt ASCII bytes, Unicode short walk, long cached UTF-16 view; missing -> NaN | ascii_scan.a, string_positions.a | methods.a, characters.a, policy.a |
| at truncates, NaN -> 0, negative relative, out-of-range -> undefined | strings_more.a, library_string_indices.a | methods.a on nested views |
| indexOf empty needle returns clamped from | search_from.a, search_from_sweep.a | methods.a |
| flagged ASCII indexOf needle longer than remaining bytes, match at first/end, absent, partial prefix mismatch | lint_runtime_search_boundaries.a | methods.a, policy.a |
| flagged ASCII indexOf with ASCII, Unicode, surrogate, NUL, literal/unflagged or sliced needle | lint_runtime_search_boundaries.a, search_halves.a | methods.a, calls.a |
| Unicode indexOf whole needle, from >0 or past end, starting on low half skips next code point | search_from.a, search_from_sweep.a | methods.a |
| needle starts low or ends high -> unit search; internal lone surrogate stays byte search | search_halves.a, lint_runtime_search_boundaries.a | surrogates.a, methods.a |
| lastIndexOf empty, too-long, absent, whole needle, first/last match, failed byte candidates, Unicode unit position | runtime_last_index_of.a, lint_runtime_search_boundaries.a | methods.a, policy.a, surrogates.a |
| lastIndexOf surrogate boundary -> unit search | runtime_last_index_of.a, strings_more.a | surrogates.a |
| other callers of the same share helper: trim, split pieces, string iteration by whole code point | shared_slices.a, maps_and_text.a, string_positions.a | existing programs already reach these callers |
| direct method syntax on strings and views | all existing string programs | all nine |
| String.prototype.method.call for all seven methods on ASCII/Unicode views | library_string_prototype.a, string_build_calls.a, lint_runtime_search_calls.a cover subsets | calls.a |
| receiver evaluated before arguments reassign it; start/end each evaluated once in order | library_string_indices.a, library_string_prototype.a, lint_runtime_search_calls.a | calls.a |
| prototype calls convert primitive number/boolean receiver to string | library_string_prototype.a, lint_runtime_search_calls.a | calls.a adds substring and charAt conversions |
| codePointAt ASCII, Unicode short/long, pair high vs low, final unpaired high | string_positions.a, string_build_caches.a | unchanged feature dependency already covered |

## Cases unavailable to an Adamic source program

- lastIndexOf(search, position): lowering's stringMethods table accepts only one argument; observed NotYet in both direct and prototype-call probes. No disagreement is claimed for a refused program.
- Direct substring(), charAt(), charCodeAt(), at() without arguments: this branch's embedded library declarations reject them (TS2554). slice() works. Prototype charAt/charCodeAt/substring omissions are supported by lowering, but the direct source forms are outside the declared subset.
- Borrowed C stack adamic_string with zero references and no literal marker: user source cannot construct runtime headers or stack storage. Existing TestRuntimeStringViews covers whole/partial borrowed-copy results.
- Near SIZE_MAX storage/header arithmetic, or non-ASCII index offsets exceeding UINT32_MAX/2: source string-size checks and available memory prevent fabricating those runtime states safely. No such allocation attempted.
- Header identity, owner pointer/count, capacity vs length, exact pin ratio, and zero allocations per cached character cannot be printed through the language's string API. The programs reach those paths; recorded counts and existing native pointer/count assertions establish their representation behavior.
- Detached methods, apply/bind forms, null/undefined receivers, object ToPrimitive conversion: outside the supported intrinsic-call subset. library_string.go only recognizes immediate prototype .call, refuses null/undefined and nonprimitive conversions.

## Commands

Every build/test shell sources /workspace/adamic-tools/env.sh. Logs are retained in evidence/ after validation finishes. Early oracle attempts found only subset refusals, which were rewritten into supported forms.

```sh
bash cloud/setup.sh
# First setup overlapped checkout of the requested branch and failed its cache warm.
# Repeated on the stable branch, with complete output saved:
bash cloud/setup.sh > /tmp/string-views-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/string_views_' -count=1 -timeout 30m > /tmp/string-views-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/string_views_' -count=1 -timeout 30m > /tmp/string-views-oracle-2.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/string_views_' -count=1 -timeout 30m > /tmp/string-views-oracle-final.log 2>&1
python3 notes/string-views/run-builds.py > /tmp/string-views-build-final.log 2>&1
# Runner invokes this separately for every string_views_*.a, then runs the binary and source on Node:
# go run ./cmd/adamic build internal/oracle/testdata/<program>.a -o /tmp/string-views-builds/<program>
# /tmp/string-views-builds/<program>
# node --disable-warning=ExperimentalWarning oracle/node.mjs /workspace/adamic/internal/oracle/testdata/<program>.a
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/string-views-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/string-views-gate.log 2>&1
```

Setup timing: Go ready 0s; clang ready 0s; Node ready 0s; submodules ready 0s; build cache warm 97s; done 97s on 5 processors. nproc: 5. cgroup cpu.max: 400000 100000.

Additional verification and mutation commands:

```sh
go vet ./... > /tmp/string-views-vet.log 2>&1
gofmt -l cmd internal > /tmp/string-views-gofmt.log
git diff --check > /tmp/string-views-diff-check.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m > /tmp/string-views-counts-check.log 2>&1
git worktree add --detach /tmp/string-views-mutant ba9c9ef
cp internal/oracle/oracle_test.go /tmp/string-views-mutant/internal/oracle/oracle_test.go
cp internal/oracle/testdata/string_views_*.a /tmp/string-views-mutant/internal/oracle/testdata/
cp internal/oracle/counts.md /tmp/string-views-mutant/internal/oracle/counts.md
# The worktree's empty cohere submodule directory is replaced with a symlink to the initialized dependency.
rmdir /tmp/string-views-mutant/cohere
ln -s /workspace/adamic/cohere /tmp/string-views-mutant/cohere
python3 notes/string-views/run-mutant.py > /tmp/string-views-mutant-runner.log 2>&1
# Runner changes one line in the isolated checkout, runs this test, restores in finally, then runs it again:
# ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/string_views_lifetime.a -count=1 -v -timeout 30m
```

The mutant changes only the final owner assignment in string_share.c, replacing adamic_retain((adamic_string *)owner) with (adamic_string *)owner. The fixture uses dynamically built short strings and nested views; immortal literals cannot hide the missing count. Source in the primary checkout stays clean throughout the isolated experiment.

## Observed verification

All nine final CLI builds exit 0 and match Node stdout and stderr exactly. The focused uncached oracle passes (1.237s). Counts regeneration passes (60.226s), adds exactly nine rows and leaves every previous row unchanged. The separate uncached counts check passes (93.562s). Vet exits 0, formatting prints nothing, and diff whitespace checks print nothing.

The one-line missing-owner-retain mutant exits 1 and is caught by AddressSanitizer heap-use-after-free in string_views_lifetime.a (the run reached native execution, not a warning or refusal). After restoration the identical uncached oracle exits 0 (1.507s), and git diff for the isolated string_share.c is empty. Full mutant diagnostics are in evidence/mutant.log.gz. This is a deliberate mutant failure, not an output difference on the submitted branch.

The broad uncached repository gate is **partial**. It was stopped after about 30 minutes, exit 143 (SIGTERM), while stage1/cohere/markdownblocks and type-aware corpus work remained. Before stopping, it passed the complete internal/native package (410.164s), complete internal/oracle package (380.614s), bridge, compiler commands, flow, freshness, fuzz, IR, loading, lowering, regexp, Unicode properties, CSS, JSON and lint packages. No natural test failure was reported before stopping. This does not claim a complete repository gate pass. The complete output is evidence/gate-partial.log.gz.

The stop command was inline Python: verify PID 5582 is the go test -count=1 -timeout 30m ./... process, enumerate its descendants with ps -eo pid=,ppid=, send SIGTERM to its six remaining children and then PID 5582. The Go command returned 143. This stopped only this task's broad test run.
