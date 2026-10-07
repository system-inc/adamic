Rebased all thirty completed slot 02 helpers onto origin/area/stage1-lint d65a8f93, containing current main 39638d9e and the requested shared harness/finding model; no new claim or helper.
Commits: previous published f82ae4ba44701f1dfb968363e25a5f057f6dd613; rebased source bf995e635c5673dcbad96848e54733ceb4d2fad3; publication branch codex/lint-helpers-02 only.
Commands: setup PASS 139s/nproc 5; complete helper oracle PASS 650.727s; selected shared harness PASS 294.490s; inventory PASS 8.255s; uncached input oracle PASS 25.172s; vet exit 0.
Mutants: all 102 fresh helper semantic witnesses caught against actual Go; the inherited emitted-JavaScript mismatch check also rejects its planted mismatch.
Not covered: full repository gate, integrated consumer findings, the separate inherited comments package, arbitrary malformed adapters, complete CFG builder and the prior external Tailwind live/corpus gaps.

# Landing

The landing-first cap makes this rebase, fresh oracle run and publication the current unit. All thirty owned helpers were already complete and pushed; no claimed rules remain unfinished. No new claim is made. The only publication target is codex/lint-helpers-02. Exact force-with-lease uses the previous published f82ae4ba because the requested rebase rewrites history. Neither main nor an area branch is pushed; integration remains external.

A wildcard fetch refreshed main, area/stage1-lint and every codex/lint-helpers* branch. The latest requested area tip d65a8f931c98655936ae04c6899f38f14862b73e includes current main 39638d9e278d38bb5aeae887f46d55a70e47aaad, 50a5f105 and harness 41eb6eab2. The rebase completed cleanly; five equivalent shared commits already in the area branch were skipped. All thirty owned .a helper source hashes match the previously published manifest. evidence/ancestry.log and helper-source-manifest.json record these facts.

The inherited root harness, finding model, profile compilation and separate comments helper package are accepted. The helper-tree difference from the previous publication consists solely of the inherited comments subtree. No shared registration, harness or protected compiler source is edited. The expected thirteen-package allocator leak-check replacement is not present in this base; none is reverted. There are no new rules, node-kind dispatch changes, regex matchers or option-pattern translations in this unit.

# Verification

Test output goes directly to evidence files without piping. Go commands source /workspace/adamic-tools/env.sh. Setup reports Go 1.27.1, clang 20.1.8, Node 24.19.0, five processors: Go/clang/Node/submodules ready in 0s; cache warm 139s; done 139s.

Commands:

- bash cloud/setup.sh > evidence/setup.log 2>&1; nproc > evidence/nproc.log: PASS, 139s and 5.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/helpers-final.log 2>&1: PASS 650.727s; all 102 compiling semantic mutants caught.
- go test ./stage1/cohere/lint -run '^(TestEmittedJavaScriptMismatch|TestDotARename|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix|TestWitnessScriptKind|TestProfileCompilation)$' -count=1 -v -timeout=20m > evidence/shared-harness.log 2>&1: PASS 294.490s, all six selected tests. These check emitted-JavaScript disagreement rejection, .a loading, full suggestion serialization, suggestions alongside applied automatic fixes, script-kind selection and profile compilation. Profile reports allocations 120/frees 120, retains 1012/releases 635, peak 86 and regions 0.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 8.255s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1: PASS 25.172s, all six fixtures, zero cache hits, six probe misses.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory ./stage1/cohere/lint > evidence/vet.log 2>&1: exit zero, empty log.

# Mutants and behavior

[evidence/mutant-witnesses.log](evidence/mutant-witnesses.log) records every fresh helper mutant name and its independent Go mismatch. The expected set is the previous completed gate's 102 mutants: four inherited option/policy mutants and ninety-eight owned mutants across all thirty helpers. Definitions and detailed behavior remain in earlier slot02/batch reports; those reports' pre-rebase commit IDs are historical. The initial inherited and first slot02 suites compare actual pinned Go, source Node and sanitized native. Batch2 onward additionally compares emitted JavaScript. Each semantic mutant must compile, finish without stderr/sanitizer errors and agree across its Adamic backends before an independent Go mismatch receives credit. Crashes and refusals do not count.

The selected shared harness adds coverage independently of those 102 helper mutants. Its emitted-JavaScript mismatch test rejects a planted mismatch after an initially agreeing run. Its result is not counted as another owned helper mutant.

# Rules and limits

No new dependency is removed in this landing unit. The rule consumers, remaining dependencies and conditional readiness effects for every completed helper remain enumerated in earlier batch RULES.md and readiness.json files. The last three helpers read, write and loop each serve array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks: twelve dependency entries, zero final blockers removed. All earlier prerequisite effects remain unchanged by the rebase.

This is helper parity plus six selected inherited harness checks, not full repository or integrated consumer findings/fixes/suggestions parity. The separate inherited comments package is outside the selected Go package gate. Common AST adapters, complete CFG behavior, malformed inputs, callback failure/concurrency and previously documented unavailable Tailwind live/corpus checks remain outside coverage. No shared-harness blocker is observed. The final fetch confirms main and the requested area tip remain unchanged; both are ancestors of the verified source. All 102 mutant names match the prior completed suite exactly.
