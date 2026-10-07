Rebased the owned rule branch onto current origin/main 39638d9e278d38bb5aeae887f46d55a70e47aaad, preserving all owned source bytes and existing foundation merge resolution.
Pre-rebase pushed tip c485ad2cbeeca456afa2d11cb49888952915d2db; rebased implementation 72635a73e04dfaba2f523e0c054e180c4f7e882b; sibling helpers freshly green and pushed at 77e7f673f.
Owned package PASS 45.425s, vet clean; aggregate PASS 520.492s: 3,016 supported cases on 314 compiler/stage1 files, 118,907,067 identical Go/source Node/emitted JS/sanitized-native finding/fix bytes.
All ten compiling aggregate mutants and all 17 named-listener mutations caught only by actual-Go comparison on every backend; resolved-path mutant also compiles and is caught on all three.
Uncovered: eight partial unregistered adapters and prior parser/JSX/filename exclusions; default shared gate still fails at profile_test.go:32:23, no full-repository gate or fresh throughput claim.

Current-main changes from b8fb957aa affect stage3 and landing documentation, not lint, compiler or oracle files. The six foundation merge conflicts preserve the already-reviewed existing resolution; no new shared-harness logic was introduced. All owned rule content was byte-equal to its previous pushed tip before adding this evidence. Compatibility validation uses isolated scratch harness f4d98cab, with current-main internal/, cmd/adamic and oracle/node.mjs restored and verified equal. No scratch compatibility edits are committed or pushed. Shared test compilation still refuses: cannot range over portFiles, a func(t *testing.T) []string; the exact default-gate log is preserved.

Commands, with /workspace/adamic-tools/env.sh sourced:
```
go test ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import -count=1 -v -timeout=15m > /tmp/wave12-39638-rules.log 2>&1
go vet ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import > /tmp/wave12-39638-rule-vet.log 2>&1
go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -timeout=15m > /tmp/wave12-39638-default.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_STAGE1_SOURCE=/workspace/scratch/wave12-landing-rules/stage1 ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestLandingExistingCandidates$' -count=1 -v -timeout=25m > /tmp/wave12-39638-isolated.log 2>&1
```
The last command runs only in isolated scratch compatibility worktree. All original and supported-corpus inputs use the pinned actual Go oracle. Historical parser and option/filename boundaries from B8FB-LANDING.md remain; green observations never include the eight unregistered partial rules as completed finding ports.

Ten aggregate mutants: orm_bare_decorator_ignored, boolean_outcome_companion_ignored, next_document_import_path_exemption_ignored, this_alias_allowed_name_ignored, optional_assertion_removes_operand_byte, constraint_suggestion_deletes_name, last_internal_owner, alias_prefix_boundary, whitelist_widened, role-suffix regex disabled. Every named declaration first kind changes to valid Unknown, compiling and exiting cleanly before real-Go listener comparison catches it. Prior valid descriptor mutation evidence is preserved, not claimed newly rerun here. Path resolution mutant drops directory resolution and is caught by actual-Go comparison. No mutant is credited for compilation or sanitizer failure.

Harness 41eb6eab2 and its DEDUP_LEDGER.md were read before future harness integration. It is not yet an ancestor of current main. When rebasing onto it, retire this branch's no-this-alias, no-non-null-asserted-optional-chain and no-unnecessary-type-constraint copies in favor of wave1-07, wave1-05 and wave1-08 respectively. Keep the named winning Next document-import and ORM copies. No batch-only rule is assigned specifically to wave1-12. No shared compiler/harness/generator file is edited for these helpers. Neither main nor area branches is pushed.

Setup Go/clang/Node/submodules ready 0s, cache warm and total 122s, nproc 5 (four-core quota). Rebase initially failed because the workspace disk was full; deleting only regenerable Go cache artifacts recovered 2.6 GB, and both rebases then succeeded. Helper suite fresh PASS 125.005s and final vet clean, all nine delivered helpers and 24 controls. Four CFG consumers each lose two new dependencies, zero final blockers. Exact counter remains blocked at bigint-return lowering; other unclaimed helpers remain.

---

Built: named listener declarations and literal boolean-outcome suffix regex retained, with the rule branch rebased onto current origin/main b8fb957aa839a9e8cb0b54279dd9864fa317bd30.
Commits: previous pushed rule 0ea61b9af7aede044c6d8f82eea96cd177ca32de; rebased implementation 79ed6a17bc1a17f445cb4daad5c56471dbd73eb7 before this evidence commit; helper sibling pushed 876616bf89a7019967b817bf135cafb66cd08a86.
Checks: owned rule package PASS 56.795s and vet clean; supported-domain aggregate PASS 519.733s, 3,016 cases over 314 compiler/stage1 sources, 118,907,067 identical actual Go/source Node/emitted JS/sanitized-native finding/fix bytes.
Mutants: all 17 named-domain Unknown controls, the valid Unknown descriptor control, 10 aggregate semantic controls including suffix regex disabled, and the resolved-path mutation are caught again by their actual-Go comparisons after clean compilation/execution.
Uncovered: shared default profile/registration/supplied-node API and eight incomplete JSX/Tailwind candidates remain parked; exact helper counter and the full repository gate remain blocked or outside the passing scope; no new helper was claimed.

Current detailed evidence: B8FB-LANDING.md and b8fb-*.log. Earlier landing history follows.

Built: converted all 17 owned supplemental listener declarations to registry kind names and replaced boolean-outcome's manual suffix matcher with the shared-table RegExp literal.
Commits: previous rule tip 7f99254378ee9fd8b50f4cfd9aa18cebe5291b2d; refreshed helper sibling 2b917cbb24fa112ddf8537814e7790f91e4877a2; this rule refresh is committed next on codex/lint-wave1-12.
Checks: 17 named listener domains / 1,201 bytes and nine rule.json domains agree with actual Go on Node, emitted JS and sanitized native; independent owned package PASS 45.219s; aggregate PASS 523.434s with 3,016 cases / 118,907,067 identical findings/fix bytes; vet clean.
Mutants: 17 compiling declaration mutations to Unknown, one valid Unknown rule.json mutation, 10 compiling aggregate semantic controls including role-suffix regex disabled, and the retained resolved-path mutation are caught by their actual-Go comparisons.
Uncovered: shared default profile/registration/supplied-node API and eight incomplete JSX/Tailwind candidates remain parked; full repository gate and broader unsupported parser domains are not claimed.

Current detailed evidence: NAMES-LANDING.md and named-*.log. Earlier landing history follows.

Parked: codex/lint-wave1-12 is rebased onto current origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 under the user's explicit shared-harness parking exception.
Commits: previous pushed parking tip 565a0d45349fec52f1c7d6de73a8cd4449a81d50; rebased implementation eb09445fb902b46a2b69fee7421b1750c743bd14 before this evidence commit; helper sibling pushed b591540e030dd043bf5c6a5aa50130930cefcf79.
Checks: independent rule packages and vet pass on current main; fresh aggregate PASS 457.911s, 118,905,738 identical Go/Node/emitted-JavaScript/sanitized-native bytes; default harness and registry retain their failures.
Mutants: all 17 numeric declaration and nine existing decision/path controls are caught; all nine aggregate semantic mutants also compile/run and are caught only by actual Go findings/fixes comparison on every backend.
Uncovered: default shared harness, partial JSX/Tailwind extraction, numeric rule.json/supplied-node migration, exact helper counter and the complete gate remain blocked or outside the proved scope.

Main advanced during the final fetch after three new helpers were tested and pushed. Both branches were rebased before this final validation. Main's delta adds Stage 3 artifacts, documentation and internal/oracle/stage3_hook_test.go; internal/load, internal/lower, internal/native, internal/javascript, cmd/adamic, lint sources and cohere are unchanged from f8013f0b. Recreating the registration foundation merge caused the same six original conflicts. Since main changed no lint files, the exact original resolution from 9f163dc4d91841438835dd181eb403a76ce372d4 was restored. The finished lint tree equals the previous pushed branch before these evidence files. No shared implementation or compiler change is introduced.

Fresh independent packages: JSX decisions PASS 8.852s, Tailwind decisions PASS 20.487s, numeric declarations/parser-gap/path checks PASS 48.353s. Registry fails in 0.185s because incomplete candidates deliberately have no descriptors; its validation mutants remain masked and are not passing evidence. Default lint test compilation fails at profile_test.go:32:23: cannot range over portFiles (func(t *testing.T) []string). Seventeen numeric listener mutations to Unknown compile/run and are caught by actual Go comparison on source Node, emitted JavaScript and sanitized native. Five JSX decision, three Tailwind decision and one resolved-path mutation are also caught. Whole-rule parity is not inferred from independent decisions.

The isolated rule checkout retains the published .a/suggestion-capable harness f4d98cab50048692781da3599131317dc569d466, excluding the eight incomplete directories there only. Its internal, cmd/adamic and oracle/node.mjs files were restored from current main into index/worktree and verified exactly equal before replay. No scratch shared change is committed. The owned landing-validation_test.go.txt is the complete aggregate test artifact; its prerequisites are the existing owned capture test artifacts recorded in F801-LANDING.md.

The aggregate selects the same 314 pinned TypeScript/compiler and current stage1 source files and 3,016 total supported fixture/witness/path/corpus rows as the earlier run. Unsupported parser recovery and JSX/filename cases remain excluded explicitly, as recorded in F801-LANDING.md. No broader scope is silently inferred from rebasing. The supported-domain output compares actual Go canonical findings and fixes against source Node, emitted JavaScript and ASan/UBSan native. Mutants run over the 190 supported fixture/witness/path rows and must compile and exit cleanly before only the output comparison catches them.

Commands, after sourcing /workspace/adamic-tools/env.sh:

```sh
go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m > /tmp/wave12-c019-default.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/rules/next-google-font-preconnect ./stage1/cohere/lint/rules/better-tailwindcss-enforce-shorthand-classes ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import ./stage1/cohere/lint/registry -count=1 -v -timeout=15m > /tmp/wave12-c019-rules.log 2>&1
go vet ./stage1/cohere/lint/rules/next-google-font-preconnect ./stage1/cohere/lint/rules/better-tailwindcss-enforce-shorthand-classes ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import > /tmp/wave12-c019-rule-vet.log 2>&1
# In the isolated published-harness checkout:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_STAGE1_SOURCE=/workspace/scratch/wave12-landing-rules/stage1 ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestLandingExistingCandidates$' -count=1 -v -timeout=25m > /tmp/wave12-c019-isolated.log 2>&1
```

The default command and combined independent/registry command exit 1 for the named shared/partial-entry blockers; independent packages and vet pass. Parking follows the user's explicit exception for named shared harness blockers (#zmh9v36), not a claimed default gate pass. When the user names the shared harness/Diagnostic landing sha, rebase and re-green this branch before additional work. No full gate or new throughput sample is claimed. Toolchain setup this turn passed in 40s (Go/clang/Node/submodules ready at 0s, cache warm 40s), nproc 5, four-core quota, 17.6 GB. No further rule or helper is claimed after main advanced.

Final aggregate result: PASS 457.911s. All 3,016 supported cases over 314 compiler/stage1 source files match 118,905,738 bytes exactly. All nine semantic mutants are caught on source Node, emitted JavaScript and sanitized native after clean compilation/execution. Complete output is c019-isolated.log. No additional exclusions or new implementation changes were introduced in this refresh.
