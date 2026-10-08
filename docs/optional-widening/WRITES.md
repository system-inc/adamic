Built checked optional writes on Lane 1's shared views, with reduced sources, loud never traps and conservative proof erasure.
Commits: reads d3226902; write implementation 46a28295; current-main merge 095be737 includes main 48c05d09; final write checkpoint contains this report.
Validation: final uncached gate passed: IR 1.664s, lower 40.482s, fixtures 29.438s, whole Node oracle 289.362s; native 299.695s and fresh 86.083s; setup 469.448s, nproc 5, quota 4.
Mutants: erase without proof, skip shape, stored is fresh, unreduced source, object covariance, nominal slot, stale tag, direct subclasses, never no-op, compound omission, virtual base only and closed numeric enum.
Not covered: complete tsc lowering, class-expression execution, callable/indexed-dictionary view contracts, and macOS; 101 Refused plus 134 NotYet remain in the 391-site schema audit.

## Behavior

A supported widening uses `(*lowering).view` from Lane 1 commit `609ed39`, merged at `3ececcf1`. Reads check absence and the complete supported target contract at the read. Writes through the widened optional use the real object's immutable declared slot certificate, independently of its current payload. The certificate must accept the written expression's static type, including finite literals, undefined, structural object mutability and nominal ancestry. A whole numeric enum retains its open numeric domain.

Both backends evaluate receiver and value in source order and check before changing the slot or its readiness. Native certificates share the object's allocation; successful overwrites do not change their logical contract. JavaScript uses the same contract ids in slot metadata. Actual record own slots and class slots are covered. Unsupported logical types get no certificate and fail closed. Frozen/accessor checks remain independent.

A pinned failure is `adamic: panic: field write failed: property 'y' on Sentinel at sentinel.a:1 has no compatible declared slot`, followed by exit 70. The message names the property, actual class or record type, and write location. Twenty-nine static probes hold source Node observations independently: twenty match Node, and nine pin the complete inserted stop. Sanitized native, release native and generated JavaScript run every probe. Assignment, compound assignment and increment all carry certificates. The reached-never write additionally pins `unreachable expression value at never-write.a:5:6`; source Node continues.

The existing fresh-write analysis now exposes whether the literal has another alias. A literal held in a local, passed, captured or reachable through another object loses that proof. Fresh writes reserve the optional slot under its target contract and erase the semantic check. A dominating nominal tag erases a write only when every possible instantiated runtime subclass has a compatible slot. Calls, callbacks and reassignment kill that fact. Read erasure additionally requires the exact physical representation. An unreachable receiver must have terminal bodies for every possible call target, including virtual overrides. General never traps apply to the source expression itself, not an impossible component nested in an inhabited source type.

The class read-erasure proof considers every loaded class and transitive subclass, including declarations in additional roots and class expressions. This rests on whole-program closed-world compilation. Separate compilation must revisit the rule. The assumption remains beside the rule in optional_widening.go and in REPORT.md. Field-name propagation is conservative; unrelated aliases can retain an otherwise unnecessary check. Class-expression execution still reaches the existing NotYet boundary. The generic subclass and direct/grandchild class read probes remain checked.

Required object writes without the new optional-slot certificate retain Lane 1's explicit source-slot fence. This unit does not silently relax that separate downcast-write policy. The merge preserves main's unique enum-tag, nominal ancestry and generic-argument cast proofs. Main's hidden-optional refusal test is replaced with a required read-check assertion, as the new ruling permits that upcast. Main's validated primitive narrowing helpers are kept out of pointer-only non-null checks.

## Stage 3

The fully composed adaptation tree from coverage commit `90ca91c` yields 247 genuine widening sites in 79 implementation files. Matching exact file, line, column, AST kind and expression against its pinned 391 entries removes 144 locations after source reduction, with no new unmatched locations. None is classified unreachable merely because a selected constituent printed never. The executing nested-component sites therefore leave the widening bucket rather than take exception (b).

Production target-schema admission gives **101 Refused, 134 NotYet, 12 ContractReady**. All 391 rows and input hashes are in [checked-stage3-391.json](evidence/checked-stage3-391.json). This is schema admission, not a claim that twelve complete files compile: whole-program alias admission and unrelated lowering failures remain. The callable and indexed dictionary families require shared view adapters absent from Lane 1; no second mechanism was built.

All stage 3 fixtures were rerun under their TypeScript conditions. The three records that held the original optional refusal now give two Refused and one NotYet: **0 of those 3 compile**. Reference spreads return to the first-spread policy, the optional host method reaches the callable-contract refusal, and structural cache reaches element access. Neither original object record was Compiles before refusal-2. Their Node observations are unchanged. Six other assertion fixtures compiled following the Lane 1 read merge and still agree with recorded Node. The current-main merge required six further exact diagnostic changes; only stage0 fields were rewritten. [Former refusal records](evidence/checked-stage3-former-refusals.json) retain the before/after audit.

Developer tools' six widening lies stop at checked reads, with repaired programs matching Node in both backends, as recorded in CHECKED.md. Function-typed developer programs remain skipped for `codex/refusal-pass-rulings`.

## Validation and mutants

All output went to log files. The census command was:

```sh
OPTIONAL_WIDENING_CONFIG=/tmp/optional-checked-tsc/src/compiler/tsconfig.json \
OPTIONAL_WIDENING_OUTPUT=/tmp/optional-checked-stage3-admission-final.jsonl \
OPTIONAL_WIDENING_ADMISSION=1 go test ./internal/lower \
-run '^TestOptionalWideningCensus$' -count=1 -v
```

It passed in 17.950s. The new write/never probes passed in 21.115s; reference-count regeneration passed in 45.773s. Setup on the merged checker pin passed: Go 0.060s, Node 0.051s, Markdown 0.212s, clang 0.415s, submodules 4.012s, build 469.118s, warm 469.409s, done 469.448s. Environment `/workspace/adamic-tools/env.sh`, Go 1.27.1, Node 24.19.0, clang 20.1.8, nproc 5, cgroup quota 4 CPUs. No Go-module fetch failure occurred; setup contains main's proxy fallback.

The mutant runner is [write-mutants.py](evidence/write-mutants.py). It rejects build, clang and sanitizer failures, restores each source in finally, and checks the intended semantic/proof failure. Erase-without-proof and shape omission change the pinned write failure; treating a stored literal as fresh fails the alias proof; omitting reduction installs an unnecessary view; covariance permits alias corruption and only traps at a later read; ignoring nominal identity accepts the wrong object; retaining tag facts across a call erases an unproved write; direct-subclass-only misses the grandchild; never no-op prints continued; compound omission delays the stop to a read; inspecting only a virtual base trusts an override that returns; closing the numeric enum incorrectly refuses a permitted write. Two draft mutants generated invalid C and were discarded; corrected anchors and representation-preserving replacements are the counted evidence.

The final restored command was:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/ir ./internal/lower ./stage3/fixtures ./internal/oracle -count=1 -timeout 30m > /tmp/optional-writes-last-gate.log 2>&1
go vet ./internal/fresh ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./stage3/fixtures ./internal/oracle > /tmp/optional-writes-last-vet.log 2>&1
```

Exit 0: IR 1.664s, lower 40.482s, fixtures 29.438s, whole uncached Linux Node oracle 289.362s. Vet, gofmt and whitespace logs are empty. The preceding broader command also ran fresh (86.083s), native (299.695s) and JavaScript (no tests). Its IR call-reader and required-object-fence failures were repaired; the entire IR and lower packages were re-greened, then the final command above re-ran the complete oracle and fixtures. No native backend or fresh-analysis source changed after those packages passed. The full repository gate and macOS are not claimed.

All twelve final mutants were caught without build or sanitizer failures and restored before the last gate. The per-mutant results and logs are under evidence/. Main was fetched again and remains `48c05d091f0a43c31cbe051b1d6578d99eeedf19`, already merged. Pushes are only to this worker's `codex/optional-widening-2` branch; no PR is opened.
