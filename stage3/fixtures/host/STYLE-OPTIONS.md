# Checker option follow-up, October 7

The host stock checker omits noImplicitReturns and
noFallthroughCasesInSwitch, following docs/0.1.md's October 7 ruling on main.
The shared stage 0 runner (fixtures_test.go, load.Load) and host/check.py
(go run ./cmd/adamic build) inherit internal/load.compilerOptions; they do
not define a second checker configuration. Use a compiler with main's ruling.
For a host-capable library compiler predating the ruling, check.py writes a
Go overlay in its log directory, removes both style options, and sets
erasableSyntaxOnly false as on main. Every stage 0 invocation uses that
overlay. The shared Go runner can use the same file with go test -overlay.
Production loader files are untouched.
Stock erasableSyntaxOnly is false to match current main's loader, and the
explicit stock Node declaration root is canonicalized so a node_modules
symlink cannot introduce duplicate declarations.

The following live tooling still explicitly enables both style flags and was
left unchanged:

- stage3/adapt/10-type-imports/adapt.cjs:78-79 (checker for import adaptation)
- stage3/adapt/20-optional-declarations/adapt.cjs:24 (checker for adaptation)
- stage3/adapt/47-host-errors/adapt.cjs:23 (checker for host-error adaptation)
- stage3/adapt/70-readonly-views/adapt.cjs:12 (adaptation checker)
- stage3/adapt/70-readonly-views/field-audit.cjs:14 (field audit)
- stage3/adapt/70-readonly-views/survey.cjs:18 (survey)
- stage3/adapt/70-readonly-views/writers.cjs:14 (writer analysis)
- stage3/adapt/75-optional-widening/coverage/run.py:50 (coverage checker config)
- stage3/adapt/75-optional-widening/coverage/probe.py:21 (coverage probe)
- stage3/triage/inspect_contracts.cjs:12 (contract inspection)
- stage3/triage/trace_origins.cjs:13 (origin analysis)

stage3/census/inventory.cjs reads upstream src/compiler/tsconfig.json;
stage3/census/config_hook.go.txt and make_overlay.py retain upstream project
options for census measurements. They can therefore inherit upstream house
style rather than independently setting Adamic options. Archived counts and
latent/data/*/fallthrough-resolution.py historical overlays remain unchanged.
stage3/drivers/tsc/corpus.py:15 lists both names as supported upstream test
options; it does not unconditionally enable them.
