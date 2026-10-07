# Proven guards and relations coverage

The coverage branch starts at origin/codex/proven-relations, c9f9d39, and merges
origin/codex/proven-type-guards, b95861f. These are sibling branches, so the merge
is necessary to test both features. Conflicts were additive: keep both sets of
count rows, and both implementation sections of docs/escape-hatches.md.

One Node/native disagreement was observed and kept only in this notes folder. Eleven new executable oracle fixtures
match source Node, generated JavaScript, sanitized native and release native.
Every fixture also builds with the CLI and runs with the same stdout, stderr and
exit code as Node. build-results.json records commands and complete observations.

## Inventory against existing oracle programs

The existing column records coverage before this work, including fixtures on both
requested branches. New filenames below omit the coverage_ prefix and .a suffix;
all eleven live under internal/oracle/testdata. Refusals live in this notes folder,
not the positive oracle. A proof rejection has no executable native output.

| Code case | Existing oracle program | New coverage or checked refusal |
|---|---|---|
| Named guard over a union of node kinds, true and false narrowing | proven_guards.a | guard_nodes, three kinds |
| Direct filter callback by declared name | proven_guards.a | guard_nodes, a union target |
| Inline expression arrow filter guard | proven_guards.a | guard_nodes |
| Function expression assigned to a variable, called and passed to filter | none with explicit predicate | guard_nodes |
| Class method guard over a parameter | none | guard_nodes |
| Negation of a trusted comparison | none inside guard body | guard_nodes |
| Parenthesized checks and literal on the left | none | guard_nodes, guard_literals |
| Logical OR checks with a union target | none | guard_nodes |
| Logical AND checks excluding two node kinds | none | guard_nodes |
| Block, empty statement, if with else, literal true and false returns | blocks/if without else in proven_guards.a | guard_nodes |
| Multiple true paths whose union equals the annotation | proven_guards.a, isAB | guard_nodes |
| typeof number versus string | proven_guards.a | assertion_forms additionally uses reverse comparison |
| typeof boolean in a mixed primitive union | none | guard_literals |
| String discriminant === and !== | proven_guards.a | guard_nodes |
| Numeric discriminant | none | guard_tags |
| Boolean discriminant | none | guard_tags |
| Numeric literal target, literal on the left | none | guard_literals |
| Boolean literal target | none | guard_literals |
| undefined equality/inequality guard | proven_guards.a | guard_literals |
| null equality guard with both outcomes | none | guard_match_null on RegExpExecArray or null |
| Nominal instanceof guard | proven_class_guards.a | existing fixture rerun |
| Imported class symbol aliases in instanceof | none | guard_imports and its helper |
| Imported guards, function aliases and guards held in fields | none | guard_imports and its helper |
| A call invalidates facts, then a new check proves the return | none | guard_imports |
| Guard failure by throw, with caught failure | none | guard_imports |
| A guard with no true returns, targeting never | none | guard_imports |
| Unknown parameter representation | none | unknown_guard.a, NotYet |
| Generic predicate with intersection target | none | generic_guard.a, refused |
| Assertions: implicit fallthrough and explicit bare return | proven_assertions.a | assertion_forms |
| Assertions: failure by throw, success and caught failure | proven_assertions.a | assertion_forms method |
| Assertions: terminal runtime panic and successful call | none | assertion_forms |
| asserts cond on a boolean, !cond branch | proven_assertions.a | existing fixture rerun |
| asserts cond using === true, else throw, caller narrowing | none | assertion_forms |
| Assertion method called through an explicitly typed instance | none | assertion_forms |
| Mutable discriminant changed to another variant before a guard | none | mutable_kind_guard.a, Node/native disagreement |
| Parameter writes, including captures | no positive program can validly use these | refusal probes |
| Invalidation by call, alias write, getter read | no positive program can retain stale facts | refusal probes |
| Reject opaque calls, aliases, computed/optional property checks | none | refusal probes |
| Reject unproven true/false paths and mismatched aggregate target | none | refusal probes |
| Reject no body, default/rest parameter, this predicate | none | refusal probes |
| Reject unverified control flow and over 256 paths | none | refusal probes |
| Reject empty/early/wrong assertion postconditions | none | refusal probes |
| Reject nonboolean asserts cond | none | refusal probes |
| Assertion returning a value / guard implicit fallthrough | none | TS2322 / TS2366 preempt verifier |
| satisfies primitive, operand evaluated once | proven_satisfies.a | existing fixture rerun |
| satisfies object literals with contextual literal types | proven_satisfies.a | relation_array_literals, guard_tags |
| satisfies array literals, preserve source type | proven_satisfies.a | relation_array_literals |
| Nested exact literals, including parenthesized origin | proven_satisfies.a without parentheses | relation_array_literals |
| Fresh array elements with optional fields absent/present | proven_satisfies.a absent-only | relation_array_literals |
| Fresh tuple with compatible optional fields present | none | relation_array_literals |
| Fresh tuple with optional field absent | none | exact_optional_tuple.a, conservatively refused |
| Array spread and readonly optional element view | none | relation_containers |
| Tuple viewed as readonly array | none | tuple_to_array.a, representation NotYet |
| Nested field, shorthand and spread with compatible fields | none in relation expressions | relation_views |
| Existing mutable slots, invariant element and field types | ordinary invariance fixtures | relation refusal probes for both relation operators |
| Readonly field becoming writable | ordinary invariance fixtures | relation refusal probe |
| Readonly array covariance | proven_satisfies.a, proven_upcasts.a | relation_containers, optional fields |
| Fresh mutable array upcast, then write wider element | none | relation_views |
| Map and Set readonly relations, identity and reads | none with these operators | relation_containers |
| Fresh Map satisfies identity type | none | relation_views |
| Mutable upcasts of slice, filter, map, concat, Array.from copies | none | relation_fresh_copies |
| Mutable upcasts of new Map and Set copies | none | relation_fresh_copies |
| Array.from on an iterable / new Array constructor | none | array_from_iterable.a / array_constructor.a, NotYet |
| Mixed primitive union elements in a filter guard | none | primitive_filter_guard.a, NotYet |
| Function parameter contravariance and result covariance | proven_upcasts.a | relation_views for satisfies and compatible optional results |
| Unsafe method parameter bivariance | no positive fixture can validly use it | relation refusal probes |
| Nominal class ancestry and interface view | proven_upcasts.a | existing fixture rerun |
| Unrelated nominal identity, literal satisfying a class | no positive fixture can validly use it | relation refusal probes |
| Generic class identical type and value identity | none | relation_views |
| Invariant class arguments with hidden optional field | none | relation refusal probe |
| Generic constraint in a relation | none | relation_views; optional_generic refusal |
| Source union and target union | proven_satisfies.a, same Shape union | relation_views, distinct compatible unions |
| Hidden optional field, direct and nested | no positive fixture can validly use it | relation refusal probes |
| Optional callback argument and return traversal | none | optional_callback_parameter/result refusal probes |
| Optional array, tuple, spread and shorthand traversal | none | optional_spread_array, optional_tuple, optional_spread, optional_shorthand refusals |
| Optional reverse compatibility of mutable elements/fields | none | relation refusal probes |
| as const and checked downcasts kept intact | existing casts.a and cast_fails.a | full gate |

The table covers observable obligations and value/call shapes, rather than claiming
all combinations of every type and syntax. Defensive nil checks, cycle bookkeeping,
missing malformed parameters and checker-internal impossible shapes are not source
language features. Nonassignable satisfies expressions are rejected by TypeScript;
unchecked downcasts continue through the existing cast path.

## Refusal programs

run_probes.py builds every .a listed in refusals.json using go run ./cmd/adamic
build with -o. It requires a nonzero exit and the listed diagnostic substring.
refusal-results.json retains complete diagnostics and exact commands for all 65
probes. Two diagnostics come from the TypeScript checker before Adamic's verifier:
returning boolean from an assertion is TS2322; implicit guard return is TS2366.
All other listed probes reach and name the expected Adamic proof obligation.

| Program | Expected diagnostic substring |
|---|---|
| predicate_unconditional_true.a | true return |
| predicate_different_variable.a | return expression |
| predicate_wider_narrowing.a | true return narrows to number, not 1 |
| predicate_partial_true_set.a | does not match |
| predicate_narrower_narrowing.a | false return |
| predicate_false_branch_lie.a | false return |
| predicate_positive_number.a | return expression |
| predicate_opaque_call.a | return expression |
| predicate_modified_parameter.a | parameter is assigned |
| predicate_captured_write.a | parameter is assigned |
| predicate_aliased_discriminant_call.a | true return |
| predicate_aliased_discriminant_write.a | true return |
| predicate_aliased_discriminant_getter.a | true return |
| predicate_assertion_wider_shape.a | normal return |
| predicate_nominal_guard_mismatch.a | true return |
| predicate_nominal_assertion_mismatch.a | normal return |
| predicate_assertion_empty.a | normal return |
| predicate_assertion_early_return.a | normal return |
| predicate_assertion_wrong_branch.a | normal return |
| predicate_assertion_call_invalidation.a | normal return |
| predicate_assert_cond_empty.a | normal return |
| predicate_assert_cond_false.a | normal return |
| predicate_assert_cond_nonboolean.a | boolean parameter |
| predicate_bodyless_signature.a | no body |
| predicate_default_argument.a | unchanged plain parameter |
| predicate_unknown_control_flow.a | return paths |
| relation_satisfies_mutable_array.a | can write Animal where Dog is read |
| relation_upcast_mutable_array.a | can write Animal where Dog is read |
| relation_satisfies_mutable_field.a | can write Animal where Dog is read |
| relation_satisfies_nested_mutable_literal.a | can write Animal where Dog is read |
| relation_satisfies_readonly_to_writable.a | readonly field pet becomes writable |
| relation_satisfies_function_parameter.a | function taking Dog |
| relation_upcast_function_parameter.a | function taking Dog |
| relation_upcast_nominal_identity.a | nominal ancestry for A |
| relation_satisfies_nominal_identity.a | nominal ancestry for A |
| relation_satisfies_hidden_optional.a | optional field count |
| relation_upcast_hidden_optional.a | optional field count |
| relation_upcast_mutable_optional_elements.a | optional field element.count |
| relation_satisfies_mutable_optional_field.a | optional field pet.count |
| relation_upcast_invariant_optional_class_argument.a | optional field type argument.count |
| relation_satisfies_nested_optional.a | optional field element.count |
| guard_delegation.a | return expression |
| guard_computed.a | return expression |
| guard_optional_chain.a | return expression |
| guard_alias.a | return expression |
| guard_rest.a | rest parameter |
| guard_this.a | no body |
| assertion_value.a | TS2322 |
| guard_implicit.a | TS2366 |
| guard_branch_call.a | branch is not a trusted |
| optional_callback_result.a | optional field count |
| optional_callback_parameter.a | optional field parameter.count |
| optional_spread.a | optional field count |
| optional_shorthand.a | optional field x.count |
| optional_spread_array.a | optional field element.count |
| optional_tuple.a | optional field element.count |
| optional_generic.a | optional field count |
| guard_path_limit.a | too many return paths |
| guard_switch.a | return paths |
| guard_for.a | return paths |
| guard_try.a | return paths |
| guard_nested_property.a | return expression |
| guard_literal_alias.a | return expression |
| guard_negative_literal.a | return expression |
| generic_guard.a | true return narrows to T, not T & Leaf |

## Cases not kept as oracle programs

All eight reproducers have Node observations and CLI diagnostics in
unsupported-results.json. No native binary exists for these, so they are compiler
limitations rather than output differences.

- nullable_primitives.a: number or null has no stage-0 representation. A nullable
  regex match can be represented, so guard_match_null covers literal null checks.
- tuple_to_array.a: tuples are object-backed, and viewing one as an array is NotYet.
- exact_optional_tuple.a: the optional relation walk compares tuple element types
  without propagating their fresh literal origins. It conservatively refuses an
  omitted optional field even in a freshly written tuple. The positive tuple gives
  each optional field an explicitly compatible value instead.

- unknown_guard.a: the direct typeof proof is accepted, but unknown cannot be
  represented as a native parameter.
- generic_guard.a: the checker flow type for the parameter remains T, so the body
  cannot prove the annotation T & Leaf. The same check on a concrete node union
  is covered by the accepted programs.

- array_from_iterable.a: Array.from lowering supports only the { length } and
  callback form. The positive fresh-copy fixture uses that form.
- array_constructor.a: new Array<T> is recognized as fresh by the relation but
  its value construction is NotYet.
- primitive_filter_guard.a: string or number array elements cannot be represented
  yet. Direct guards over the same primitive union are covered by existing and
  new programs; filter guards over represented node unions are covered too.

Bodyless predicate type annotations cannot become callable positive programs by
merely declaring an alias: the verifier requires a body. this/default/rest guards,
computed/optional property checks and opaque predicate delegation remain refused.
Loops, switches and try paths are not verified. These are refusal coverage, not
missing runtime validators or claimed new language support.

## Mutation

At internal/lower/predicates.go:249, replace the trusted &&/|| combination
`return p.check(binary.Left) && p.check(binary.Right)` with `return false`.
coverage_guard_nodes.a fails before native compilation, naming an untrusted return
expression at line 6. mutant.log records the exact test and output. Restore the
line and the same fixture passes. No compiler changes remain. This proves the new
program depends on branch code; it is not a mutation caught by clang warnings.
The first mutation attempt used an unconfigured system go and did not run a test;
the configured-toolchain attempt is the evidence retained in mutant.log.

## Commands and setup

From /workspace/adamic, run bash cloud/setup.sh, then source
/workspace/adamic-tools/env.sh. nproc printed 5. The successful setup timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (80s)
setup: done in 80s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

The initial setup overlapped branch checkout and failed cache warming on a
transient undefined provenRelation helper. The completed-source retry above passed.
/opt/adamic-tools/env.sh was absent; the setup-selected path above is used.

Validation commands (all test output captured to logs and then read):

```sh
source /workspace/adamic-tools/env.sh
python3 notes/proven-guards-relations/run_probes.py > /tmp/proven-refusals.log 2>&1
python3 notes/proven-guards-relations/run_builds.py > /tmp/proven-builds.log 2>&1
python3 notes/proven-guards-relations/run_unsupported.py > /tmp/proven-unsupported.log 2>&1
python3 notes/proven-guards-relations/run_disagreement.py > /tmp/proven-disagreement.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(coverage_|proven_)' -count=1 -timeout 30m -v > /tmp/proven-final-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/proven-counts-update.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m > /tmp/proven-counts-check.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_guard_nodes[.]a' -count=1 -timeout 30m -v
# Previous command run once with mutation (fails), then restored (passes).
gofmt -l cmd internal > /tmp/proven-format.log
go vet ./... > /tmp/proven-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/proven-full-gate.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/proven-final-oracle-all.log 2>&1
git diff --check
```

Before the final oracle run, the coverage-only oracle was run three times while
correcting console argument types and moving unsupported shapes to notes. The
first two failed at load/lowering, the third passed all eight. The added import pair passed independently. The final 16-fixture
run includes the five existing feature programs and uses the finished sources.
The mutable-kind disagreement remains unchanged in notes; no output differences were hidden by rewriting a program.

The fresh-copy fixture was initially blocked by an inferred never[] concat
argument, iterable Array.from and new Array construction. concat now has a
represented element; the unsupported constructor forms remain separate .a
reproducers. This fixture independently passed after those compile-time limits
were isolated. The full repository gate started with the first eight fixtures;
the entire oracle is rerun on the final eleven and regenerated counts separately.

## Final observed results

- All 11 CLI builds and executions match Node: build-results.json.
- All 65 refusal diagnostic checks pass: refusal-results.json.
- All 8 unsupported-shape reproducers fail at the expected compile-time limit;
  Node runs each successfully: unsupported-results.json.
- The final 16-fixture uncached feature oracle passes: oracle.log.
- The entire final uncached oracle passes in 199.894s: oracle-all.log. It includes
  allocation counts, sanitizer comparisons, release builds, leak checks and the
  input/stream checks on the final eleven new fixtures plus the existing corpus.
- Count regeneration passes in 29.078s: counts-update.log. Only new rows and
  registration order changed; no existing recorded number changed.
- The verifier mutation fails, restoration passes: mutant.log and
  mutant-restored.log. internal/lower has no remaining diff.
- Formatting, vet and git diff --check pass with empty logs.

The broad uncached repository gate completed with one failure: the unrelated
TestMarkdownUnicodeWidths could not import its absent pinned npm scratch
dependencies from /tmp/adamic-markdown-width. Every other package passed. After
installing the three documented pinned dependencies, the exact failing test was
rerun uncached and passed, including its mutation checks. The original broad
run is preserved as full-gate.log; the successful retry is width-recheck.log.

```sh
npm install --prefix /tmp/adamic-markdown-width --ignore-scripts --no-audit --no-fund emoji-regex@10.6.0 get-east-asian-width@1.6.0 narrow-emojis@0.0.3 > /tmp/proven-width-install.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/markdownblocks -run '^TestMarkdownUnicodeWidths$' -count=1 -timeout 30m -v > /tmp/proven-width-recheck.log 2>&1
```

## Observed disagreement

mutable_kind_guard.a changes a mutable kind from leaf to branch, without giving
that object a branch label. The pure user-defined guard is accepted and selects
that new kind. Node and the JavaScript backend print `branch undefined\n` and exit
0 with empty stderr. Sanitized and release native both print no stdout and exit
70 with `adamic: panic: compiler bug: a field the checker proved is there is missing\n`.

Likely root cause: internal/lower/class.go:446 emits the discriminant field store
without proving that the whole object is now the other variant. The newly admitted
guard at internal/lower/predicates.go:160 trusts the checker narrowing, which assumes
that complete-variant invariant. This is an inference about responsibility, not a
claim that the branch introduced the underlying mutable-tag write hole.

run_disagreement.py temporarily registers the notes program with the actual oracle,
requires its observed disagreement, removes registration, then repeats the CLI
build and all three executions. disagreements.json contains exact commands and
outputs; mutable-kind-oracle.log contains the sanitizer and release comparison.
The program is never added to oracle/testdata or its count table. The eleven
retained positive fixtures still agree; the entire final oracle passed with only
those positive registrations.
