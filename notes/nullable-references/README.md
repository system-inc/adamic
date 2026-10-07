# Nullable reference coverage

Base: origin/codex/nullable-references at 98ad7e1. Seven added programs agree on Node,
native (sanitized and release), and the JavaScript backend. No differing programs.

## Case inventory

Each reference kind below means string, object, array, map and class, with a present
value and its statically declared null or undefined empty case.

| Code path or use | Existing oracle evidence | Added evidence |
| --- | --- | --- |
| One reference plus one empty case; strict null and undefined comparisons; typeof | nullable_references_generics.a, nullable_references_strings.a | nullable_coverage_{string,object,array,map,class}.a |
| String literal members sharing one nullable representation | No nullable literal-union program found | nullable_coverage_calls.a: red, blue, null |
| Null and undefined comparisons with the constant on the left, including !== | Existing nullable programs put the constant on the right | nullable_coverage_calls.a |
| String spelling in console, templates and concatenation; Number; !; empty string preserved by ?? | nullable_references_strings.a | nullable_coverage_calls.a checks side-effectful operands execute once |
| Property optional reads, class method optional calls, coalescing | nullable_references_generics.a, nullable_references_slots.a | Per-kind programs; maps use coalescing because their optional reads are NotYet |
| Optional interface function-property call and skipped argument evaluation | No nullable interface-call program found | nullable_coverage_calls.a |
| Function parameters, returns, captures, generic functions, generic classes, nested collection generic keys | nullable_references_generics.a, nullable_references_slots.a, generic_values.a | Per-kind observations exercise both generic empty-case instantiations; nullable_coverage_nested_generics.a adds nullable collection arguments and structural-field inference |
| Field storage of strings; readonly generic fields of other kinds | nullable_references_slots.a, nullable_references_generics.a | All five kinds in mutable fields, present to empty and empty to present |
| Nullable array elements, direct iteration | nullable_references_slots.a covers strings, objects, classes, maps | All five kinds, including arrays of arrays, push and index replacement |
| Nullable map values, direct iteration | nullable_references_slots.a strings; nullable_references_map_narrowed.a match arrays | All five kinds, insertion and replacement |
| Missing Map.get and past-end array reads of reference-or-undefined elements | Existing general reference fixtures have partial coverage; map_narrowed covers match-array missing lookup | All five kinds; missing reads passed to functions, then generics, strict comparisons, typeof and ?? |
| JSON empty-case schema, top-level/field/array, present function omission | nullable_references_json.a | Already covered |
| typeof null RegExp match/exec results | nullable_references_typeof_match.a | Already covered |
| Two empty cases, defaulted nullable arguments, changed nullable views, assertions and loose equality | internal/lower/nullable_test.go | Deliberate refusals, not agreeing runnable fixtures |

## Unsupported cases

Index reads of (R | null)[] and Map<K, R | null>.get produce R | null | undefined.
The branch refuses them even inline in comparisons, typeof, templates or calls.
It cannot distinguish both empty values in one pointer; has() and repeated get()
do not provide a checker-visible presence refinement. Direct iteration is tested.

Optional Map.get calls and optional Map.size reads are NotYet. The map fixture
uses (value ?? fallback).size. Defaulted or optional nullable parameters also
need both empty cases. Nullable scalar numbers/booleans, Weak, and multiple
non-string reference alternatives are outside this one-reference representation.
Object/class JSON and non-string reference coercion remain NotYet because present
values need complete shape or coercion semantics. These are existing limitations.

## Mutation proof

Changed internal/lower/expression.go:451 from the typeof-null fallback "object"
to "undefined". nullable_coverage_calls.a compiled, then failed both native and
JavaScript stdout comparisons. Its first line changed from "object 1" to
"undefined 1". Restored the line before final validation.

## Validation

Tool environment: source /workspace/adamic-tools/env.sh. Node v24.19.0,
Go go1.27.1, clang 20.1.8, nproc 5. Stable-checkout setup timing:
go, clang, node and submodules ready at 0s; build cache warm at 80s; done at 80s.
The initial setup overlapped checkout and failed its cache warm; the stable rerun passed.

Commands and complete validation logs are reported with the final result.
