# combine inference gap, c4vhbep

No compiled program differed from Node. Of 96 programs, 10 compile, 60 are Refused, and 26 are NotYet. All 10 compiled programs match Node byte for byte on stdout, stderr, and exit code on both checked native and JavaScript backends. Every checked native run exits 0 with empty stderr and no ASan/UBSan finding.

No inferred append program in this census compiles. Six explicit append controls and four concatenate programs compile. This is bounded evidence, not a claim that every possible program in this family is safe. In particular, the exact tsc-isArray cohort has no executable program.

The reported `{}[]` is a synthetic constraint-check slot. The checker resolves append with the enclosing T, including at the actual refusal. Adamic substitutes the append binder's `{}` constraint while judging an inferred binding, then refuses T[] against that slot. There is no observed append body instantiated at `{}`.

## Scope and provenance

Base: origin/main at `5ae1e2401`. Report branch: `typescript/combine-gap`. No compiler changes. Pinned dependency hashes, the TypeScript 6.0.3 core.ts source URL and digest, and tool versions are in evidence/provenance.json and evidence/toolchain.log.

Each .ts file is standalone. Every file retains append's three overload signatures and implementation from core.ts:921 and the tsc predicate `value is readonly unknown[]`. The original `combine` cohort retains all three combine overloads and its complete implementation. The explicit controls spell both append arms as append<T>. Concatenate is a minimal mutable generic spread helper, avoiding its independent readonly overload-result refusal. Printing uses String(value) and a separate typeof line because the house console declaration accepts one string. For object elements the printed value is the n field; the read variant also reads that field arithmetically.

Cohorts:

- No prefix: the requested 36 fixtures, with tsc's exact `isArray(value: any)` implementation calling Array.isArray. The full combine call uses [11],31; ["alpha"],"beta"; or [{n:11}],{n:31}. Each driver prints every element of its result and its typeof.
- `unknown-`: 36 supplementary fixtures. Only isArray's parameter changes from any to unknown. This avoids the independent any NotYet and preserves the predicate and disputed call arguments.
- `array-result-unknown-`: 24 supplementary alias/read/stored/second fixtures. In addition, the probe declares T[] | undefined and returns undefined for missing inputs, isolating the array arm and avoiding the independent void-read refusal. Both inferred and explicit versions receive these changes.

Variants isolate the append arm, a local append alias, storage in T[] followed by indexing, a second generic function, or the concatenate arm. The read driver uses number arithmetic, toUpperCase, or a field read after a presence check. The full combine fixtures preserve both append arms; concatenate fixtures exercise only two arrays. Every cohort includes number, string, and { n: number } instantiations. Supplementary rows are not presented as exact-isArray results.

## Per-program verdicts

`n/a` means no executable exists, not a successful Node comparison. Reason codes apply to checked-build findings too:

- I: Refused, `a value of type T[] seen as {}[], which can write {} where T is read` (adamic/invariant-mutable).
- C: Refused, combine overload 1 promises T[] | undefined, which the implementation result T | T[] | undefined cannot serve.
- A: NotYet, `a value of type any`, at isArray's parameter.
- V: Refused, probe's inferred instance is read as T = void by the signature-readback path, and its body returns a value (adamic/no-void-read).
- O: NotYet, a narrowed union member's object tag cannot be checked with typeof.
- G: NotYet, a generic function value without a concrete contextual signature.
- clean: checked native executed with ASan/UBSan and no finding; stdout, stderr, and exit code match Node on both backends.

Full located reasons are in results.json. Per-backend stdout, stderr, and exit codes for every compiled program are in evidence/.

| Program | Verdict | Native matches Node | JavaScript matches Node | Checked-build findings / stop |
| --- | --- | --- | --- | --- |
| [array-result-unknown-explicit-alias-number.ts](array-result-unknown-explicit-alias-number.ts) | NotYet | n/a | n/a | G |
| [array-result-unknown-explicit-alias-object.ts](array-result-unknown-explicit-alias-object.ts) | NotYet | n/a | n/a | G |
| [array-result-unknown-explicit-alias-string.ts](array-result-unknown-explicit-alias-string.ts) | NotYet | n/a | n/a | G |
| [array-result-unknown-explicit-read-number.ts](array-result-unknown-explicit-read-number.ts) | compiles | y | y | clean |
| [array-result-unknown-explicit-read-object.ts](array-result-unknown-explicit-read-object.ts) | NotYet | n/a | n/a | O |
| [array-result-unknown-explicit-read-string.ts](array-result-unknown-explicit-read-string.ts) | compiles | y | y | clean |
| [array-result-unknown-explicit-second-number.ts](array-result-unknown-explicit-second-number.ts) | compiles | y | y | clean |
| [array-result-unknown-explicit-second-object.ts](array-result-unknown-explicit-second-object.ts) | NotYet | n/a | n/a | O |
| [array-result-unknown-explicit-second-string.ts](array-result-unknown-explicit-second-string.ts) | compiles | y | y | clean |
| [array-result-unknown-explicit-stored-number.ts](array-result-unknown-explicit-stored-number.ts) | compiles | y | y | clean |
| [array-result-unknown-explicit-stored-object.ts](array-result-unknown-explicit-stored-object.ts) | NotYet | n/a | n/a | O |
| [array-result-unknown-explicit-stored-string.ts](array-result-unknown-explicit-stored-string.ts) | compiles | y | y | clean |
| [array-result-unknown-inferred-alias-number.ts](array-result-unknown-inferred-alias-number.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-alias-object.ts](array-result-unknown-inferred-alias-object.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-alias-string.ts](array-result-unknown-inferred-alias-string.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-read-number.ts](array-result-unknown-inferred-read-number.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-read-object.ts](array-result-unknown-inferred-read-object.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-read-string.ts](array-result-unknown-inferred-read-string.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-second-number.ts](array-result-unknown-inferred-second-number.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-second-object.ts](array-result-unknown-inferred-second-object.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-second-string.ts](array-result-unknown-inferred-second-string.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-stored-number.ts](array-result-unknown-inferred-stored-number.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-stored-object.ts](array-result-unknown-inferred-stored-object.ts) | Refused | n/a | n/a | I |
| [array-result-unknown-inferred-stored-string.ts](array-result-unknown-inferred-stored-string.ts) | Refused | n/a | n/a | I |
| [explicit-alias-number.ts](explicit-alias-number.ts) | NotYet | n/a | n/a | A |
| [explicit-alias-object.ts](explicit-alias-object.ts) | NotYet | n/a | n/a | A |
| [explicit-alias-string.ts](explicit-alias-string.ts) | NotYet | n/a | n/a | A |
| [explicit-combine-number.ts](explicit-combine-number.ts) | Refused | n/a | n/a | C |
| [explicit-combine-object.ts](explicit-combine-object.ts) | Refused | n/a | n/a | C |
| [explicit-combine-string.ts](explicit-combine-string.ts) | Refused | n/a | n/a | C |
| [explicit-concatenate-number.ts](explicit-concatenate-number.ts) | NotYet | n/a | n/a | A |
| [explicit-concatenate-object.ts](explicit-concatenate-object.ts) | NotYet | n/a | n/a | A |
| [explicit-concatenate-string.ts](explicit-concatenate-string.ts) | NotYet | n/a | n/a | A |
| [explicit-read-number.ts](explicit-read-number.ts) | NotYet | n/a | n/a | A |
| [explicit-read-object.ts](explicit-read-object.ts) | NotYet | n/a | n/a | A |
| [explicit-read-string.ts](explicit-read-string.ts) | NotYet | n/a | n/a | A |
| [explicit-second-number.ts](explicit-second-number.ts) | NotYet | n/a | n/a | A |
| [explicit-second-object.ts](explicit-second-object.ts) | NotYet | n/a | n/a | A |
| [explicit-second-string.ts](explicit-second-string.ts) | NotYet | n/a | n/a | A |
| [explicit-stored-number.ts](explicit-stored-number.ts) | NotYet | n/a | n/a | A |
| [explicit-stored-object.ts](explicit-stored-object.ts) | NotYet | n/a | n/a | A |
| [explicit-stored-string.ts](explicit-stored-string.ts) | NotYet | n/a | n/a | A |
| [inferred-alias-number.ts](inferred-alias-number.ts) | Refused | n/a | n/a | I |
| [inferred-alias-object.ts](inferred-alias-object.ts) | Refused | n/a | n/a | I |
| [inferred-alias-string.ts](inferred-alias-string.ts) | Refused | n/a | n/a | I |
| [inferred-combine-number.ts](inferred-combine-number.ts) | Refused | n/a | n/a | I |
| [inferred-combine-object.ts](inferred-combine-object.ts) | Refused | n/a | n/a | I |
| [inferred-combine-string.ts](inferred-combine-string.ts) | Refused | n/a | n/a | I |
| [inferred-concatenate-number.ts](inferred-concatenate-number.ts) | NotYet | n/a | n/a | A |
| [inferred-concatenate-object.ts](inferred-concatenate-object.ts) | NotYet | n/a | n/a | A |
| [inferred-concatenate-string.ts](inferred-concatenate-string.ts) | NotYet | n/a | n/a | A |
| [inferred-read-number.ts](inferred-read-number.ts) | Refused | n/a | n/a | I |
| [inferred-read-object.ts](inferred-read-object.ts) | Refused | n/a | n/a | I |
| [inferred-read-string.ts](inferred-read-string.ts) | Refused | n/a | n/a | I |
| [inferred-second-number.ts](inferred-second-number.ts) | Refused | n/a | n/a | I |
| [inferred-second-object.ts](inferred-second-object.ts) | Refused | n/a | n/a | I |
| [inferred-second-string.ts](inferred-second-string.ts) | Refused | n/a | n/a | I |
| [inferred-stored-number.ts](inferred-stored-number.ts) | Refused | n/a | n/a | I |
| [inferred-stored-object.ts](inferred-stored-object.ts) | Refused | n/a | n/a | I |
| [inferred-stored-string.ts](inferred-stored-string.ts) | Refused | n/a | n/a | I |
| [unknown-explicit-alias-number.ts](unknown-explicit-alias-number.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-alias-object.ts](unknown-explicit-alias-object.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-alias-string.ts](unknown-explicit-alias-string.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-combine-number.ts](unknown-explicit-combine-number.ts) | Refused | n/a | n/a | C |
| [unknown-explicit-combine-object.ts](unknown-explicit-combine-object.ts) | Refused | n/a | n/a | C |
| [unknown-explicit-combine-string.ts](unknown-explicit-combine-string.ts) | Refused | n/a | n/a | C |
| [unknown-explicit-concatenate-number.ts](unknown-explicit-concatenate-number.ts) | compiles | y | y | clean |
| [unknown-explicit-concatenate-object.ts](unknown-explicit-concatenate-object.ts) | NotYet | n/a | n/a | O |
| [unknown-explicit-concatenate-string.ts](unknown-explicit-concatenate-string.ts) | compiles | y | y | clean |
| [unknown-explicit-read-number.ts](unknown-explicit-read-number.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-read-object.ts](unknown-explicit-read-object.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-read-string.ts](unknown-explicit-read-string.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-second-number.ts](unknown-explicit-second-number.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-second-object.ts](unknown-explicit-second-object.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-second-string.ts](unknown-explicit-second-string.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-stored-number.ts](unknown-explicit-stored-number.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-stored-object.ts](unknown-explicit-stored-object.ts) | Refused | n/a | n/a | V |
| [unknown-explicit-stored-string.ts](unknown-explicit-stored-string.ts) | Refused | n/a | n/a | V |
| [unknown-inferred-alias-number.ts](unknown-inferred-alias-number.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-alias-object.ts](unknown-inferred-alias-object.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-alias-string.ts](unknown-inferred-alias-string.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-combine-number.ts](unknown-inferred-combine-number.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-combine-object.ts](unknown-inferred-combine-object.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-combine-string.ts](unknown-inferred-combine-string.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-concatenate-number.ts](unknown-inferred-concatenate-number.ts) | compiles | y | y | clean |
| [unknown-inferred-concatenate-object.ts](unknown-inferred-concatenate-object.ts) | NotYet | n/a | n/a | O |
| [unknown-inferred-concatenate-string.ts](unknown-inferred-concatenate-string.ts) | compiles | y | y | clean |
| [unknown-inferred-read-number.ts](unknown-inferred-read-number.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-read-object.ts](unknown-inferred-read-object.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-read-string.ts](unknown-inferred-read-string.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-second-number.ts](unknown-inferred-second-number.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-second-object.ts](unknown-inferred-second-object.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-second-string.ts](unknown-inferred-second-string.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-stored-number.ts](unknown-inferred-stored-number.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-stored-object.ts](unknown-inferred-stored-object.ts) | Refused | n/a | n/a | I |
| [unknown-inferred-stored-string.ts](unknown-inferred-stored-string.ts) | Refused | n/a | n/a | I |

## Where the {} slot comes from

References below are file:line at the base commit; no proposed patch.

1. `internal/lower/invariance.go:710` formats the shared refusal. Its actual caller is `internal/lower/constraint_relation.go:174`, after the inferred-binding argument check at `internal/lower/constraint_relation.go:129`. It is not the ordinary contextual-type walk at `internal/lower/invariance.go:449`.
2. `internal/lower/type_argument_bindings.go:50` obtains the checker's resolved signature. `internal/lower/type_argument_bindings.go:124` reads its mapper and `internal/lower/type_argument_bindings.go:139` obtains the bound type and constraint with the checker's instantiateType. It does not infer new call type arguments.
3. `internal/lower/constraint_relation.go:74` sends an inferred binding to the argument check. `internal/lower/constraint_relation.go:93` deliberately replaces the selected binder with its constraint. `internal/lower/constraint_relation.go:96` makes that mapper. `internal/lower/constraint_relation.go:108` constructs both the actual resolved parameter (`given`, T[]) and the constraint-substituted parameter (`slot`, {}[]). `internal/lower/constraint_relation.go:129` judges the argument against slot. `internal/lower/constraint_relation.go:174` emits the refusal. This substitution creates the {} instance seen in the diagnostic; it does not select an append code instance.
4. `internal/lower/constraint_relation.go:71` routes a written type argument to a type-versus-constraint check instead, explaining why append<T> avoids this particular argument-slot refusal.
5. `internal/lower/lower.go:94` runs refusal before declarations and body lowering. The captured stack has no active lowering type mapper. The signature queried at that exact refusal still has parameters T[] and T | undefined. See evidence/refusal-trace.log. The independent signature and contextual-type trace agrees: signature-trace.log. The injected trace changes no branch or type argument; its extra signature query occurs after a refusal has already been selected. The stack's invariance.go line numbers include the temporary overlay insertion; the base references above do not.
6. `internal/lower/generic.go:25` also reads GetResolvedSignature for actual generic function lowering. `internal/lower/generic.go:49` and `internal/lower/generic.go:52` recover a substitution by matching declared parameter/return types against resolved parameter/return types with inferTypes. This is structural readback of a resolved signature, not fresh inference from call argument expressions. `internal/lower/generic.go:148` installs the body mapper. That path was not reached for the disputed inferred append calls.

## Other consumers of the same source

Actual binding-vector readers, beyond the constraint check:

- `internal/lower/records_declared_keys.go:55` uses the same typeArgumentBindings list for declared-key checks (.a; this census is .ts).
- `internal/lower/class_generic_calls.go:31`, `internal/lower/class_generic_calls.go:37`, `internal/lower/class_generic_calls.go:46` read the resolved signature mapper for class-dependent generic calls and constrained-call nominal checks. These and type_argument_bindings are the only production readers of the signature's private mapper field.
- `internal/lower/generic.go:25`, `internal/lower/generic.go:49`, `internal/lower/generic.go:52` read back actual generic function instances from resolved signatures. Direct ordinary calls enter at `internal/lower/expression.go:1359`; generic generator calls enter at `internal/lower/generators.go:321`.
- `internal/lower/generic_function_value.go:18`, `internal/lower/generic_function_value.go:22` share that readback for function values, with a contextual signature supplied by the pinned checker rather than a call's resolved signature.

For completeness, every other direct production GetResolvedSignature consumer in internal/lower is listed below. These read resolved parameter/result types, predicates, declaration identity, or overload metadata; they are not all new binding-vector inference engines.

- `internal/lower/any_relations.go:207`
- `internal/lower/any_relations.go:225`
- `internal/lower/census_overload_proof.go:235`
- `internal/lower/census_predicate_marker.go:55`
- `internal/lower/census_small.go:302`
- `internal/lower/dead_zone_class.go:291`
- `internal/lower/detached_own.go:247`
- `internal/lower/generators.go:311`
- `internal/lower/library_node.go:477`
- `internal/lower/narrowed_nullish.go:83`
- `internal/lower/placeholder.go:216`
- `internal/lower/predicate_false_flow.go:220`
- `internal/lower/predicates.go:604`
- `internal/lower/predicates_branded.go:173`
- `internal/lower/predicates_checked.go:262`
- `internal/lower/predicates_checked.go:306`
- `internal/lower/predicates_checked.go:489`
- `internal/lower/predicates_function_type.go:84`
- `internal/lower/predicates_proof.go:1031`
- `internal/lower/records_divergence.go:118`
- `internal/lower/records_divergence.go:221`
- `internal/lower/speculation.go:26`
- `internal/lower/void_containers.go:309`
- `internal/lower/void_widening.go:106`

## Reproduction and checks

The .go.txt files are test overlays in package oracle. They call the existing lowered, onNode, onJavaScriptBackend, nativelyUncached, and disagreement helpers. They introduce no replacement runner. The census uses load.Load and lower.Lower with the repository checker settings and enforces the a-check header convention: exact refused.What for Refused, checked for compiled or NotYet. All 96 final headers were verified against the recorded verdicts.

`internal/oracle/oracle_test.go:703` builds through native.Build with Sanitize: true. `internal/native/native.go:123` supplies all house flags, including `internal/native/native.go:171`'s -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all and UBSan source-path stripping. Linux uses ASAN_OPTIONS=detect_leaks=0, exactly as the oracle comparison does. This census claims address and undefined-behavior checking, not a separate leak census. All native and Node subprocesses finish under the oracle's bounded execution helpers.

From the repository root, with pinned submodules and the normal house tool PATH:

```sh
python3 stage3/census/combine-gap/prepare-overlays.py /tmp/combine-gap-overlays
ADAMIC_GATE_UNCACHED=1 go test -overlay /tmp/combine-gap-overlays/oracle.json ./internal/oracle -run '^TestCombineGap$' -count=1 -v > /tmp/combine-gap-oracle.log 2>&1
go test -overlay /tmp/combine-gap-overlays/signatures.json ./internal/oracle -run '^TestCombineGapSignatures$' -count=1 -v > /tmp/combine-gap-signatures.log 2>&1
go test -overlay /tmp/combine-gap-overlays/refusal.json ./internal/oracle -run '^TestCombineGapRefusalTrace$' -count=1 -v > /tmp/combine-gap-refusal.log 2>&1
```

All three tests passed. evidence/oracle.log is the final 96-program run. No compiler source changes are committed. No process started for this task remains running.
