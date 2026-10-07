Built three .a CFG helpers: read, write and loop; four consumers each, twelve prerequisite entries across four rules, zero final blockers; rebased all thirty retained helpers onto current main b8fb957a.
Commits: original published claim cc54b055; original implementation d0f8c162; rebased implementation 3025ecc6410685a74f21682070c0449f68c26fc9; branch codex/lint-helpers-02.
Commands: rebased setup 151s/nproc 5; isolated oracle PASS 25.029s; fresh complete helper oracle PASS 598.813s; inventory PASS 7.990s; uncached input oracle PASS 2.385s; inherited-static oracle PASS 1.623s; vet exit 0.
Mutants: twelve new compiling semantic mutants caught against actual Go on source Node, emitted JavaScript and sanitized native; all ninety prior mutants caught again, one hundred and two total on the rebased source.
Not covered: full repository gate, whole-rule findings/fixes/suggestions, complete CFG builder, arbitrary malformed adapters, callback failures/concurrency or prior external Tailwind live/corpus gaps.

# Landing and ownership

All twenty-seven earlier helpers were pushed and landing-ready at 6435cd02 on origin/main c01907a7, with the complete helper oracle PASS 638.918s and all ninety compiling semantic mutants caught. The branch was clean. A wildcard fetch inspected all nineteen claim files across twenty origin codex/lint-helpers* branches. Every larger concrete symbol was reserved; the original comments bundle and remaining rule-local strict-options label stayed excluded. These three tie the highest available concrete count, four consumers each. Claim cc54b05517821c1c898b3cee34b8537ae24d7cab was pushed before code.

The immediate refresh and final ownership audit find the three complete symbolic claims only on this slot. evidence/claims.log and claims-final.log record the audits. Source d0f8c162b7d95d4c0fecab854d9ee8336699d759 passed the isolated 25.029s suite and complete 644.331s gate on c01907a7. The final fetch revealed main had advanced during that gate to b8fb957aa839a9e8cb0b54279dd9864fa317bd30, adding the inherited static-field fix. The branch's fifty-one commits rebased cleanly. The exact helper-tree diff from d0f8c162 to 3025ecc6 is zero. All thirty helper source hashes remain unchanged; evidence/helper-source-manifest.json records them and evidence/source.log records ancestry/source/lease details.

The full helper oracle and supporting checks were rerun against the new main. The final wildcard fetch confirms main remains b8fb957a and is an ancestor of the verified source. Publication targets only codex/lint-helpers-02, using an exact force-with-lease against the last published cc54b055 because the requested rebase rewrites its history. No main or area branch is pushed and no pull request is opened. Only this slot's claim, batch files and standalone slot02_batch10_test.go change. No shared registration, test harness or protected compiler file is edited.

The main update changes internal/native/emit_objects.go and its existing oracle fixture/count registration. Those changes were accepted in the rebase. No shared lint leak-check replacement or harness/Diagnostic change arrived in this main revision, and none was reverted. No new rule is introduced, so no rule.json kinds, dispatcher, node-kind string relevance comparison or context.report change is needed. No regexp matcher or pattern translation is introduced.

# Actual Go behavior and coverage

read and write invoke a present hook exactly once with the original builder and node pointers, including a nil node and an unreachable current block. Neither filters on node kind or reachability. loop invokes a present hook exactly once only for a nonnil node. All three preserve callback arguments and arbitrary state changes performed by the callback itself. The numeric arena adapter carries original builder/node identity, with -1 representing a nil node. present projects that builder's actual hook registration; it must not be guessed by the caller.

Observed: every consuming rule contributes actual runtime fixture sources, 2,119 captures total. The Go Core and React rule packages pass during capture. The oracle parses 2,077 distinct sources including controls, builds 3,282 graph roots with actual IndexRoots/Build, and observes 10,661 hook-method entries. The methods do not inspect node contents, so method/node-kind/reachability call shapes deduplicate to fifteen, retaining actual parser pointers for six representative node kinds. Every observed call shape is expanded across hook presence, both builder identities and two callback modes. Additional controls cover nil and every representative node kind for all three operations, both reachabilities and both callback modes, 456 total replay cases.

The actual Go replay invokes private read/write/loop methods. Its callback independently checks received builder and node pointer identity, records calls in order, emits into the original current block, and optionally switches that builder to the other block before another emit. The complete builder/block state and dependency trace agree with source Node, emitted JavaScript and ASan/UBSan native. All backends must exit zero without stderr. The final isolated suite, pre-rebase full suite and fresh rebased full suite catch every new semantic mutant.

Temporary overlays insert capture calls at method entry, preserving every original Go method body. No cohere worktree file is edited. Pinned cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db. Expected Want is removed before Adamic receives inputs. Capturing every fixture and graph root includes roots beyond each rule's filtered graph selection; it is helper coverage, not instrumentation of every consumer's exact filtered graph invocation or whole-rule findings parity. Node-content representatives are justified by the helpers' opaque forwarding contract, not an inference that their consuming rules ignore syntax.

Capture artifacts regenerate byte for byte. Source SHA-256 a1693be8bd8b2e385e79206d208166de3608b9d325e7aa8b2e96f22683824edf; coverage SHA-256 06098e3a3b2c055f9f017ed10c62e864bc1fd186fe83c8f4f533cb578fd5d00f. evidence/reproducibility.log records before/after hashes.

Inferred: twelve prerequisite entries are removed across four distinct rules. No rule loses its final listed helper blocker. RULES.md names every consumer. readiness.json subtracts only this slot's thirty retained helpers without presuming other workers' unlanded work. Readiness remains conditional on the frozen inventory's common AST adapter.

# Commands and outputs

All test output goes directly to logs, never through a pipe. Source /workspace/adamic-tools/env.sh before Go commands. Go 1.27.1, clang 20.1.8, Node 24.19.0.

- Initial bash cloud/setup.sh > evidence/pre-rebase-setup.log 2>&1: PASS 61s. Go ready 0s; clang, Node and submodules ready 1s; build cache warm 61s; done 61s on five processors. nproc prints 5.
- python3 .../batch10/testdata/regenerate.py > evidence/regeneration.log 2>&1: all four consumers, 2,119 rows, Go Core/React PASS. Package output is in evidence/capture.log. Second regeneration and source/coverage hash comparison: PASS, identical bytes.
- go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch10$' -count=1 -v -timeout=20m > evidence/isolated.log 2>&1: PASS 25.029s, all twelve new compiling semantic mutants caught.
- Pre-rebase complete helper oracle: PASS 644.331s, all 102 compiling semantic mutants caught; supporting inventory PASS 6.098s, uncached input oracle PASS 2.093s, vet exit zero. Those logs retain the pre-rebase- prefix.
- After rebase, bash cloud/setup.sh > evidence/setup.log 2>&1: PASS 151s. Go, clang, Node and submodules ready 0s; build cache warm 151s; done 151s on five processors.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/helpers-final.log 2>&1: PASS 598.813s, all 102 compiling semantic mutants caught again against current main.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 7.990s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1: PASS 2.385s, all six fixtures, zero cache hits and six probe misses.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^inherited_static_field_read[.]a$' -count=1 -v > evidence/inherited-static-oracle.log 2>&1: PASS 1.623s. The newly landed fixture matches source Node, emitted JavaScript, sanitized native and release native; zero cache hits, three native misses and two Node misses. No new static-field implementation or mutant credit is claimed by this helper worker.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet.log 2>&1: exit zero, empty log.
- git diff --exit-code d0f8c162 3025ecc6 -- stage1/cohere/lint/helpers: exit zero; all helper-tree content unchanged by rebase.
- git merge-base --is-ancestor origin/main HEAD and git diff --check: exit zero.

# Every new mutant and its witness

Each final mutant compiles and finishes without stderr or sanitizer findings on all three Adamic backends. Those backends must agree before an independent Go difference counts. Temporary copies contain mutations; production helpers remain unchanged.

| Mutation | Actual Go witness |
|---|---|
| Read ignores hook presence | Go makes zero calls with an absent hook; mutant invokes the test callback. |
| Read skips nil nodes | Go forwards nil to a present hook; mutant makes no call. |
| Read calls twice | Go records one callback and emit; mutant records two. |
| Read substitutes builder zero | Go preserves builder one; trace and mutated builder/block state differ. |
| Write ignores hook presence | Go makes zero calls with an absent hook; mutant invokes the callback. |
| Write skips nil nodes | Go forwards nil to a present hook; mutant skips it. |
| Write calls twice | Go records one callback; mutant records two with additional state changes. |
| Write replaces node with nil | Go preserves the actual node pointer; trace and event values differ. |
| Loop ignores hook presence | Go skips an absent hook; mutant calls it for a nonnil node. |
| Loop removes the nil-node guard | Go skips nil even with a present hook; mutant invokes the callback. |
| Loop calls twice | Go records one callback; mutant records two. |
| Loop substitutes builder zero | Go preserves builder one; trace and original-builder mutation differ. |

[evidence/mutant-witnesses.log](evidence/mutant-witnesses.log) lists every executed mutant name and its first fresh independent Go mismatch, all 102. Their names are checked byte for byte against the successful pre-rebase suite. The ninety previous definitions remain in earlier slot02/batch reports. Original option/policy and first slot02 tests cover Go/source Node/native; batch2 onward additionally compares emitted JavaScript. No extra emitted-JavaScript coverage is claimed for original tests.

# Limits

The full repository gate, integrated rule findings/fixes/suggestions, complete CFG builder and shared runtime integration remain outside this unit. Numeric/arena/hook-presence adapter validation, arbitrary malformed inputs, callback panic/error behavior, concurrent hook replacement and previously documented unavailable Tailwind live/corpus gates are not covered. Earlier claim and implementation SHAs in reports remain historical pre-rebase identifiers; their evidence is preserved. No shared-harness or compiler blocker was encountered. This unit completes exactly three new helpers, re-greens all thirty on current main and makes no further claim.
