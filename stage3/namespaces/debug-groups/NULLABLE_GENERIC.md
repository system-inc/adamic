# Debug nullable generic parameter

Site debug.ts:255:37; exact reduced source in nullable_generic.a.
Node prints 3, null rejected, undefined rejected on three lines. The source
mutant drops the null check, prints null instead of null rejected and is caught
by Node. The original parameter and branches remain checker-valid.

Production diagnostic:

```text
adamic: /workspace/adamic/stage3/namespaces/debug-groups/nullable_generic.a:2:34: stage 0 can't lower a value of type T | null | undefined yet
```

NotYet is sound for the current ABI. General null | undefined | T needs distinct
absence tags; references' null-pointer undefined and MaybeNumber's optional
presence bit cannot silently encode three states. Narrow sound acceptance
requires a tagged representation and specialization of T, checked reads, and
matching ownership in both backends. A closed-world specialization excluding
one absence must prove it at every admitted call; the public signature alone
cannot provide that proof. No such ABI was added in this namespace unit.

TestDebugNullableGenericBoundary pins that limitation. Source Node mutation
is independent semantic evidence, not a native compiler mutant or acceptance
claim. This group remains 1.
