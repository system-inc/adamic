1 refusal still lets miscompiles through (2 witnesses); 4 over-refused programs across 3 guards; 13 safe fixtures and 38 refusal witnesses added on coverage/class-refusals, based on ad1571edd147e59cddfbaa9464c0ad6ffd4f852d.

Counts mean distinct guards for the first number and distinct programs for over-refusals. The private-static guard has no new finding in this matrix. Compiler sources are restored to the requested main revision. Findings are review probes, not oracle fixtures.

## Findings: observations

Every program is linked below. Native release means the actual `go run ./cmd/adamic build` entry point, using the compiler's default clang `-O2`. Sanitized means that same entry point with `--sanitize`, enabling ASan and UBSan together, with LeakSanitizer at exit. Successful observations here exit 0 with empty stderr.

| Program | Node stdout | Native release stdout | Sanitized native stdout | JavaScript backend stdout | Responsible code |
|---|---|---|---|---|---|
| [Conditional receiver, overridden next](iterator_return_conditional.a) | `0,100,200\n` | `0,1,2\n` | `0,1,2\n` | `0,1,2\n` | `internal/lower/iteration_origin.go:400`, called by `iteration.go:218` |
| [Conditional receiver, hidden close](iterator_conditional_close.a) | `0\nclosing:return\nclosing:return\n0\n` | `0\n0\n` | `0\n0\n` | `0\n0\n` | `internal/lower/iteration_origin.go:400`, called by `iteration.go:218` |

The conditional-return change is `return this.index >= 0 ? this : this;` in the base `[Symbol.iterator]()` method. Both arms return the receiver. Neither native build reported a sanitizer error in these two programs.

For the following four programs, **both main native builds refuse**, exit 1, before execution. The JavaScript backend refuses identically. The counterfactual is main with only the identified guard reverted. All three counterfactual backends match Node, exit 0, and emit no stderr.

| Program | Node stdout | Main refusal | Counterfactual release, sanitized and JavaScript stdout | Responsible code |
|---|---|---|---|---|
| [Unused arrow in a base method](super_unused_arrow.a) | `safe\n` | `a field initializer reading an uninitialized field through Base.describe -> Derived.score -> Derived.count` | `safe\n` | `internal/lower/class_inheritance.go:617`, following the method at line 601 |
| [Unreachable branch in a base method](super_dead_branch.a) | `safe\n` | Same initializer refusal | `safe\n` | `internal/lower/class_inheritance.go:617`, following the method at line 601 |
| [Only a base instance passed to an iterator helper](iterator_base_parameter.a) | `0,1,2\n` | `adamic/iterator-receiver-origin` | `0,1,2\n` | `internal/lower/iteration.go:224`, with descendant scan at `iteration_origin.go:439` |
| [Keys of a separate plain object](keys_unrelated.a) | `value\n1,2\n` | `adamic/symbol-key-view` | `value\n1,2\n` | `internal/lower/class_features.go:64` |

Full refusal text and native build statuses are in [observations.json](observations.json). Full counterfactual observations are in [counterfactuals.json](counterfactuals.json).

## Inferences from the code

The iterator receiver recognizer handles a direct `this` and variable aliases, then returns false for a conditional expression. That prevents the new guard from running, so protocol calls use the base implementations.

The initializer walker visits every child, including an arrow body that is never called and a branch that is unreachable. Both over-refusal programs finish the initializer without reading the future field.

The iterator descendant scan sees a protocol-changing subclass even when no instance of it is created. The origin proof cannot trace the helper's parameter to its only call site's exact base instance.

The keys guard compares shapes throughout all modules, without tracing the argument's origin. The unrelated iterable has a compatible `value` field, so it blocks keys of a separate plain object.

These are explanations of the observed probes, not claims about every possible program.

## Coverage and limits

The 13 safe fixtures cover no override; fields declared before a super method or getter uses them; an inherited field available before its derived redeclaration; exact base and unchanged subclass iterators; a closed iterator class tree used through a parameter; a factory returning a separate base iterator; plain object keys, an explicit copy and a class descriptor; private instance delegation; and private static storage. Every fixture matched source Node, actual CLI release and sanitized builds, and the JavaScript backend. The shared oracle also checks their leaks and counts.

The 38 new witnesses cover parentheses, an immediately invoked arrow, getters, deeper methods/getters, generic forwarding, interfaces, unions, optional access and calls, for-of, spread, Array.from, destructuring, aliases, added nesting, private writes and static getters. Of these, 35 hit the new named refusals. Three hit existing NotYet diagnostics: a structural iterator signature, a nested function declaration in the keys probe, and an optional Object.keys call. Those tests explicitly expect NotYet, rather than attributing them to a new guard. The original eight Node-backed probes remain covered by TestClassWrongOutput103/106/107/108.

Other exploratory observations, not counted as findings:

- [Deferred super arrow](super_deferred.a): Node prints `score 6\n`. Main refuses. Reverting the initializer guard produces invalid C in both native builds (`adamic_local_5_this` undeclared); the backend panics with `ReferenceError: local_5_this is not defined`, exit 70. It has not been shown to compile correctly, so it is not counted as an over-refusal.
- [Super argument read](super_ignored_argument.a): the checker rejects it with TS2729 before lowering. It does not demonstrate a guard bypass.
- [Arrow returning the receiver](iterator_return_arrow.a): main gives NotYet, `a function returning this`.
- [Private identifier in an optional chain](private_optional_checker_rejected.a): the checker gives TS18030. The valid optional-holder probe is a separate refused witness. Private identifiers cannot be exposed by an interface alone; the interface witness instead carries a Box instance.

The initial scratch JavaScript runs used Node without the oracle runtime resolver and failed import resolution. Those observations were discarded. Every accepted probe's JavaScript leg was rerun through `oracle/node.mjs`; those corrected outcomes are the ones recorded in observations.json.

## Mutation proof

Each mutation was applied independently and restored in a `finally` block. No final compiler-source diff remains.

| Guard reverted | Exact mutation | Original test failure | Actual original program result with mutation |
|---|---|---|---|
| Initializer | Restore `initializerReads` from 5348895b; remove now-unused strings import | All 3 subtests of TestClassWrongOutput103 fail: wanted refusal, got nil | Number and getter: Node `score NaN\n`, all backends `score 1\n`. String: Node `caught TypeError\n`; release SIGSEGV; sanitized UBSan null string member access, exit 1; backend catches TypeError |
| Iterator receiver | Make the new condition in `iteration.go:218` false | Both TestClassWrongOutput107 subtests fail: wanted refusal, got nil | Override probe uses base numbers and base:return in all backends; hidden-close probe drops both close messages |
| Structural keys | Make the new condition in `class_features.go:54` false | TestClassWrongOutput108 fails: wanted refusal, got nil | All backends expose `__adamic_symbol_iterator` in the first keys result; Node does not |
| Private static access | Make the new target condition in `class_static.go:497` false | Field subtest of TestClassWrongOutput106 gets nil; method subtest panics in lowering | Field: both native builds panic about a missing field, exit 70; backend prints `undefined\ns1\nhidden s1\n` instead of Node's `s1\nt2\nhidden t2\n`. Method: all compiler entry points panic, `ir: virtual call has no target set` |

The new TestClassRefusalsSyntaxCoverage was also run under each mutation, filtered to the corresponding family. All four runs exit 1. The optional-super witness falls back to the existing optional-call NotYet under the old initializer analysis; its test still fails because the targeted initializer refusal is gone. Other existing NotYet exclusions are not claimed as killed by a new guard. [Evidence logs](evidence/) retain all original and new witness failures. The successful over-refusal counterfactuals were measured with the same independent mutations; the initializer counterfactual was additionally run on the unused-arrow and dead-branch programs.

## Exactly what was run

From `/workspace/adamic`, read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the complete `git diff 5348895b ad1571ed`, all six commit messages in `5348895b..ad1571ed`, and the complete current five touched compiler-source files. The initial checkout did not contain those commits, so `git fetch origin` preceded `git switch -c coverage/class-refusals ad1571ed`.

Toolchain setup:

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
```

Setup completed in 315.390 seconds; nproc is 5, cgroup quota is 4 CPUs. Go 1.27.1, clang 20.1.8, Node v24.19.0. Every subsequent compiler and test shell sourced that env file. The cloud-environment runtime skill was used to inspect setup and networking context.

Scratch probe work, with exact scripts preserved under evidence/:

```sh
python3 /tmp/class-refusals/generate.py
go build -o /tmp/class-refusals/adamic ./cmd/adamic
python3 /tmp/class-refusals/run.py > /tmp/class-refusals/matrix.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run TestClassWrongOutput -count=1 -timeout 30m > /tmp/class-refusals/original.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 /tmp/class-refusals/mutate.py > /tmp/class-refusals/mutants.log 2>&1
python3 /tmp/class-refusals/supplement.py
python3 /tmp/class-refusals/run-supplement.py > /tmp/class-refusals/supplement.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 /tmp/class-refusals/mutate-super-boundaries.py > /tmp/class-refusals/super-boundaries.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 /tmp/class-refusals/mutate-coverage.py > /tmp/class-refusals/coverage-mutants.log 2>&1
```

For each of 61 scratch cases, the runners ran source Node and compiler `js`. For each accepted case they ran both native build commands below and executed each produced binary. Generated JavaScript was executed through the source runner after correcting its initial runtime resolution. The first 53 cases and 8 supplemental cases are all recorded, with exact sources, in observations.json.

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs <probe.a>
/tmp/class-refusals/adamic js <probe.a>
go run ./cmd/adamic build <probe.a> -o <binary>
go run ./cmd/adamic build <probe.a> -o <binary> --sanitize
<binary>
node --disable-warning=ExperimentalWarning oracle/node.mjs <generated.mjs>
```

Mutation runners used `go run ./cmd/adamic js` and the same two build forms for the original probes and over-refusal boundaries. Each original mutation ran `go test ./internal/oracle -run '^TestClassWrongOutput10N$' -count=1 -timeout 3m` for its named test (103, 107, 108 or 106). The new-witness mutations ran `go test ./internal/lower -run 'TestClassRefusalsSyntaxCoverage/FAMILY' -count=1 -timeout 3m` for super, iterator, keys and private. Every test wrote to a log that was then read.

Both main native build forms were additionally invoked against all four review over-refusal programs to record their exact refusals. A scratch inline Python command initially had an indentation error and ran no builds; it was corrected and rerun. A formatting invocation initially lacked the setup env and could not find gofmt; the sourced invocation below succeeded.

Final validation commands:

```sh
gofmt -w internal/lower/class_refusals_coverage_test.go internal/oracle/class_refusals_coverage_test.go
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestClassRefusals|TestClassWrongOutput|TestNativeAgreesWithNode/internal/oracle/testdata/class_refusals_safe_' -count=1 -timeout 30m > /tmp/class-refusals/focus.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/class-refusals/counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -count=1 -timeout 30m > /tmp/class-refusals/packages.log 2>&1
go vet ./... > /tmp/class-refusals/vet.log 2>&1
gofmt -l cmd internal > /tmp/class-refusals/gofmt.log
git diff --check > /tmp/class-refusals/diff-check.log
git diff --cached --check
```

Focused tests passed: lower 0.991s, oracle 4.208s. Counts update passed in 18.474s, adding only the 13 new rows. Complete uncached packages passed: lower 36.326s, oracle 133.479s. `go vet ./...` passed with no diagnostics; gofmt and whitespace checks printed nothing. These complete package runs include the recorded-count check, source and backend Node, native release, ASan, UBSan and leak checks. The staged whitespace check additionally caught a trailing blank line in a review probe. It was removed, and the staged check then passed. The full repository test gate was not run.
