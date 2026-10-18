# Step 20 iteration fixtures

Source reductions use TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, from Microsoft's public TypeScript
repository. The source locations in `outcomes.json` and each fixture refer to
that source. Reductions keep the loop's type and operation, replace compiler
objects with small concrete values, and print the yielded values. They do not
copy code from the cohere submodule.

`baseline-outcomes.json` records Node output and the exact compiler outcome on
`dcdbb9098f77f30ad41790c56df1bd63ad462b63`. `outcomes.json` records the current
acceptance contract. A fixture that stops at NotYet or Refused is still held to
its executable source behavior on Node; it is registered with the native oracle
only when it lowers. Source acceptance is checked before lowering.

- `object_iteration.a` reduces the binder's NodeArray loop and retains its
  inherited readonly-array interface and optional metadata declaration.
- `iterable_view.a` reduces core's arrayFrom without erasing Iterable.
- `generator.a` and `delegated_generator.a` specialize singleIterator and
  flatMapIterator, preserving yield and yield* respectively.
- `collections.a` specializes Map keys, values, saved entries, Set values and
  Set entries from core, builderState and executeCommandLine.
- `strings.a` specializes arrayFrom to strings; supplementary and lone
  surrogates are authored probes of the string specialization.
- `object_binding.a` reduces the relatedInformation loop in program.
  `object_binding_stress.a` keeps the relatedInformation and renamed-property
  destructuring shapes, adding mutation, captures, exits and ownership probes.
- `user_forwarding.a` reduces core's custom Set iterator forwarding.
- `accessor_binding.a`, `array_view_weak.a`, and the three intrinsic write
  fixtures are authored soundness counterexamples.
- `array_view_stress.a` keeps the generic inherited-array shape, adding live
  mutation, retained RHS, per-iteration captures, dynamic strings and exits.
- `test262_array_views.a` adapts four original array iteration cases to the
  inherited readonly-array view. Assertions are preserved as throwing numeric
  comparisons. The original cases also run independently on Node in default
  and strict modes. This covers four cases, not the complete iteration suite.

The test262 source revision is `2e0a56762801e275a9fdf96dc49d90ba0cddcf63`.
Its BSD license is in `TEST262-LICENSE`. Microsoft's TypeScript source is
licensed under Apache-2.0; reductions retain source attribution in their headers.

Run the scoped snapshot and acceptance tests with output redirected to logs:

```sh
go test ./internal/oracle -run '^TestStep20IterationOutcomes$' -count=1 -v
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/stage3/fixtures/iteration/' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
```

A snapshot change requires review of its exact reason and independent Node
output. `ADAMIC_STEP20_RECORD` writes observations to an external file; ordinary
runs never silently refresh expected outcomes.

The first implementation accepts the four inherited readonly-array fixtures.
`array_view_variance.a` is an authored counterexample: a readonly container still
exposes mutable element fields, and widening their writable type must be Refused.
The three intrinsic-write fixtures now require the ruled refusal reason.

Run the four implementation mutants in an isolated scratch directory:

```sh
python3 stage3/fixtures/iteration/check_array_view_mutants.py /tmp/step20-mutants
```

The live-snapshot mutant must disagree with Node in both backends. The strong
Weak-slot mutant must fail under ASan. The override and mutable-element mutants
must violate their specific refusal contracts. A compiler build failure is not
an accepted mutant kill.
