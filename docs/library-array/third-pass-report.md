Built: proven Array intrinsic aliases and metadata, oversized map errors, function shift errors, and throwing sort/toSorted callbacks.
Commits: claims 556c33a and de2a5ad; implementation fa291a9.
Commands and outputs: validity Array 242 to 247 pass, 386 to 381 refused, zero disagreements; Linux packages, Array oracle, counts and vet pass.
Mutants: ten method mutants and two guard mutants caught only by Node stdout; 33 retained mutant executions also caught.
Uncovered: 349 language first blockers and 32 other-library first blockers; full repository functional gate not run.

## Measurement

Branch codex/library-array-3 starts at second-pass tip bcf803d, as requested.
origin/main was fetched but not merged: the requested base is the second-pass
branch. The unchanged independent validity runner is af12899, built from the
separate /workspace/array2-validity checkout; Adamic builds from this branch's
sources. Test262 is c8c798898646638cd0c24879f8e0374e847e7d74, stock tsc 6.0.3,
adaptation enabled. The baseline was measured before implementation edits.

| Survey | Pass | Disagreement | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| Before | 242 | 0 | 386 | 0 | 610 |
| After | 247 | 0 | 381 | 0 | 610 |

All 242 baseline passes remain passing. All five new passes agree with Node.

Only refused is work owed. Both complete surveys contain 1844 not-typescript
outcomes and 3082 total inputs. The intermediate measurement was 246 pass,
zero disagreement, 382 refused, zero crashed and 610 skipped, before the
never-returning comparator overload. Full tables, newly passing paths and the
386-row original-to-final refusal ledger are archived alongside this report.

## Classification and implementation

Claim commits precede compiler edits. Reviewing actual source split the inherited
16 Array blockers into six Array library blockers, seven language blockers and
three other-library blockers. One additional claimed language item was actually
a TypeScript-valid never-returning sort callback and was claimed as Array work
before implementation. The final Array first-blocker count is zero; this is a
classification of this pinned corpus, not a claim of universal Array support.

Unreassigned const/let aliases have no runtime slot only when a whole-module
symbol-use proof restricts them to intrinsic observations or supported calls.
This preserves typeof, name, length and nonconstructible prototype metadata.
Escapes, shorthand storage, writes, identity, exports, alias chains, early reads
and hoisted readers remain refused. Ordinary intrinsic identity and call
semantics are retained; no user closure pretends to be the built-in function.

Generic map.call ports the oversized non-array ArraySpeciesCreate failure:
all supplied argument effects run, a proven immutable numeric length exceeds
the Array length bound, callable validation succeeds, and RangeError occurs
before any callback. Sparse ordinary mapping remains refused. Only exact
power-of-two exponentiation is used in Math.pow proofs. Shorthand escape,
mutable/aliased receivers and unproven lengths cannot enter this specialization.

Generic shift.call ports the zero-length function receiver's readonly SetProperty
TypeError, preserving V8's exact short function source including UTF-8 comments.
Typed, long, stored, nonzero-arity and generator functions remain refused.

Sort/toSorted accept represented callbacks whose proven return type is never.
The native adapter does not read a number from the callback's void return slot;
existing TimSort propagates the thrown exception before using the adapter result.
The JavaScript backend already propagates it with the correct algorithm.
Optional number sorting retains existing undefined partitioning. Optional
reference sorting remains refused, including numeric callbacks: omitting that
guard demonstrably calls a comparator which Node does not call.

Nine positive .a fixtures cover metadata, receiver errors, argument evaluation,
length boundaries, callback effects, heap ownership, empty/singleton copies,
optional numeric partitioning and abrupt completion. Two permanent negative .a
fixtures cover escaped map lengths and optional reference sorting. Allocations
and frees balance for all nine positives. Small shared hooks only register
proven aliases and carry the sort callback proof. The integration-owned
library_array_mutant_test.go, lower.go, native.go, emit.go and oracle_test.go
remain untouched.

V8 13.6.233.17 ports are named in THIRD_PARTY_NOTICES.md:
src/builtins/array-map.tq, src/builtins/array-shift.tq and
third_party/v8/builtins/array-sort.tq. No approximation replaces a missing
representation or algorithm.

Join metadata now lowers, but its corpus input next refuses a string plus
undefined in an unreachable branch. Function shift now lowers, but its corpus
input later constructs a readonly length using Object.defineProperty. Those
remaining inputs are language work. third-pass-handoff.md lists every language
family, counts and one-line reproducer for @system_adamic; the JSON preserves
observed messages and explains masked blockers. The 32 other-library inputs
need Boolean/String/Number/Date/Error subclass constructors, Object.freeze or
Object.prototype.toString and are outside this Array slice.

## Mutants

Every new method-family mutant below compiles, exits zero, has empty stderr and
no ASan/UBSan/leak report. Only the source-on-Node stdout comparison kills it.
TestLibraryArrayNewFamilyMutants retains these permanently. Two source guard
mutants were applied separately and restored in finally; subsequent refusal
tests pass, and no temporary test remains.

| Mutation | Fixture that catches it |
|---|---|
| aliased isArray classification | library_array_alias_is_array.a |
| aliased every result | library_array_alias_every.a |
| join intrinsic own prototype | library_array_alias_join.a |
| aliased includes nullish error | library_array_alias_includes.a |
| aliased indexOf boundary | library_array_alias_search.a |
| aliased lastIndexOf negative bound | library_array_alias_search.a |
| map oversized RangeError message | library_array_map_oversized.a |
| shift function readonly message | library_array_shift_function.a |
| sort never callback skipped | library_array_sort_never.a |
| toSorted never callback skipped | library_array_to_sorted_never.a |
| Remove shorthand-value escape proof | library_array_refused/library_array_map_length_escape.a |
| Remove optional-reference partition proof | library_array_refused/library_array_sort_never_optional_references.a |

The map guard mutant prints `RangeError Invalid array length` while Node prints
`mapped`. The optional-reference sort guard mutant prints `compare`, `Error stop`
and `|aa` while Node prints `aa|`. Both native artifacts exit zero with clean
sanitizers. Restored lowering refuses both programs explicitly.

The 33 retained executions also run in the Linux gates: twelve lower-level
second-pass mutants, thirteen first-pass oracle mutants (including with bounds),
and eight shrinking-search mutants. All are checked against Node; nonoptional
shrinking fixtures additionally check the intended inserted stop. The optional
unconditional-stop mutants deliberately exit 70 and are caught by Node's
continuing output. These are retained regression checks, separate from the
new method mutants' clean-exit requirement.

| Retained lower-level mutation | Catching fixture |
|---|---|
| constructor argument order | library_array_dense_construction.a |
| reduceRight index stride | library_array_reductions.a |
| reduce saved length boundary | library_array_reductions.a |
| reduce empty TypeError message | library_array_reductions.a |
| generic nullish TypeError message | library_array_receiver_calls.a |
| generic primitive every result | library_array_receiver_calls.a |
| from string iteration order | library_array_from_string.a |
| string mutation readonly error | library_array_string_mutations.a |
| generic forward boundary | library_array_search.a |
| generic backward negative sign | library_array_search.a |
| isArray classification | library_array_metadata.a |
| copyWithin undefined end | library_array_copy_within.a |

| Retained oracle mutation | Catching fixture |
|---|---|
| iterator index | library_array_iterators.a |
| join separator | library_array_join.a |
| method metadata | library_array_metadata.a |
| with replacement | library_array_with.a |
| flatMap order | library_array_flat_map.a |
| spliced copy | library_array_spliced.a |
| flat order | library_array_flat.a |
| copyWithin write | library_array_copy_within.a |
| search direction | library_array_search.a |
| copy reversal | library_array_copy.a |
| default sort ordering | library_array_copy.a |
| reverse callback order | library_array_find_last.a |
| with bounds omitted | library_array_with.a |
| findLastIndex missing-index skip, native and JavaScript | library_array_find_last_shrinks.a |
| findLast missing-index skip, native and JavaScript | library_array_find_last_value_shrinks.a |
| optional numeric unconditional stop, native and JavaScript | library_array_find_shrinks_optional_numbers.a |
| optional Leaf unconditional stop, native and JavaScript | library_array_find_shrinks_optional_leaves.a |

## Linux validation and limits

Exact commands are in third-pass-reproduce.md. Test outputs are log files, never
piped. Archived logs include setup, package gates, Node-only mutants, guard
restoration, counts, vet and the refusal census.

- `go test ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/flow ./internal/load`: pass; lower 31.276s, native 155.639s, flow 108.301s, load 1.753s; JavaScript and IR have no package tests.
- `go test ./internal/lower -count=1 -run TestLibraryArray`: pass, 10.742s, all 22 method mutants and the refusal checks, including generator functions.
- Uncached filtered Array/related oracle plus retained mutants: pass, 7.464s; native, emitted JavaScript and Node comparisons include every new positive fixture.
- Counts update: pass, 37.805s. Counts verification: pass, 36.078s. Every added fixture has allocations equal to frees.
- `go vet ./...`: pass, empty output.
- Cohere types-only on exact temporary source mirrors with project strict options and prelude: pass, 13 checked in 0.07s. The pinned Cohere cannot select .a; repository sources stay .a. The full mirror lint/format gate fails with 48 findings and 11 files that would change, including existing prelude findings and purpose-built calls/let aliases/comparator forms. The .a selection check reports no files to check. Neither full style gate is claimed green.

`bash cloud/setup.sh` succeeded. Timing lines: Go ready 0s; clang ready 0s;
Node ready 0s; submodules ready 0s; build cache warm 61s; total 61s. `nproc`: 5;
cgroup cpu.max: 400000 100000; memory 17.6 GB. Tool versions: Go 1.27.1,
clang 20.1.8, Node 24.19.0. No setup workaround was required. The shallow clone's
main-only fetch refspec required explicitly fetching the requested branch refs.

Full repository functional tests were not run. Unsupported language foundations,
other libraries, generic sparse mapping, escaping intrinsic values and optional
reference sort partitioning remain refused. The requested runner retains one
checker diagnostic mismatch in refused although tsc rejects with another
message; the survey is reported unchanged, rather than adjusting its outcome.
