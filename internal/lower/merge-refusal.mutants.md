# Class merge refusal mutant

On Linux, a temporary Go overlay removed only the registration of rule 2075.
The run selected `^TestClassMergeRefusal$` and exited 1. The registration remains
in the branch.

| Red test | Go assertion | Fixture declaration | Mutant result |
| --- | --- | --- | --- |
| `TestClassMergeRefusal/optional_write` | `merge_refusal_test.go:43` | `main.a:5:1` | Lowered successfully. |
| `TestClassMergeRefusal/optional_read` | `merge_refusal_test.go:43` | `main.a:2:1` | Lowered successfully. |
| `TestClassMergeRefusal/required_unused` | `merge_refusal_test.go:43` | `main.a:2:1` | Lowered successfully. |
| `TestClassMergeRefusal/namespace_unused` | `merge_refusal_test.go:43` | `main.a:2:1` | The existing namespace rule returned `NotYet` for a namespace merged with a runtime value. |
