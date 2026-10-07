Built collapse.FrameworkStaticReading in one .a file: six prerequisite entries removed across six Tailwind rules, zero final blockers alone.
Commits: fa760462 claimed before code; b07ef13e/bed0dfb8 are the withdrawn NewTheme duplicate, not additional delivery.
Commands: parity PASS 15.99s, 3,036 cases, 893 registrations including three controls, 75,326 identical bytes; owned package vet clean; six external Node oracle fixtures PASS 10.132s.
Mutants: wrong count, missing name reported found, dropped order, copied order storage and reused sorted record all ran cleanly and failed Go byte comparison on source Node, emitted JavaScript and sanitized native.
Not covered: full rule integration, independently compiled node-conversion/property-sort helpers, full repository gate and unavailable external Tailwind/corpus tests.

## API and proof boundary

frameworkStaticReading takes a name, a map of declaration identities and explicitly supplied nodesFromDeclarations/propertySort dependencies. Object-constrained declaration views preserve present nil/empty lists independently of missing keys; sort results carry the Go nil bit plus values for order storage. This helper implements only lookup, dispatch and returning the exact count/order with a fresh reading record and shared order storage. Externally owned helpers are not copied or guessed.

The real Go constructor table has 890 pinned registrations. The oracle adds nil declarations, an allocated empty declaration list and a value-absent declaration as three controls. Every entry is sorted by real Go PropertySort. Declaration identities and those dependency results drive the isolated Adamic helper, while the comparison answers come from the unmodified Go FrameworkStaticReading logic. The test overlay adds one observational statement after sorting to retain its order storage and record fields; it does not change the algorithm or supplied inputs. Mutating returned order storage is observed through the original sort result, while mutating the returned count must leave the original sort record unchanged. Both properties have independent clean-executing mutants.

Inputs include source-derived names from all 157 actual fixture sources across all six consumers, every pinned registration, altered unknown names, empty/Unicode/NUL/case controls and the three supplemental declarations. All 3,036 results match Go byte for byte on source .a Node, emitted JavaScript and ASan/UBSan native. The callback answers isolate this helper and do not prove independent compilation or integration of those other helpers.

Consumers:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Each loses one helper dependency, not its final blocker. NewTheme was withdrawn to slot 05's earlier claim and is not counted. Historical constructor source exists only in bed0dfb8; its evidence remains separately labelled.

## Commands and limits

```bash
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/lint/helpers/from-wave1-06 -run '^TestFrameworkStaticReading$' -count=1 -v -timeout=20m > /tmp/helpers06-static-alias-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/from-wave1-06 > /tmp/helpers06-static-vet.log 2>&1
python3 stage1/cohere/lint/helpers/from-wave1-06/testdata/regenerate.py > /tmp/helpers06-static-regenerate.log 2>&1
go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > /tmp/helpers06-oracle.log 2>&1
```

Capture records all six consumers before external-engine skips. Its separate rule-package gate fails the same eight external Tailwind/live/corpus checks listed in REPORT.md and evidence/capture.log; it is not counted as passed. The raw capture is reused only for the exact same pinned consumer cohort and is regenerated after changing its symbol metadata. Arbitrary invalid UTF-8, concurrent mutation of the supplied table, arbitrary callback semantics and malformed adapters are outside the typed immutable-table boundary.

Toolchain setup and original failures are documented in REPORT.md. The first alias observer failed to assign Go Sort directly to Reading; the observer now explicitly copies the matching fields. That compile failure is not a credited mutant. Shared harness, compiler, other workers' helper files and cohere's worktree remain untouched.
