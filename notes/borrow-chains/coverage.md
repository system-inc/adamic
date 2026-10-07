# Borrow chain coverage

Base: origin/codex/borrow-chains at 395b8af. No compiler changes are retained.
New programs: borrow_chain_coverage_*.a in internal/oracle/testdata.
All source strings used as lifetime witnesses are constructed at runtime.

The inventory below distinguishes existing source coverage from additional focused
coverage. Existing means a program in testdata at 395b8af, not merely a Go IR test.
Names without a prefix refer to existing programs. Names in bold refer to the new
borrow_chain_coverage_<name>.a programs. Oracle registration is explicit in
oracle_test.go; putting a file in testdata alone does not run it.

| Code case | Existing program | Added focused program |
|---|---|---|
| Strong Property rooted at Read, borrowed parameter | borrow_chain_argument, borrow_chain_walk | **depths** |
| Stable owned local root | class_oct6_subclass_holder | **depths**, **return**, **spread** |
| Two and three fields in one expression, intermediate borrowed locals | borrow_chain_argument has two; borrow_chain_walk splits ancestors into locals | **depths** |
| Defined through narrowed optional receiver | borrow_defined_lent_field, literal_optional_shapes | **exclusions**, **parents** |
| Optional receiver refuses chain provenance | optional_class_method, weak_parent | **exclusions** |
| Weak receiver/read ends chain | weak_parent, weak_narrowed | Existing coverage retained |
| Method property is not a field chain | borrow_chain_listener, class_as_interface | **method** |
| Non-Property declaration/expression, call-result root | borrow_element, class_oct6_subclass_holder | **return** reads returned objects too |
| No current function/top-level expression | borrow_chain_walk builds globals | **values** reads a global at top level |
| Global root refuses | borrow_defined_lent_field | **exclusions** |
| Captured root refuses | borrow_defined_lent_field | **exclusions**; **callbacks** has a writing captured root |
| Different-function root | closures, borrow_defined_lent_field | **exclusions** returns a closure reading its outer parameter |
| Assigned root refuses, including enclosing blocks | borrow_chain_reassigned, param_assigned_in_try | **parents**, **scope** |
| Captured destination refuses | borrow_chain_capture | Existing coverage retained |
| Assigned destination refuses | self_assignments (general field reads) | **exclusions** |
| Closure function excluded from declaration planner | borrow_chain_capture, closures | **exclusions**, **bounded** |
| Nonreference destination is not borrowed | borrow_chain_walk numeric total | **values** number and boolean fields |
| String reference destination | borrow_chain_argument, borrow_chain_walk | **depths**, **values** |
| Object reference destination | borrow_chain_walk | **depths**, **return** |
| Array reference destination | borrow_chain_walk | **values**, **parents** |
| Map reference destination | collections (general fields) | **values** |
| Closure reference destination | closures (general fields) | **values**, including identity without an unknown call |
| Missing optional reference (NULL) | literal_optional_shapes | **values**, **exclusions** |
| Weak handle value/target | weak_parent, weak_narrowed | Existing coverage retained |
| Boxed Union destination excluded from lendable | unions uses union locals | Cannot lower a string-or-number field; see limits |
| Matching SetProperty, direct | borrow_chain_write | **scope** |
| Matching SetProperty through direct method | class_oct6_subclass_holder changes a held child | **method** |
| Matching SetProperty through virtual CallTargets | borrow_chain_override, borrow_chain_store | Existing coverage retained |
| Name matches unrelated holder too | No focused chain program found | **guards** |
| Write before declaration also refuses (whole function) | No focused chain program found | **guards** |
| Write in later iteration/finally | borrow_element_throw has array invalidation in catch | **scope**, **finally** |
| Named call, every target read-only | borrow_chain_argument | **depths**, **parents** |
| Recursive named targets checked once, entire body | borrow_chain_walk | **parents** also iterates ancestors |
| Structural method/function call, unknown targets refuse | borrow_chain_listener, borrow_chain_unknown | **values** invokes a function field outside its positive inspector |
| Known CallClosure targets | call_targets_closure (general borrowing) | **bounded** |
| ArrayMap targets | visits (general callbacks) | **bounded** |
| ArrayVisit targets | visits (general callbacks) | **callbacks** |
| ArrayReduce targets | visits (general callbacks) | **callbacks** |
| ArrayFrom targets | array_from_length (general callbacks) | **callbacks** uses supported { length } form |
| MapForEach targets | reuse_foreach_global (general mutation) | **callbacks** |
| ArraySort targets | call_targets_sort (general borrowing) | **callbacks** |
| Unknown closure targets refuse | borrow_chain_unknown | Existing coverage retained |
| ObjectLiteral without spread allowed | borrow_chain_walk | **depths**, **return** |
| ObjectLiteral spread refuses (reuse may remove owner) | reuse_spread_method, reuse_weak_during_spread | **spread** |
| ArrayLiteral allowed | borrow_chain_walk | **callbacks**, **return** |
| ArrayPush allowed, stored value needs own count | borrow_chain_walk constructs arrays | **return** pushes a chain into a returned array |
| MakeClosure allowed, captured values still retained | borrow_chain_capture | **bounded**, **callbacks** |
| MakeError allowed, operands walked | borrow_element_throw | **finally** |
| Other impure/runtime operation refuses | borrow_map_overwrite, borrow_element | **guards** uses Map.set |
| Root consumed or moved: emission refuses; planner marks lending | borrow_chain_store, borrow_element_super_move | **spread** passes a local root to an owned-return helper before the child is used |
| Root reused by object spread | reuse_spread_method_alias | **spread** exercises object copying around a held child |
| Root reused by array plan | reuse_arrays, borrow_element | A plain native array has no user-defined strong fields; root-array guard has no distinct source-level strong-field witness |
| Kept/stored result owns a count | borrow_chain_store | **return**, **spread** |
| Owned return of chained string or object | class_oct6_release (method-return lifetime) | **return** |
| Closure capture of chained value survives holder scope | borrow_chain_capture | Existing coverage retained |
| Throw/exceptional cleanup with chain alive | No focused field-chain program found | **finally**, **scope** |
| Loop walking children recursively | borrow_chain_walk | **parents** walks indexed children |
| Loop walking parents iteratively | weak_parent uses recursive weak ancestors | **parents** walks strong ancestors with reassigned cursor |

The unchanged pureKind and consumes lists are inherited, not new cases introduced
by this branch. Their existing source witnesses include lent_reads (reads, lengths,
indexing, maps and snapshot order), borrow_defined_lent and
borrow_defined_lent_field (pass-through narrowing), casts and cast_fails (checked
casts), unions (boxes and type tests), optional_strings and optional_numbers
(coalescing), strings/string_positions/searches (string consumers),
library_math_number_* (numeric consumers), objects/has_own (field consumers), and
class_inheritance (instance tests). Every operand of allowed operations is still
walked. The added guards/changes probes put mutations among those operations.

## Limits observed

A string | number field read failed lowering at values.a:8:16:
"stage 0 can't lower a field of type string | number yet". The attempted field
was removed; scalar number/boolean and optional-reference fields remain.
This is a compile-time limit, not an output disagreement.

Throwing the chained value itself cannot be an oracle fixture: exceptions.go only
accepts new Error at the throw site or rethrow of the surrounding catch binding.
A field holding Error is NotYet; string/object fields are refused. The finally
program instead throws new Error with a chained message and covers cleanup and
continued use across catch/finally. No claim of direct field-value throw coverage.

Array.from([1, 2], callback) is NotYet (only { length } is lowered). The callbacks
program uses Array.from({ length: 2 }, callback), covering the branch's ArrayFrom
case without requiring a compiler change.

The consumed/moves/spreads/arrays checks inspect internal reuse plans. A source
program can exercise transfer/reuse and observable lifetime, but cannot request
one particular internal plan independently of the others. No manufactured IR
or compiler changes were added to force unreachable combinations.

## Mutation

Changed exactly one line in internal/native/borrow.go, then restored it:

    if write, ok := statement.(ir.SetProperty); ok && names[write.Name] && false {

The uncached oracle for borrow_chain_coverage_method.a exited 1, reporting ASan
heap-use-after-free in inspect when reading the saved leaf after Root.replace.
Node printed "method1 new2\n"; the sanitized native run printed nothing and
exited 1. This is a deliberately mutated compiler failure, not a disagreement
on the requested branch. The final compiler file has no diff from 395b8af.

## Commands run

Every build/test shell sourced /workspace/adamic-tools/env.sh, the setup's actual
installation path. Setup printed Go ready 0s, clang ready 0s, Node ready 1s,
submodules ready 1s, build cache warm 109s, done 109s. nproc printed 5.

Repository inspection: read CLAUDE.md, README.md, docs/0.1.md and docs/memory.md;
fetched both branch refs, read their diff, the three branch commit messages, the
changed implementation/test/report/tool files, the oracle harness and counts
harness, and searched existing .a programs. The initial main ref was stale and
was refreshed before using the branch diff as the coverage inventory.

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
git fetch origin codex/borrow-chains:refs/remotes/origin/codex/borrow-chains
git fetch origin main:refs/remotes/origin/main
git diff origin/main...origin/codex/borrow-chains
git log --format=fuller origin/main..origin/codex/borrow-chains
git switch -c coverage/borrow-chains origin/codex/borrow-chains

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/borrow_chain_coverage_' -count=1 -timeout 30m > /tmp/borrow-coverage-oracle-final.log 2>&1
for file in internal/oracle/testdata/borrow_chain_coverage_*.a; do
  out=/tmp/$(basename "$file" .a)
  printf '%s\n' "$file"
  go run ./cmd/adamic build "$file" -o "$out" && "$out" || exit
done > /tmp/borrow-coverage-builds-final.log 2>&1

go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/borrow-coverage-counts-complete.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/borrow-coverage-gate-final.log 2>&1
go vet ./... > /tmp/borrow-coverage-vet.log 2>&1
gofmt -l cmd internal > /tmp/borrow-coverage-gofmt.log
git diff --check

go run ./cmd/adamic c internal/oracle/testdata/borrow_chain_coverage_values.a > /tmp/borrow-coverage-values.c
go run ./cmd/adamic c internal/oracle/testdata/borrow_chain_coverage_depths.a > /tmp/borrow-coverage-depths.c
go run ./cmd/adamic build /tmp/borrow-coverage-throw.a -o /tmp/borrow-coverage-throw > /tmp/borrow-coverage-throw.log 2>&1
go run ./cmd/adamic build /tmp/borrow-coverage-union.a -o /tmp/borrow-coverage-union > /tmp/borrow-coverage-union.log 2>&1
```

The mutation was run by an inline Python driver with exactly the change and Go
command in mutant.py, writing /tmp/borrow-coverage-mutant.log. Its full output is
retained in mutant.txt. To reproduce after sourcing the toolchain:

```sh
python3 notes/borrow-chains/mutant.py
```

The focused oracle and standalone loop were repeated while writing the programs:
initial mixed union field failed lowering; then Array.from(array, callback) failed
lowering; after fixes they passed. Intermediate logs used
/tmp/borrow-coverage-oracle.log and /tmp/borrow-coverage-builds.log. Counts updates
also ran to /tmp/borrow-coverage-counts.log and
/tmp/borrow-coverage-counts-final.log; the final complete update has all 12 rows.
An earlier full gate to /tmp/borrow-coverage-gate.log was terminated with SIGTERM
when the positive callback fixtures were refined; it is superseded by the final
gate, not claimed as a pass. The union CLI diagnostic probe was first rejected by
the checker's console argument type; using a template made it type-correct and
confirmed the lowering limit. Neither unsupported temporary program was committed.

Generated C verifies that list, map and callback declarations in values borrow
without retains or scope releases. depths records zero retains overall.

## Intermediate flow-range finding

The full gate caught an additional, non-oracle issue in the first writing-callback
probe. Node, native and JavaScript all agreed, but the flow range verifier reported:

    function 1 (changes): child$2 was mutated at instruction 2 (order 3), outside its range [1, 2)

The intermediate probe was:

```ts
function changes(root: Root): void {
 const child = root.child;
 const text = child.text;
 [1].forEach((n: number): void => { root.child.text = `changed${n}`; });
 console.log(text);
}
```

The final probe saves root.child.text directly, preserving the two-field read,
writing callback and old-string observation without the redundant object alias.
No flow/compiler fix is included. The redundant alias form remains a known gap
in flow's range inference; the branch's native borrow guard correctly kept it
alive. This finding must not be interpreted as a backend output disagreement.

A focused range-test invocation on the simplified probe alone failed the
harness's "no mutation was seen in any program" assertion: it tracks no mutable
object local in that form. The whole flow package was therefore rerun, rather
than accepting an empty mutation run. Exact follow-up commands:

```sh
go test ./internal/flow -run 'TestEveryMutationIsInItsRange/programs/../oracle/testdata/borrow_chain_coverage_callbacks.a' -count=1 -timeout 30m > /tmp/borrow-coverage-flow-focused.log 2>&1
go test ./internal/flow -count=1 -timeout 30m > /tmp/borrow-coverage-flow-final.log 2>&1
```

Counts, the focused uncached oracle and every standalone build/run were repeated
with the final direct-string version, using the commands above and the same
complete/final log paths.

## Final results

Zero backend disagreements. Twelve programs added, each registered in the oracle
and counts table. All final standalone builds/runs succeeded (builds.txt).
Final focused uncached oracle PASS 6.417s (oracle.txt). Final complete count update
PASS 52.653s (counts.txt). Complete flow package after simplifying the callback
PASS 137.588s (flow.txt). Vet, gofmt listing and git diff --check were clean.
The one-line field-write mutant was caught by ASan, and the compiler was restored.

The broad uncached repository gate is incomplete, not a claimed full pass. It
reported complete native PASS 239.371s and complete oracle PASS 213.352s, along
with all completed non-stage1 packages except the intermediate alias form's flow
failure. After the final flow rerun passed, the remaining stage1 CSS/JSON/lint
processes were stopped with SIGTERM to the Go test driver (exit 143).
The gate invocation had already failed on the intermediate fixture, so it could
not supply a clean final full-gate result. Its captured output is gate.txt. No
claim is made that every stage1 corpus package finished. Exact termination:

```sh
kill -TERM 20225
```

The final focused oracle and count update ran after all fixture edits. No compiler
source or pre-existing count row differs from the feature branch base. The known
flow alias gap above is documented rather than repaired as part of this coverage
change.
