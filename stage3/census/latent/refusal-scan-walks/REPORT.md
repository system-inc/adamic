# Refusal scan investigation, task 2bz5m7v

Initial source finding, before changes: base origin/main 42dde4620.

The proposed repeated per-rule file registration is not present. First-error
`latentRun` calls `fresh().latentRefuse(file)` once per file before the unit
loop (`stage3/census/latent/lower.go.txt:273`, unit loop at 275).
`refusalRegistry.register` installs a closure at init time, not a source walk
(`internal/lower/refusal_rules.go:33`). Its closure calls the actual rule, so
cumulative samples attributed to register can include arbitrarily deep rule
helpers. Production refuse already dispatches every rule from one shared
syntax visitor (`internal/lower/refusals.go:102`) and has a separate contracts
visitor (lines 72-87), plus module preflights (lines 56-69).

The census clone removes each original visitor child walk and installs exactly
one deferred child walk (`refusalrewrite/rewrite.go:139-178`). It deliberately
continues after refusals and skips diagnosed top-level function bodies. The
per-unit latentRegister is separate declaration registration, not refusal-rule
registration. No compiler or overlay behavior has been changed.

Profiling and rule-helper attribution are pending. A single shared dispatch
walk already exists, so proposing that as a new lowering fix would be incorrect.
