Built three .a helpers: CallExpressionSource (five consumers), MatchIgnoringCase and HasAttributeNamed (four each); thirteen prerequisite entries, one final blocker removed.
Commits: published claims d9b239a, ownership evidence d82cca3; implementation bf389ea31efb6b8001afbd75612853e89498acad, expanded queries 3956633c43de66ef8c7ee131801ade0c3c5a5e5d; branch codex/lint-helpers-02 on current main e8ba3d5.
Commands: setup 41s/nproc 5; isolated helper gate PASS 88.831s; complete helper gate PASS 477.730s; final expanded-query gate PASS 81.783s; inventory PASS 4.365s; uncached input oracle PASS 1.561s; vet exit 0.
Mutants: eleven new compiling semantic mutants caught against actual Go on source Node, emitted JavaScript and sanitized native; all fifty-three prior mutants caught again, sixty-four total.
Not covered: whole-rule findings/fixes/suggestions, full repository gate, arbitrary malformed AST adapters and invalid UTF-8; common AST adapter readiness remains conditional.

# Landing and ownership

Before claiming this batch, the previous eighteen helpers were published at 402a608 on current origin/main e8ba3d5 with complete helper oracle PASS 395.124s and all fifty-three compiling semantic mutants caught. The branch was clean and main was its ancestor. An explicit wildcard fetch inspected all seventeen claim files across eighteen origin codex/lint-helpers* branches. The original comments bundle in shared HELPERS.md and the rule-local strict-options ledger entry were excluded. CallExpressionSource had the highest unclaimed concrete consumer count, five; the JSX helpers tied the next count, four. Claims d9b239a were pushed before writing code.

A post-publication fetch exposed concurrent slot 05 claims for CallExpressionSource and HasAttributeNamed. Slot 02 d9b239af40dcf9632c3d55b2bbfb4b3474454ecb was committed at 03:54:07 UTC; slot 05 f846de1f640356fda6f11d263dfe3cf4c8ec922f at 03:54:09 UTC. Under the established earliest-claim rule, slot 02 retains both. This evidence was published separately in d82cca3. The final refresh at slot 05 69bdfe7 confirms its explicit withdrawal of both later overlapping claims. MatchIgnoringCase has no competing claim. No fourth helper is reserved. The final claim refresh is recorded in evidence/claims-final.log.

Only the owned claim, batch directory and standalone slot02_batch7_test.go change. Shared registration, shared harness and protected compiler files are untouched. Only codex/lint-helpers-02 is pushed; no main or area branch and no pull request.

# Behavior and external evidence

CallExpressionSource preserves Go's import/import.defer predicates, require's exact single argument, string/no-substitution-template restriction and separate source-presence flag. import.source, computed arguments, ordinary member calls and callee parentheses decline. Empty literal sources remain present. Dynamic imports may contain additional arguments. Decoded NUL, newlines and supplementary characters compare as decimal UTF-16 units.

MatchIgnoringCase implements Go strings.EqualFold with rune decoding and a sorted Unicode simple-fold table generated from Go unicode.SimpleFold. It performs no multi-character expansion or Unicode normalization. The generator visits every Unicode scalar, producing all 1,512 noncanonical mappings for Unicode 17.0.0. The test regenerates the complete mapping and production table in temporary files; table and version drift reject the run. Actual cohere MatchIgnoringCase, rather than the generated table, decides expected verdicts. Kelvin K, long s, sigma cycles, supplementary Deseret letters, lengths, embedded NUL/newlines, dotted/dotless I, sharp s, ligatures and canonical-normalization differences are covered.

HasAttributeNamed reuses the delivered AttributeName helper. Nil/kind/list guards, nil property/name edges, empty identifiers, spreads and namespaced names are exercised. Initializers do not affect presence. Match arguments and ordered callback trace must agree with real Go. Exact, case-insensitive, always-false and stateful second-call matchers expose order and short-circuit behavior independently of a pure comparator.

Observed: all thirteen consumers have actual runtime captures. The fixture contains 578 rule/file/source rows, 522 distinct parsed sources including controls, 10,571 projected nodes, 13,880 folding/name pairs and 623 attribute targets. The final targeted gate tests twenty-one requested names in each of four matcher modes, including every literal query used by the consuming rules. The external Go oracle calls the actual exported helpers at pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. No helper implementation is replaced. Temporary capture overlays observe testing entry points without editing the cohere worktree. Expected Want is removed from the Adamic runtime input.

The Next, Nexus and React Go rule packages all pass during capture. This is an external Go rule-package gate; the new Adamic ports are verified as helpers, not integrated native rule implementations. Capture requires every consumer and a zero Go test exit. Sources, coverage, generated fold JSON and the .a table regenerate byte for byte; SHA-256 evidence is in reproducibility.log. A final capture rerun with the strict zero-exit check also passes.

Inferred from the frozen ledger: thirteen prerequisites are removed across thirteen distinct rules. CallExpressionSource removes the final listed helper blocker for nexus/import-no-forbidden-source. The other twelve retain blockers. RULES.md lists every consumer, and readiness.json subtracts only this slot's twenty-one delivered helpers, without assuming other workers' branches have landed. This is conditional helper readiness under the inventory's common AST adapter assumption.

# Commands and outputs

All test output is written directly to logs, never piped. Source /workspace/adamic-tools/env.sh before Go/build commands. Evidence paths are relative to this batch directory.

- bash cloud/setup.sh > evidence/setup.log 2>&1: PASS. Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 41s, done 41s on five processors. nproc prints 5. Go 1.27.1, clang 20.1.8, Node 24.19.0.
- python3 .../batch7/testdata/regenerate_folds.py > evidence/fold-generation.log 2>&1: Unicode 17.0.0, 1,512 mappings.
- python3 .../batch7/testdata/regenerate.py > evidence/regeneration.log 2>&1: 578 sources, all thirteen consumers; Next, Nexus and React packages PASS. Final strict capture in evidence/final-regeneration.log also passes; package output is in capture.log.
- Capture/folding regeneration and byte equality: evidence/reproducibility.log, PASS. Source SHA 689e4c22bd7ee0aad75221236bdae94c5e89a7e26b9272a9db419732901e79a3; coverage SHA 7f2f29fd021144b5e56554eb93782ba5b986af9c8b402614b0236fb890562c98; fold JSON SHA 976ea014910efd3ee57f739deb634668a062c9f2fc3f1a973f197a09d8d83180; generated .a helper SHA 32fd95fe4a11d34b1f53cfe7be4ed2999d66af0f5211de905779018d12159dfa.
- go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch7$' -count=1 -v -timeout=20m > evidence/isolated.log 2>&1: PASS 88.831s, all eleven new semantic mutants caught.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/helpers-final.log 2>&1: PASS 477.730s, all sixty-four semantic mutants caught; includes the added full-scalar table drift verification.
- Final consumer-query expansion: go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch7$' -count=1 -v -timeout=20m > evidence/expanded-consumer-queries.log 2>&1: PASS 81.783s; real Go, source Node, emitted JavaScript and sanitized native agree on all twenty-one queries, and all eleven new semantic mutants are caught again. Expanded vet in evidence/vet-expanded.log exits zero.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 4.365s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1: PASS 1.561s, all six fixtures, zero cache hits and six probe misses.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet.log 2>&1: exit 0, empty log.

# Every new semantic mutant

All mutants use temporary source copies. Source Node, emitted JavaScript and ASan/UBSan native must compile and execute successfully, exit zero without stderr, and agree with each other before a difference from actual Go is credited. Compile failures, crashes and sanitizer reports are not mutant kills.

| Mutation | Actual Go witness |
|---|---|
| Require accepts multiple arguments | require('one','two') declines in Go; mutant returns one. |
| Treat import.source as deferred import | import.defer('deferred') is present in Go and absent in mutant. |
| Drop no-substitution template sources | import with a plain template returns template in Go; mutant declines. |
| Empty literal source means absent | Go returns found=true for import(''); mutant returns false. |
| Drop Kelvin simple-fold mapping | Go folds K and Kelvin sign together; mutant differs. |
| Drop long s simple-fold mapping | Go folds S and long s together; mutant differs. |
| Drop supplementary Deseret mapping | Go folds the supplementary capital/small pair; mutant differs. |
| Ignore unequal string lengths | One side's extra suffix makes Go false; mutant returns true. |
| Reverse match callback arguments | Candidate/wanted traces differ on a key attribute and empty target. |
| Reverse attribute iteration | Go traces x before key; mutant traces key before x. |
| Call matcher for unnamed attributes | Spread/name-declined properties add an extra call and change presence. |

The fifty-three previous semantic mutants are rerun in the complete package. Their individual mutations and Go witnesses are documented in ../landing/REPORT.md and batch2 through batch6 reports; helpers-final.log records every mismatch again. They include all four inherited options/policy checks, three first-batch name/cache checks, four reader/splitting/namespace checks, five factory/math checks, five prefix/underscore checks, fourteen lookup/root/sort checks and eighteen stylesheet checks. Original first-batch comparisons cover Go/source Node/native; later batches also cover emitted JavaScript. No additional emitted-JavaScript coverage is claimed for the original tests.

# Limits and corrected attempt

An initial capture invocation retained the previous batch's environment-key spelling and produced no captured rows. The completeness assertion rejected it. The slot-local key was corrected, and subsequent captures include all thirteen consumers with passing Go packages. No compiler or shared harness workaround was needed. Final review added explicit download, jsx, is and dangerouslySetInnerHTML query strings after the complete package gate. Production helpers are byte-identical to that passing gate. The affected batch is retested against the expanded actual Go queries; unrelated tests are not repeated.

No full repository gate, whole-rule findings/fixes/suggestions parity, arbitrary AST graph coverage or adapter implementation is claimed. Caller-owned projections must preserve exact Go kinds, text, edges and dense argument/property lists. Invalid arena edges refuse. Go's manually malformed require node with a nil argument list panics before its later helper guard; that shape is outside the valid parser-node contract. Arbitrary invalid UTF-8 byte strings, independent JSON normalization of unpaired UTF-16 surrogates and panicking matcher callbacks are outside the tested projection. Unicode tables are tied to the explicit Go version and refuse drift in validation. The twelve still-blocked rules keep their residual dependencies in the handoff ledger.
