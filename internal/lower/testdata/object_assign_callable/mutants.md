# Mutants for Object.assign's callable-source refusal (#k8aak04)

Each mutant is one edit, run on a Linux box against internal/lower, then put back.

The fixtures named `*--assign-function.ts` are Correctness's logging-only assign-function fixtures (typescript/record-handle-fixtures 420caaea, `stage3/census/record-handle-fixtures/adamic-logging`). In kinds 3, 9 and 10, each `declare const enum` stands as its value type, since an ambient enum is refused ahead of Object.assign and the enum is used only as a type. `class-dog.ts` and `class-box.ts` are Language's class copies; class-box.ts is the string-log variant.

| Mutant | Edit | Red |
|---|---|---|
| (a) The rule unregistered | internal/lower/library_object.go: the `refusalRules.register(160, "objectAssignCallableSourceRefusal", ...)` line removed, and its line in refusal_rules_test.go with it | PENDING |
| (b) Unions judged by their common signatures | internal/lower/library_object.go, `callableOrConstructable`: the union and intersection walk removed, so a union is judged by `GetSignaturesOfType` alone, which gives only signatures every member shares | PENDING |
