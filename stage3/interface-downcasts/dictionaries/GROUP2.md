Built shared dictionary source read hooks; seven source controls now run generated C and JavaScript against Node.
Commits: this checkpoint on codex/views-dictionaries, parent e9a0682932f89442a4c06ff36a593ed019cfc8d0; integration remains ba59427c.
Validation: source oracles, focused dictionary/record/lazy/array regressions, vet, ranking and whitespace checks pass.
Mutants: all three requested source mutations caught in each backend; the eleven prior component/guard mutations also pass.
Not covered: richer CompilerOptions elements and dynamic record production/writes; 29 candidate pairs / 228 reads remain pending.

Revised whole-family working date: October 17, 2026, 23:00 UTC. The October 19
conditional handoff estimate is superseded: shared hooks are now this lane's work,
with no waiting on owner permission. This is an estimate for per-pair evidence,
not a claim that whole-program tsc compilation is available. Candidate counts are
accepted as the working inventory; exact reachability remains unmeasured.

## Source support

The shared lazy classifier interns string-index dictionary descriptors and their
recursive children. Indexed and named reads use the declared selected value's
contract, evaluate receiver and key once, and check each read. Descendant fields
are registered in the shared conservative policy, including reads lowered before
a cast. Dictionary descriptors never become source writable-slot certificates.
Unsupported key domains and element families retain demanded-read refusals.

The C source adapter keeps fixed-shape producers in their original objects and
checks actual physical slot types/readiness before interpreting the payload.
Boxed unions, maybe-number storage, reference heap kinds and semantic null are
classified independently of the target. Existing record wrappers dispatch through
record.c with independent shape and reference-storage checks. Missing prototype
members reuse record.c's existing refusal. Ordinary missing keys are undefined and
must satisfy the selected declared type. JavaScript uses own data descriptors and
shared readiness metadata without invoking getters.

The seven controls are options-good, options-wrong, options-read-wrong,
nested-good, nested-wrong, array-wrong and array-read-wrong. All are .a sources
under components/, despite that historical directory name. The new source test
uses interfaceFixture, native.C/native.Build and javascript.JavaScript through the
ordinary oracle helpers. It is compiler output, unlike GROUP1's manual adapters.
Successful sanitized native runs include the leak check; negative runs pin exit
70 and the full named message. Valid 42/undefined and nested name match source Node.
Wrong object, wrong array and later entry.name failures are exact in both backends.

This establishes the broad primitive/object read path needed by the highest-read
family. It does not complete ParsedCommandLine.options/CompilerOptions: its full
CompilerOptionsValue union includes arrays and further dictionaries. No row is
marked complete on the strength of the smaller Options witness.

## Mutants

Each source mutant modifies emitted calls while leaving production files intact:

- Skip dictionary membership: isolated options-read-wrong prints object and exits
  zero in both backends; the original pinned exit-70 read catches it.
- Accept array as object: isolated array-read-wrong prints object and exits zero in
  both backends; the lookup's pinned array refusal catches it.
- Drop the nested field check: nested-wrong prints 42 and exits zero in both
  backends instead of the exact entry.name refusal. The native mutant converts
  the observed numeric slot to printable string storage after removing the
  declared string check, to avoid an invalid pointer reinterpretation. The
  JavaScript mutant reads the field directly. Both are executable semantic mutants.

Initial mutation harnesses accidentally encountered the union-to-string guard
and unsafe later array-as-object consumption. The final isolated lookup controls
remove those confounders; initial output is preserved separately. The native
transitive mutant's representation conversion is explicit, not hidden as a pure
metadata change. The prior eleven component/guard mutants still pass independently.

## Validation

Output is saved in logs/group2-*.log. Completed commands:

```text
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewDictionarySource' -count=1 -v -timeout 10m
ok internal/oracle 3.391s; seven controls, six source mutations
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run 'TestViewDictionary|TestCheckedViewDictionary|TestRecordsAgainstNode|TestLazyView|TestCheckedViewLazy|TestCheckedView.*Array' -count=1 -v -timeout 10m
ok lower 0.082s; native 14.590s; javascript 0.004s [no tests to run]; oracle 24.595s
```

The earlier broader command also selected TestInterfaceCastRuntimeMutants. Its
existing twice_operand case failed with 'mutant changed no cast'; that remains
unresolved, and no full gate or broad cast gate is claimed. The same initial run
caught component-harness duplicate anchors/definitions after extending the runtime;
those harnesses now isolate their original helper and pass in the final run.
Vet for ir/flow/lower/native/javascript/oracle, candidate ranking --check and
whitespace checks complete separately. nproc remains 5. Setup timing remains the
recorded integration setup in GROUP1; no new toolchain setup was needed.

## Next work and integrator

Take this lane's new tip through the user. Every shared hunk is named in
checked-views-plan.md under the dictionary lane. No other lane was merged.
Next groups cover helper/generic/callback/field propagation, richer array/union
and nested dictionary elements, followed by dynamic producers/writes using the
existing record table. They are implementation work, not requested handoffs.
