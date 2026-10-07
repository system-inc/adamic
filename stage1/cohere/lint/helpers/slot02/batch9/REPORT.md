Built three .a CFG helpers: setCycleBarrier, linkWithCycleBarrier and link; four consumers each, twelve prerequisite entries across four rules, zero final blockers.
Commits: claim 194914c7; implementation 2d341265dca0b71ccd12d2c3030a49fedf70b83c; branch codex/lint-helpers-02 on current origin/main c01907a7.
Commands: setup 53s/nproc 5; final isolated oracle PASS 82.920s; complete helper oracle PASS 638.918s; inventory PASS 11.514s; uncached input oracle PASS 4.098s; vet exit 0.
Mutants: thirteen new compiling semantic mutants caught against actual Go on source Node, emitted JavaScript and sanitized native; all seventy-seven prior mutants caught again, ninety total.
Not covered: full repository gate, whole-rule findings/fixes/suggestions, arbitrary malformed adapters, externally aliased slice headers, the external appendSuccessor implementation or prior Tailwind live/corpus gaps.

# Landing and ownership

All twenty-four earlier helpers were pushed and landing-ready at 3c541615 on current main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. The fresh complete helper oracle passed 493.875s with all seventy-seven compiling mutants. The branch was clean. A wildcard fetch inspected all nineteen claim files across twenty origin codex/lint-helpers* branches. Every larger concrete symbol was reserved; the original comment bundle and remaining rule-local strict-option label stayed excluded. These three tie the highest available concrete count, four consumers each. Claim 194914c7 was pushed before helper code.

The immediate post-publication refresh and final wildcard refresh find the retained symbols only on this slot. evidence/claims.log and claims-final.log record the audits. Main remains c01907a7 and is an ancestor of the verified implementation. Only this slot's claim, batch files and standalone slot02_batch9_test.go change. All twenty-four prior helper hashes remain unchanged; evidence/helper-source-manifest.json records all twenty-seven implementations. No shared registration, test harness or protected compiler source changes. No main/area branch is pushed, and no pull request is opened.

No new rule is introduced, so there is no rule.json kinds registration, global dispatch, kind-string relevance test, context.report change or Diagnostic-model integration change. No regexp matcher is introduced. The announced shared allocator-aware leak-check replacement has not arrived in this main base; no shared leak check was reverted or altered.

# Go behavior and coverage

setCycleBarrier preserves the nil-source and negative-index guards. A nil barrier slice stays nil on false, otherwise allocates false values through successor length. A present slice grows to successor length with false values and preserves earlier entries, then the requested index is replaced. The numeric arena adapter preserves block identities and explicit nil-versus-present barrier storage.

linkWithCycleBarrier returns -1 without dependency calls for either nil endpoint. Otherwise its returned index is the old successor length, before append. It calls appendSuccessor and then setCycleBarrier with unchanged identities/index/value, and only a reachable source sets the target incoming flag to true. Self and duplicate edges remain present; existing target incoming true survives an unreachable source. link forwards both original block identities exactly once with barrier false and discards the return value.

appendSuccessor remains an explicit externally owned dependency, not a fourth helper implementation. The barrier setter is also injected into the linking helper and wired to the delivered setter in the isolated suite. The Go replay initializes the two inline successor slots before invoking the actual methods. Capacity and externally retained Go slice-header aliases are outside this private-state projection. Valid setter indices must fit the resulting slice; invalid arena/numeric/index adapters are outside parity coverage.

Observed: all four consumers provide 2,119 runtime source captures. The actual Go Core and React rule packages pass during capture. The oracle parses 2,077 distinct sources, including controls, and builds 3,282 graph roots using actual IndexRoots/Build with empty hooks. Method-entry snapshots deduplicate to 226 observed states. Explicit nil/self/duplicate/size/index/storage/reachability/incoming controls produce 6,658 total replay cases.

This is graph-helper coverage over every consuming rule fixture, including roots beyond each rule's filtered listener selection. It is not instrumentation of each rule's exact filtered graph invocation or whole-rule findings parity. Capture/trace calls are inserted at method entry by a temporary overlay; the original Go method bodies remain unchanged. Replays call the actual private methods. Pinned cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db. No cohere worktree file is edited. Expected Want is removed before Adamic reads the input corpus.

Actual Go, source Node, emitted JavaScript and ASan/UBSan native agree on return values, ordered dependency traces, successor identities, reachability, incoming flags, barrier presence and contents. All backends must exit zero without stderr. The final isolated gate and complete package gate both catch every new semantic mutant.

Capture artifacts regenerate byte for byte. Source SHA-256 a1693be8bd8b2e385e79206d208166de3608b9d325e7aa8b2e96f22683824edf; coverage SHA-256 69ca088dbc27bbacf7c6af84994e4ce88468c3dc956229c22a01f274c57a15f6. evidence/reproducibility.log contains before/after hashes.

Inferred: twelve prerequisite entries are removed across four distinct rules. No rule loses its final listed helper blocker. RULES.md names every consumer. readiness.json subtracts only this slot's twenty-seven retained helpers, without presuming other workers' unlanded prerequisites. Readiness remains conditional on the frozen inventory's common AST adapter.

# Commands and outputs

All test output goes directly to logs, never through a pipe. Source /workspace/adamic-tools/env.sh before Go commands. Go 1.27.1, clang 20.1.8, Node 24.19.0.

- bash cloud/setup.sh > evidence/setup.log 2>&1: PASS. Go ready 0s; clang, Node and submodules ready 1s; build cache warm 53s; done 53s on five processors. nproc prints 5.
- python3 .../batch9/testdata/regenerate.py > evidence/regeneration.log 2>&1: all four consumers, 2,119 rows, Go Core/React PASS. Ordinary package output is in evidence/capture.log. Second regeneration and source/coverage hash comparison: PASS, identical bytes.
- go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch9$' -count=1 -v -timeout=20m > evidence/isolated-final.log 2>&1: PASS 82.920s; all thirteen compiling semantic mutants caught.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/helpers-final.log 2>&1: PASS 638.918s; all ninety compiling semantic mutants caught.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 11.514s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1: PASS 4.098s; all six fixtures, zero cache hits and six probe misses.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet.log 2>&1: exit zero, empty log.
- git merge-base --is-ancestor origin/main HEAD and git diff --check: exit zero.

# Every new mutant and its witness

Each final mutant compiles and finishes without stderr or sanitizer findings on all three Adamic backends. Those backends must agree before an independent Go difference counts. Temporary copies hold mutations; production helpers remain unchanged.

| Mutation | Actual Go witness |
|---|---|
| Allocate a nil barrier slice for false at a valid index | Go retains nil storage; mutant changes the presence flag. |
| Pad a growing barrier slice with true | Go preserves false padding; untouched padded entries differ. |
| Negate the stored barrier bit | Go stores true for the observed edge; mutant stores false. |
| Refuse index zero alongside negative indices | Go allocates and sets the zero edge; mutant leaves storage nil. |
| Return old edge index plus one | Go returns the exact old successor length; return values differ. |
| Propagate incoming from unreachable sources | Go propagates only from reachable sources; incoming state differs. |
| Clear target incoming for a reachable source | Go sets incoming true; mutant leaves false. |
| Negate the forwarded barrier | Go preserves the supplied bit in the dependency trace and storage. |
| Plain link forwards true | Go forwards false; trace/storage differs. |
| Nil endpoint returns -2 | Go returns -1; return values differ with no dependency calls. |
| Plain link delegates twice | Go records one with/append/barrier sequence; mutant records two. |
| Barrier callback precedes append for safe false/nil storage | Go trace records append first; mutant reverses the order while finishing. |
| Plain link replaces target with source | Go preserves target identity; dependency trace and edges differ. |

[evidence/mutant-witnesses.log](evidence/mutant-witnesses.log) lists all ninety executed mutant names and their first independent Go mismatches. The seventy-seven previous definitions remain in the prior slot02/batch reports. Original option/policy and first slot02 tests cover Go/source Node/native; batch2 onward additionally compares emitted JavaScript. No extra emitted-JavaScript coverage is claimed for original tests.

# Superseded work and limits

One draft nil-allocation mutant removed the guard unconditionally and crashed at invalid indices. Its failed run is preserved in evidence/superseded-crashing-mutant.log and receives no semantic credit. The final mutant changes only valid-index behavior and finishes on every backend. The first corrected ten-mutant suite passed 52.087s in evidence/isolated.log; the final suite adds explicit with-call tracing and three further mutants, then passes 82.920s and the complete 638.918s gate. No production implementation changed after its green final isolated run.

No full repository gate, integrated rule findings/fixes/suggestions, complete CFG builder or externally owned appendSuccessor storage optimization is claimed. Arbitrary malformed/numeric/arena adapters, out-of-range Go panics, externally aliased Go slice headers and full dataflow/path analysis remain outside this unit. Previously documented Tailwind live/corpus gaps remain outside coverage. This unit completes exactly three new helpers and makes no further claim.
