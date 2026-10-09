# Runtime feature coverage

Base: origin/codex/lint-runtime-fixes at 8c171be2163f8bad0410e7061520812f766b863d.
Reviewed CLAUDE.md, README.md, docs/0.1.md and docs/memory.md, the full
origin/main...origin/codex/lint-runtime-fixes diff, all five branch commit messages,
and all eight changed files. Searched existing oracle programs before adding these five.
The oracle uses an explicit fixtures list, so each program is registered in oracle_test.go.

## Cases and existing programs

Paths below are relative to internal/oracle/testdata. Existing means on the base branch.
The rows describe source-level coverage, not instrumented branch coverage or speed measurements.

| Runtime case | Existing program using it | Additional coverage |
|---|---|---|
| Release NULL, including clearing optional state and scope exit | undefined_strings.a, undefined_references.a | lint_runtime_release_chain.a clears an empty and a populated global chain; emitted C releases the NULL global at exit |
| Release count-zero immortal strings and booleans | strings.a, unions.a | literal children in lint_runtime_release_chain.a; boolean union assignments in lint_runtime_release_shared.a |
| Release still-shared heap values without destroying them | self_assignments.a, collections.a, closures.a | lint_runtime_release_shared.a drops 255 of 256 holders before reading the survivor; 256 discarded identity returns directly release each shared text, object, array, map and closure |
| Release the last owner, then drain children iteratively | long_chain.a has one million plain links | lint_runtime_release_chain.a has 100,000 links, each with a fresh string and array, then another drain after the queue empties |
| Grow the destruction queue beyond its initial 64 entries | size_class_churn.a drops arrays of 4,000 distinct objects; long_chain.a is narrow | lint_runtime_release_chain.a frees an array of 256 distinct last-owner roots, all outside regions |
| String destruction without an index | strings.a | runtime-built labels and captured text in both release programs |
| String destruction with an index | shared_slices.a, long_literals.a | long mixed search strings in lint_runtime_search_boundaries.a |
| String slice destruction drops its byte owner | shared_slices.a, shared_slice_append.a | shared and nested slices in lint_runtime_search_boundaries.a; equal-header slices |
| Plain object destruction: reference and numeric fields, absent and immortal children | objects.a, long_chain.a | links with shared label/array children and undefined next pointers |
| Class destruction: inherited fields and arrays | class_oct6_release.a | already covered |
| Array destruction: reference elements, NULL elements, numeric elements, metadata | collections.a, undefined_elements.a, regexp_exec.a, regexp_match.a | shared object arrays and wide roots; labels alongside NULL and immortal elements |
| Map destruction: held keys/values, scalar values and shared objects | collections.a, library_map_set.a | 256 holders of one object-valued map |
| Cell destruction: reference and scalar captured values | closures.a, method_closures.a | shared closure reads a reassigned captured heap string |
| Closure destruction drops captured cells | closures.a | 256 holders of one closure, then final scope destruction |
| Number-box destruction | unions.a | 64 numeric union assignments per churn, across three churns; generated C calls adamic_box_number |
| Boolean-box handling | unions.a | repeated boolean union assignments; boxes are immortal, so no heap boolean can reach final-owner destruction |
| Weak handle dropped and target forgotten | weak_parent.a, weak_narrowed.a, reuse_weak_after_reuse.a | already covered |
| Map iterator destruction lets go of map and decrements iterating | library_map_set_next.a, map_iteration.a | already covered |
| lastIndexOf empty needle, including empty haystack | strings_more.a, runtime_last_index_of.a | explicitly printed beside indexOf on empty haystacks and shared slices |
| Whole-character byte search: ASCII, two-byte, four-byte needles | runtime_last_index_of.a, shared_slices.a | direct and intrinsic calls, full-string match at byte zero, shared and nested slices |
| Three-byte BMP needle on backward byte path | no explicit lastIndexOf example found | lint_runtime_search_boundaries.a uses 中中文 |
| Internal lone high/low surrogate on byte path | runtime_last_index_of.a has all slices of a lone-surrogate text | longer x-surrogate-y strings, with repeated candidates |
| Needle starts with low surrogate or ends with high surrogate: UTF-16 fallback | runtime_last_index_of.a, search_halves.a | sliced needles and intrinsic calls with either half of 😀 |
| Needle longer in bytes than haystack | runtime_last_index_of.a includes empty haystack cases | entire heap haystack plus one byte; empty haystack with nonempty needles |
| Matching first byte, mismatching following bytes | runtime_last_index_of.a has absent searches | repeated prefixes followed by ?; backward scan must continue |
| Absent needle with no matching first byte | runtime_last_index_of.a | missing on originals and slices |
| Match at zero, middle, and last possible byte; result measured in UTF-16 | runtime_last_index_of.a, long_literals.a | whole haystack as needle, inserted ! marker and suffix piece |
| indexOf default/from-position, supplementary low-half start and clamping | search_from.a, search_from_sweep.a | intrinsic indexOf from position 2, beside reverse search |
| Direct string method call | searches.a, strings_more.a | both new search programs |
| Intrinsic String.prototype.indexOf.call | library_string_prototype.a | ASCII, non-ASCII, half-surrogate needles, primitive receiver conversion |
| Intrinsic String.prototype.lastIndexOf.call | none found | lint_runtime_search_calls.a, including a numeric receiver |
| Receiver retained while argument evaluation overwrites its variable | generic lending fixtures, but no explicit reverse search found | lint_runtime_search_calls.a uses a heap receiver in both direct and intrinsic calls |
| Equality: both NULL, left NULL, right NULL | narrowed_compared.a uses undefined comparisons, often specialized to NULL checks | lint_runtime_equal_headers.a passes both operands through optional string parameters; generated C calls adamic_string_equal |
| Equality: same header, immortal or heap | strings.a, shared_slices.a | heap copies and shared slice compared with themselves through parameters |
| Equality: differing lengths | strings.a | appended ! versus original |
| Equality: distinct nonempty headers with equal bytes | unions.a, shared_slices.a | independent repeats, copies and shared slices, ASCII/WTF-8/multibyte |
| Equality: distinct headers with equal length but different bytes | unions.a | final a versus final b |
| Equality: distinct zero-length heap headers | from_codes.a compares an empty result with a literal; no independent empty concat pair found | two empty concat allocations and an empty literal versus an allocation |
| Equality: canonicalized surrogate pair bytes and sliced halves | from_codes.a, shared_slices.a | joined high/low halves equal 😀; each sliced half equals its lone-surrogate literal |
| Equality call routes: ===, !==, Object.is, union equality, map key comparison | strings.a, library_object_is.a, unions.a, shared_slices.a | first three compared on every optional-string pair in lint_runtime_equal_headers.a |

## Limits and refused forms

A missing *match* is covered. An omitted search argument is a different case:
the bundled declaration requires a string needle. Both omitted-needle intrinsic
probes fail TS2554; explicit undefined needles fail TS2345. No native program is produced.
The direct lastIndexOf('a', 1) probe is accepted by the checker but lowering says
"stage 0 can't lower lastIndexOf with these arguments yet" (object.go's stringMethods table).
String.prototype.lastIndexOf.apply is refused as an unbound method read.
The scratch probes are /tmp/lint-runtime-builds/{last_position,last_missing,index_missing,
last_undefined,index_undefined,search_apply}.a, with separate refusal logs.

The draining == true early return in release_last has no source-level trigger found:
free_one drops children through let_go rather than adamic_release; the runtime's
free callbacks do not call user code. It cannot be forced by an ordinary Adamic
program without modifying the runtime. Allocation/reallocation failure and reference
count exhaustion are not deterministic closed oracle inputs. A heap boolean with a
last owner cannot be made: both boolean boxes have immortal count-zero headers.
The identical-header shortcut and the empty-drain/noinline changes preserve output;
these programs validate behavior and lifetime, not whether memcmp was skipped or
whether a compiler preserved noinline, and make no performance claim.

## Mutation proof

Changed exactly one real runtime line in string_search_impl.h:

```diff
- for (size_t offset = string->length - search->length + 1; offset-- > 0;) {
+ for (size_t offset = string->length - search->length + 1; offset-- > 1;) {
```

The uncached oracle for lint_runtime_search_boundaries.a exited 1 with "stdout differs".
When needle == text, Node printed `0 0 0 0 -1 -1` and mutant native printed
`0 -1 0 -1 -1 -1`. Both native modes compiled and finished with exit 0 and empty
stderr; this was not a compiler warning or a sanitizer failure. A try/finally
restored the source verbatim. The five-program oracle then passed again.
The full mutant log is /tmp/lint-runtime-mutant.log.

## Setup and commands

Every test/build shell sources /workspace/adamic-tools/env.sh, the path setup wrote.
Go 1.27.1, clang 20.1.8, Node v24.19.0; nproc = 5, CPU quota = 4 CPUs.
Successful setup: go ready (0s), clang ready (0s), node ready (0s), submodules ready (0s),
build cache warm (79s), done in 79s. The first setup overlapped a branch switch and
failed cache warming on removed files; it was rerun successfully on the stable branch.

```sh
bash cloud/setup.sh > /tmp/adamic-setup.log 2>&1
bash cloud/setup.sh > /tmp/adamic-setup-retry.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
git fetch origin main codex/lint-runtime-fixes
git fetch origin codex/lint-runtime-fixes:refs/remotes/origin/codex/lint-runtime-fixes
git diff origin/main...origin/codex/lint-runtime-fixes
git log --format=fuller origin/main..origin/codex/lint-runtime-fixes
git show a24561a --
git switch -c coverage/lint-runtime-fixes origin/codex/lint-runtime-fixes
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/lint_runtime_' -count=1 -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/lint_runtime_search_boundaries.a$' -count=1 -timeout 30m
# The preceding single-fixture command ran under the mutant; all-five command ran again after restoration.
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
```

The all-five command was repeated while replacing unsupported multi-step optional
chains and mixed scalar containers with supported narrowing and union locals.
The final value audit added three-byte BMP text and separate empty concat headers;
counts and the all-five oracle were repeated after those additions. A generated-C
audit also added discarded identity returns to exercise repeated public releases
of the same shared values; the affected oracle, counts and CLI build were rerun. The first generic closure
identity call hit a specialization collision with an object (clang rejected the
pointer types), so its final version uses a typed closure helper. The full gate
also exposed a five-minute flow trace timeout on the local deep chain: the tracer
recursively snapshots its growing contents at each point. The chain is now global,
as in the existing million-link fixture, preserving its full depth. Flow and oracle
checks were rerun after this adjustment. Logs:
/tmp/lint-runtime-oracle.log, /tmp/lint-runtime-restored.log,
/tmp/lint-runtime-final-oracle.log, /tmp/lint-runtime-counts.log,
/tmp/lint-runtime-gofmt.log, /tmp/lint-runtime-vet.log, /tmp/lint-runtime-gate.log.
All test output went to files and was read afterward.

Every fixture was also explicitly built and run with the following loop;
stdout and stderr matched Node and all commands exited 0:

```sh
for file in internal/oracle/testdata/lint_runtime_*.a; do
    name=$(basename "$file" .a)
    go run ./cmd/adamic build "$file" -o "/tmp/lint-runtime-builds/$name"
    "/tmp/lint-runtime-builds/$name" > "/tmp/lint-runtime-builds/$name.native.stdout" 2> "/tmp/lint-runtime-builds/$name.native.stderr"
    node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" > "/tmp/lint-runtime-builds/$name.node.stdout" 2> "/tmp/lint-runtime-builds/$name.node.stderr"
    cmp "/tmp/lint-runtime-builds/$name.native.stdout" "/tmp/lint-runtime-builds/$name.node.stdout"
    cmp "/tmp/lint-runtime-builds/$name.native.stderr" "/tmp/lint-runtime-builds/$name.node.stderr"
done
```

Each build's diagnostics went to /tmp/lint-runtime-builds/$name.build.log.
Also ran `go run ./cmd/adamic c` on the equality, shared-release and chain programs,
writing /tmp/lint-runtime-{equal,shared,chain}.c and inspecting runtime calls.
Every unsupported probe was passed to `go run ./cmd/adamic build <probe.a> -o <probe>`;
all exited 1 with the expected checker/lowering refusals.

Additional flow verification commands (output read from the named logs):

```sh
go test ./internal/flow -run 'TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/lint_runtime_|TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/lint_runtime_|TestEveryMutationIsInItsRange/programs/../oracle/testdata/lint_runtime_' -count=1 -timeout 30m > /tmp/lint-runtime-flow.log 2>&1
go test ./internal/flow -count=1 -timeout 30m > /tmp/lint-runtime-flow-complete.log 2>&1
git diff --check
```

Final results: zero output differences, five added programs. All five pass the
uncached oracle, sanitizer/leak checks and explicit CLI builds/runs against Node.
Counts regeneration passes and changes only the five new rows. gofmt and go vet
pass. The final affected flow checks pass (4.608s), and the complete flow package
passes (87.728s).

The original complete uncached gate is partial: its process disappeared when the
execution environment reconnected, and there is no final exit status. Its log
records passes for native (312.662s), oracle (279.945s), Unicode (878.887s), the
bridge and all other completed core packages. Its earlier flow failures were the
new local chain's trace timeouts, resolved by the final global-chain version and
the complete flow rerun above. Stage-1 results were recorded through json; lint,
markdownblocks and later packages have no completed result in that log. This is
not a claim of a complete repository gate pass.
