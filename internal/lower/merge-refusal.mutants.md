# Class merge refusal mutant

On Linux, a temporary Go overlay removed only the registration of rule 2075.
The lower run selected `^TestClassMerge(Refusal|Allowed)$` and exited 1.
The registration remains in the branch.

| Red test | Go assertion | Fixture declaration | Mutant result |
| --- | --- | --- | --- |
| `TestClassMergeRefusal/optional_write` | `merge_refusal_test.go:45` | `main.a:5:1` | The optional-write rule returned `NotYet` at `main.a:2:43` because the class holds no `f` slot. |
| `TestClassMergeRefusal/optional_read` | `merge_refusal_test.go:45` | `main.a:2:1` | Lowered successfully. |
| `TestClassMergeRefusal/required_unused` | `merge_refusal_test.go:45` | `main.a:3:1` | Lowered successfully. |
| `TestClassMergeRefusal/namespace_unused` | `merge_refusal_test.go:45` | `main.a:2:1` | The existing namespace rule returned `NotYet` for a namespace merged with a runtime value. |
| `TestClassMergeRefusal/namespace_class_merge_twin` | `merge_refusal_test.go:45` | `main.a:1:12` | The existing namespace rule returned the same `NotYet`. |

`TestClassMergeAllowed/duplicate_field` and `TestClassMergeAllowed/inherited_field`
still agreed with Node, printing `7` and `8` respectively.
The separate oracle run selected `TestMergedFieldAdamicInitializedAgreesWithNode`
and exited 0. Its `merge_present.a` fixture still agreed with Node and printed `7`.
