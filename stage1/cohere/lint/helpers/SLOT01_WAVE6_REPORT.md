Built: normalizeUtilityDefinition, normalizeValueFunctionArguments and RegisterFrameworkVariants in .a; 18 dependency entries across six rules.\
Commits: claims 2cdece8, bde463f, bcaf9aa; implementations 4cddf6c, 883ade6, 5fd641d, all pushed; argument rewriter withdrawn.\
Commands and outputs: full helper package PASS 314.507s; vet exit 0; filtered uncached oracle PASS 1.047s; resumed setup 17s, nproc 5.\
Mutants: ten new compiling semantic mutants and one exact-integer refusal mutant caught; all 38 prior/inherited checks rerun and caught.\
Not covered: whole-rule parity, dynamic fixture generation, cross-slot callback integration, the full Go int64 order domain or full repository test gate; zero additional final blockers.

# Slot 01 sixth batch report

## Delivered behavior

Previous fifteen retained helpers were tested and pushed before this batch; this helper unit has no outstanding rule claims. Three new production helpers occupy separate .a files. No shared registration generator, shared rule harness, compiler ownership file or existing rule entry point was edited. The frozen readiness ledger remains unchanged.

The definition wrapper preserves nil-definition no-op and forwards a present definition's nodes to its explicit normalization dependency. Nil is represented by a bound flag; unrelated arena values remain observable during nil controls. The walker descends children before checking each parent, including children under non-declaration parents. It preserves declaration/value-present/nonempty guards, the exact raw --value(/--modifier( substring bail, and in-place value updates. ParseValue, normalizeValueFunctionNodes and ValueToCss remain an explicit rewrite callback. Actual Go decides every mutation and supplies dependency observations rather than duplicating other workers' parsers here.

An observational Go build overlay appends the raw value to a trace immediately before the actual walker's ParseValue call. It preserves the original branches, parser calls and mutations. The Go worktree is untouched. This makes child-first order and the early substring bail observable; a parent-first mutant differs in Go invocation order even where final rewritten values would be identical. Source Node, sanitized native and emitted JavaScript all match both mutation values and the actual Go trace.

Framework replay preserves an existing registration's name/order while replacing kind, copies new records with their explicit orders, and advances lastOrder only for a newly inserted larger order. An existing root with an incoming order 1000 does not advance the counter. Shared and negative orders are retained. Caller mutation after replay does not alter stored records; repeated replay changes kinds as Go does while preserving positions. Captures include the complete actual Go FrameworkVariantRegistrations table, not only hand-built order controls.

The numeric adapter explicitly refuses non-safe integers with `NotYet: framework variant order outside exact integer range`. JavaScript/native number cannot preserve every Go int64 value; the guard refuses instead of silently rounding. The refusal mutant accepts 9007199254740993 as 9007199254740992. This is a documented integer-adapter limit, separate from semantic Go stdout comparisons on exact orders. Valid consumer-table orders and captured controls are exact.

## Claims and ownership

Wildcard fetch and complete claims checks included every origin codex/lint-helpers* branch before each reservation. Final fetch checked 18 matching branches and found only slot 01 claims for the three retained helpers. All larger named helpers were reserved, including the comment bundle in base HELPERS.md; each selected helper tied the largest unclaimed concrete count at six. Generic strict-option decoding remains rule-local schema work rather than a named shared helper. Fetched claim snapshots are retained under evidence/slot01-wave6/claims-*.json. No fourth retained helper was claimed.

Initial argument-rewriter claim 3c87ca1 at 02:30:44 UTC lost to the newly published codex/lint-helpers-from-lint-wave1-12 claim 61519ec at 02:30:27. Its passing local file and tests were removed before any implementation commit. argument.log is explicitly withdrawn historical evidence, not delivered coverage or credited mutants. Registry factory, Register and AttachComparison were found reserved by slot 03 and never claimed here.

Replacement definition claim 2cdece8 was pushed before code; implementation 4cddf6c passed and was pushed before walker claim bde463f. Walker implementation 883ade6 passed its mutation and ordering tests and was pushed before replay claim bcaf9aa. Replay implementation 5fd641d passed and is pushed. Additional actual-table capture lives in slot-owned Go testdata and is included with final evidence. Normal pushes only; no PR or history rewriting.

## Commands and observations

All test stdout/stderr went directly to log files, never through pipes. Evidence paths below are relative to stage1/cohere/lint/helpers; commands ran at repository root using that complete prefix. Build shells source /workspace/adamic-tools/env.sh.

- `bash cloud/setup.sh > /tmp/lint-helpers-01-wave6-setup.log 2>&1`: PASS, copied to evidence/slot01-wave6/setup.log. Timing lines: go ready 0s, clang ready 0s, node ready 0s, submodules ready 0s, build cache warm 17s, done 17s on five processors. nproc 5, cgroup cpu.max 400000 100000, 17.6 GB.
- `go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/slot01-wave6/final.log 2>&1`: PASS 314.507s, 46 top-level tests.
- `go vet ./... > evidence/slot01-wave6/vet.log 2>&1`: exit 0, empty log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > evidence/slot01-wave6/oracle.log 2>&1`: PASS 1.047s, six fixture/probe misses.
- Targeted ten-test run before the integer guard: PASS 57.721s, targeted.log. Guarded replay/refusal run: PASS 18.487s, framework.log. The actual-table replay rerun passed in 19.437s and is retained in framework-table.log; final.log includes the final corpus and all eleven new checks.
- Initial wrapper: PASS 11.141s, definition.log; initial walker group: PASS 31.679s, walker.log; isolated parent-first mutant: PASS 5.142s, order.log.
- `gofmt -l` for slot-owned Go files and `git diff --check`: empty output.

The resumed cloud-runtime skill reported current enforced network policy and ready tool variables. No credentials were printed or guessed. Setup and all ordinary Git pushes succeeded.

Each tree corpus reads all six consuming fixture files, 625 complete string expressions plus controls, producing 1,284 nil/present tree batches with six arena nodes each. Wrapper observations comprise 7,704 UTF-16 output lines. Walker observations comprise 9,641 lines, including actual Go call-count/order traces. Inputs are controlled Go arenas hosting fixture strings, not a claim that every string is a parsed CSS stylesheet. Shape controls include declaration children, descendants of a comment, absent values, empty values, lowercase/uppercase/spaced function spellings, nested functions, quoted calls and Unicode/escape/newline arguments.

The final registry corpus includes all six consumers, the same complete string expressions plus collision/Unicode controls and the actual Go framework table. It records three snapshots per batch: initial replay, caller mutation without replay, and replay of the mutated input. Initial counter values 3 and 9 test both advancement and keeping a larger existing counter. Exact final batch/output counts are printed in final.log and framework-table.log. Actual Go maps provide expected names, kinds and orders; no ordering assumption about Go map iteration is used. Queries follow captured input order, including duplicates.

These are bounded helper corpora. Complete Go string concatenations, labels and messages are retained; dynamically constructed fixtures and whole-rule findings/fixes are not executed. Semantic mutants must compile, run successfully and have empty stderr before a Go stdout difference is credited. Compiler failures, panic or sanitizer findings do not count.

## Every new mutant and catch

| Helper | Mutation | Independent witness |
|---|---|---|
| normalizeUtilityDefinition | remove nil guard | line 4: mutant rewrites unrelated arena value; actual Go nil leaves it unchanged |
| normalizeUtilityDefinition | skip forwarding normalization | line 10: actual Go rewrites the child; mutant leaves raw argument spacing |
| normalizeValueFunctionArguments | skip child recursion | line 11: actual Go normalizes nested modifier; mutant leaves it raw |
| normalizeValueFunctionArguments | remove value-present guard | line 7377: mutant changes absent-value declaration; Go preserves it |
| normalizeValueFunctionArguments | remove declaration-kind guard | line 7373: mutant changes comment value; Go preserves it |
| normalizeValueFunctionArguments | remove substring bail | line 29: mutant trace has two calls, Go one |
| normalizeValueFunctionArguments | parent before child | line 7380: mutant first trace is parent fixture, Go nested modifier |
| RegisterFrameworkVariants | overwrite existing order | line 2: mutant order 1000, Go 3 |
| RegisterFrameworkVariants | do not advance lastOrder | line 1: mutant lastOrder 3, Go 5 |
| RegisterFrameworkVariants | store incoming object by reference | line 26: caller mutation changes mutant kind to tampered; Go still static |
| integer adapter refusal | remove exact-integer guard | compiled mutant accepts 9007199254740993 as 9007199254740992; source/native/emitted-JS baselines refuse |

The first ten are successful compiling semantic mutants caught by actual Go output. The eleventh proves an explicit representability refusal rather than inventing a Go boolean for an unsupported numeric adapter. Full regression also reruns all 38 prior/inherited checks, documented in SLOT01_REPORT.md and SLOT01_WAVE2_REPORT.md through SLOT01_WAVE5_REPORT.md. final.log retains their witnesses. Totals: 47 compiling semantic mutants plus two domain-refusal mutants.

## Every consumer and remaining dependencies

All three helpers remove one dependency each from each of these six rules:

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

Eighteen dependency entries are removed across six distinct consumers. No additional rule becomes fully helper-ready from this batch or from adding it to previous slot 01 deliveries. The cumulative slot-only final helper blockers removed remain the three structure rules from the first batch. slot01_wave6_readiness.json lists every helper's consumers and residual dependencies after this batch and after all eighteen retained slot 01 helpers. Unmerged helpers from concurrent branches are not silently credited, and no rule implementation status is changed.

Not covered: whole-rule findings/spans/fixes/suggestions, shared profile compilation or registration, native integration of the parser/normalizer callbacks or arbitrary cross-slot arenas, dynamic fixture construction, malformed raw UTF-8 or isolated UTF-16 surrogates, cyclic/non-null-invalid trees beyond index refusals, concurrent registry mutation, or the full Go int64 order domain. Registry input uses exact safe integers and explicitly refuses unsupported orders. The full repository go test ./... gate was not run; verification runs the entire touched helper package, repository-wide vet and a filtered uncached external oracle.
