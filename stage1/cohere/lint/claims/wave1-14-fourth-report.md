Built: reproducible blockers for all three fourth-batch claims; no partial rule port was registered.
Commits: prior work pushed at a6b67854; pre-code claim 4d5ef284; evidence commit follows.
Commands and outputs: blocker probes PASS 15.048s; original Go tests PASS 0.015s; vet exit 0; setup 17s, nproc 5.
Mutants: none, because no fourth-batch rule implementation could satisfy the shared repair contract.
Not covered: these rules' four-way parity, compiler/stage1 comparisons, semantic mutants, native/Node/Go findings rates or the full repository gate.

## Selection and claim

Pushed all existing work first: Everything up-to-date. Fetched every origin head
without recursive submodule fetching. Scanned 320 origin refs, main's implementation
selections and direct claim files. All helper-ready names remained represented in
claims, including earlier existing-origin-port skips. The next three eligible
syntax-ready inventory entries were, in order:

1. @typescript-eslint/no-unnecessary-type-constraint
2. @typescript-eslint/prefer-as-const
3. @typescript-eslint/prefer-enum-initializers

Claim 4d5ef284 was pushed before creating the owned validation harness. The audit
in wave1-14-fourth-evidence/selection.json reconstructs this worker's claim file
at a6b67854, before the update, and asserts that these are the exact first three.
It includes the main and inventory pins and per-rule claim matches. Reports and
logs are excluded from claim classification.

## Observed shared-contract blockers

The probes build the existing independent Go serializer with virtual selections
of the real, unmodified upstream rules. No production source, shared list or
cohere submodule file is changed. Each source first runs successfully in count
mode with exactly one finding, establishing valid parsing and listener execution.
Full serialization then reaches a specific unsupported shape guard and exits 2.

| Rule | Valid input | Observed serializer refusal |
|---|---|---|
| no-unnecessary-type-constraint | `function f<T extends any>() {}` | `panic: unexpected suggestion shape` |
| prefer-as-const | `let foo: 'bar' = 'bar';` | `panic: unexpected fix shape` |
| prefer-enum-initializers | `enum E { A }` | `panic: unexpected suggestion shape` |

The exact rule bodies explain the observations:

- Type constraint reports the parameter name, but its suggestion removes the
  separate range from name.End() through constraint.End(). The shared serializer
  requires the suggestion's single edit to equal the diagnostic range.
- As-const annotations produce two distinct safe fix edits: remove the colon and
  type annotation, then insert ` as const` after the initializer. The model and
  serializer support one replacement. Its as-expression arm alone fits the model,
  but porting only that arm would omit required findings and repairs.
- Enum initializers offer three ordered suggestions: position, position plus one,
  and the member's written name as a string. Finding holds one suggestion string;
  the serializer requires exactly one suggestion. Selecting one would discard
  observable Go behavior and the author's choice.

The current Linter.fixed also applies diagnostic start/end rather than the
independent editStart/editEnd fields. These existing fields therefore do not
supply a working independent repair-range contract.

Every original Go test for the three rules passed separately, including exact
fixed sources and applied suggestion rewrites. Go cohere remains pinned at
715ba94f3608a6500086b1076ce5cb7e51b836db. Counts alone are not presented as rule parity.

Inference: completing these ports requires shared support for independent edit
ranges, multiple fixes and ordered suggestion arrays, carried through application,
serialization and differential comparison. CLAUDE.md says: "Never edit a dispatch,
oracle, corpus or copied-file list." The unit owns its rules and claim files, not
those shared production contracts. No partial listener, placeholder descriptor,
findings-only port or incorrect repair is registered. All three remain claimed
and explicitly blocked.

## Commands and evidence

Source /workspace/adamic-tools/env.sh in each shell. Go 1.27.1, clang 20.1.8,
Node 24.19.0. nproc 5; cpu.max 400000 100000; memory 17.6 GB.

All tests wrote output directly to log files:

```
python3 stage1/cohere/lint/claims/wave1-14-fourth-evidence/validate.py > /tmp/lint-wave1-14-fourth-run.log 2>&1
# Runs go test -overlay=<printed overlay> ./stage1/cohere/lint -run '^TestWave14Fourth' -count=1 -v -timeout=20m
GOFLAGS=-overlay=<printed overlay> bash cloud/setup.sh > <owned>/setup.log 2>&1
go vet -overlay=<printed overlay> ./... > <owned>/vet.log 2>&1
# From cohere/:
go test ./internal/lint/rules/typescript -run '^(TestNoUnnecessaryTypeConstraint|TestPreferAsConst|TestPreferEnumInitializers)' -count=1 -v -timeout=10m > <owned>/upstream.log 2>&1
```

The temporary compatibility overlay is inherited from the earlier ports: .a
module discovery/imports/mutants/copying and the profile_test.go range-over-function
typo. It injects only the owned blocker test and prior owned helper functions.
The production compatibility patch remains unapplied, as previously reported.

Results: three independently built blocker probes PASS 15.048s; original Go rule
tests PASS 0.015s; repository vet exits 0 with an empty log. Setup uses the known
compatibility workaround and succeeds: Go ready 0s, clang ready 1s, Node ready 1s,
submodules ready 1s, cache warm 17s, done 17s on 5 processors.

No new Adamic module was written; there is no .ts implementation. No semantic
mutant or throughput number is claimed for an unbuilt rule. Node source, emitted
JavaScript, sanitized native and the compiler/stage1 corpora were not compared for
these three rules. The full repository gate was not run; the focused touched-package
blocker tests, filtered independent Go rule tests and repository vet were run.
