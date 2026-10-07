Built owned, reproducible blocker probes for the next three claimed TypeScript rules; no complete ports were registered.
Commits: prior work 37eee07 pushed; third-batch claim e18b1c2 pushed before probes; report and evidence are in the final third-batch commit.
Commands: setup failed inherited profile_test.go compilation; nproc=5; upstream selected tests PASS 0.017s; all three serializer probes explicitly refused their repair shapes.
Mutants: three Go-only repair-output mutants are checked by byte comparison; these are witness sensitivity checks, not Node/emitted-JavaScript/native port mutants.
Not covered: port/backend parity, corpus comparisons, port mutants, throughput, full gate; shared repair infrastructure prevents satisfying that bar in owned rule directories.

## Selection

Pushed all existing work, fetched 320 origin refs, and examined claim Markdown across every ref (39 distinct blobs), plus production source/registrations on origin/main ef3d907ecdc4c771b016f7d9c52372def057a340. All 46 helper-ready names are represented in claims. The first three eligible inventory `syntax ready for AST/API adaptation` entries are @typescript-eslint/no-unnecessary-type-constraint, @typescript-eslint/prefer-as-const and @typescript-eslint/prefer-enum-initializers. Selection follows the previous batch's original 46-plus-54 queue. Claim audit JSON enumerations are evidence, not reservations; claim Markdown name mentions are conservatively reserved, including skipped dispositions. See wave1-15-third-selection.json. No additional rules were taken after these three.

## Observed blockers

All probes use unchanged cohere rule bodies at 715ba94f3608a6500086b1076ce5cb7e51b836db. Collection and serialization are copied verbatim from the shared oracle; only its original main function is renamed to allow the owned probe entry point. A Go overlay adds the probe without editing cohere. Each witness produces exactly one real diagnostic, then the existing serializer exits 2 on the diagnostic's repair shape.

| Rule | Witness | Observed repair | Serializer result |
| --- | --- | --- | --- |
| no-unnecessary-type-constraint | `function data<T extends any>() {}` | Finding 14:15; one suggestion deleting 15:27 | unexpected suggestion shape |
| prefer-as-const | `let value: 'bar' = 'bar';` | Finding 11:16; deletion 9:16 and insertion 24:24 | unexpected fix shape |
| prefer-enum-initializers | `enum D { Up }` | Finding 9:11; three suggestions replacing 9:11 with Up = 0, Up = 1, or Up = 'Up' | unexpected suggestion shape |

The serializer accepts exactly one fix equal to the diagnostic range, or exactly one suggestion containing one edit equal to that range. Finding has one edit interval/replacement and one suggestion string. It cannot preserve a list of fixes or a list of suggestions. It has a separate edit interval, but the current oracle serialization still refuses the constraint-removal interval. These are upstream behavioral requirements, not arbitrary options.

Inference: complete findings-and-repairs parity cannot be certified through the existing shared contract. Removing suggestions, collapsing fixes or registering a finding-only candidate would lose upstream output. The unit's ownership instructions forbid edits to shared finding, serializer, dispatch and copy lists. No incomplete listener or fake success is registered. The claimed rules remain reserved and explicitly blocked. Implementing a rule-local serializer would not integrate its results into the shared driver contract.

## Checks and mutants

Source /workspace/adamic-tools/env.sh, then from the repository root:

```
python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/probe.py --scratch /tmp/lint-wave1-15-third > /tmp/lint-wave1-15-third-run.log 2>&1
go -C cohere test ./internal/lint/rules/typescript -run '^(TestNoUnnecessaryTypeConstraint|TestPreferAsConst|TestPreferEnumInitializers)' -count=1 -v -timeout=10m > /tmp/lint-wave1-15-third/upstream.log 2>&1
```

The selected upstream suite passed in 0.017s, including its exact suggestions, multiple edits and applied-fix assertions. Probes assert exactly one upstream finding; shape inspection exits 0 and serialization exits 2 with the named refusal. No test output was piped.

Each Go-only mutant is applied through a separate scratch overlay, compiles successfully and exits 0 in shape-inspection mode. Full shape-output byte comparison detects it:

- Constraint removal: replace its nonempty edit interval with an empty interval at the same start.
- As-const: change the insertion from ` as const` to ` as mutable`.
- Enum initializers: change the second suggested position from index+1 to index+2.

These prove the blocker witnesses observe real repairs. They do not fulfill the requested per-port mutant bar because no port exists for these rules. The unmodified serializer failure is an explicit limitation, not a successful Go/port differential comparison.

`bash cloud/setup.sh` reported Go, clang, Node and submodules ready at 0s each; test-cache warming failed at stage1/cohere/lint/profile_test.go:32 (`cannot range over portFiles`, now a function). No final warm/done timing was printed. Toolchain remained usable: Go 1.27.1, Node 24.19.0, clang 20.1.8, nproc=5. Sourced the actual environment path and ran the isolated upstream and overlay probes. The initial attempt to redirect a documentation read to /tmp was denied by the read-only sandbox; documentation was then read directly. It changed no file and was unrelated to test results.

No .ts Adamic implementation was written. Files ending .ts.txt are raw TypeScript test inputs; Go probe source is kept as .go.txt outside ordinary package discovery. No shared compiler, lint infrastructure or submodule files were changed. No PR was opened.

Findings/s for native, Node and Go are unmeasured for this batch: there is no complete port to time, and Go count-only mode would bypass precisely the failing repair serialization. Compiler/stage1 corpus parity, emitted JavaScript, sanitized native and full gate are not claimed. The prior batch's rates and passes remain in wave1-15-next-report.md and do not certify these newly claimed rules.

Evidence is in ../rules/typescript-no-unnecessary-type-constraint/evidence/; probe.py reproduces all three rules and three witness-sensitivity mutants.
