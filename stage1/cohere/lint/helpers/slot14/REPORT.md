Built: two bounded helper candidates, both withdrawn to earlier claims; zero executable helpers delivered.
Commits: 2b2947ff integer claim, 71e76328 replacement claim; final evidence and withdrawals follow.
Checks: actual Go matches source Node, emitted JavaScript and sanitized native for 4,761 integer inputs and 1,012 static-node cases.
Mutants: ignored negative sign and erased value presence both compile, exit zero with clean stderr and differ from Go on all three runtimes.
Limits: missing private live Tailwind fixtures block the required six-consumer gate; no active helper claim remains and no further helper is taken.

The foundation is origin/codex/lint-helpers at 95100eb440b47f3f18e960c6f5b49cadf0dc1d9d, with pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. Previous rule work remains pushed through 9d2c673b on codex/lint-wave1-14. Read the foundation helpers README and parsed its complete readiness ledger before ranking concrete symbols. The largest unclaimed symbol count was six; the generic strict-option label is per-rule work and comments remain owned by the original bundle.

The initial selection inspected 389 origin refs and five unique claim blobs. All-head fetch before implementation found no visible duplicate, but later published commits revealed earlier claims: slot 05's leadingInteger claim 652db0c7 at 02:30:32 UTC precedes ours at 02:30:56. Slot 09's static-node claim 4d352341 at 02:35:21 precedes ours at 02:36:10. Final refresh inspected 416 refs and 17 claim blobs. Neither implementation is delivered or credited. Archived .a.txt sources are raw reproduction witnesses, not importable helper modules. The existing shared helpers and their tests were not edited.

Both symbols have these six consumers in the frozen readiness ledger:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Each symbol would remove six prerequisite occurrences and zero final blockers alone. This worker reports zero removals because earlier owners retain both helpers. The ledger is not updated.

Bounded observations: leadingInteger compares 738 controls extracted from all six consumer test sources plus 4,023 sign, Unicode, signed-64/wrapping and deterministic random controls. All four paths match 160,778 bytes over 4,761 inputs. Exact Go integer results use four base-65536 limbs, avoiding JavaScript rounding, and distinguish no digits from zero. The sign mutant is caught only by output comparison.

Static-node comparison uses all 890 entries in the upstream FrameworkStaticDeclarations table plus nil/empty arrays and every property/value/presence/important combination in independent controls: 1,012 cases, 139,208 identical bytes. It preserves order, declaration kind, all meaningful declaration fields and fresh nodes/backing arrays. Go's nil container/context fields are checked by its oracle; fields outside a declaration kind's set are not exposed by the candidate. Empty Go slice object identity is not a Go observable; Adamic empty-array freshness is checked locally. The presence mutant is caught only by output comparison. Input strings are decoded Unicode, not arbitrary invalid Go UTF-8 bytes.

The independent Go oracles call the real private helpers through added virtual test files using Go overlays. No Go helper implementation is copied into an oracle. The live capture overlay adds a recording call to a temporary copy of variant.go and runs original consumer test suites. All six record zero leadingInteger calls. Canonical and unknown suites exit 1 through explicit fixture coverage guards. The other four exit 0 largely through skipped fixtures, which is not live coverage. Logs identify no installed tailwindcss and missing /Users/kirkouimet/Projects/ahra/app/_theme/styles/theme.css; canonical placement also depends on unavailable corpus files. These external repositories cannot be substituted with invented fixtures and called the original consumer gate. This blocks the requested all-consumer validation, so no subsequent claim is held.

Reproduction after sourcing /workspace/adamic-tools/env.sh, with every test output directed to a log:

```sh
python3 stage1/cohere/lint/helpers/slot14/validate.py > /tmp/helper14-run.log 2>&1
python3 stage1/cohere/lint/helpers/slot14/validate_nodes.py > /tmp/helper14-nodes-run.log 2>&1
go vet ./... > /tmp/helper14-vet.log 2>&1
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestCountsAreRecorded' -count=1 -timeout 30m > /tmp/helper14-filtered.log 2>&1
```

The validators reconstruct withdrawn modules only inside their own temporary directories. Exact upstream consumer commands and Go overlays are in validate.py. [Integer evidence](evidence/comparison.log), [node evidence](evidence/nodes.log), [six-consumer coverage](evidence/coverage.json), [framework cases](evidence/nodes-cases.json). Native correctness uses ASan/UBSan and default Linux leak checking; clean stderr and zero exit are required before any mutation is credited. No throughput claim is made for withdrawn candidates.

Setup passed: go, clang, Node and submodules each 0s; build cache warm 79s; done in 80s. nproc 5, cgroup quota 400000/100000; Go 1.27.1, clang 20.1.8, Node 24.19.0. Repository vet exits zero with empty output; gofmt -l cmd internal is empty. Filtered external oracle passes in 14.968s. The cohere CLI refuses the candidate .a paths as outside its TypeScript/JavaScript program, so no self-lint pass is claimed. No full repository gate or live consumer findings/fixes gate passed. An existing untracked shared .generated directory is left untouched and excluded from commits.
