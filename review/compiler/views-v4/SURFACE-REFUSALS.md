# Escaped adapter surface refusals

Dynamic argument-list apply, primitive call thisArg, primitive apply thisArg, and nullable primitive call thisArg are refused before emission, with a source path and an actionable fix. Node runs each misfit witness successfully. Ordinary dynamic/primitive forms previously reported NotYet; these now report Refused with fix. The old nullable-receiver exception was narrowed to explicit undefined, avoiding admission of a primitive union through the object ABI.

Commands: every go test used -count=1 -timeout 90s and an external timeout 90.

- ./internal/oracle -run '^TestV4EscapeAdapter(DynamicApply|Primitive(Call|Apply)|NullablePrimitiveCall)Refusal$' -v: green, 0.384s; leaves 0.34s, 0.36s, 0.36s, 0.37s.
- ./internal/oracle -run '^TestV4EscapeAdapter(Call|Apply|Dynamic|Primitive|Nullable)': green; see surface-refusals-controls.log.
- ./internal/ir -run TestCallTargetReaders: green, 20.748s.

No backend is reached by the unsupported programs. Existing supported call/apply checks and their native/JavaScript omission mutants remain green. No persistent fixtures or count rows changed. Setup total 66.671s, build 66.472s, node 0.024s, Go 0.024s, dependencies 0.082s, submodules 0.098s, clang 0.211s; nproc=5, cgroup quota=4.
