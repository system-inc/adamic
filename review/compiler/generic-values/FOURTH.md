Built and pushed the ruled higher-rank refusal unit #td6c9vy; completed census and admission-delta evidence #25fpja8 toward step 16.
Commits: d34d7634 (refusal), 5511dac3 (binder priority), 62ede3ef (stored-slot boundary); this evidence commit changes only docs/review.
Checks: full internal/lower passes 44.100s; focused controls 0.387s; TestCallTargetReaders 22.760s; counts refresh 87.426s; full admission proof passes 94.879s (1,103 programs, five admissions, none omitted).
Mutants: false results/identity, alias escape/readiness, missing rank/nested binders, traversal/priority, emitted comparer, IR/loader output, source hash/root coverage all caught; details below.
Uncovered: explicit instantiation expressions, nested generic callable escapes and unknown callable descriptors remain NotYet; overloaded contexts without one concrete signature remain refused; this checker-rejected census cannot certify whole tsc or retired bytes.

The implementation sequence is 000991f1 comparer/default values, e2dca338 immutable declaration aliases, then the three higher-rank commits above. d34d7634 and both corrective commits each passed their own checks and were pushed to compiler/generic-values. The final compiler revision is 62ede3ef50b7822cc2d072b4fa6ab5bbeb53b10b, on requested dependency 50654a407f9bcc3a7fb821da88088b9d99e6a728. No other worker branch was merged. The admission tool/generator comes from detached compiler/admission-delta at 543925aa5e950830507f6907f70e32d125beef11, using the same cohere pin, and is not merged.

The supplied ruling is marked ruled in docs/generic-function-values.md; its higher-rank subset refusal is recorded in docs/0.1.md. No task-thread reader was available, so the supplied ruling text is the authority. Const aliases preserve binders for direct calls and still check readiness; escapes specialize only from a concrete contextual slot. Adapter identity is source declaration identity across specializations, compatible with represented V4 views. An argument/result type graph is not the binders of the containing monomorphic callable; its own slots, annotations and escapes receive their own checks. Finite traversal of actual stored object/container slots remains conservative, explicitly refusing an incomplete graph with no known rank signature. Known free binders take precedence over that bound.

Census observations

The guarded full latent census uses all 79 byte-identical compiler source files from the original step 16 ledger, reconstructed from pinned TypeScript 050880ce and stage3 adaptations at 8cb5e7c1. Base/head load the same complete compiler directory, with 324 identical checker diagnostics. Only the 26 surveyed roots are selected for attempts; every root identity, status and available body span matches the historical ledger and dependency run. The wrapper reuses the existing scratch census/refusal/statement rewriters, with two newer outer metadata hooks recorded before their unchanged refusal visitor. It restores dependency lower/load sources for the base overlay and blanks newly added lower files. Production Load/Lower and usable IR output are disabled. This is isolated-lowering observation, never backend execution of checker-rejected tsc.

The family changes from 26 roots/14 sites to 11 roots/eight sites. Seven old positions lose the observed generic-value blocker: core.ts:220, core.ts:2378, core.ts:2442, factory/parenthesizerRules.ts:676, program.ts:1076, sourcemap.ts:821 and transformers/classFields.ts:2071. A later parenthesizer site at line 691 is newly observed and lacks one concrete contextual signature. All eleven core.ts:220 roots pass that value site. Three standalone default roots still lack a concrete outer caller; three path roots lack value-site context; Debug's generic stack marker remains unfixed; enter's identity-only use has no slot; the state-machine slot names TOuterState, TState and TResult in its Refused diagnostic.

Seven roots have no observed isolated-lowering blocker, not seven certified whole programs. Both runs retain ten identical computed-property module-scan panics and 29 identical isolated-statement panics, plus other recovery boundaries that can hide children. The 4,031 historical hidden bytes receive no retirement credit. census-comparison.md lists all 26 rows; census-comparison.json retains every selected-root lowering finding; census-base/head.jsonl.gz retain all raw records and metadata. The parser verifies source bytes, exact diagnostic headers, complete root coverage, body/status metadata and unchanged recovered panic boundaries before computing counts.

Admission proof

The complete gate reports verdict pass, complete_corpus true, 1,103 classified programs: 964 accepted by both, 134 refused by both, five newly accepted, zero new refusals or compiler failures. Budget zero selects every admission; sampling_size five and omitted zero. The five are 01_comparer.a, 02_index_default.a, 03_utility_default.a, 04_alias.a, and the unchanged generics/04_function_value.a now admitted by this compiler. Every admission matches source Node, JavaScript and native stdout/exit, with empty stderr. The normal local oracles additionally check sanitizers and leaks. The manifest extends the pinned official generator with all 20 .a files in generic-values/ and generics/, including the refused higher-rank control. Git verifies every program blob; the result pins compiler/generator/manifest hashes. admission_manifest.py reproduces the extension from Git.

The proof's compiler and program source revision is 62ede3ef. This final evidence commit changes only docs/review and adds no program source or compiler code, so it preserves the admission set, runtime outputs and fixture corpus proven at that revision. Evidence hashes cannot include their own future commit hash; the recorded source SHA is intentional.

Mutants and catchers (all actually run)

| Mutant | Catcher and observation |
| --- | --- |
| False specialized comparer body in each of the three comparer/default reductions | TestGenericValueWrongResult, TestGenericValueIndexWrongResult, TestGenericValueUtilityWrongResult: Node stdout differs in native and JavaScript, with valid exit and no sanitizer/leak failure |
| Distinct source identity per adapter | TestGenericValueWrongIdentity: equality, Map/Set and array output differs from Node in both backends |
| Alias escapes into a polymorphic box slot | TestGenericValueAliasEscapeMutant: source Node prints 4 x, escape is refused with a concrete-slot fix |
| Remove alias readiness | TestGenericValueAliasEarlyRead with recorded overlay: early/4 becomes 4/4 in both backends and the test fails |
| Disable higher-rank collector | Four TestHigherRank leaves fail: missing target binders/path or unintended admission, 0.253s |
| Declared-default instance returns 17 | TestGenericValueDeclaredDefaultWrongResult: both backends disagree with Node, valid execution, 1.16s |
| Restore traversal-bound priority over known binders | TestHigherRankBindersBeforeTraversalLimit: box.run's U/V hidden by generic traversal refusal; expected failure 0.072s |
| Traverse monomorphic callable argument/result graphs | TestGenericValueRecursiveArgumentContract: concrete admission incorrectly refused; expected failure 0.041s |
| Remove nested FunctionType check | TestHigherRankNestedCallableContract: planted run.poly<U> is admitted; expected failure 0.096s |
| Emitted comparer instances both return false | Actual admission-delta gate verdict fail, 5.922s: JavaScript stdout differs, Node/native agree, every execution exits zero; filtered diff run is only the mutant, not the complete green proof |
| Measurement returns non-nil IR | Existing census output guard fails with measurement returned usable IR |
| Production loader exposes a measurement program | Existing census output guard fails with measurement loader exposed an output program |
| Change one source hash | Census parser rejects source bytes differ from historical ledger |
| Remove one root | Census parser rejects root coverage changed |

Commands and evidence

All command output is redirected to logs; Go commands have -timeout and/or an outer hard timeout. No full gate or whole oracle package was run. Full internal/lower was explicitly requested. Each newly added test leaf is below 60s; new recursive comparer control 0.37s, nested-rank refusal 0.04s and bounded-object binder refusal below 0.1s. Other individual durations are recorded in FIRST/SECOND/THIRD reports and their verbose logs. Setup completed in 39.279s with GOPROXY=https://proxy.golang.org|direct; nproc is five with a four-CPU quota. Setup timing lines are in setup-turn2.log.gz. Physical pinned Node declarations were installed with npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund after a cache symlink failed declaration identity.

Final compiler checks:
- timeout 120 go test ./internal/lower -timeout 90s
- timeout 120 go test ./internal/lower -run '^(TestHigherRank|TestGenericValueRecursiveArgumentContract)' -timeout 90s -v
- timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s
- timeout 120 go vet ./internal/lower
- prior final counts: go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 210s -args -update-counts (87.426s, no rank fixture row because it is refused)
- final oracle/mutant selectors and overlays are recorded in THIRD, THIRD-followup, THIRD-contract and their logs; no permanent fixture was added in the corrective commits.

Census reproduction:
- reconstruct the pinned adapted compiler source and verify all 79 ledger hashes;
- timeout 120 python3 review/compiler/generic-values/census_overlay.py REPOSITORY SCRATCH_OVERLAY
- timeout 120 go build -buildvcs=false -overlay=SCRATCH_OVERLAY/overlay.json -o CENSUS ./stage3/census/latent/tool
- timeout 600 env LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 CENSUS ADAPTED/src/compiler OUTPUT_JSONL
- repeat with the dependency lower/load source view, as implemented by census_overlay.py;
- timeout 60 python3 review/compiler/generic-values/census_compare.py ADAPTED BASE_JSONL HEAD_JSONL OUTPUT_JSON
- timeout 180 python3 stage3/census/latent/audit_output_guards.py REPOSITORY HEAD_OVERLAY SCRATCH_GUARD_MUTANTS

Admission reproduction:
- timeout 60 python3 review/compiler/generic-values/admission_manifest.py REPOSITORY 62ede3ef > MANIFEST
- timeout 600 /tmp/generic-values-admission-delta --base 50654a40 --head 62ede3ef --manifest MANIFEST --manifest-generator cloud/admission-corpus/manifest.py --manifest-generator-revision 543925aa --json --workers 4 --compile-timeout 45s --timeout 10s
- mutant: same source revisions/manifest, --corpus-filter diff, real cached lowering companions, and admission_wrong_result.py as --head-binary; GENERIC_VALUES_HEAD_COMPILER points at the real head compiler. The wrapper changes only the two emitted comparer returns and delegates every other command.

The first admission checkout failed because /tmp lacked space. Its log is preserved. Remove only this task's interrupted checkout and put TMPDIR on /workspace/generic-values-proof-tmp; the complete proof then passes. An early census attempt hit its 300s bound after 25 roots; the final complete run uses a watched 600s bound. No partial run is reported as complete.

Lane checks after commits d34d7634/5511dac3/62ede3ef pass in 15.6s/14.4s/15.4s: gofmt/tools on 285 Go files, t.Parallel on 23 packages, .a-check 100 programs, vet 23 packages. Earlier first-unit lanes skipped vet after their short bound; explicit vet across lower/ir/javascript/native/oracle passed. The final evidence lane output and push SHA are reported in the final response.
