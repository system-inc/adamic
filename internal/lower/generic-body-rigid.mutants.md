# A generic body keeps its type parameter rigid (#p0fy047)

Inside a generic body a mutable slot holding a type parameter, seen as one holding the parameter's
constraint or anything wider, is refused (docs/rulings.md, the generic body and the invariant-mutable
entry). No code of its own holds this. The invariance walk already keeps T rigid wherever it asks tsc
whether the wider side's slot can be written into the narrower side's, since tsc never takes a `{}`
into a `T`: the container arm (`widenedArguments`), the writable-field arm (`widenedProperties`), the
record arm (`widened`'s `recordElement` branch) and the function relation (`widenedParameters`).
`internal/oracle/generic_body_rigid_test.go` pins it, and each mutant below puts the substitution the
ruling forbids (T judged by its constraint) into one arm, or drops one check, so the pins show which
arm holds which road. c4v-6's nested arm (`boundArgumentsRefusal`, #c4vhbep) is mutant 2.

Each mutant is a sed run in the pins-and-mutants proof job (the job's command is in the head's
commit message), checked as applied, run over `TestGenericBodyKeepsTypeParameterRigid`,
`TestInferredBindingNestedTypeParameter` and `TestConstraintRelationSites`, and restored. A row
is red in both its `.a` and `.ts`.

| Mutant | Where | Red rows |
| --- | --- | --- |
| m1, a T at a container's slot judged by its base constraint | invariance.go:391, `widenedArguments` | rigid b01, b02, r1, r2, r3, r6, r7 |
| m2, c4v-6's nested arm dropped | constraint_relation.go:173, `boundArgumentsRefusal` | rigid c01, c03, q01, q02, s01; inferred combine_number, combine_object, combine_string, p01, p02, p03; sites a5, a5b |
| m3, the function relation's assignability test dropped | invariance.go:235, `widenedParameters` | measured at -4 |
| m4, a T in a writable field judged by its base constraint | invariance.go:256, `widenedProperties` | measured at -4 |
| m5, a T in a record's element judged by its base constraint | invariance.go:160, `widened`'s record arm | measured at -4 |
| m6, a T at a container's slot judged by its own constraint | invariance.go:391, `widenedArguments` | measured at -4 |

m1 leaves b03 (held by the field arm, m4), r4 (the record arm, m5) and r8 alone. r8 is
`<T extends U, U>`: `GetBaseConstraintOfType` answers nil for a T whose constraint is an
unconstrained U, so m1's substitution doesn't happen, and m6 puts in the immediate constraint U
instead. m1 and m2 were measured in J-7w47hh2 at 1d928432, where q03's stop file still read
"measure", so q03 is left out of their rows until they are measured again at -4. r5 and r5b stop under
method-signature-style's tag, which is `wideningRefusal`'s fix text for the function relation
(invariance.go:1164). No registry rule carries that name, so m3 is the proof that a soundness check
holds them.
