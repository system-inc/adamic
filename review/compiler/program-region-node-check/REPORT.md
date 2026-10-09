Added selector member-site checks for host fixture 14 and an empty-selection control for the pinned scout parse.a; parser Node membership cannot be certified on this host witness.
Base: runtime/step06e-tsc f2de9280c19f9b427d0cc5142accb0ba940365be; delivery: compiler/program-region-node-check.
Focused selector tests, Node/backend sanitizer and leak oracle, and vet pass; lane results are recorded below.
Applicable cell-drop, acyclic-selection and contextual-adoption mutants fail; requested Node and NodeArray drops survive because neither type exists in fixture 14.
The requested step 24 tree-membership claim remains unproved; it needs a real parser allocation witness or the step 32 tsc.ts build.

Fixture 14 is exactly stage3/fixtures/host/14_getCurrentDirectory.a, SHA256 6ceeaf4f6b75ce26c683602d061d1b64977beb1fb6ef540d9bd572dd98131763. It imports node:fs, node:path, node:os and node:process and instantiates memoize<string>. It declares no parser Node or NodeArray and allocates no parser tree. The four members reported by runtime are closure/cell allocations, not parser nodes. No compiler, selector, runtime or fixture source changed here. No branch was merged and no runtime branch is a push target.

Host member allocation sites, all in stage3/fixtures/host/14_getCurrentDirectory.a:

| Member | Type | Source site |
|---|---|---|
| captured callback cell | (() => string) or undefined | 13:28 |
| captured value cell | string | 14:9 |
| returned memoization closure | () => string | 15:12 |
| input callback closure | () => string | 23:37 |

TestProgramRegionHost14MemberSites reads the selector's programPlan.cells, programPlan.types and programPlan.functions, then checks the effective allocation flags produced from that plan. The input callback's contextual type selects its allocation even though its function declaration flag is false. Both cells and both closures are asserted, with source locations printed. Six selected checker type identities support these four allocation sites; neither of the four allocations is a parser Node or NodeArray. The runtime's original report called the input closure a member; that is confirmed by contextual allocation selection, not by its function flag alone.

TestProgramRegionScoutParseSelectionEmpty reconstructs the scout driver and its dependency modules from 9e062aac7a55e117199c1cc5fbb7a0d59bedf07a. Source bytes and per-file SHA256 are in parse-source.json; only import extensions and reconstructed file names change from .ts to .a. It checks the selector's type, cell, function and class member sets, all empty. It does not run the 77-input parser benchmark again or claim this acyclic tree is a member. The source bundle is a test input, not a cohere submodule copy or a new oracle fixture.

Tests and times: timeout 90 go test ./internal/lower -run '^(TestProgramRegionHost14MemberSites|TestProgramRegionScoutParseSelectionEmpty)$' -count=1 -timeout 60s -v passed. Host leaf 0.35s; scout leaf 8.16s including source reconstruction and loading. ADAMIC_GATE_UNCACHED=1 timeout 90 go test ./internal/oracle -run '^TestProgramRegionHost14AgreesWithNode$' -count=1 -timeout 60s -v passed; leaf 36.35s including native runtime builds. The host prints true followed by true, as Node does, in both backends with ASan/UBSan and leak checks. Counts are allocations 11, frees 7, retains 18, releases 34, peak 8, regions 4, matching runtime's recorded witness. timeout 90 go vet ./internal/lower passed. No oracle fixture was added or changed, so counts.md was unchanged and global counts regeneration was not needed.

Mutants and limits:

- drop-Node removes selected identities displayed as Node from programRegionSelectionMode. Host test passes: there is no matching identity. This survivor proves the supplied fixture cannot hold the requested Node claim.
- drop-NodeArray removes selected NodeArray identities from the same selector. Host test passes for the same reason. This is not a certified check or a retired mutant.
- drop-host-callback-cell changes programMembership.cells output for callback. TestProgramRegionHost14MemberSites fails its four allocation-site assertion. A separate host oracle run is refused during cycle validation; that refusal is recorded but is not used as the proof of the member-count assertion.
- select-acyclic-type selects one otherwise counted candidate. TestProgramRegionScoutParseSelectionEmpty fails with one selected type identity. It never reaches native emission.
- drop-contextual-closure-adoption changes the native emitter's adoption condition from expression.ProgramRegion || target.ProgramRegion to target.ProgramRegion in an overlay only. Node/backend behavior, sanitizers and leak checks pass, then TestProgramRegionHost14AgreesWithNode fails its member assertion: allocations 11, frees 8, regions 3 instead of 4. This proves that assertion can fail without a compiler error or a runtime crash.

Commands, logs and outcomes are in mutants-results.json and logs/. Mutant source files are .go.txt, generated C evidence is .c.txt. run-mutants.py reproduces the selector mutants; the contextual-adoption overlay and exact command are recorded separately. Initial selector test development inspected function flags alone and found only three sites; the corrected test includes contextual allocation selection and finds the fourth. Initial logs are preserved, current pass claims use selection-final.log and host-oracle.log.

Setup: GOPROXY=https://proxy.golang.org|direct; timeout 300 bash cloud/setup.sh succeeded. Timing lines: Go 0.023s, Node 0.023s, markdown 0.066s, submodules 0.068s, clang 0.162s, Go build 47.800s, deferred test binaries 47.994s, cache warm 47.995s, done 48.028s. nproc=5, CPU quota=4. Every test command has a 90-second outer limit and a 60-second Go timeout. No full-package gate ran.

Choice: keep fixture 14 unchanged and do not synthesize Node/NodeArray types, force their membership by name, or present closure counts as parser-tree evidence. A replacement witness must contain actual tsc Node and NodeArray allocations in the owning strongly connected graph. Until that witness or the step 32 build exists, task #wny5r9y is incomplete. These tests cover its genuine host witness and acyclic control only.
